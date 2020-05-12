package fintypes

import (
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"math"
	"sort"
	"time"
)

/*
	huobi的futureAccount成员
	AccountRights float64 //账户权益
	KeepDeposit   float64 //保证金
	ProfitReal    float64 //已实现盈亏
	ProfitUnreal  float64
	RiskRate      float64 //保证金率
*/
// 合约账户余额
/**
期货的保证金分“开仓保证金”和“持仓保证金”（大家习惯性翻译为“初始保证金”和“维持保证金”）。
比如开仓时需要5000美元/手才能开仓，而开仓后则账户上只需要保持4000美元以上即可“维持”这个头寸，这4000叫“持仓保证金”。
除了中国以外的期货市场是直接按照合约比例（比如8％，经纪公司加3个百分点控制风险）来收取保证金，其他的国家大部分都是这种开仓和持仓保证金制度。
*/
//WalletBalance         gdecimal.Decimal // 钱包余额 等于MarginBalance
//PositionInitialMargin gdecimal.Decimal // 当前的仓位占用保证金，等于InitialMargin
/*ContractBalance struct {
	InitialMargin     gdecimal.Decimal // 开仓保证金，也叫初始保证金，也就是当前仓位占用的保证金，你下好单之后哪怕不操作这个值也会动态变化
	MaintMargin       gdecimal.Decimal // 持仓保证金，也叫维持保证金（maintenance margin）
	MarginBalance     gdecimal.Decimal // 保证金总额，包括未被持仓占用的和被持仓占用的，相当于现货中的Total，等于InitialMargin+MaxWithdrawAmount
	MaxWithdrawAmount gdecimal.Decimal // 最大可提现，或者叫剩余可用保证金额度，这个可能和Free概念不同
	UnrealizedProfit  gdecimal.Decimal // 未实现的盈亏，也就是浮盈/浮亏
}

// basic balance
// total = free + locked
// net = free + locked - borrowed - interest
SpotBalance struct {
	Free     gdecimal.Decimal `json:"Free"`     // available，可操作买卖，但不一定可以提现
	Locked   gdecimal.Decimal `json:"Locked"`   //
	Borrowed gdecimal.Decimal `json:"Borrowed"` // to repay, only available in margin account
	Interest gdecimal.Decimal `json:"Interest"` // only available in margin account

	// 扩展信息，用于回测
	BorrowHist []Borrow // 现货杠杆才有
}*/
/*


// 需要还的
func (sb SpotBalance) ToRepay() gdecimal.Decimal {
	return sb.Borrowed.Add(sb.Interest)
}

func (sb SpotBalance) IsZero() bool {
	return sb.Free.IsZero() && sb.Interest.IsZero() && sb.Locked.IsZero() && sb.Borrowed.IsZero()
}

// TODO: 目前的实现可能不严谨
func (sa ContractBalance) IsZero() bool {
	return sa.InitialMargin.IsZero() && sa.MaintMargin.IsZero() && sa.MarginBalance.IsZero()
}

// Add function will change content of receiver, so pointer required
func (sb *SpotBalance) Add(toAdd SpotBalance) {
	sb.Free = sb.Free.Add(toAdd.Free)
	sb.Borrowed = sb.Borrowed.Add(toAdd.Borrowed)
	sb.Locked = sb.Locked.Add(toAdd.Locked)
	sb.Interest = sb.Interest.Add(toAdd.Interest)
}

func (cb *ContractBalance) Add(toAdd ContractBalance) {
	cb.InitialMargin = cb.InitialMargin.Add(toAdd.InitialMargin)
	cb.MaintMargin = cb.MaintMargin.Add(toAdd.MaintMargin)
	cb.MarginBalance = cb.MarginBalance.Add(toAdd.MarginBalance)
	cb.MaxWithdrawAmount = cb.MaxWithdrawAmount.Add(toAdd.MaxWithdrawAmount)
	//cb.PositionInitialMargin = cb.PositionInitialMargin.Add(toAdd.PositionInitialMargin)
	cb.UnrealizedProfit = cb.UnrealizedProfit.Add(toAdd.UnrealizedProfit)
	//cb.WalletBalance = cb.WalletBalance.Add(toAdd.WalletBalance)
}

// 这是一个精简转换
func (cb *ContractBalance) ToBasic() SpotBalance {
	return SpotBalance{
		Free:     cb.MaxWithdrawAmount,
		Locked:   cb.InitialMargin,
		Borrowed: gdecimal.Zero,
		Interest: gdecimal.Zero,
	}
}*/

