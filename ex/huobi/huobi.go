package huobi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/net/ghttp"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shawnwyckoff/gopkg/apputil/gparam"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/net/ghttputils"
)

/**
docs: https://huobiapi.github.io/docs/spot/v1/cn/
*/

type Client struct {
	config     *fintypes.ExProperty
	httpClient *http.Client
	accessKey  string
	secretKey  string
	proxy      string
}

func New(apiKey, secretKey, proxy string, c gtime.Clock, email string) (*Client, error) {
	hb := &Client{}
	hb.config = &fintypes.ExProperty{
		Name:                   fintypes.Huobi,
		Email:                  email,
		MaxDepth:               150,
		PairDelimiter:          "",
		PairDelimiterLeftTail:  nil,
		PairDelimiterRightHead: nil,
		PairNormalOrder:        true,
		PairUpperCase:          false,
		PairsSeparator:         ",",
		Periods: map[fintypes.Period]string{fintypes.Period1Min: "1min", fintypes.Period5Min: "5min", fintypes.Period15Min: "15min",
			fintypes.Period30Min: "30min", fintypes.Period1Hour: "60min", fintypes.Period4Hour: "4hour",
			fintypes.Period1Day: "1day", fintypes.Period1Week: "1week", fintypes.Period1MonthFUZZY: "1mon",
			fintypes.Period1YearFUZZY: "1year",
		},
		OrderStatus: map[fintypes.OrderStatus]string{
			fintypes.OrderStatusNew:             "submitted",
			fintypes.OrderStatusPartiallyFilled: "partial-filled", fintypes.OrderStatusFilled: "filled",
			fintypes.OrderStatusPartiallyCanceled: "partial-canceled", fintypes.OrderStatusCanceled: "canceled", fintypes.OrderStatusCanceling: "canceling",
		},
		RateLimits: map[fintypes.ExApi]time.Duration{
			fintypes.ExApiGetKline: time.Second / 10,
			fintypes.ExApiGetFill:  time.Second / 10,
		},
		MarketEnabled: map[fintypes.Market]bool{
			fintypes.MarketSpot:   true,
			fintypes.MarketFuture: true,
			fintypes.MarketPerp:   true,
		},
		Clock:          c,
		IsBackTestEx:   false,
		TradeBeginTime: time.Date(2017, 10, 1, 20, 20, 20, 20, gtime.TimeZoneAsiaShanghai),
	}

	hb.httpClient = http.DefaultClient
	if err := ghttp.SetProxy(hb.httpClient, proxy); err != nil {
		return nil, err
	}
	hb.accessKey = apiKey
	hb.secretKey = secretKey
	hb.proxy = proxy
	return hb, nil
}

// exchange custom settings
func (hb *Client) Config() *fintypes.ExProperty {
	return hb.config
}

// get all supported pairs, min trade amount
func (hb *Client) GetMarketInfo() (*fintypes.MarketInfo, error) {
	res := &fintypes.MarketInfo{Infos: map[fintypes.PairM]fintypes.PairInfo{}}

	for market, enabled := range hb.config.MarketEnabled {
		if !enabled {
			continue
		}
		baseUrl, ok := apiPathMap[market][apiUrlBase]
		if !ok {
			return nil, gerror.New("apiUrlBase for %s not found", market)
		}
		subUrl, ok := apiPathMap[market][apiUrlMarketInfo]
		if !ok {
			return nil, gerror.New("apiUrlMarketInfo for %s not found", market)
		}
		reqUrl := baseUrl + subUrl
		respMap, err := ghttp.GetMap(reqUrl, hb.proxy, time.Minute)
		if err != nil {
			return nil, err
		}
		if respMap["status"].(string) != "ok" {
			return nil, errors.New(respMap["err-code"].(string))
		}

		data, ok := respMap["data"].([]interface{})
		if !ok {
			return nil, errors.New("nil data in map")
		}
		for _, v := range data {
			vMap := v.(map[string]interface{})
			unitSym := vMap["base-currency"].(string)
			quoteSym := vMap["quote-currency"].(string)
			state := vMap["state"].(string)
			amountPrecision := (vMap["amount-precision"]).(int)
			pricePrecision := (vMap["price-precision"]).(int)
			pairInfo := fintypes.PairInfo{}
			pairInfo.Enabled = "online" == state
			pairInfo.UnitPrecision = amountPrecision
			pairInfo.QuotePrecision = pricePrecision
			pairInfo.UnitMin = gdecimal.NewFromFloat64((vMap["min-order-amt"]).(float64))
			pairInfo.UnitStep = pairInfo.UnitMin
			res.Infos[fintypes.NewPair(unitSym, quoteSym).SetM(market)] = pairInfo
		}
	}

	return res, nil
}

func parseMarketMargin(s string) (fintypes.Market, fintypes.Margin, bool) {
	switch s {
	case "spot":
		return fintypes.MarketSpot, fintypes.MarginNo, true
	case "margin":
		return fintypes.MarketSpot, fintypes.MarginIsolated, true
	case "super-margin":
		return fintypes.MarketSpot, fintypes.MarginCross, true
	default:
		return fintypes.MarketError, fintypes.MarginError, false
	}
}

