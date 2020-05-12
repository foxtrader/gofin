package fintypes

import (
	"encoding/json"
	"github.com/shawnwyckoff/gopkg/apputil/gtest"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gjson"
	"testing"
	"time"
)

// example: LTC/USDT
func newTestDepth(sort bool) Depth {
	r := &Depth{}
	r.Sells = append(r.Sells, OrderBook{Price: gdecimal.NewFromInt(13), Amount: gdecimal.NewFromInt(1)})
	r.Sells = append(r.Sells, OrderBook{Price: gdecimal.NewFromInt(12), Amount: gdecimal.NewFromInt(1)})
	r.Sells = append(r.Sells, OrderBook{Price: gdecimal.NewFromInt(11), Amount: gdecimal.NewFromInt(1)})
	r.Buys = append(r.Buys, OrderBook{Price: gdecimal.NewFromInt(10), Amount: gdecimal.NewFromInt(1)})
	r.Buys = append(r.Buys, OrderBook{Price: gdecimal.NewFromInt(8), Amount: gdecimal.NewFromInt(1)})
	r.Buys = append(r.Buys, OrderBook{Price: gdecimal.NewFromInt(9), Amount: gdecimal.NewFromInt(1)})
	if sort {
		r.Sort()
	}
	return *r
}

func TestDepth_MarketBuyDetect(t *testing.T) {
	d := newTestDepth(true)

	dt := d.MarketBuyDetect(gdecimal.NewFromInt(11))
	expected := DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(1),
		QuoteDealAmount: gdecimal.NewFromInt(11),
		Price1:          gdecimal.NewFromInt(11),
		PriceReach:      gdecimal.NewFromInt(11),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketBuyDetect error1")
		return
	}

	dt = d.MarketBuyDetect(gdecimal.NewFromInt(23))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(2),
		QuoteDealAmount: gdecimal.NewFromInt(23),
		Price1:          gdecimal.NewFromInt(11),
		PriceReach:      gdecimal.NewFromInt(12),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketBuyDetect error2")
		return
	}

	dt = d.MarketBuyDetect(gdecimal.NewFromInt(36))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(3),
		QuoteDealAmount: gdecimal.NewFromInt(36),
		Price1:          gdecimal.NewFromInt(11),
		PriceReach:      gdecimal.NewFromInt(13),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketBuyDetect error3")
		return
	}

	dt = d.MarketBuyDetect(gdecimal.NewFromInt(37))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(3),
		QuoteDealAmount: gdecimal.NewFromInt(36),
		Price1:          gdecimal.NewFromInt(11),
		PriceReach:      PricePiercing,
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketBuyDetect error4")
		return
	}
}

func TestDepth_MarketSellDetect(t *testing.T) {
	d := newTestDepth(true)

	dt := d.MarketSellDetect(gdecimal.NewFromInt(1))
	expected := DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(1),
		QuoteDealAmount: gdecimal.NewFromInt(10),
		Price1:          gdecimal.NewFromInt(10),
		PriceReach:      gdecimal.NewFromInt(10),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketSellDetect error1")
		return
	}

	dt = d.MarketSellDetect(gdecimal.NewFromInt(2))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(2),
		QuoteDealAmount: gdecimal.NewFromInt(19),
		Price1:          gdecimal.NewFromInt(10),
		PriceReach:      gdecimal.NewFromInt(9),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketSellDetect error2")
		return
	}

	dt = d.MarketSellDetect(gdecimal.NewFromInt(3))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(3),
		QuoteDealAmount: gdecimal.NewFromInt(27),
		Price1:          gdecimal.NewFromInt(10),
		PriceReach:      gdecimal.NewFromInt(8),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketSellDetect error3")
		return
	}

	dt = d.MarketSellDetect(gdecimal.NewFromInt(4))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(3),
		QuoteDealAmount: gdecimal.NewFromInt(27),
		Price1:          gdecimal.NewFromInt(10),
		PriceReach:      PricePiercing,
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketSellDetect error4")
		return
	}
}