type (
	// 一条现货杠杆借贷记录
	Borrow struct {
		Time      time.Time        // 借入时间
		Principal gdecimal.Decimal // 本金
	}

	// 资产属性
	AssetProperty struct {
		Market           Market // 属于哪个市场
		Margin           Margin
		CustomSubAccName string // 自定义子账户名称
		Asset            string
	}

	// 资产数量
	AssetAmount struct {
		// 沿用现货的成员
		Free     gdecimal.Decimal `json:"Free"`     // available，可操作买卖，但不一定可以提现
		Locked   gdecimal.Decimal `json:"Locked"`   //
		Borrowed gdecimal.Decimal `json:"Borrowed"` // to repay, only available in margin account
		Interest gdecimal.Decimal `json:"Interest"` // only available in margin account
	}

	// 余额记录，一个特定市场特定杠杆类型下特定资产对应一个BTBalance
	Balance struct {
		AssetProperty          // 资产属性
		AssetAmount            // 资产数量
		BorrowHist    []Borrow // 现货杠杆才有，用于回测
	}
)

func NewAP(market Market, margin Margin, asset string) AssetProperty {
	return AssetProperty{
		Market:           market,
		Margin:           margin,
		CustomSubAccName: "",
		Asset:            asset,
	}
}

// 某一笔借款的利息
func (b *Borrow) Interest(tm time.Time, interestRateDaily gdecimal.Decimal) gdecimal.Decimal {
	days := int(math.Ceil(float64(tm.Sub(b.Time)) / float64(gtime.Day)))
	return interestRateDaily.MulInt(days)
}

func (b *AssetAmount) Add(toAdd AssetAmount) {
	b.Free = b.Free.Add(toAdd.Free)
	b.Locked = b.Locked.Add(toAdd.Locked)
	b.Borrowed = b.Borrowed.Add(toAdd.Borrowed)
	b.Interest = b.Interest.Add(toAdd.Interest)
}

// 需要还的
func (b *AssetAmount) ToRepay() gdecimal.Decimal {
	return b.Borrowed.Add(b.Interest)
}

// 净值
func (b *AssetAmount) Net() gdecimal.Decimal {
	return b.Free.Add(b.Locked).Sub(b.Borrowed).Sub(b.Interest)
}

func (b *Balance) IsZero() bool {
	return b.Free.IsZero() && b.Interest.IsZero() && b.Locked.IsZero() && b.Borrowed.IsZero()
}

func (b *Balance) BorrowedByHist() gdecimal.Decimal {
	r := gdecimal.Zero
	for _, v := range b.BorrowHist {
		r = r.Add(v.Principal)
	}
	return r
}

func (b *Balance) InterestByHist(tm time.Time, interestRateDaily gdecimal.Decimal) gdecimal.Decimal {
	totalInterest := gdecimal.Zero
	for _, v := range b.BorrowHist {
		days := int(math.Ceil(float64(tm.Sub(v.Time)) / float64(gtime.Day)))
		totalInterest = totalInterest.Add(interestRateDaily.MulInt(days))
	}
	return totalInterest
}

// 验证资产种类，如果Market、Margin、CustomSubAccName、Asset不同则返回错误
func (b *Balance) VerifyAssetProperty(ap AssetProperty) error {
	if b.Market != ap.Market {
		return gerror.Errorf("different market %s vs %s", b.Market, ap.Market)
	}
	if b.Margin != ap.Margin {
		return gerror.Errorf("different margin %s vs %s", b.Margin, ap.Margin)
	}
	if b.CustomSubAccName != ap.CustomSubAccName {
		return gerror.Errorf("different CustomSubAccName %s vs %s", b.CustomSubAccName, ap.CustomSubAccName)
	}
	if b.Asset != ap.Asset {
		return gerror.Errorf("different Asset %s vs %s", b.Asset, ap.Asset)
	}
	return nil
}

func (b *Balance) AssetPropertyEquals(ap AssetProperty) bool {
	err := b.VerifyAssetProperty(ap)
	return err == nil
}

func (b *Balance) Add(toAdd Balance) error {
	if err := b.VerifyAssetProperty(toAdd.AssetProperty); err != nil {
		return err
	}
	pba := &b.AssetAmount
	pba.Add(toAdd.AssetAmount)
	b.BorrowHist = append(b.BorrowHist, toAdd.BorrowHist...)
	sort.Sort(b)
	return nil
}

// 按borrows的时间排序
func (b Balance) Len() int { return len(b.BorrowHist) }
func (b Balance) Less(i, j int) bool {
	return b.BorrowHist[i].Time.Before(b.BorrowHist[j].Time)
}
func (b Balance) Swap(i, j int) {
	b.BorrowHist[i], b.BorrowHist[j] = b.BorrowHist[j], b.BorrowHist[i]
}