type accId struct {
	market fintypes.Market
	margin fintypes.Margin
	id     string
}

func (hb *Client) getAccountIds() ([]accId, error) {
	var ids []accId
	for market, enabled := range hb.config.MarketEnabled {
		if !enabled {
			continue
		}
		baseUrl, ok := apiPathMap[market][apiUrlBase]
		if !ok {
			return nil, gerror.New("apiUrlBase for %s not found", market)
		}
		subUrl, ok := apiPathMap[market][apiUrlMarketInfo]
		if !ok {
			return nil, gerror.New("apiUrlMarketInfo for %s not found", market)
		}
		params := &url.Values{}
		hb.signForm(params, "GET", baseUrl, subUrl)
		reqUrl := baseUrl + subUrl + "?" + params.Encode()
		respMap, err := ghttp.GetMap(reqUrl, hb.proxy, time.Minute)
		if err != nil {
			return nil, err
		}
		if respMap["status"].(string) != "ok" {
			return nil, errors.New(respMap["err-code"].(string))
		}

		data := respMap["data"].([]interface{})
		for _, v := range data {
			vMap := v.(map[string]interface{})
			market, margin, ok := parseMarketMargin(vMap["type"].(string))
			if !ok {
				// 未识别的type，越过
				continue
			}
			item := accId{
				market: market,
				margin: margin,
				id:     vMap["id"].(string),
			}
			ids = append(ids, item)
		}
	}

	return ids, nil
}

func getAccId(ids []accId, market fintypes.Market, margin fintypes.Margin) (string, bool) {
	for _, v := range ids {
		if v.market == market && v.margin == margin {
			return v.id, true
		}
	}
	return "", false
}

// get account info includes all currency balances
func (hb *Client) GetAccount() (*fintypes.Account, error) {
	ids, err := hb.getAccountIds()
	if err != nil {
		return nil, err
	}

	res := fintypes.NewEmptyAccount()

	accId, ok := getAccId(ids, fintypes.MarketSpot, fintypes.MarginNo)
	if !ok {
		return nil, gerror.New("MarketSpot MarginNo account id not found")
	}
	baseUrl, ok := apiPathMap[fintypes.MarketSpot][apiUrlBase]
	if !ok {
		return nil, gerror.New("apiUrlBase not found")
	}
	subUrl, ok := apiPathMap[fintypes.MarketSpot][apiUrlAccountBalance]
	if !ok {
		return nil, gerror.New("apiUrlAccountIds not found")
	}
	params := &url.Values{}
	params.Set("accountId-id", accId)
	hb.signForm(params, "GET", baseUrl, subUrl)
	reqUrl := baseUrl + subUrl + "?" + params.Encode()
	respMap, err := ghttp.GetMap(reqUrl, hb.proxy, time.Minute)
	if err != nil {
		return nil, err
	}
	if respMap["status"].(string) != "ok" {
		return nil, gerror.Errorf("result status %s != ok", respMap["err-code"].(string))
	}
	dataMap := respMap["data"].(map[string]interface{})
	if dataMap["state"].(string) != "working" {
		return nil, gerror.Errorf("data state %s != working", dataMap["state"].(string))
	}

	list := dataMap["list"].([]interface{})

	for _, v := range list {
		vMap := v.(map[string]interface{})
		currencyStr := vMap["currency"].(string)
		typeStr := vMap["type"].(string)
		balance, err := gdecimal.NewFromString(vMap["balance"].(string))
		if err != nil {
			return nil, err
		}

		assetProperty := fintypes.AssetProperty{}
		assetProperty.Market = fintypes.MarketSpot
		assetProperty.Margin = fintypes.MarginNo
		assetProperty.Asset = strings.ToUpper(currencyStr)
		switch typeStr {
		case "trade":
			res.SetFreeAmount(assetProperty, balance)
		case "frozen":
			res.SetLockedAmount(assetProperty, balance)
		default:
			return nil, gerror.Errorf("invalid typeStr %s", typeStr)
		}
	}

	/*
		baseUrl := apiPathMap[fintypes.MarketFuture][apiUrlAccountBalance]
		var resp []struct {
			Symbol            string  `json:"symbol"`
			MarginBalance     float64 `json:"margin_balance"`
			MarginPosition    float64 `json:"margin_position"`
			ProfitUnreal      float64 `json:"profit_unreal"`
			WithdrawAvailable float64 `json:"withdraw_available"`
		}
		params := &url.Values{}
		if err := hb.doRequest(baseUrl, params, &resp); err != nil {
			return nil, err
		}
		for _, v := range resp {
			fintypes.AssetAmount{
				MaxWithdrawAmount:     gdecimal.NewFromFloat64(v.WithdrawAvailable),
				PositionInitialMargin: gdecimal.NewFromFloat64(v.MarginPosition),
				UnrealizedProfit:      gdecimal.NewFromFloat64(v.ProfitUnreal),
				WalletBalance:         gdecimal.NewFromFloat64(v.MarginBalance),
			}
		}
	*/

	return res, nil
}

