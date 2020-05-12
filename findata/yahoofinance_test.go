package findata

import (
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/container/gjson"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"testing"
)

func TestYahooFinance_GetKlineEx(t *testing.T) {
	yf, _ := NewYahooFinance("")

	beginDate, _ := gtime.NewDate(2018, 1, 2)
	beginTime := beginDate.ToTimeUTC()
	info, err := yf.GetKlineEx(fintypes.Nyse, fintypes.MarketSpot, fintypes.NewPair("ANTM", "USD"), fintypes.Period1Day, &beginTime)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(gjson.MarshalStringDefault(info, false))

	beginDate, _ = gtime.NewDate(2018, 1, 2)
	beginTime = beginDate.ToTimeUTC()
	info, err = yf.GetKlineEx(fintypes.Nasdaq, fintypes.MarketSpot, fintypes.NewPair("AAPL", "USD"), fintypes.Period1Day, &beginTime)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(gjson.MarshalStringDefault(info, false))

	beginDate, _ = gtime.NewDate(2018, 1, 2)
	beginTime = beginDate.ToTimeUTC()
	info, err = yf.GetKlineEx(fintypes.Sse, fintypes.MarketSpot, fintypes.NewPair("600519", "CNY"), fintypes.Period1Day, &beginTime)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(gjson.MarshalStringDefault(info, false))

	beginDate, _ = gtime.NewDate(2018, 1, 2)
	beginTime = beginDate.ToTimeUTC()
	info, err = yf.GetKlineEx(fintypes.PlatformOpen, fintypes.MarketSpot, fintypes.IndexToPairP(fintypes.IndexDJI).Pair(), fintypes.Period1Day, &beginTime)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(gjson.MarshalStringDefault(info, false))
}
