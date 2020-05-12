package huobi

import (
	"github.com/foxtrader/gofin/fintypes"
)

type (
	ApiUrl string
)

const (
	ApiUrlBase                   ApiUrl = "base"
	ApiUrlMarketInfo             ApiUrl = "market-info"
	ApiUrlAccount                ApiUrl = "account"
	ApiUrlAccountInfo            ApiUrl = "account-info"
	ApiUrlDepth                  ApiUrl = "depth"
	ApiUrlTicks                  ApiUrl = "ticks"
	ApiUrlKline                  ApiUrl = "kline"
	ApiUrlFills                  ApiUrl = "fills"
	ApiUrlMMarginBorrowable      ApiUrl = "margin-borrowable"
	ApiUrlMCrossMarginBorrowable ApiUrl = "cross-margin-borrowable"
	ApiUrlMMarginBorrow          ApiUrl = "mmargin-borrow"
	ApiUrlMCrossMarginBorrow     ApiUrl = "mcross-margin-borrow"
	ApiUrlMMarginRepay           ApiUrl = "mmargin-repay"
	ApiUrlMCrossMarginRepay      ApiUrl = "mcross-margin-repay"
	ApiUrlTransfer               ApiUrl = "transfer"
	ApiUrlTrade                  ApiUrl = "trade"
	ApiUrlOrders                 ApiUrl = "orders"
	ApiUrlOpenOrders             ApiUrl = "open-orders"
	ApiUrlOrder                  ApiUrl = "order"
	ApiUrlCancelOrder            ApiUrl = "cancel-order"
)

var ApiPathMap = map[fintypes.Market]map[ApiUrl]string{
	fintypes.MarketSpot: {
		ApiUrlBase:        "https://api.huobi.pro",
		ApiUrlMarketInfo:  "/v1/common/symbols",
		ApiUrlAccountInfo: "/v1/account/accounts",
		ApiUrlAccount:     "/v1/account/accounts/%s/balance",
		ApiUrlDepth:       "/market/depth?symbol=%s&type=step0&depth=%d",
		ApiUrlTicks:       "/market/tickers",
		ApiUrlKline:       "/market/history/kline?period=%s&size=%d&symbol=%s",
		ApiUrlFills:       "/market/history/trade?symbol=%s&size=%d",
		ApiUrlTransfer:    "/v1/dw/transfer-in/margin",
		ApiUrlTrade:       "/v1/order/orders/place",
		ApiUrlOrders:      "/v1/order/orders",
		ApiUrlOrder:       "/v1/order/orders/%s",
		ApiUrlCancelOrder: "/v1/order/orders/%s/submitcancel",
	},
	MarketMargin: {
		ApiUrlBase:                   "https://api.huobi.pro",
		ApiUrlAccountInfo:            "/v1/account/accounts",
		ApiUrlAccount:                "/v1/account/accounts/%s/balance",
		ApiUrlDepth:                  "/market/depth?symbol=%s&type=step0&depth=%d",
		ApiUrlKline:                  "/market/history/kline?period=%s&size=%d&symbol=%s",
		ApiUrlFills:                  "/market/history/trade?symbol=%s&size=%d",
		ApiUrlMMarginBorrowable:      "/v1/margin/loan-info",
		ApiUrlMCrossMarginBorrowable: "/v1/cross-margin/loan-info",
		ApiUrlMMarginBorrow:          "/v1/margin/orders",
		ApiUrlMCrossMarginBorrow:     "/v1/cross-margin/orders",
		ApiUrlMMarginRepay:           "/v1/margin/orders/%s/repay",
		ApiUrlMCrossMarginRepay:      "/v1/cross-margin/orders/%s/repay",
		ApiUrlTransfer:               "/v1/dw/transfer-out/margin",
		ApiUrlTrade:                  "/v1/order/orders/place",
		ApiUrlOrders:                 "/v1/order/orders",
		ApiUrlCancelOrder:            "/v1/order/orders/%s/submitcancel",
	},
	fintypes.MarketFuture: {
		ApiUrlBase:        "https://api.hbdm.com",
		ApiUrlDepth:       "/market/depth?symbol=%s&type=step0",
		ApiUrlKline:       "/market/history/kline?period=%s&size=%d&symbol=%s&from=%d&to=%d",
		ApiUrlFills:       "/market/history/trade?symbol=%s&size=%d",
		ApiUrlTransfer:    "/v1/futures/transfer",
		ApiUrlOrders:      "/api/v1/contract_order_info",
		ApiUrlOpenOrders:  "/api/v1/contract_openorders",
		ApiUrlCancelOrder: "/api/v1/contract_cancel",
		ApiUrlOrder:       "/api/v1/contract_order_detail",
	},
	fintypes.MarketPerp: {
		ApiUrlBase:        "https://swap-vip.btcquant.pro",
		ApiUrlAccountInfo: "swap-api/v1/swap_account_info",
		ApiUrlDepth:       "/swap-ex/market/depth?contract_code=%s&type=step0",
		ApiUrlKline:       "/swap-ex/market/history/kline?period=%s&size=%d&contract_code=%s&from=%d&to=%d",
		ApiUrlFills:       "/swap-ex/market/history/trade?contract_code=%s&size=%d",
		ApiUrlOrders:      "/swap-api/v1/swap_order_info",
		ApiUrlOpenOrders:  "/swap-api/v1/swap_openorders",
		ApiUrlOrder:       "/swap-api/v1/swap_order_detail",
		ApiUrlCancelOrder: "/swap-api/v1/swap_cancel",
	},
}