// get open order books
func (hb *Client) GetDepth(market fintypes.Market, pair fintypes.Pair, limit int) (*fintypes.Depth, error) {
	var depth *fintypes.Depth
	hb.setBaseUrl(market)
	if fintypes.MarketSpot == market || MarketMargin == market {
		n := 5
		if limit <= 5 {

			n = 5
		} else if limit <= 10 {
			n = 10
		} else if limit <= 20 {
			n = 20
		} else {
			n = 150
		}
		url := hb.baseUrl + ApiPathMap[market][ApiUrlDepth]
		urlStr := fmt.Sprintf(url, pair.Format("", false), n)
		respmap, err := ghttputils.HttpGet(hb.httpClient, urlStr)
		if err != nil {
			return nil, err
		}
		if "ok" != respmap["status"].(string) {
			return nil, errors.New(respmap["err-msg"].(string))
		}

		tick, _ := respmap["tick"].(map[string]interface{})

		depth = hb.parseDepthData(tick, limit)
		mills := utils.ToInt64(tick["ts"].(float64))
		depth.Time = time.Unix(int64(mills/1000), int64(mills%1000)*int64(time.Millisecond))
	} else if fintypes.MarketFuture == market || fintypes.MarketPerp == market {
		urlStr := hb.baseUrl + fmt.Sprintf(ApiPathMap[market][ApiUrlDepth], strings.ToUpper(string(pair)))
		ret, err := ghttputils.HttpGet(hb.httpClient, urlStr)
		if err != nil {
			return nil, err
		}

		s := ret["status"].(string)
		if s == "error" {
			return nil, errors.New(ret["err_msg"].(string))
		}

		tick, _ := ret["tick"].(map[string]interface{})
		depth = hb.parseDepthData(tick, limit)
		mills := utils.ToInt64(tick["ts"].(float64))
		depth.Time = time.Unix(int64(mills/1000), int64(mills%1000)*int64(time.Millisecond))
	} else {
	}

	return depth, nil
}

// get all ticks
func (hb *Client) GetTicks() (map[fintypes.PairM]fintypes.Tick, error) {
	hb.setBaseUrl(market)
	url := hb.baseUrl + "/market/tickers"
	respmap, err := ghttputils.HttpGet(hb.httpClient, url)
	if err != nil {
		return nil, err
	}
	if respmap["status"].(string) == "error" {
		return nil, errors.New(respmap["err-msg"].(string))
	}
	data, ok := respmap["data"].([]interface{})
	if !ok {
		return nil, errors.New("response format error")
	}

	ticks := make(map[PairExt]*fintypes.Tick)
	for _, v := range data {
		_sym := v.(map[string]interface{})
		_symbol := _sym["symbol"].(string)
		_close := _sym["close"].(float64)
		_low := _sym["low"].(float64)
		_high := _sym["high"].(float64)
		_vol := _sym["vol"].(float64)

		tick := new(fintypes.Tick)
		tick.Time = time.Now()
		tick.Last = gdecimal.NewFromFloat64(_close)
		tick.Buy = gdecimal.Zero
		tick.Sell = gdecimal.Zero
		tick.High = gdecimal.NewFromFloat64(_high)
		tick.Low = gdecimal.NewFromFloat64(_low)
		tick.Volume = gdecimal.NewFromFloat64(_vol)
		tpair, _ := fintypes.ParsePairCustom(_symbol, hb.config)
		pairExt := tpair.SetM(market)
		ticks[pairExt] = tick
	}
	for k, v := range ticks {
		pairTicks.Ticks[k] = *v
	}

	return pairTicks, nil
}

// get candle bars
func (hb *Client) GetKline(market fintypes.Market, target fintypes.Pair, period fintypes.Period, since *time.Time) (*fintypes2.Kline, error) {
	var klines *fintypes2.Kline
	hb.setBaseUrl(market)
	if fintypes.MarketSpot == market || MarketMargin == market {
		url := hb.baseUrl + ApiPathMap[market][ApiUrlKline]
		symbol := target.Format("", false)
		periodSec := period.ToSeconds()
		times := int64(time.Since(*since).Seconds()) / periodSec
		if times > 2000 {
			times = 2000
		}
		periodSecStr := period.String()
		respmap, err := ghttputils.HttpGet(hb.httpClient, fmt.Sprintf(url, periodSecStr, times, symbol))
		if err != nil {
			return nil, err
		}
		data, ok := respmap["data"].([]interface{})
		if !ok {
			return nil, errors.New("response format error")
		}

		klines = &fintypes2.Kline{}
		klines.Pair = PairExt(string(target))
		klines.Period = period
		for _, e := range data {
			item := e.(map[string]interface{})
			klines.Items = append(klines.Items, fintypes2.Bar{
				O: gdecimal.NewFromFloat64(item["open"].(float64)),
				C: gdecimal.NewFromFloat64(item["close"].(float64)),
				H: gdecimal.NewFromFloat64(item["high"].(float64)),
				L: gdecimal.NewFromFloat64(item["low"].(float64)),
				V: gdecimal.NewFromFloat64(item["vol"].(float64)),
				T: time.Unix(int64(item["id"].(float64)), 0),
			})
		}
	} else if fintypes.MarketFuture == market || fintypes.MarketPerp == market {
		url := hb.baseUrl + ApiPathMap[market][ApiUrlKline]
		symbol := target.Format("", true)
		periodSec := period.ToSeconds()
		times := int64(time.Since(*since).Seconds()) / periodSec
		periodSecStr := period.String()

		respmap, err := ghttputils.HttpGet(hb.httpClient, fmt.Sprintf(url, periodSecStr,
			times, symbol, since.Unix(), time.Now().Unix())) // 默认到现在为止
		if err != nil {
			return nil, err
		}

		data, ok := respmap["data"].([]interface{})
		if !ok {
			return nil, errors.New("response format error")
		}

		klines = new(fintypes2.Kline)
		klines.Pair = PairExt(string(target))
		klines.Period = period
		for _, e := range data {
			item := e.(map[string]interface{})
			klines.Items = append(klines.Items, fintypes2.Bar{
				O: gdecimal.NewFromFloat64(item["open"].(float64)),
				C: gdecimal.NewFromFloat64(item["close"].(float64)),
				H: gdecimal.NewFromFloat64(item["high"].(float64)),
				L: gdecimal.NewFromFloat64(item["low"].(float64)),
				V: gdecimal.NewFromFloat64(item["vol"].(float64)),
				T: time.Unix(int64(item["id"].(float64)), 0),
			})
		}
	} else {
	}

	return klines, nil
}

