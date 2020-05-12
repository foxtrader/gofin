package huobi

import (
	"github.com/foxtrader/gofin/fintypes"
	"testing"
	"time"

	"github.com/shawnwyckoff/gopkg/apputil/glogs"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/stretchr/testify/assert"
)

var (
	apikey    = "a72631b3-6d1e289e-a64b10cd-bg2hyw2dfg"
	secretkey = "8df51ad1-e477e214-cad7fb3d-1fc64"
)

var hbpro, _ = New(apikey, secretkey, "")

func init() {
	glogs.Log.SetLevel(glogs.DEBUG)
}

func TestHuobiPro_Config(t *testing.T) {
	info := hbpro.Config()
	t.Log(info)
}

func TestHuobiPro_GetAccountInfo(t *testing.T) {
	// info, err := hbpro.GetAccountInfo(MarketTypeMargin)
	info, err := hbpro.GetAccountInfo(fintypes.MarketSpot)
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetMarketInfo(t *testing.T) {
	info, err := hbpro.GetMarketInfo(fintypes.MarketSpot)
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetAccount(t *testing.T) {
	info, err := hbpro.GetAccount(fintypes.MarketSpot)
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetDepth(t *testing.T) {
	info, err := hbpro.GetDepth(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt"), 5) // 测试 现货/杠杆 交易
	// info, err := hbpro.GetDepth(MarketTypeFuture, NewPair("BTC-USD", "USD"), 5) // 测试 交割/永续 交易
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetTicks(t *testing.T) {
	info, err := hbpro.GetTicks(fintypes.MarketSpot)
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetKline(t *testing.T) {
	st, _ := time.ParseDuration("-1500s")
	tm := time.Now().Add(st)
	period, _ := fintypes.ParsePeriod("5min")
	info, err := hbpro.GetKline(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt"), period, &tm) // 测试 现货/杠杆 交易
	// info, err := hbpro.GetKline(MarketTypeFuture, NewPair("BTC-USD", "USD"), period, &tm) // 测试 交割/永续 交易
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetFills(t *testing.T) {
	var fromId int64
	info, err := hbpro.GetFills(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt"), &fromId, 3) // 测试 现货/杠杆 交易
	// info, err := hbpro.GetFills(MarketTypeFuture, NewPair("BTC-USD", "USD"), &fromId, 3) // 测试 交割/永续 交易
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetBorrowable(t *testing.T) {
	info, err := hbpro.GetBorrowable(fintypes.NewPair("btc", "usdt"), Unit, MMargin)
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_Borrow(t *testing.T) {
	err := hbpro.Borrow(fintypes.NewPair("btc", "usdt"), Quote, MMargin, gdecimal.NewFromFloat64(0.5))
	assert.Nil(t, err)
	t.Log(err)
}

func TestHuobiPro_Repay(t *testing.T) {
	err := hbpro.Repay("1116237737", MMargin, gdecimal.NewFromFloat64(0.5)) // order ID不存在
	assert.Nil(t, err)
	t.Log(err)
}

func TestHuobiPro_Transfer(t *testing.T) {
	err := hbpro.Transfer(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt"), Quote, gdecimal.NewFromFloat64(0.5))
	assert.Nil(t, err)
	t.Log(err)
}

func TestHuobiPro_Trade(t *testing.T) {
	info, err := hbpro.Trade(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt"), TradeTypeSideLimitBuy,
		gdecimal.NewFromFloat64(0.01), gdecimal.NewFromFloat64(8855.5))
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetAllOrders(t *testing.T) {
	info, err := hbpro.GetAllOrders(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt")) // 测试 现货/杠杆 交易
	// info, err := hbpro.GetAllOrders(MarketTypeFuture, NewPair("BTC-USD", "USD")) // 测试 交割 交易
	// info, err := hbpro.GetAllOrders(MarketTypePerp, NewPair("BTC-USD", "USD"))   // 测试 永续 交易
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetOpenOrders(t *testing.T) {
	info, err := hbpro.GetOpenOrders(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt")) // 测试 现货/杠杆 交易
	// info, err := hbpro.GetOpenOrders(MarketSpot, NewPair("BTC-USD", "USD")) // 测试 交割 交易
	// info, err := hbpro.GetOpenOrders(MarketSpot, NewPair("BTC-USD", "USD")) // 测试 永续 交易
	assert.Nil(t, err)
	t.Log(info)
}

func TestHuobiPro_GetOrder(t *testing.T) {
	ord, err := hbpro.GetOrder(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt"), "1116237737") // order ID不存在
	assert.Nil(t, err)
	t.Log(ord)
}

func TestHuobiPro_CancelOrder(t *testing.T) {
	err := hbpro.CancelOrder(fintypes.MarketSpot, fintypes.NewPair("btc", "usdt"), "1116237737") // order ID不存在
	assert.Nil(t, err)
}
