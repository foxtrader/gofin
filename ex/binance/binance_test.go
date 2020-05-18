package binance

import (
	"fmt"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/apputil/gtest"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gjson"
	"os"
	"testing"
	"time"
)

func TestBinance_GetTicks(t *testing.T) {
	ex, err := New("", "", "socks5://127.0.0.1:7448", nil, "")
	gtest.Assert(t, err)
	ticks, err := ex.GetTicks()
	gtest.Assert(t, err)
	if len(ticks) < 100 {
		gtest.PrintlnExit(t, "GetTicks error1")
	}
}

func TestBinance_GetAccount(t *testing.T) {
	bnc, err := New(os.Getenv("BNC_ACCESS"), os.Getenv("BNC_SECRET"), "socks5://127.0.0.1:7448", nil, "")
	if err != nil {
		t.Error(err)
		return
	}

	acc, err := bnc.GetAccount()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(acc)
}

func TestBinance_Trade(t *testing.T) {
	bnc, err := New("", "", "socks5://127.0.0.1:7448", nil, "")
	if err != nil {
		t.Error(err)
		return
	}

	oi, err := bnc.Trade(fintypes.MarketSpot, fintypes.MarginNo, 1, fintypes.HOLO.Against(fintypes.ETH), fintypes.OrderSideBuyLong, fintypes.OrderTypeLimit, gdecimal.NewFromInt(10000), gdecimal.NewFromFloat64(0.0000023), gdecimal.Zero)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(oi)
}

// TODO: finish test
func TestBinance_GetMarketInfo(t *testing.T) {
	bnc, err := New("", "", "socks5://127.0.0.1:7448", nil, "")
	if err != nil {
		t.Error(err)
		return
	}

	mi, err := bnc.GetMarketInfo()
	if err != nil {
		t.Error(err)
		return
	}

	//fmt.Println(mi.Pairs())
	fmt.Println(gjson.MarshalStringDefault(mi, true))
}

func TestBinance_GetKline(t *testing.T) {
	ex, err := New("", "", "socks5://127.0.0.1:7448", nil, "")
	if err != nil {
		t.Error(err)
		return
	}

	// check perp market
	since := time.Date(2019, 9, 8, 17, 57, 0, 0, time.UTC)
	ks, err := ex.GetKline(fintypes.MarketPerp, fintypes.BTC.Against(fintypes.USDT), fintypes.Period1Min, &since)
	gtest.Assert(t, err)
	if !ks.Items[0].T.Equal(since) {
		t.Errorf("perp verify error")
		return
	}

	// normal check
	since = ex.Property().TradeBeginTime
	firstBTCUSDTTime := time.Date(2017, 8, 17, 04, 00, 00, 0, time.UTC)
	ks, err = ex.GetKline(fintypes.MarketSpot, fintypes.BTC.Against(fintypes.USDT), fintypes.Period1Min, &since)
	if err != nil {
		t.Error(err)
		return
	}
	if !ks.Items[0].T.Equal(firstBTCUSDTTime) {
		t.Errorf("timestamp error")
		return
	}
	if ks.Len() != 1000 {
		t.Errorf("kline length %d != 1000", ks.Len())
		return
	}

	// check some special trade pairs
	// 早期交易大量缺失的，AE/BTC, TRX/BNB
	//since = time.Date(2018, 5, 22, 6, 45, 0, 0, time.UTC)
	firstBTCTUSDTime := time.Date(2019, 3, 13, 4, 0, 0, 0, time.UTC)
	ks, err = ex.GetKline(fintypes.MarketSpot, fintypes.BTC.Against(fintypes.TUSD), fintypes.Period1Min, &since)
	if !ks.Items[0].T.Equal(firstBTCTUSDTime) {
		t.Errorf("timestamp2 error")
		return
	}
	if ks.Len() != 999 {
		t.Errorf("kline2 length %d != 999", ks.Len())
		return
	}
	ks, err = ex.GetKline(fintypes.MarketSpot, "TRX/BNB", fintypes.Period1Min, &ex.Property().TradeBeginTime)
	if ks.Len() != 999 {
		t.Errorf("kline3 length %d != 999", ks.Len())
		return
	}

	// check since time is included in the returned results
	since = time.Date(2019, 3, 13, 4, 1, 0, 0, time.UTC)
	ks, err = ex.GetKline(fintypes.MarketSpot, fintypes.BTC.Against(fintypes.TUSD), fintypes.Period1Min, &since)
	if !ks.Items[0].T.Equal(since) {
		t.Errorf("since verify error")
		return
	}
}

/*
func TestBinance_SubFilled(t *testing.T) {
	ex, err := New("", "", "socks5://127.0.0.1:1086")
	if err != nil {
		t.Error(err)
		return
	}
	retC, _, _, errC, err := ex.SingleRoutineSubFills(NewPair2(BTC, USDT))
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		for {
			e := <-errC
			fmt.Println(e)
		}
	}()
	for {
		v := <-retC
		fmt.Println(v)
	}
}*/

/*
func TestBinance_GetFills(t *testing.T) {
	time.Sleep(time.Second)
	bnc, err := New("341yq8QV3OyhQCBK9dRVTsnrz45Dcul2EqWL10RA4n9kwzMvG36oXS4UlIyJcMrS", "", "socks5://127.0.0.1:1086", nil)
	if err != nil {
		t.Error(err)
		return
	}

	fs, err := bnc.GetFills(NewPair2(BTC, USDT), &FillOption{BeginId: 1, IdLimit: 100})
	if err != nil {
		t.Error(err)
		return
	}
	for _, v := range fs {
		fmt.Println(v)
	}
	if len(fs) != 100 {
		t.Errorf("returns %d records, but 100 expected", len(fs))
		return
	}

	maxFill, err := bnc.MaxFill(NewPair2(BTC, USDT))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(maxFill)
}
*/

/*
func TestBinance_SubMarketStat(t *testing.T) {
	ex, err := New("", "", "socks5://127.0.0.1:7448", nil, "")
	if err != nil {
		t.Error(err)
		return
	}
	retC, _, _, errC, err := ex.SubMarketStat(time.Second * 5)
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		for {
			e := <-errC
			fmt.Println(e)
		}
	}()
	for {
		v := <-retC
		fmt.Println(v.T)
		for k, v := range v.items {
			fmt.Println(k, v)
		}
		fmt.Println("---")
	}
}*/
/*
func TestBinance_SubKline(t *testing.T) {
	ex, err := New("", "", "socks5://127.0.0.1:1086")
	if err != nil {
		t.Error(err)
		return
	}
	retC, _, _, errC, err := ex.SubKline(NewPair2(BTC, USDT), Period1Min)
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		for {
			e := <-errC
			fmt.Println(e)
		}
	}()
	for {
		v := <-retC
		fmt.Println(v)
	}
}
*/