// get exchange filled trades history
func (hb *Client) GetFills(market fintypes.Market, target fintypes.Pair, fromId *int64, limit int) ([]fintypes.Fill, error) {
	var fills []fintypes.Fill
	hb.setBaseUrl(market)
	url := hb.baseUrl + ApiPathMap[market][ApiUrlFills]
	urlStr := fmt.Sprintf(url, target.Format("", false), limit)
	respmap, err := ghttputils.HttpGet(hb.httpClient, urlStr)
	if err != nil {
		return nil, err
	}
	if respmap["status"].(string) == "error" {
		return nil, errors.New(respmap["err-msg"].(string))
	}
	tickmap, ok := respmap["data"].([]interface{})
	if !ok {
		return nil, errors.New("tick assert error")
	}

	for _, e := range tickmap {
		tickInter := e.(map[string]interface{})
		itemInters := tickInter["data"].([]interface{})
		for _, z := range itemInters {
			item := z.(map[string]interface{})
			fills = append(fills, fintypes.Fill{
				Id:      int64(tickInter["id"].(float64)),
				Time:    time.Unix(int64(item["ts"].(float64)/1000), 0),
				Price:   gdecimal.NewFromFloat64(item["price"].(float64)),
				UnitQty: gdecimal.NewFromFloat64(item["amount"].(float64)),
				Side:    SideType(item["direction"].(string)),
			})
		}
	}

	return fills, nil
}

// margin account borrowable
func (hb *Client) GetBorrowable(target fintypes.Pair, tp UnitQuoteType, margin MarginType) (gdecimal.Decimal, error) {
	var path string
	params := &url.Values{}
	if MMargin == margin {
		path = "/v1/margin/loan-info"
		params.Set("symbols", target.Format("", false))
	} else {
		path = "/v1/cross-margin/loan-info"
	}
	hb.setBaseUrl(MarketMargin)
	hb.buildPostForm("GET", path, params)
	urlStr := hb.baseUrl + path + "?" + params.Encode()
	respmap, err := ghttputils.HttpGet(hb.httpClient, urlStr)
	if err != nil {
		return gdecimal.Zero, err
	}
	if respmap["status"].(string) == "error" {
		return gdecimal.Zero, errors.New(respmap["err-msg"].(string))
	}

	tickmap, ok := respmap["data"].([]interface{})
	if !ok {
		return gdecimal.Zero, errors.New("tick assert error")
	}
	asset := strings.ToLower(target.Dist(tp))
	var loanableAmt gdecimal.Decimal
	if MMargin == margin {
		for _, e := range tickmap {
			tickInter := e.(map[string]interface{})
			items := tickInter["currencies"].([]interface{})
			for _, v := range items {
				item := v.(map[string]interface{})
				log.Print(item["currency"].(string))
				if asset == item["currency"].(string) {
					loanableAmt, _ = gdecimal.NewFromString(item["loanable-amt"].(string))
					break
				}
			}
		}
	} else {
		for _, e := range tickmap {
			item := e.(map[string]interface{})
			if asset == item["currency"].(string) {
				loanableAmt, _ = gdecimal.NewFromString(item["loanable-amt"].(string))
				break
			}
		}
	}

	return loanableAmt, nil
}