func TestDepth_MarketBuyDetectEx(t *testing.T) {
	d := newTestDepth(true)

	dt := d.MarketBuyDetectEx(gdecimal.NewFromInt(11), gdecimal.NewFromFloat64(0.1), 8, gdecimal.NewFromInt(1))
	expected := DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(1),
		QuoteDealAmount: gdecimal.NewFromInt(11),
		Price1:          gdecimal.NewFromInt(11),
		PriceReach:      gdecimal.NewFromInt(11),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketBuyDetectEx error1")
		return
	}

	dt = d.MarketBuyDetectEx(gdecimal.NewFromInt(23), gdecimal.NewFromFloat64(0.1), 8, gdecimal.NewFromInt(1))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(2),
		QuoteDealAmount: gdecimal.NewFromInt(23),
		Price1:          gdecimal.NewFromInt(11),
		PriceReach:      gdecimal.NewFromInt(12),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketBuyDetectEx error2")
		return
	}

	dt = d.MarketBuyDetectEx(gdecimal.NewFromInt(36), gdecimal.NewFromFloat64(0.1), 8, gdecimal.NewFromInt(1))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(2),
		QuoteDealAmount: gdecimal.NewFromInt(23),
		Price1:          gdecimal.NewFromInt(11),
		PriceReach:      gdecimal.NewFromInt(12),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketBuyDetectEx error3")
		return
	}

	dt = d.MarketBuyDetectEx(gdecimal.NewFromInt(37), gdecimal.NewFromFloat64(0.1), 8, gdecimal.NewFromInt(1))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(2),
		QuoteDealAmount: gdecimal.NewFromInt(23),
		Price1:          gdecimal.NewFromInt(11),
		PriceReach:      gdecimal.NewFromInt(12),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketBuyDetectEx error4")
		return
	}

	depthString := `{"T":"2019-02-18T11:44:00Z","Sells":[{"Price":"0.001436","Amount":"6502773"},{"Price":"0.001436","Amount":"6502773"}],"Buys":[{"Price":"0.0014276","Amount":"6502773"},{"Price":"0.001426","Amount":"6502773"}]}`
	if err := json.Unmarshal([]byte(depthString), &d); err != nil {
		t.Error(err)
		return
	}
	dt = d.MarketBuyDetectEx(gdecimal.NewFromInt(10000), gdecimal.NewFromFloat64(0.1), 22, gdecimal.NewFromFloat64(0.000001))
	// dt.UnitDealAmount.Mul(dt.PriceReach).String() == "10000.0000000000000000000308"
	// 必须带上精度，否则算下来要大于原始输入quoteAmount了
	if dt.UnitDealAmount.Mul(dt.PriceReach).WithPrec(8).EqualInt(10000) == false {
		gtest.PrintlnExit(t, "DetectedQuote %s(%s x %s) != Original Quote Amount 10000", dt.UnitDealAmount.Mul(dt.PriceReach).String(), dt.UnitDealAmount.String(), dt.PriceReach.String())
	}

	precision := 8
	lot := gdecimal.NewFromInt(1)
	dt = d.MarketBuyDetectEx(gdecimal.NewFromInt(10000), gdecimal.NewFromFloat64(0.1), precision, lot)
	if dt.UnitDealAmount.Mul(dt.PriceReach).WithPrec(precision).String() != "9999.999568" {
		t.Errorf("MarketBuyDetectEx2 error2")
	}

	// 下面这个测试用例涉及精度问题
	d = Depth{
		Time: time.Now(),
		DepthRawData: DepthRawData{
			Sells: OrderBookList{OrderBook{Price: gdecimal.NewFromFloat64(6273.03627303), Amount: gdecimal.NewFromInt(10000000000000)}},
			Buys:  OrderBookList{OrderBook{Price: gdecimal.NewFromFloat64(6273.02372697), Amount: gdecimal.NewFromInt(10000000000000)}},
		},
	}
	quoteAmount, err := gdecimal.NewFromString("5141.73808735821834851056117427")
	gtest.Assert(t, err)
	precision = 8
	lot = gdecimal.NewFromFloat64(0.000001)
	dt = d.MarketBuyDetectEx(quoteAmount, gdecimal.NewFromFloat64(0.01), precision, lot)
	if dt.UnitDealAmount.Mul(dt.PriceReach).Trunc(precision, lot.Float64()).String() != "5141.73181899" {
		gtest.PrintlnExit(t, "DetectedQuote %s(%s x %s) != Expected Quote Amount 5141.73181941", dt.UnitDealAmount.Mul(dt.PriceReach).String(), dt.UnitDealAmount.String(), dt.PriceReach.String())
	}
	if dt.UnitDealAmount.Mul(dt.PriceReach).GreaterThan(quoteAmount) {
		gtest.PrintlnExit(t, "DetectedQuote %s(%s x %s) > Original Quote Amount %s", dt.UnitDealAmount.Mul(dt.PriceReach).String(), dt.UnitDealAmount.String(), dt.PriceReach.String(), quoteAmount)
	}
}

func TestDepth_MarketSellDetectEx(t *testing.T) {
	d := newTestDepth(true)

	dt := d.MarketSellDetectEx(gdecimal.NewFromInt(1), gdecimal.NewFromFloat64(0.1))
	expected := DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(1),
		QuoteDealAmount: gdecimal.NewFromInt(10),
		Price1:          gdecimal.NewFromInt(10),
		PriceReach:      gdecimal.NewFromInt(10),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketSellDetectEx error1")
		return
	}

	dt = d.MarketSellDetectEx(gdecimal.NewFromInt(2), gdecimal.NewFromFloat64(0.1))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(2),
		QuoteDealAmount: gdecimal.NewFromInt(19),
		Price1:          gdecimal.NewFromInt(10),
		PriceReach:      gdecimal.NewFromInt(9),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketSellDetectEx error2")
		return
	}

	dt = d.MarketSellDetectEx(gdecimal.NewFromInt(3), gdecimal.NewFromFloat64(0.1))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(2),
		QuoteDealAmount: gdecimal.NewFromInt(19),
		Price1:          gdecimal.NewFromInt(10),
		PriceReach:      gdecimal.NewFromInt(9),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketSellDetectEx error3")
		return
	}

	dt = d.MarketSellDetectEx(gdecimal.NewFromInt(4), gdecimal.NewFromFloat64(0.1))
	expected = DepthTolerance{
		UnitDealAmount:  gdecimal.NewFromInt(2),
		QuoteDealAmount: gdecimal.NewFromInt(19),
		Price1:          gdecimal.NewFromInt(10),
		PriceReach:      gdecimal.NewFromInt(9),
	}
	if !dt.Equal(expected) {
		t.Errorf("MarketSellDetectEx error4")
		return
	}
}

func TestDepth_Sort(t *testing.T) {
	withOrder := newTestDepth(true)
	outOfOrder := newTestDepth(false)
	outOfOrder.Sort()
	if gjson.MarshalStringDefault(withOrder, false) != gjson.MarshalStringDefault(outOfOrder, false) {
		gtest.PrintlnExit(t, "2 depths should equal")
	}
}
