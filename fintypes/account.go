package fintypes

import (
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gjson"
	"sort"
	"strings"
	"time"
)

type (
	Account struct {
		Balances []Balance
	}

	AccountSnapshot struct {
		Time   time.Time
		Asset  Asset       `json:"Asset, omitempty"` // 计算Total时用的的Fiat，通常是USD
		Total  AssetAmount // 统计得来的法币总值，用于回测计算收益指标等
		Detail Account
	}
)

func NewEmptyAccount() *Account {
	return &Account{}
}

func NewTestAccount(assets []string, totalInUSD gdecimal.Decimal, ticks Ticks) (*Account, error) {
	if len(assets) == 0 {
		return nil, errors.Errorf("no assets")
	}

	eachUSD := totalInUSD.DivInt(len(assets))
	r := NewEmptyAccount()

	for _, v := range assets {
		price, err := ticks.GetUSDPrice(MarketSpot, v)
		if err != nil {
			return nil, err
		}
		r.SetAmount(AssetProperty{
			Market:           MarketSpot,
			Margin:           MarginNo,
			CustomSubAccName: "",
			Asset:            v,
		}, AssetAmount{
			Free: eachUSD.Div(price),
		})
	}
	return r, nil
}

// 几倍风险，1表示满仓，2表示2倍风险
/*unc (a *Account) GetPosition(market Market, pair Pair) gdecimal.Decimal {
	unitAmount := gdecimal.Zero
	quoteAmount := gdecimal.Zero

	for _, v := range a.Balances {
		if v.Market == market && strings.ToUpper(v.Asset) == strings.ToUpper(pair.Unit()) {
			unitAmount = v.AssetAmount.Net()
		}
	}

}*/

// 根据资产名称获取总数量
func (a *Account) GetAmountByName(asset string) AssetAmount {
	res := &AssetAmount{}

	for _, v := range a.Balances {
		if strings.ToLower(asset) == strings.ToLower(v.Asset) {
			res.Add(v.AssetAmount)
		}
	}
	return *res
}

// 根据资产属性（market、margin、名称等等）获取数量
func (a *Account) GetAmountByProperty(ap AssetProperty) AssetAmount {
	for _, v := range a.Balances {
		if v.AssetPropertyEquals(ap) {
			return v.AssetAmount
		}
	}
	return AssetAmount{}
}

// upsert new balance item
func (a *Account) SetAmount(ap AssetProperty, amount AssetAmount) {
	for i := range a.Balances {
		if a.Balances[i].AssetPropertyEquals(ap) {
			a.Balances[i].AssetAmount = amount
			return
		}
	}
	a.Balances = append(a.Balances, Balance{
		AssetProperty: ap,
		AssetAmount:   amount,
	})
}

// upsert free amount
func (a *Account) SetFreeAmount(ap AssetProperty, freeAmount gdecimal.Decimal) {
	for i := range a.Balances {
		if a.Balances[i].AssetPropertyEquals(ap) {
			a.Balances[i].AssetAmount.Free = freeAmount
			return
		}
	}
	a.Balances = append(a.Balances, Balance{
		AssetProperty: ap,
		AssetAmount:   AssetAmount{Free: freeAmount},
	})
}

// upsert locked amount
func (a *Account) SetLockedAmount(ap AssetProperty, lockedAmount gdecimal.Decimal) {
	for i := range a.Balances {
		if a.Balances[i].AssetPropertyEquals(ap) {
			a.Balances[i].AssetAmount.Locked = lockedAmount
			return
		}
	}
	a.Balances = append(a.Balances, Balance{
		AssetProperty: ap,
		AssetAmount:   AssetAmount{Locked: lockedAmount},
	})
}

func (a *Account) String() string {
	return a.JsonString()
}

func (a *Account) JsonString() string {
	return gjson.MarshalStringDefault(a, true)
}

func (a *Account) addBalance(toAdd Balance) {
	for k, v := range a.Balances {
		if v.AssetPropertyEquals(toAdd.AssetProperty) {
			pb := &a.Balances[k]
			pb.Add(toAdd)
			a.Balances[k] = *pb
			return
		}
	}

	a.Balances = append(a.Balances, toAdd)
}

func (a *Account) AddAccount(toAdd Account) {
	for _, v := range toAdd.Balances {
		a.addBalance(v)
	}
}

func (a *Account) AddLock(ap AssetProperty, toAdd gdecimal.Decimal) {
	//fmt.Println("addLock", toAdd)
	for k, v := range a.Balances {
		if v.AssetPropertyEquals(ap) {
			a.Balances[k].Locked = a.Balances[k].Locked.Add(toAdd)
			return
		}
	}

	a.Balances = append(a.Balances, Balance{
		AssetProperty: ap,
		AssetAmount: AssetAmount{
			Locked: toAdd,
		},
	})
}