// margin account borrow
func (hb *Client) Borrow(target fintypes.Pair, tp UnitQuoteType, margin MarginType, amount gdecimal.Decimal) error {
	var path string
	params := &url.Values{}
	hb.setBaseUrl(MarketMargin)
	err0 := hb.setNewHuobiId(MarketMargin)
	if nil != err0 {
		return err0
	}
	params.Set("accountId-id", hb.accountId)
	if MMargin == margin {
		path = "/v1/margin/orders"
	} else {
		path = "/v1/cross-margin/orders"
	}
	hb.buildPostForm("POST", path, params)
	urlStr := hb.baseUrl + path + "?" + params.Encode()
	asset := strings.ToLower(target.Dist(tp))
	postData := make(map[string]string)
	postData["symbol"] = target.Format("", false)
	postData["currency"] = asset
	postData["amount"] = amount.String()
	respData, err := ghttputils.HttpPostForm4(hb.httpClient, urlStr, postData, nil)
	if err != nil {
		return err
	}

	var respmap map[string]interface{}
	json.Unmarshal(respData, &respmap)
	status := respmap["status"].(string)
	if status == "error" {
		return errors.New(respmap["err-msg"].(string))
	}

	return nil
}

// margin account repay
func (hb *Client) Repay(orderId string, margin MarginType, amount gdecimal.Decimal) error {
	var path string
	gparams := &url.Values{}
	hb.setBaseUrl(MarketMargin)
	err0 := hb.setNewHuobiId(MarketMargin)
	if nil != err0 {
		return err0
	}
	gparams.Set("accountId-id", hb.accountId)
	if MMargin == margin {
		path = fmt.Sprintf("/v1/margin/orders/%s/repay", orderId)
	} else {
		path = fmt.Sprintf("/v1/cross-margin/orders/%s/repay", orderId)
	}
	hb.buildPostForm("POST", path, gparams)
	urlStr := hb.baseUrl + path + "?" + gparams.Encode()
	postData := make(map[string]string)
	postData["order-id"] = orderId
	postData["amount"] = amount.String()
	respData, err := ghttputils.HttpPostForm4(hb.httpClient, urlStr, postData, nil)
	if err != nil {
		return err
	}

	var respmap map[string]interface{}
	json.Unmarshal([]byte(respData), &respmap)
	status := respmap["status"].(string)
	if status == "error" {
		return errors.New(respmap["err-msg"].(string))
	}

	return nil
}

// transfer between spot and margin account
func (hb *Client) Transfer(market fintypes.Market, target fintypes.Pair, tp UnitQuoteType, amount gdecimal.Decimal) error {
	if market != fintypes.MarketSpot && market != MarketMargin {
		return errors.New("Access denied")
	}

	gparams := &url.Values{}
	hb.setBaseUrl(market)
	err0 := hb.setNewHuobiId(market)
	if nil != err0 {
		return err0
	}
	gparams.Set("accountId-id", hb.accountId)
	path := ApiPathMap[market][ApiUrlTransfer]
	hb.buildPostForm("POST", path, gparams)
	urlStr := hb.baseUrl + path + "?" + gparams.Encode()
	asset := strings.ToLower(target.Dist(tp))
	postData := make(map[string]string)
	postData["symbol"] = target.Format("", false)
	postData["currency"] = asset
	postData["amount"] = amount.String()
	respData, err := ghttputils.HttpPostForm4(hb.httpClient, urlStr, postData, nil)
	if err != nil {
		return err
	}

	var respmap map[string]interface{}
	json.Unmarshal(respData, &respmap)
	if nil == respmap["data"] {
		return errors.New(respmap["err-msg"].(string))
	}
	transferId := int(respmap["data"].(float64))
	if transferId <= 0 {
		return errors.New(respmap["err-msg"].(string))
	}

	return nil
}

// limit-buy, limit-sell, market-buy, market-sell
// when market-buy/market-sell, price will be ignored
// amount: always unit amount, not quote amount, whether trade type is buy or sell.
func (hb *Client) Trade(market fintypes.Market, target fintypes.Pair, side TradeTypeSide, amount, price gdecimal.Decimal) (*fintypes.OrderId, error) {
	if market != fintypes.MarketSpot && market != MarketMargin {
		return nil, errors.New("Access denied")
	}
	strId, err := hb.placeOrder(market, target, amount.String(), price.String(), side.CustomFormat(hb.config))
	if err != nil {
		return nil, err
	}
	res := fintypes.NewOrderId(market, target, strId)
	return &res, nil
}

// get all my history orders' info
func (hb *Client) GetAllOrders(market fintypes.Market, target fintypes.Pair) ([]fintypes.Order, error) {
	var orders []fintypes.Order
	var err error
	if fintypes.MarketSpot == market || MarketMargin == market {
		orders, err = hb.getProOrders(market, QueryOrdersParams{
			fintypes.Pair: target,
			States:        "partial-canceled,filled",
			Size:          100,
			Direct:        "next",
		})
	} else if fintypes.MarketFuture == market || fintypes.MarketPerp == market {
		orders, err = hb.getContractOrders(market, QueryOrdersParams{
			Types:  "all",
			Symbol: strings.ToUpper(string(target)),
		})
	} else {
	}

	return orders, err
}

// get all my unfinished orders' info
func (hb *Client) GetOpenOrders(market fintypes.Market, target fintypes.Pair) ([]fintypes.Order, error) {
	var orders []fintypes.Order
	var err error
	if fintypes.MarketSpot == market || MarketMargin == market {
		orders, err = hb.getProOrders(market, QueryOrdersParams{
			fintypes.Pair: target,
			Size:          100,
			States:        "pre-submitted,submitted,partial-filled",
		})
	} else if fintypes.MarketFuture == market || fintypes.MarketPerp == market {
		orders, err = hb.getContractOrders(market, QueryOrdersParams{
			Types:  "unfinished",
			Symbol: strings.ToUpper(string(target)),
		})
	} else {
	}

	return orders, err
}

