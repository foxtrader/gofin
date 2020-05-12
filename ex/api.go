package ex

import (
	"github.com/foxtrader/gofin/ex/binance"
	"github.com/foxtrader/gofin/fintypes"
	fintypes2 "github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"strings"
	"time"
)

type (

	// Ex interface
	Ex interface {
		// exchange custom properties
		Property() *fintypes.ExProperty

		// get all supported pairs, min trade amount...
		GetMarketInfo() (*fintypes.MarketInfo, error)

		// get account info includes all currency balances.
		GetAccount() (*fintypes.Account, error)

		// get max open order books
		GetDepth(market fintypes.Market, target fintypes.Pair) (*fintypes.Depth, error)

		// get candle bars
		GetKline(market fintypes.Market, target fintypes.Pair, period fintypes.Period, since *time.Time) (*fintypes2.Kline, error)

		// get all ticks
		GetTicks() (map[fintypes.PairM]fintypes.Tick, error)

		// margin account borrowable
		GetBorrowable(margin fintypes.Margin, asset string) (gdecimal.Decimal, error)

		// margin account borrow
		Borrow(margin fintypes.Margin, asset string, amount gdecimal.Decimal) error

		// margin account repay
		Repay(margin fintypes.Margin, asset string, amount gdecimal.Decimal) error

		// transfer between different sub accounts
		Transfer(asset string, amount gdecimal.Decimal, from, to fintypes.SubAcc) error

		// amount: always unit amount, not quote amount, whether buy or sell
		// price: when market-buy/market-sell, price will be ignored
		Trade(market fintypes.Market, margin fintypes.Margin, leverage int, target fintypes.Pair, side fintypes.OrderSide, orderType fintypes.OrderType, amount, price gdecimal.Decimal) (*fintypes.OrderId, error)

		// get all my history orders' info
		GetAllOrders(market fintypes.Market, margin fintypes.Margin, target fintypes.Pair) ([]fintypes.Order, error)

		// get all my unfinished orders' info
		GetOpenOrders(market fintypes.Market, margin fintypes.Margin, target fintypes.Pair) ([]fintypes.Order, error)

		// get order info by id
		GetOrder(id fintypes.OrderId) (*fintypes.Order, error)

		// cancel unfinished order by id
		CancelOrder(id fintypes.OrderId) error

		// get exchange match results history, not history of current account but whole market
		//GetFills(market Market, target Pair, since Since) ([]Fill, error)
	}
)

// email is required in living trading, but not required in kline spider
func NewEx(name fintypes.Platform, apiKey, apiSecret, proxy string, c gtime.Clock, email string) (Ex, error) {
	switch strings.ToLower(name.String()) {
	case strings.ToLower(fintypes.Binance.String()):
		return binance.New(apiKey, apiSecret, proxy, c, email)
	default:
		return nil, gerror.Errorf("unsupported exchange(%s)", name.String())
	}
}
