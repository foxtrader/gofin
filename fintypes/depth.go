package fintypes

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"

	"sort"
	"time"
)

type (
	OrderBook struct {
		Price  gdecimal.Decimal `json:"Price"`
		Amount gdecimal.Decimal `json:"Amount"`
	}

	OrderBookList []OrderBook

	// 盘口承受力
	DepthTolerance struct {
		UnitDealAmount  gdecimal.Decimal
		QuoteDealAmount gdecimal.Decimal
		Price1          gdecimal.Decimal
		PriceReach      gdecimal.Decimal // 上/下探价格, -1 means 击穿
	}
)

var (
	PricePiercing = gdecimal.NewFromInt(-1)
)

func (obl OrderBookList) Len() int {
	return len(obl)
}

func (obl OrderBookList) Swap(i, j int) {
	obl[i], obl[j] = obl[j], obl[i]
}

func (obl OrderBookList) Less(i, j int) bool {
	return obl[i].Price.LessThan(obl[j].Price)
}

func (dt *DepthTolerance) Equal(cmp DepthTolerance) bool {
	if !dt.UnitDealAmount.Equal(cmp.UnitDealAmount) {
		return false
	}
	if !dt.QuoteDealAmount.Equal(cmp.QuoteDealAmount) {
		return false
	}
	if !dt.Price1.Equal(cmp.Price1) {
		return false
	}
	if !dt.PriceReach.Equal(cmp.PriceReach) {
		return false
	}
	return true
}

// verify price and unit amount
func (dt *DepthTolerance) Verify() error {
	if dt.PriceReach.Equal(PricePiercing) {
		return errors.Errorf("invalid *DepthTolerance because of piercing price")
	}
	if !dt.PriceReach.IsPositive() || !dt.UnitDealAmount.IsPositive() {
		return errors.Errorf("invalid price(%s) or unitAmount(%s)", dt.PriceReach.String(), dt.UnitDealAmount.String())
	}
	return nil
}

// 删除挂单价格或者数量为0的单子
func RemoveInvalidOrders(input []OrderBook) []OrderBook {
	var r []OrderBook

	for _, v := range input {
		if v.Price.IsPositive() && v.Amount.IsPositive() {
			r = append(r, v)
		}
	}

	return r
}

// 获取挂单中币的总量
func GetTotalAmount(input []OrderBook) gdecimal.Decimal {
	r := gdecimal.Zero
	for _, v := range input {
		r = r.Add(v.Amount)
	}
	return r
}

// 获取挂单中（如果都成交）基准货币的总量
func GetTotalPrice(input []OrderBook) gdecimal.Decimal {
	r := gdecimal.Zero
	for _, v := range input {
		r = r.Add(v.Amount.Mul(v.Price))
	}
	return r
}

// Raw depth data without timestamp.
// Useful in remake new Depth like struct.
type DepthRawData struct {
	Sells OrderBookList `json:"Sells"` // Asks
	Buys  OrderBookList `json:"Buys"`  // Bids
}

func (drd *DepthRawData) GetMinSell() (price, amount gdecimal.Decimal, err error) {
	if drd.Sells.Len() == 0 {
		return gdecimal.Zero, gdecimal.Zero, errors.Errorf("Empty sells")
	}
	cpy := drd.Sells
	sort.Sort(cpy)
	return cpy[0].Price, cpy[0].Amount, nil
}

func (drd *DepthRawData) GetMaxBuy() (price, amount gdecimal.Decimal, err error) {
	if drd.Buys.Len() == 0 {
		return gdecimal.Zero, gdecimal.Zero, errors.Errorf("Empty sells")
	}
	cpy := drd.Buys
	sort.Sort(cpy)
	maxIdx := cpy.Len() - 1
	return cpy[maxIdx].Price, cpy[maxIdx].Amount, nil
}

func (drd *DepthRawData) Equal(cmp *DepthRawData) bool {
	if cmp == nil {
		return false
	}
	if len(drd.Buys) != len(cmp.Buys) || len(drd.Sells) != len(cmp.Sells) {
		return false
	}
	for k, v := range drd.Buys {
		if v.Price != cmp.Buys[k].Price || v.Amount != cmp.Buys[k].Amount {
			return false
		}
	}
	for k, v := range drd.Sells {
		if v.Price != cmp.Sells[k].Price || v.Amount != cmp.Sells[k].Amount {
			return false
		}
	}
	return true
}

// Depth = open order books.
type Depth struct {
	Time time.Time `json:"T"` // T
	DepthRawData
}

