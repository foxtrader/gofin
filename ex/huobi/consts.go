package huobi

import (
	"github.com/foxtrader/gofin/fintypes"
)

type (
	hbApiUrl string
)

const (
	apiUrlBase                   hbApiUrl = "base"
	apiUrlMarketInfo             hbApiUrl = "market-info"
	apiUrlAccountIds             hbApiUrl = "accounts-info"
	apiUrlAccountBalance         hbApiUrl = "account-balance"
	apiUrlDepth                  hbApiUrl = "depth"
	apiUrlTicks                  hbApiUrl = "ticks"
	apiUrlKline                  hbApiUrl = "kline"
	apiUrlFills                  hbApiUrl = "fills"
	apiUrlMMarginBorrowable      hbApiUrl = "margin-borrowable"
	apiUrlMCrossMarginBorrowable hbApiUrl = "cross-margin-borrowable"
	apiUrlMMarginBorrow          hbApiUrl = "mmargin-borrow"
	apiUrlMCrossMarginBorrow     hbApiUrl = "mcross-margin-borrow"
	apiUrlMMarginRepay           hbApiUrl = "mmargin-repay"
	apiUrlMCrossMarginRepay      hbApiUrl = "mcross-margin-repay"
	apiUrlTransferIn             hbApiUrl = "transfer-in"
	apiUrlTransferOut            hbApiUrl = "transfer-out"
	apiUrlTrade                  hbApiUrl = "trade"
	apiUrlOrders                 hbApiUrl = "orders"
	apiUrlOpenOrders             hbApiUrl = "open-orders"
	apiUrlOrder                  hbApiUrl = "order"
	apiUrlCancelOrder            hbApiUrl = "cancel-order"
)

var apiPathMap = map[fintypes.Market]map[hbApiUrl]string{
	fintypes.MarketSpot: {
		apiUrlBase:                   "https://api.huobi.pro",
		apiUrlMarketInfo:             "/v1/common/symbols",
		apiUrlAccountIds:             "/v1/account/accounts",
		apiUrlAccountBalance:         "/v1/account/accounts/%s/balance",
		apiUrlDepth:                  "/market/depth?symbol=%s&type=step0&depth=%d",
		apiUrlKline:                  "/market/history/kline?period=%s&size=%d&symbol=%s",
		apiUrlFills:                  "/market/history/trade?symbol=%s&size=%d",
		apiUrlMMarginBorrowable:      "/v1/margin/loan-info",
		apiUrlMCrossMarginBorrowable: "/v1/cross-margin/loan-info",
		apiUrlMMarginBorrow:          "/v1/margin/orders",
		apiUrlMCrossMarginBorrow:     "/v1/cross-margin/orders",
		apiUrlMMarginRepay:           "/v1/margin/orders/%s/repay",
		apiUrlMCrossMarginRepay:      "/v1/cross-margin/orders/%s/repay",
		apiUrlTransferIn:             "/v1/dw/transfer-in/margin",
		apiUrlTransferOut:            "/v1/dw/transfer-out/margin",
		apiUrlTrade:                  "/v1/order/orders/place",
		apiUrlOrders:                 "/v1/order/orders",
		apiUrlCancelOrder:            "/v1/order/orders/%s/submitcancel",
	},
	fintypes.MarketFuture: {
		apiUrlBase:           "https://api.hbdm.com",
		apiUrlAccountBalance: "/v1/contract_account_info",
		apiUrlDepth:          "/market/depth?symbol=%s&type=step0",
		apiUrlKline:          "/market/history/kline?period=%s&size=%d&symbol=%s&from=%d&to=%d",
		apiUrlFills:          "/market/history/trade?symbol=%s&size=%d",
		apiUrlTransferIn:     "/v1/futures/transfer",
		apiUrlOrders:         "/api/v1/contract_order_info",
		apiUrlOpenOrders:     "/api/v1/contract_openorders",
		apiUrlCancelOrder:    "/api/v1/contract_cancel",
		apiUrlOrder:          "/api/v1/contract_order_detail",
	},
	fintypes.MarketPerp: {
		apiUrlBase:           "https://swap-vip.btcquant.pro",
		apiUrlAccountBalance: "swap-api/v1/swap_account_info",
		apiUrlDepth:          "/swap-ex/market/depth?contract_code=%s&type=step0",
		apiUrlKline:          "/swap-ex/market/history/kline?period=%s&size=%d&contract_code=%s&from=%d&to=%d",
		apiUrlFills:          "/swap-ex/market/history/trade?contract_code=%s&size=%d",
		apiUrlOrders:         "/swap-api/v1/swap_order_info",
		apiUrlOpenOrders:     "/swap-api/v1/swap_openorders",
		apiUrlOrder:          "/swap-api/v1/swap_order_detail",
		apiUrlCancelOrder:    "/swap-api/v1/swap_cancel",
	},
}
