package fintypes

import (
	"fmt"
	"github.com/shawnwyckoff/gopkg/apputil/gtest"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"testing"
	"time"
)

func TestPairExt_Verify(t *testing.T) {
	tcs := gtest.NewCaseList()
	tcs.New().Input(pairExt("BTC/USDT.1min.spot.binance")).Expect(true)
	tcs.New().Input(pairExt("BTC/USDT.1min.binance")).Expect(true)
	tcs.New().Input(pairExt("BTC/USDT.spot.binance")).Expect(true)
	tcs.New().Input(pairExt("BTC/USDT.binance")).Expect(true)
	tcs.New().Input(pairExt("BTC/USDT")).Expect(true)
	tcs.New().Input(pairExt("BTC_USDT")).Expect(true)
	tcs.New().Input(pairExt("BTC($)USDT")).Expect(true) // 注意，Pair中的特殊符号不算错误

	tcs.New().Input(pairExt("BTC/USDT.6min.spot.binance")).Expect(false)
	tcs.New().Input(pairExt("BTC/USDT.1min.sport.binance")).Expect(false)
	tcs.New().Input(pairExt("BTC/USDT.1min.sport.BinancE")).Expect(false)
	tcs.New().Input(pairExt("BTC2/USDT.1min.sport.BinancE")).Expect(false)

	for _, v := range tcs.Get() {
		s := v.Inputs[0].(pairExt)
		expect := v.Expects[0].(bool)
		err := s.Verify()
		if expect == true && err != nil {
			gtest.Assert(t, err)
		}
		if expect == false && err == nil {
			gtest.PrintlnExit(t, "invalid pairExt %s verified passed", s.String())
		}
	}
}

func TestPairExt_Market(t *testing.T) {
	tcs := gtest.NewCaseList()
	tcs.New().Input(pairExt("BTC/USDT.1min.spot.binance")).Expect(MarketSpot)
	tcs.New().Input(pairExt("BTC/USDT.1min.perp.binance")).Expect(MarketPerp)
	tcs.New().Input(pairExt("BTC/USDT.spot.binance")).Expect(MarketSpot)
	tcs.New().Input(pairExt("BTC/USDT.future.huobi")).Expect(MarketFuture)
	tcs.New().Input(pairExt("BTC/USDT.huobi")).Expect(MarketError)

	for _, v := range tcs.Get() {
		s := v.Inputs[0].(pairExt)
		expect := v.Expects[0].(Market)
		got := s.Market()
		if expect != got {
			gtest.PrintlnExit(t, "pairExt %s, Market expect %s but got %s", s.String(), expect, got)
		}
	}
}

func newBinanceExConfig() ExProperty {
	cc := ExProperty{
		Name:                  Binance,
		MaxDepth:              100,
		PairDelimiter:         "",
		PairDelimiterLeftTail: []string{"BULL", "BEAR"},
		PairNormalOrder:       true,
		PairUpperCase:         true,
		MakerFee:              gdecimal.NewFromFloat64(0.001),
		TakerFee:              gdecimal.NewFromFloat64(0.001),
		FillRateLimit:         time.Second, // 真实的数值应该是 time.Second / 4 但那样会出现大量奇怪的超时问题，所以临时改成1s
		KlineRateLimit:        time.Second / 100 * 127,
	}
	cc.Periods = make(map[Period]string)
	cc.Periods[Period1Min] = "1m"
	cc.Periods[Period3Min] = "3m"
	cc.Periods[Period5Min] = "5m"
	cc.Periods[Period15Min] = "15m"
	cc.Periods[Period30Min] = "30m"
	cc.Periods[Period1Hour] = "1h"
	cc.Periods[Period2Hour] = "2h"
	cc.Periods[Period4Hour] = "4h"
	cc.Periods[Period6Hour] = "6h"
	cc.Periods[Period8Hour] = "8h"
	cc.Periods[Period12Hour] = "12h"
	cc.Periods[Period1Day] = "1d"
	cc.Periods[Period1Week] = "1w"
	cc.Periods[Period1MonthFUZZY] = "1M"
	cc.MarketEnabled = map[Market]bool{}
	cc.MarketEnabled[MarketSpot] = true
	cc.MarketEnabled[MarketFuture] = false
	cc.MarketEnabled[MarketPerp] = true
	cc.TradeBeginTime = time.Date(2017, 7, 14, 00, 00, 00, 0, time.UTC) // this time is approximation, more exact time seem like 2017-07-14 04:00:00 +0000 UTC

	return cc
}