func (d Depth) String() string {
	buf, err := json.Marshal(d)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (d *Depth) Sort() {
	sort.Sort(sort.Reverse(d.Buys))
	sort.Sort(d.Sells)
}

func (d *Depth) ParseSellOnePrice() gdecimal.Decimal {
	return d.Sells[0].Price
}

func (d *Depth) ParseBuyOnePrice() gdecimal.Decimal {
	return d.Buys[0].Price
}

func (d *Depth) ParseBuyerTotal() (avgPrice, totalAmount gdecimal.Decimal) {
	totalPrice := gdecimal.Zero
	for _, v := range d.Buys {
		totalAmount = totalAmount.Add(v.Amount)
		totalPrice = totalPrice.Add(v.Amount.Mul(v.Price))
	}
	return totalPrice.Div(totalAmount), totalAmount
}

func (d *Depth) ParseSellerTotal() (avgPrice, totalAmount gdecimal.Decimal) {
	totalPrice := gdecimal.Zero
	for _, v := range d.Sells {
		totalAmount = totalAmount.Add(v.Amount)
		totalPrice = totalPrice.Add(v.Amount.Mul(v.Price))
	}
	return totalPrice.Div(totalAmount), totalAmount
}

// 砸盘探测，能砸多深
func (d *Depth) MarketSellDetectEx(unitAmount, slippageAllowed gdecimal.Decimal) DepthTolerance {
	r := DepthTolerance{}
	if d.DepthRawData.Buys.Len() > 0 {
		r.Price1 = d.DepthRawData.Buys[0].Price
	}
	r.PriceReach = PricePiercing
	minPriceAllowed := r.Price1.Mul(gdecimal.One.Sub(slippageAllowed)) // 最小忍受价格

	leftUnit := unitAmount // example: BTC in BTC/USDC
	for _, orderBook := range d.DepthRawData.Buys {

		// 超过最小忍受价格即退出，太便宜了不卖
		if orderBook.Price.LessThan(minPriceAllowed) {
			break
		}

		dealUnit := gdecimal.Min(orderBook.Amount, leftUnit)
		r.UnitDealAmount = r.UnitDealAmount.Add(dealUnit)
		r.QuoteDealAmount = r.QuoteDealAmount.Add(dealUnit.Mul(orderBook.Price))
		r.PriceReach = orderBook.Price
		leftUnit = leftUnit.Sub(dealUnit)
		if leftUnit.LessThanOrEqual(gdecimal.Zero) {
			break
		}
	}
	return r
}

// 拉盘探测，能拉多高
// note: 如果不给出Prec，可能导致模拟成交的时候由于除法除不尽导致结果不精确而稍稍大于绝对精确值的情况，从而产生连锁反应
// FIXME: precision 确定一下是unit的precision吗，同理也给marketSell确认一下
func (d *Depth) MarketBuyDetectEx(quoteAmount gdecimal.Decimal, slippageAllowed gdecimal.Decimal, precision int, lot gdecimal.Decimal) DepthTolerance {
	r := DepthTolerance{}

	if d.DepthRawData.Sells.Len() > 0 {
		r.Price1 = d.DepthRawData.Sells[0].Price
	}
	r.PriceReach = PricePiercing
	maxPriceAllowed := r.Price1.Mul(gdecimal.One.Add(slippageAllowed)) // 最大忍受价格

	leftQuote := quoteAmount // example: USDC
	for _, orderBook := range d.DepthRawData.Sells {

		// 超过最大可接受价格即退出，太贵了不买
		if orderBook.Price.GreaterThan(maxPriceAllowed) {
			break
		}

		dealQuote := gdecimal.Min(orderBook.Amount.Mul(orderBook.Price), leftQuote)
		//if precision == nil || lot == nil {
		//	r.UnitDealAmount = r.UnitDealAmount.Add(dealQuote.Div(orderBook.Price))
		//} else {
		// NOTE: DivRound可能有问题，导致除以后结果偏大，相当于DivRoundUp了，然后就导致UnitDealAmount x PriceReach > quoteAmount，也就是新的unit数量大于原始的unit数量
		//r.UnitDealAmount = r.UnitDealAmount.Add(dealQuote.DivRound(orderBook.Price, *precision).Trunc(*precision, *lot))
		r.UnitDealAmount = r.UnitDealAmount.Add(dealQuote.Div(orderBook.Price).Trunc(precision, lot.Float64()))
		//}
		r.QuoteDealAmount = r.QuoteDealAmount.Add(dealQuote)
		r.PriceReach = orderBook.Price
		leftQuote = leftQuote.Sub(dealQuote)
		if leftQuote.LessThanOrEqual(gdecimal.Zero) {
			break
		}
	}

	return r
}

// 砸盘探测，能砸多深
func (d *Depth) MarketSellDetect(unitAmount gdecimal.Decimal) DepthTolerance {
	r := DepthTolerance{}
	if d.DepthRawData.Buys.Len() > 0 {
		r.Price1 = d.DepthRawData.Buys[0].Price
	}
	r.PriceReach = PricePiercing

	leftUnit := unitAmount // example: BTC
	for _, orderBook := range d.DepthRawData.Buys {
		dealUnit := gdecimal.Min(orderBook.Amount, leftUnit)
		r.UnitDealAmount = r.UnitDealAmount.Add(dealUnit)
		r.QuoteDealAmount = r.QuoteDealAmount.Add(dealUnit.Mul(orderBook.Price))
		leftUnit = leftUnit.Sub(dealUnit)
		if leftUnit.LessThanOrEqual(gdecimal.Zero) {
			r.PriceReach = orderBook.Price
			break
		}
	}
	return r
}

// 拉盘探测，能拉多高
func (d *Depth) MarketBuyDetect(quoteAmount gdecimal.Decimal) DepthTolerance {
	r := DepthTolerance{}

	if d.DepthRawData.Sells.Len() > 0 {
		r.Price1 = d.DepthRawData.Sells[0].Price
	}
	r.PriceReach = PricePiercing

	leftQuote := quoteAmount // example: USDC
	for _, orderBook := range d.DepthRawData.Sells {
		dealQuote := gdecimal.Min(orderBook.Amount.Mul(orderBook.Price), leftQuote)
		r.UnitDealAmount = r.UnitDealAmount.Add(dealQuote.Div(orderBook.Price))
		r.QuoteDealAmount = r.QuoteDealAmount.Add(dealQuote)
		leftQuote = leftQuote.Sub(dealQuote)
		if leftQuote.LessThanOrEqual(gdecimal.Zero) {
			r.PriceReach = orderBook.Price
			break
		}
	}
	return r
}