// get order info by id
func (hb *Client) GetOrder(market fintypes.Market, pair fintypes.Pair, orderId fintypes.OrderId) (*fintypes.Order, error) {
	var order *fintypes.Order
	var err error
	hb.setBaseUrl(market)
	if fintypes.MarketSpot == market || MarketMargin == market {
		path := "/v1/order/orders/" + string(orderId)
		params := url.Values{}
		hb.buildPostForm("GET", path, &params)
		respmap, err := ghttputils.HttpGet(hb.httpClient, hb.baseUrl+path+"?"+params.Encode())
		if err != nil {
			return nil, err
		}

		if respmap["status"].(string) != "ok" {
			return nil, errors.New(respmap["err-code"].(string))
		}

		datamap := respmap["data"].(map[string]interface{})
		order = hb.parseOrder(market, datamap)
	} else if fintypes.MarketFuture == market || fintypes.MarketPerp == market {
		var path string
		params := &url.Values{}
		postData := make(map[string]string)
		if fintypes.MarketFuture == market {
			path = ApiPathMap[fintypes.MarketFuture][ApiUrlOrder]
			postData["order_id"] = string(orderId)
			postData["symbol"] = string(pair)
		} else if fintypes.MarketPerp == market {
			path = ApiPathMap[fintypes.MarketPerp][ApiUrlOrder]
			postData["order_id"] = string(orderId)
			postData["contract_code"] = string(pair)
			postData["order_type"] = "1"
		}
		hb.buildPostForm("POST", path, params)
		urlStr := hb.baseUrl + path + "?" + params.Encode()
		respData, err := ghttputils.HttpPostForm4(hb.httpClient, urlStr, postData, nil)
		if err != nil {
			return nil, err
		}

		var respmap map[string]interface{}
		json.Unmarshal([]byte(respData), &respmap)
		if respmap["status"].(string) != "ok" {
			return nil, errors.New(respmap["err-code"].(string))
		}
		datamap := respmap["data"].(map[string]interface{})
		order = hb.parseContractOneOrder(market, datamap)
	} else {
	}

	return order, err
}

// cancel unfinished order by id
func (hb *Client) CancelOrder(market fintypes.Market, pair fintypes.Pair, orderId fintypes.OrderId) error {
	hb.setBaseUrl(market)
	if fintypes.MarketSpot == market || MarketMargin == market {
		path := fmt.Sprintf("/v1/order/orders/%s/submitcancel", string(orderId))
		params := url.Values{}
		hb.buildPostForm("POST", path, &params)
		resp, err := ghttputils.HttpPostForm3(hb.httpClient, hb.baseUrl+path+"?"+params.Encode(), toJson(params),
			map[string]string{"Content-Type": "application/json", "Accept-Language": "zh-cn"})
		if err != nil {
			return err
		}

		var respmap map[string]interface{}
		err = json.Unmarshal(resp, &respmap)
		if err != nil {
			return err
		}

		if respmap["status"].(string) != "ok" {
			return errors.New(string(resp))
		}
	} else if fintypes.MarketFuture == market || fintypes.MarketPerp == market {
		var path string
		params := &url.Values{}
		postData := make(map[string]string)
		if fintypes.MarketFuture == market {
			path = ApiPathMap[fintypes.MarketFuture][ApiUrlCancelOrder]
			postData["order_id"] = string(orderId)
			postData["symbol"] = string(pair)
		} else if fintypes.MarketPerp == market {
			path = ApiPathMap[fintypes.MarketPerp][ApiUrlCancelOrder]
			postData["order_id"] = string(orderId)
			postData["contract_code"] = string(pair)
		} else {
		}
		hb.buildPostForm("POST", path, params)
		urlStr := hb.baseUrl + path + "?" + params.Encode()
		respData, err := ghttputils.HttpPostForm4(hb.httpClient, urlStr, postData, nil)
		if err != nil {
			return err
		}

		var respmap map[string]interface{}
		json.Unmarshal([]byte(respData), &respmap)
		if respmap["status"].(string) != "ok" {
			return errors.New(respmap["err-code"].(string))
		}
		if nil == respmap["data"] {
			return errors.New(respmap["err-msg"].(string))
		}
	} else {
	}

	return nil
}

func (hb *Client) parseDepthData(tick map[string]interface{}, size int) *fintypes.Depth {
	bids, _ := tick["bids"].([]interface{})
	asks, _ := tick["asks"].([]interface{})

	depth := new(fintypes.Depth)
	n := 0
	for _, r := range asks {
		var dr fintypes.OrderBook
		rr := r.([]interface{})
		dr.Price = gdecimal.NewFromFloat64(rr[0].(float64))
		dr.Amount = gdecimal.NewFromFloat64(rr[1].(float64))
		depth.Sells = append(depth.Sells, dr)
		n++
		if n == size {
			break
		}
	}

	n = 0
	for _, r := range bids {
		var dr fintypes.OrderBook
		rr := r.([]interface{})
		dr.Price = gdecimal.NewFromFloat64(rr[0].(float64))
		dr.Amount = gdecimal.NewFromFloat64(rr[1].(float64))
		depth.Buys = append(depth.Buys, dr)
		n++
		if n == size {
			break
		}
	}

	return depth
}