// 增加余额，如果已经有相关条目就着做加法，如果没有则新增
func (a *Account) AddFree(ap AssetProperty, toAdd gdecimal.Decimal) {
	for k, v := range a.Balances {
		if v.AssetPropertyEquals(ap) {
			a.Balances[k].Free = a.Balances[k].Free.Add(toAdd)
			return
		}
	}

	a.Balances = append(a.Balances, Balance{
		AssetProperty: ap,
		AssetAmount: AssetAmount{
			Free: toAdd,
		},
	})
}

// 换算成USD
// 计算的用途是统计回报的相关指标
func (a *Account) ExchangeToUSD(ticks Ticks, ignoreNotFound bool) (*AssetAmount, error) {
	res := &AssetAmount{}

	for _, balance := range a.Balances {
		unit := balance.Asset
		if balance.Market.IsContract() {

		} else {
			price, err := ticks.GetUSDPrice(MarketSpot, unit)
			if err != nil {
				if ignoreNotFound {
					continue
				}
				return nil, err
			}
			res.Free = res.Free.Add(balance.Free.Mul(price))
			res.Locked = res.Locked.Add(balance.Locked.Mul(price))
			res.Borrowed = res.Borrowed.Add(balance.Borrowed.Mul(price))
			res.Interest = res.Interest.Add(balance.Interest.Mul(price)) // FIXME 这里用Spot有没有问题
		}
	}
	return res, nil
}

// 清除为0的成员，用于Transfer等操作之后清除没必要显示的成员
func (a *Account) RemoveZeroMembers() {
	var cache []Balance
	for _, v := range a.Balances {
		if !v.IsZero() {
			cache = append(cache, v)
		}
	}
	a.Balances = cache
}

func (a *Account) getIndex(sap SubAccParsed, asset string) int {
	for k, v := range a.Balances {
		if v.Market == sap.Market && v.Margin == sap.Margin && v.CustomSubAccName == sap.Name && v.Asset == asset {
			return k
		}
	}
	return -1
}

func (a *Account) Transfer(asset string, amount gdecimal.Decimal, from, to SubAcc) error {
	// 校验和解析参数
	if amount.IsPositive() == false {
		return gerror.Errorf("invalid transfer amount %s", amount.String())
	}
	saFrom, err := from.Parse()
	if err != nil {
		return err
	}
	saTo, err := to.Parse()
	if err != nil {
		return err
	}

	// 查询并检查源钱包的数组下标
	saFromIdx := a.getIndex(saFrom, asset)
	if saFromIdx < 0 {
		return gerror.Errorf("can't find from SubAcc (%s) with Asset(%s)", from.String(), asset)
	}

	// 查询并检查目的钱包的数组下标
	saToIdx := a.getIndex(saTo, asset)
	if saFromIdx == saToIdx {
		return gerror.Errorf("can't transfer asset from (%s) to (%s), they are the same", from.String(), to.String())
	}

	// 查询源钱包余额是否足够，若足够则执行转账的第一步：减法
	if a.Balances[saFromIdx].Free.LessThan(amount) {
		return gerror.Errorf("insufficient account balance, %s required, but %s in account", amount, a.Balances[saFromIdx].Free.String())
	} else {
		a.Balances[saFromIdx].Free = a.Balances[saFromIdx].Free.Sub(amount)
	}

	// 执行转账的第二步：加法
	if saToIdx < 0 { // 已有列表中不存在目的钱包，把资产转到新建钱包即可

		a.Balances = append(a.Balances, Balance{AssetAmount: AssetAmount{
			Free: amount,
		},
		})
	} else { // 已有列表中存在目的钱包，需要把转出资产挪到目的钱包
		p := &a.Balances[saToIdx]
		p.Add(Balance{AssetAmount: AssetAmount{Free: amount}})

	}

	// 清除全为0的钱包
	a.RemoveZeroMembers()
	return nil
}

// 只支持现货杠杆
func (a *Account) Borrow(tm time.Time, margin Margin, asset string, amount gdecimal.Decimal) error {
	if amount.IsPositive() == false {
		return gerror.Errorf("borrow %s %s is invalid", asset, amount.String())
	}

	// 有余额资产的借贷
	for k := range a.Balances {
		if a.Balances[k].Market == MarketSpot && a.Balances[k].Margin == margin && a.Balances[k].Asset == asset {
			a.Balances[k].Free = a.Balances[k].Free.Add(amount)
			a.Balances[k].Borrowed = a.Balances[k].Borrowed.Add(amount)
			a.Balances[k].BorrowHist = append(a.Balances[k].BorrowHist, Borrow{
				Time:      tm,
				Principal: amount,
			})
			return nil
		}
	}

	// 无余额资产的借贷
	a.Balances = append(a.Balances, Balance{
		AssetProperty: AssetProperty{
			Market:           MarketSpot,
			Margin:           margin,
			CustomSubAccName: "",
			Asset:            asset,
		},
		AssetAmount: AssetAmount{
			Free:     amount,
			Borrowed: amount,
		},
		BorrowHist: []Borrow{{
			Time:      tm,
			Principal: amount,
		},
		}})
	return nil
}