func TestToPair_String(t *testing.T) {

	type testItem struct {
		a        Asset
		b        Asset
		expected string
	}
	var testItems []testItem
	testItems = append(testItems,
		testItem{BTC, USDT, "BTC/USDT"},
		testItem{BTC, USD, "BTC/USD"},
		testItem{XAU, USD, "XAU/USD"},
		testItem{HOLO, TUSD, "HOT/TUSD"},
		testItem{BTC, CNY, "BTC/CNY"},
		testItem{BTC, BNB, "BTC/BNB"},
		testItem{ZEC, USD, "ZEC/USD"},
	)

	for _, v := range testItems {
		s := v.a.Against(v.b)
		if s.String() != v.expected {
			t.Errorf("expected %s, but get %s", v.expected, s.String())
			return
		}
	}

}

func TestParsePairCustom(t *testing.T) {
	cl := gtest.NewCaseList()
	cl.New().Input("BTCUSDT").Expect(BTC.Against(USDT))
	cl.New().Input("ZECUSD").Expect(ZEC.Against(USD))
	cl.New().Input("XRPEUR").Expect(XRP.Against(EUR))
	cl.New().Input("ETHBTC").Expect(ETH.Against(BTC))
	cl.New().Input("BULLUSDT").Expect(Pair("BULL/USDT"))
	cl.New().Input("XRPBULLUSDT").Expect(Pair("XRPBULL/USDT"))
	cl.New().Input("ETHBULLUSDT").Expect(Pair("ETHBULL/USDT"))
	cl.New().Input("ETHBULLBUSD").Expect(Pair("ETHBULL/BUSD"))
	cl.New().Input("ETHBEARBUSD").Expect(Pair("ETHBEAR/BUSD"))

	for _, v := range cl.Get() {
		bec := newBinanceExConfig()
		pair, err := ParsePairCustom(v.Inputs[0].(string), &bec)
		if err != nil {
			t.Errorf("string(%s) parse error:%s", v.Inputs[0].(string), err.Error())
			return
		}
		if pair.String() != v.Expects[0].(Pair).String() {
			t.Errorf("expected %s, but get %s", v.Expects[0].(Pair).String(), pair.String())
			return
		}
	}
}

func TestFindSamePairs(t *testing.T) {
	toTest := map[Platform][]Pair{}
	huobiPairs := []Pair{HOLO.Against(ETH), HOLO.Against(BTC), HOLO.Against(USDT), LTC.Against(ETH), LTC.Against(BTC)}
	gatePairs := []Pair{HOLO.Against(ETH), HOLO.Against(BTC), ZEC.Against(USDT), LTC.Against(BTC)}
	bncPairs := []Pair{HOLO.Against(ETH), HOLO.Against(BTC), HOLO.Against(USDT), LTC.Against(BTC), LTC.Against(PAX)}
	toTest[Huobi] = huobiPairs
	toTest[Gate] = gatePairs
	toTest[Binance] = bncPairs

	res := FindSamePairs(toTest)
	if len(res[Pair("HOT/ETH")]) != 3 || len(res[Pair("HOT/BTC")]) != 3 || len(res[Pair("LTC/BTC")]) != 3 || len(res[Pair("HOT/USDT")]) != 2 {
		gtest.PrintlnExit(t, "FindSamePairs error1")
	}
}

func TestParsePairExtString(t *testing.T) {
	s := "ETH/USDT.spot"
	pair, period, market, platform, err := ParsePairExtString(s)
	gtest.Assert(t, err)

	if pair.String() != "ETH/USDT" || market == nil || *market != MarketSpot {
		gtest.PrintlnExit(t, "ParsePairExtString error")
	}
	fmt.Println(pair, period, platform, market)
}