func (hb *Client) signForm(postForm *url.Values, reqMethod, baseUrl, path string) error {
	postForm.Set("AccessKeyId", hb.accessKey)
	postForm.Set("SignatureMethod", "HmacSHA256")
	postForm.Set("SignatureVersion", "2")
	postForm.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05"))
	domain := strings.Replace(baseUrl, "https://", "", len(baseUrl))
	payload := fmt.Sprintf("%s\n%s\n%s\n%s", reqMethod, domain, path, postForm.Encode())
	sign, _ := gparam.GetParamHmacSHA256Base64Sign(hb.secretKey, payload)
	postForm.Set("Signature", sign)
	return nil
}

func (hb *Client) placeOrder(market fintypes.Market, pair fintypes.Pair, amount, price, orderType string) (string, error) {
	path := ApiPathMap[market][ApiUrlTrade]
	gparams := &url.Values{}
	hb.setBaseUrl(market)
	err0 := hb.setNewHuobiId(market)
	if nil != err0 {
		return "", err0
	}
	gparams.Set("accountId-id", hb.accountId)
	hb.buildPostForm("POST", path, gparams)
	urlStr := hb.baseUrl + path + "?" + gparams.Encode()
	postData := make(map[string]string)
	postData["account-id"] = hb.accountId
	postData["amount"] = amount
	postData["symbol"] = pair.Format("", false)
	postData["type"] = orderType

	if fintypes.MarketSpot == market {
		postData["source"] = "spot-api"
	} else if MarketMargin == market {
		postData["source"] = "margin-api"
	} else {
	}

	switch orderType {
	case string(TradeTypeSideLimitBuy), string(TradeTypeSideLimitSell):
		postData["price"] = price
	}

	resp, err := ghttputils.HttpPostForm4(hb.httpClient, urlStr, postData, nil)
	if err != nil {
		return "", err
	}

	respmap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		return "", err
	}

	if respmap["status"].(string) != "ok" {
		return "", errors.New(respmap["err-code"].(string))
	}

	return respmap["data"].(string), nil
}

func (hb *Client) getProOrders(market fintypes.Market, queryparams QueryOrdersParams) ([]fintypes.Order, error) {
	params := url.Values{}
	path := ApiPathMap[market][ApiUrlOrders]
	var orders []fintypes.Order
	hb.setBaseUrl(market)
	params.Set("symbol", queryparams.Pair.Format("", false))
	params.Set("states", queryparams.States)

	if queryparams.Direct != "" {
		params.Set("direct", queryparams.Direct)
	}

	if queryparams.Size > 0 {
		params.Set("size", fmt.Sprint(queryparams.Size))
	}

	hb.buildPostForm("GET", path, &params)
	respmap, err := ghttputils.HttpGet(hb.httpClient, fmt.Sprintf("%s%s?%s", hb.baseUrl, path, params.Encode()))
	if err != nil {
		return nil, err
	}

	if respmap["status"].(string) != "ok" {
		return nil, errors.New(respmap["err-code"].(string))
	}

	datamap := respmap["data"].([]interface{})
	for _, v := range datamap {
		ordmap := v.(map[string]interface{})
		ord := hb.parseOrder(market, ordmap)
		orders = append(orders, *ord)
	}

	return orders, nil
}

func (hb *Client) getContractOrders(market fintypes.Market, queryparams QueryOrdersParams) ([]fintypes.Order, error) {
	var orders []fintypes.Order
	var path string
	params := &url.Values{}
	hb.setBaseUrl(market)
	if "all" == queryparams.Types {
		path = ApiPathMap[market][ApiUrlOrders]
	} else if "unfinished" == queryparams.Types {
		path = ApiPathMap[market][ApiUrlOpenOrders]
	} else {
	}
	hb.buildPostForm("POST", path, params)
	urlStr := hb.baseUrl + path + "?" + params.Encode()
	postData := make(map[string]string)
	if fintypes.MarketFuture == market {
		postData["symbol"] = queryparams.Symbol // 例如：BTC
	} else if fintypes.MarketPerp == market {
		postData["contract_code"] = queryparams.Symbol // 例如：BTC-USD
	} else {
	}
	respData, err := ghttputils.HttpPostForm4(hb.httpClient, urlStr, postData, nil)
	if err != nil {
		return nil, err
	}
	respmap := make(map[string]interface{})
	json.Unmarshal(respData, &respmap)
	status := respmap["status"].(string)
	if status == "error" {
		return nil, errors.New(respmap["err-msg"].(string))
	}

	var data []interface{}
	if "all" == queryparams.Types {
		data = respmap["data"].([]interface{})
	} else if "unfinished" == queryparams.Types {
		response := respmap["data"].(map[string]interface{})
		data = response["orders"].([]interface{})
	} else {
	}

	for _, v := range data {
		ordmap := v.(map[string]interface{})
		ord := hb.parseContractOrder(market, ordmap)
		orders = append(orders, ord)
	}

	return orders, nil
}