func (a *Account) Repay(tm time.Time, interestRateDaily gdecimal.Decimal, margin Margin, asset string, toRepay gdecimal.Decimal) error {
	if toRepay.IsPositive() == false {
		return gerror.Errorf("repay %s %s is invalid", asset, toRepay.String())
	}

	// 获取待还资产在数组中的下标
	idx := -1
	for k, v := range a.Balances {
		if v.Market == MarketSpot && v.Margin == margin && v.Asset == asset {
			idx = k
			break
		}
	}

	// 该资产压根不存在，自然没法Repay
	if idx < 0 {
		return gerror.Errorf("can't find balance(%s, %s, %s)", MarketSpot, margin, asset)
	}

	src := a.Balances[idx]
	// 按时间排序，方便后面还款的时候先还最早的借款
	sort.Sort(src)
	// 如果填写金额大于可用额度，自动调整
	toRepay = gdecimal.Min(src.Free, toRepay)

	// 全部还清
	borrowAndInterest := src.BorrowedByHist().Add(src.InterestByHist(tm, interestRateDaily))
	if toRepay.GreaterThanOrEqual(borrowAndInterest) {
		src.Free = src.Free.Sub(borrowAndInterest)
		src.Borrowed = gdecimal.N0
		src.BorrowHist = nil
		a.Balances[idx] = src
		return nil
	}

	// 只还一部分
	totalToRepay := toRepay       // 全部可还
	totalRepaied := gdecimal.Zero // 全部已还，含本金和利息
	var newBorrowList []Borrow    // 还款之后，尚未还清的借款列表
	for _, v := range a.Balances[idx].BorrowHist {
		if totalToRepay.IsPositive() { // 还有可用于还款的资金
			currToRepay := v.Principal.Add(v.Interest(tm, interestRateDaily)) // 当前这笔借款的待还，包括本金和利息
			currRepaied := gdecimal.Min(totalToRepay, currToRepay)            // 计算当前记录的实际还款
			currRepaiedScale := currRepaied.Div(currToRepay)                  // 当前记录的还款比例，1就是100%全额还清，小于1就是只能换一部分
			if currRepaiedScale.LessThanInt(1) {                              // 未达100%归还，也就是部分还款
				v.Principal = v.Principal.Mul(gdecimal.One.Sub(currRepaiedScale)) // 部分还款之后的未还本金
				newBorrowList = append(newBorrowList, v)                          // 部分还款时需要把剩余未还本金再次归档
			}

			totalToRepay = totalToRepay.Sub(currRepaied) // 全部可还
			totalRepaied = totalRepaied.Add(currRepaied) // 全部已还，含本金和利息
		} else { // 还款资金用完了
			newBorrowList = append(newBorrowList, v) // 原封不动再次归档
		}
	}

	// 更新src
	src.Free = src.Free.Sub(totalRepaied) // 从可用余额中扣除全部已还本金和利息
	src.Borrowed = src.Borrowed.Sub(totalRepaied)
	src.BorrowHist = newBorrowList
	a.Balances[idx] = src
	return nil
}

func (a *Account) Lock(market Market, margin Margin, asset string, amount gdecimal.Decimal) error {
	for k, v := range a.Balances {
		if v.Market == market && v.Margin == margin && v.Asset == asset {
			if v.Free.LessThan(amount) {
				return gerror.Errorf("Free(%s) < LockAmount(%s)", v.Locked.String(), amount.String())
			}
			a.Balances[k].Free = a.Balances[k].Free.Sub(amount)
			a.Balances[k].Locked = a.Balances[k].Locked.Add(amount)
			return nil
		}

	}
	return nil
}

func (a *Account) Unlock(ap AssetProperty, amount gdecimal.Decimal) error {
	for k, v := range a.Balances {
		if v.AssetPropertyEquals(ap) {
			if v.Locked.LessThan(amount) {
				return gerror.Errorf("Locked(%s) < UnlockAmount(%s)", v.Locked.String(), amount.String())
			}
			a.Balances[k].Locked = a.Balances[k].Locked.Sub(amount)
			a.Balances[k].Free = a.Balances[k].Free.Add(amount)
			return nil
		}
	}
	return nil
}

func (as *AccountSnapshot) Add(toAdd AccountSnapshot) {
	as.Detail.AddAccount(toAdd.Detail)
	as.Total.Add(toAdd.Total)
}