func (hb *Client) doRequest(path string, params *url.Values, data interface{}) error {
	type BaseResponse struct {
		Status  string          `json:"status"`
		Ch      string          `json:"ch"`
		Ts      int64           `json:"ts"`
		ErrCode int             `json:"err_code"`
		ErrMsg  string          `json:"err_msg"`
		Data    json.RawMessage `json:"data"`
	}

	hb.buildPostForm("POST", path, params)
	jsonD, _ := utils.ValuesToJson(*params)
	log.Println(string(jsonD))
	var ret BaseResponse

	resp, err := ghttputils.HttpPostForm3(hb.httpClient,
		hb.baseUrl+path+"?"+params.Encode(), string(jsonD),
		map[string]string{"Content-Type": "application/json", "Accept-Language": "zh-cn"})

	if err != nil {
		return err
	}

	log.Println(string(resp))
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		return err
	}

	if ret.Status != "ok" {
		return errors.New(fmt.Sprintf("%d:[%s]", ret.ErrCode, ret.ErrMsg))
	}

	return json.Unmarshal(ret.Data, data)
}

func (hb *Client) parseOrder(market fintypes.Market, ordmap map[string]interface{}) *fintypes.Order {
	pair, _ := fintypes.ParsePairCustom(ordmap["symbol"].(string), hb.config)
	ord := fintypes.Order{
		Id:         fintypes.NewOrderId(market, pair, ordmap["id"].(string)),
		Time:       time.Unix(ToInt64(ordmap["created-at"].(float64)/1000), 0),
		Pair:       pair,
		TypeSide:   CustomFormatTradeTypeSide(ordmap["type"].(string), hb.config),
		Price:      gdecimal.NewFromFloat64(ordmap["price"].(float64)),
		Amount:     gdecimal.NewFromFloat64(ordmap["amount"].(float64)),
		Status:     CustomFormatTradeStatus(ordmap["state"].(string), hb.config),
		AvgPrice:   gdecimal.Zero,
		DealAmount: gdecimal.NewFromFloat64(ordmap["field-amount"].(float64)),
		Fee:        gdecimal.NewFromFloat64(ordmap["field-fees"].(float64)),
	}

	dealAmount := ord.DealAmount.Float64()
	if dealAmount > 0.0 {
		ord.AvgPrice = gdecimal.NewFromFloat64(ToFloat64(ordmap["field-cash-amount"]) / dealAmount)
	}

	return &ord
}

func (hb *Client) parseContractOrder(market fintypes.Market, ordmap map[string]interface{}) fintypes.Order {
	pair := fintypes.NewPair(ordmap["contract_code"].(string), "USD")
	ord := fintypes.Order{
		Id:         fintypes.NewOrderId(market, pair, ordmap["order_id_str"].(string)),
		Time:       time.Unix(ToInt64(ordmap["created_at"].(float64)/1000), 0),
		Pair:       pair,
		TypeSide:   CustomFormatTradeTypeSide(ordmap["direction"].(string)+"-"+ordmap["order_price_type"].(string), hb.config),
		Price:      gdecimal.NewFromFloat64(ordmap["price"].(float64)),
		Amount:     gdecimal.NewFromFloat64(ordmap["volume"].(float64)),
		Status:     TradeStatusMap[int(ordmap["status"].(float64))],
		AvgPrice:   gdecimal.NewFromFloat64(ordmap["trade_avg_price"].(float64)),
		DealAmount: gdecimal.Zero,
		Fee:        gdecimal.NewFromFloat64(ordmap["fee"].(float64)),
	}

	return ord
}

func (hb *Client) parseContractOneOrder(market fintypes.Market, ordmap map[string]interface{}) *fintypes.Order {
	pair := fintypes.NewPair(ordmap["contract_code"].(string), "USD")
	item := ordmap["trades"].([]interface{})
	trade := item[0].(map[string]interface{})
	ord := fintypes.Order{
		Id:         fintypes.NewOrderId(market, pair, trade["id"].(string)),
		Time:       time.Unix(ToInt64(trade["created_at"].(float64)/1000), 0),
		Pair:       pair,
		TypeSide:   CustomFormatTradeTypeSide(ordmap["direction"].(string)+"-"+ordmap["order_price_type"].(string), hb.config),
		Price:      gdecimal.NewFromFloat64(ordmap["price"].(float64)),
		Amount:     gdecimal.NewFromFloat64(ordmap["volume"].(float64)),
		Status:     TradeStatusMap[6],
		AvgPrice:   gdecimal.Zero,
		DealAmount: gdecimal.Zero,
		Fee:        gdecimal.Zero,
	}

	return &ord
}

func adaptPerpSymbol(pair fintypes.Pair) string {
	pairStr := pair.Unit() + "-" + pair.Quote()

	return pairStr
}

func toJson(params url.Values) string {
	parammap := make(map[string]string)
	for k, v := range params {
		parammap[k] = v[0]
	}
	jsonData, _ := json.Marshal(parammap)
	return string(jsonData)
}
