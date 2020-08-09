package binance

/**

币安的现货杠杠只有全仓，没有逐仓

注意，Margin和Spot是共享的盘口和K线，所以二者是一回事

time in force : 订单的有效时间.
它的选项有:
Day (9:30am-4pm)
OPG (开市交叉时段)
IOC(立刻执行或者立刻取消)
GTC(取消前都有效)
GTX (未在指定证券市场成交前都有效)
EXT (在交易时间外都有效)
NOW(即时)
*/

import (
	"context"
	"github.com/adshao/go-binance"
	"github.com/adshao/go-binance/futures"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gnum"
	"github.com/shawnwyckoff/gopkg/net/ghttp"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	in               *binance.Client
	inPerp           *futures.Client
	name             fintypes.Platform
	supportedPeriods []fintypes.Period
	property         fintypes.ExProperty
	marketInfoCache  fintypes.MarketInfo
	marketInfoUpdate time.Time
}

// email is required in living trading, but not required in kline spider
func New(accessKey, secretKey, proxy string, c gtime.Clock, email string) (*Client, error) {
	cc := fintypes.ExProperty{
		Name:                  fintypes.Binance,
		MaxDepth:              100,
		MaxFills:              1000,
		PairDelimiter:         "",
		PairDelimiterLeftTail: []string{"BULL", "BEAR"},
		PairNormalOrder:       true,
		PairUpperCase:         true,
	}
	cc.RateLimits = map[fintypes.ExApi]time.Duration{}
	cc.RateLimits[fintypes.ExApiGetKline] = time.Second / 100 * 127
	cc.RateLimits[fintypes.ExApiGetFill] = time.Second // 真实的数值应该是 time.Second / 4 但那样会出现大量奇怪的超时问题，所以临时改成1s
	cc.Periods = make(map[fintypes.Period]string)
	cc.Periods[fintypes.Period1Min] = "1m"
	cc.Periods[fintypes.Period3Min] = "3m"
	cc.Periods[fintypes.Period5Min] = "5m"
	cc.Periods[fintypes.Period15Min] = "15m"
	cc.Periods[fintypes.Period30Min] = "30m"
	cc.Periods[fintypes.Period1Hour] = "1h"
	cc.Periods[fintypes.Period2Hour] = "2h"
	cc.Periods[fintypes.Period4Hour] = "4h"
	cc.Periods[fintypes.Period6Hour] = "6h"
	cc.Periods[fintypes.Period8Hour] = "8h"
	cc.Periods[fintypes.Period12Hour] = "12h"
	cc.Periods[fintypes.Period1Day] = "1d"
	cc.Periods[fintypes.Period1Week] = "1w"
	cc.Periods[fintypes.Period1MonthFUZZY] = "1M"
	cc.MarketEnabled = map[fintypes.Market]bool{}
	cc.MarketEnabled[fintypes.MarketSpot] = true
	cc.MarketEnabled[fintypes.MarketPerp] = true
	cc.TradeBeginTime = time.Date(2017, 7, 14, 00, 00, 00, 0, time.UTC) // this time is approximation, more exact time seem like 2017-07-14 04:00:00 +0000 UTC

	ex := Client{}
	ex.in = binance.NewClient(accessKey, secretKey)
	ex.inPerp = futures.NewClient(accessKey, secretKey)
	if proxy != "" {
		if err := ghttp.SetProxy(ex.in.HTTPClient, proxy); err != nil {
			return nil, err
		}
		if err := ghttp.SetProxy(ex.inPerp.HTTPClient, proxy); err != nil {
			return nil, err
		}
	}
	ex.name = fintypes.Binance
	ex.property = cc
	ex.marketInfoUpdate = gtime.ZeroTime
	ex.property.Email = email
	if c == nil {
		ex.property.Clock = gtime.GetSysClock()
	} else {
		ex.property.Clock = c
	}

	return &ex, nil
}

func (ex *Client) binanceOrderToApiOrder(accType fintypes.Market, margin fintypes.Margin, src *binance.Order) (*fintypes.Order, error) {
	if src == nil {
		return nil, errors.Errorf("nil input binance.Order")
	}
	p, err := fintypes.ParsePairCustom(src.Symbol, ex.Property())
	if err != nil {
		return nil, err
	}

	res := fintypes.Order{}
	res.Pair = p
	res.Market = accType
	res.Margin = margin
	res.Time = gtime.EpochMillisToTime(src.Time)
	res.Id = fintypes.NewOrderId(accType, margin, p, gnum.ToString(src.OrderID))
	if src.StopPrice != "" {
		res.StopPrice, err = gdecimal.NewFromString(src.StopPrice)
		if err != nil {
			return nil, err
		}
	}
	res.Price, err = gdecimal.NewFromString(src.Price)
	if err != nil {
		return nil, err
	}
	res.Amount, err = gdecimal.NewFromString(src.OrigQuantity)
	if err != nil {
		return nil, err
	}
	res.AvgPrice = gdecimal.NewFromInt(-1)
	if src.Side == binance.SideTypeBuy {
		res.Side = fintypes.OrderSideBuyLong
	} else if src.Side == binance.SideTypeSell {
		res.Side = fintypes.OrderSideSellShort
	} else {
		return nil, errors.Errorf("unsupported OrderSide(%s)", src.Side)
	}
	if src.Type == binance.OrderTypeLimit {
		res.Type = fintypes.OrderTypeLimit
	} else if src.Type == binance.OrderTypeStopLossLimit {
		res.Type = fintypes.OrderTypeStopLimit
	} else if src.Type == binance.OrderTypeMarket {
		res.Type = fintypes.OrderTypeMarket
	} else {
		return nil, errors.Errorf("unsupported OrderType(%s)", src.Type)
	}

	switch src.Status {
	case binance.OrderStatusTypeNew:
		res.Status = fintypes.OrderStatusNew
	case binance.OrderStatusTypePartiallyFilled:
		res.Status = fintypes.OrderStatusPartiallyFilled
	case binance.OrderStatusTypeFilled:
		res.Status = fintypes.OrderStatusFilled
	case binance.OrderStatusTypeCanceled:
		res.Status = fintypes.OrderStatusCanceled
	case binance.OrderStatusTypePendingCancel:
		res.Status = fintypes.OrderStatusCanceling
	case binance.OrderStatusTypeRejected:
		res.Status = fintypes.OrderStatusRejected
	case binance.OrderStatusTypeExpired:
		res.Status = fintypes.OrderStatusExpired
	default:
		return nil, errors.Errorf("unsupported order status(%s)", src.Status)
	}
	res.Fee = gdecimal.NewFromInt(-1)
	res.DealAmount, err = gdecimal.NewFromString(src.ExecutedQuantity)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (ex *Client) binanceMarginAllOrderToOrder(src *binance.MarginAllOrder) *binance.Order {
	if src == nil {
		return nil
	}
	res := &binance.Order{}
	res.OrderID = src.ID
	res.Price = src.Price
	res.OrigQuantity = src.Quantity
	res.CummulativeQuoteQuantity = src.QuoteQuantity
	res.Symbol = src.Symbol
	res.Time = src.Time
	return res
}

/**
// Future order only
type Order struct {
	ReduceOnly       bool            `json:"reduceOnly"`
	CumQuantity      string          `json:"cumQty"`
	CumQuote         string          `json:"cumQuote"`
	WorkingType      WorkingType     `json:"workingType"`
}
// Spot/Margin order only
type Order struct {
	CummulativeQuoteQuantity string          `json:"cummulativeQuoteQty"`
	IcebergQuantity          string          `json:"icebergQty"`
	IsWorking                bool            `json:"isWorking"`
}
*/
func (ex *Client) binancePerpOrderToApiOrder(market fintypes.Market, margin fintypes.Margin, src *futures.Order) (*fintypes.Order, error) {
	res := &fintypes.Order{}

	if src == nil {
		return nil, errors.Errorf("nil input binance.Order")
	}
	pair, err := fintypes.ParsePairCustom(src.Symbol, ex.Property())
	if err != nil {
		return nil, err
	}

	res.Pair = pair
	res.Market = market
	res.Margin = margin
	res.Time = gtime.EpochMillisToTime(src.Time)
	res.Id = fintypes.NewOrderId(market, margin, pair, gnum.ToString(src.OrderID))
	if src.StopPrice != "" {
		res.StopPrice, err = gdecimal.NewFromString(src.StopPrice)
		if err != nil {
			return nil, err
		}
	}
	res.Price, err = gdecimal.NewFromString(src.Price)
	if err != nil {
		return nil, err
	}
	res.Amount, err = gdecimal.NewFromString(src.OrigQuantity)
	if err != nil {
		return nil, err
	}
	res.AvgPrice = gdecimal.NewFromInt(-1)
	if src.Side == futures.SideTypeBuy {
		res.Side = fintypes.OrderSideBuyLong
	} else if src.Side == futures.SideTypeSell {
		res.Side = fintypes.OrderSideSellShort
	} else {
		return nil, errors.Errorf("unsupported OrderSide(%s)", src.Side)
	}
	if src.Type == futures.OrderTypeLimit {
		res.Type = fintypes.OrderTypeLimit
	} else if src.Type == futures.OrderTypeStop {
		res.Type = fintypes.OrderTypeStopLimit
	} else if src.Type == futures.OrderTypeMarket {
		res.Type = fintypes.OrderTypeMarket
	} else {
		return nil, errors.Errorf("unsupported OrderType(%s)", src.Type)
	}

	switch src.Status {
	case futures.OrderStatusTypeNew:
		res.Status = fintypes.OrderStatusNew
	case futures.OrderStatusTypePartiallyFilled:
		res.Status = fintypes.OrderStatusPartiallyFilled
	case futures.OrderStatusTypeFilled:
		res.Status = fintypes.OrderStatusFilled
	case futures.OrderStatusTypeCanceled:
		res.Status = fintypes.OrderStatusCanceled
	case futures.OrderStatusTypeRejected:
		res.Status = fintypes.OrderStatusRejected
	case futures.OrderStatusTypeExpired:
		res.Status = fintypes.OrderStatusExpired
	default:
		return nil, errors.Errorf("unsupported order status(%s)", src.Status)
	}
	res.Fee = gdecimal.NewFromInt(-1)
	res.DealAmount, err = gdecimal.NewFromString(src.ExecutedQuantity)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ex *Client) Property() *fintypes.ExProperty {
	return &ex.property
}

func (ex *Client) GetMarketInfo() (*fintypes.MarketInfo, error) {
	mi := fintypes.MarketInfo{Infos: map[fintypes.PairM]fintypes.PairInfo{}}

	exInfo, err := ex.in.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	exInfoPerp, err := ex.inPerp.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	// process spot market info
	for _, symbol := range exInfo.Symbols {

		// "123456" is a binance testing symbol that can be enabled/disabled as needed
		if symbol.Symbol == "123456" {
			continue
		}

		p, err := fintypes.ParsePairCustom(symbol.Symbol, &ex.property)
		if err != nil {
			return nil, err
		}
		spotInfo := fintypes.PairInfo{}
		spotInfo.MakerFee = gdecimal.NewFromFloat64(0.001) // FIXME 目前暂时统一填写0.001，以后可能更改
		spotInfo.TakerFee = gdecimal.NewFromFloat64(0.001) // FIXME 目前暂时统一填写0.001，以后可能更改
		spotInfo.Enabled = symbol.IsSpotTradingAllowed
		spotInfo.MarginCrossEnabled = symbol.IsMarginTradingAllowed // margin shares same PairInfo with spot
		if spotInfo.MarginCrossEnabled {
			spotInfo.MinLeverage = 3
			spotInfo.MaxLeverage = 3
		}
		spotInfo.UnitPrecision = symbol.BaseAssetPrecision
		spotInfo.QuotePrecision = symbol.QuotePrecision
		for _, filterMap := range symbol.Filters {
			if ft, ok := filterMap["filterType"]; ok && ft == "LOT_SIZE" {
				spotInfo.UnitMin, err = gdecimal.NewFromString(filterMap["minQty"].(string))
				if err != nil {
					return nil, err
				}
				spotInfo.UnitStep, err = gdecimal.NewFromString(filterMap["stepSize"].(string))
				if err != nil {
					return nil, err
				}
			}
			if ft, ok := filterMap["filterType"]; ok && ft == "PRICE_FILTER" {
				spotInfo.QuoteStep, err = gdecimal.NewFromString(filterMap["minPrice"].(string))
				if err != nil {
					return nil, err
				}
			}
		}
		mi.Infos[p.SetM(fintypes.MarketSpot)] = spotInfo
	}

	// process perp market info
	for _, symbol := range exInfoPerp.Symbols {

		// "123456" is a binance testing symbol that can be enabled/disabled as needed
		if symbol.Symbol == "123456" {
			continue
		}

		p, err := fintypes.ParsePairCustom(symbol.Symbol, &ex.property)
		if err != nil {
			return nil, err
		}
		perpInfo := fintypes.PairInfo{}
		perpInfo.MakerFee = gdecimal.NewFromFloat64(0.001) // FIXME 目前暂时统一填写0.001，以后可能更改
		perpInfo.TakerFee = gdecimal.NewFromFloat64(0.001) // FIXME 目前暂时统一填写0.001，以后可能更改
		perpInfo.UnitPrecision = symbol.QuantityPrecision
		perpInfo.QuotePrecision = symbol.PricePrecision
		perpInfo.MaintMarginPercent, err = gdecimal.NewFromString(symbol.MaintMarginPercent)
		if err != nil {
			return nil, err
		}
		perpInfo.RequiredMarginPercent, err = gdecimal.NewFromString(symbol.RequiredMarginPercent)
		if err != nil {
			return nil, err
		}
		perpInfo.Enabled = true
		for _, filterMap := range symbol.Filters {
			if ft, ok := filterMap["filterType"]; ok && ft == "LOT_SIZE" {
				perpInfo.UnitMin, err = gdecimal.NewFromString(filterMap["minQty"].(string))
				if err != nil {
					return nil, err
				}
				perpInfo.UnitStep, err = gdecimal.NewFromString(filterMap["stepSize"].(string))
				if err != nil {
					return nil, err
				}
			}
			if ft, ok := filterMap["filterType"]; ok && ft == "PRICE_FILTER" {
				perpInfo.QuoteStep, err = gdecimal.NewFromString(filterMap["minPrice"].(string))
				if err != nil {
					return nil, err
				}
			}
		}

		// insert perp market info into result
		mi.Infos[p.SetM(fintypes.MarketPerp)] = perpInfo
	}

	// cache it
	ex.marketInfoCache = mi
	ex.marketInfoUpdate = ex.property.Clock.Now()

	return &mi, nil
}

// FIXME 永续合约的暂时没有获取，因为Account还没有稳定
// 注意，币安只有全仓现货杠杆，没有逐仓现货杠杆
// binance account API support total balance in BTC, but doesn't return total balance in fiat, you need to calculate it by yourself
func (ex *Client) GetAccount() (*fintypes.Account, error) {

	spotAcc, err := ex.in.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	marginAcc, err := ex.in.NewGetMarginAccountService().Do(context.Background())
	if err != nil && !strings.Contains(err.Error(), "code=-3003,") { // margin account doesn't enabled error message: <APIError> code=-3003, msg=Margin account does not exist.
		return nil, err
	}

	/*perpAcc, err := ex.inPerp.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}*/

	r := fintypes.NewEmptyAccount()
	for _, v := range spotAcc.Balances {
		sa := fintypes.Balance{}
		sa.Market = fintypes.MarketSpot
		sa.Margin = fintypes.MarginNo
		sa.CustomSubAccName = ""
		sa.Asset = strings.ToUpper(v.Asset)
		free, err := gdecimal.NewFromString(v.Free)
		if err != nil {
			return nil, err
		}
		locked, err := gdecimal.NewFromString(v.Locked)
		if err != nil {
			return nil, err
		}
		sa.Free = free
		sa.Locked = locked
		if sa.IsZero() {
			continue
		}
		r.Balances = append(r.Balances, sa)
	}

	if marginAcc != nil {
		for _, v := range marginAcc.UserAssets {
			sa := fintypes.Balance{}
			sa.Market = fintypes.MarketSpot
			sa.Margin = fintypes.MarginNo
			sa.CustomSubAccName = ""
			sa.Asset = strings.ToUpper(v.Asset)
			free, err := gdecimal.NewFromString(v.Free)
			if err != nil {
				return nil, err
			}
			locked, err := gdecimal.NewFromString(v.Locked)
			if err != nil {
				return nil, err
			}
			borrowed, err := gdecimal.NewFromString(v.Borrowed)
			if err != nil {
				return nil, err
			}
			interest, err := gdecimal.NewFromString(v.Interest)
			if err != nil {
				return nil, err
			}
			sa.Free = free
			sa.Locked = locked
			sa.Borrowed = borrowed
			sa.Interest = interest
			if sa.IsZero() {
				continue
			}

			r.Balances = append(r.Balances, sa)
		}
	}

	/**
	description:
	Wallet SpotBalance = Total Net Transfer + Total Realized Profit + Total Net Funding Fee - Total Commission.
	钱包余额 ＝ 总共净划入 + 总共已实现盈亏 + 总共净资金费用 - 总共手续费

	NOTE: future account返回了当前账户在市场上所有交易对的position，暂时忽略了它

	注意：
	InitialMargin/MaintMargin/PositionInitialMargin/UnrealizedProfit 这几个成员在你下好单之后哪怕不操作这个值也会动态变化
	position中的"leverage": "20", // 当前设置的杠杆倍率，可以在界面上调整
	*/

	/*
		if perpAcc != nil {
			for _, v := range perpAcc.Assets {
				cb := Balance{}
				cb.Market =           MarketSpot
				cb.Margin =           MarginNo
				cb.CustomSubAccName = ""
				cb.Asset =           strings.ToUpper(v.Asset)
				cb.InitialMargin, err = gdecimal.NewFromString(v.InitialMargin)
				if err != nil {
					return nil, err
				}
				cb.MaintMargin, err = gdecimal.NewFromString(v.MaintMargin)
				if err != nil {
					return nil, err
				}
				cb.MarginBalance, err = gdecimal.NewFromString(v.MarginBalance)
				if err != nil {
					return nil, err
				}
				cb.MaxWithdrawAmount, err = gdecimal.NewFromString(v.MaxWithdrawAmount)
				if err != nil {
					return nil, err
				}
				//cb.PositionInitialMargin, err = gdecimal.NewFromString(v.PositionInitialMargin)
				//if err != nil {
				//	return nil, err
				//}
				cb.UnrealizedProfit, err = gdecimal.NewFromString(v.UnrealizedProfit)
				if err != nil {
					return nil, err
				}
				//cb.WalletBalance, err = gdecimal.NewFromString(v.WalletBalance)
				//if err != nil {
				//	return nil, err
				//}
				if cb.IsZero() {
					continue
				}

				r.Balances = append(r.Balances, cb)
			}
		}*/

	return r, nil
}

func (ex *Client) GetDepth(market fintypes.Market, target fintypes.Pair) (*fintypes.Depth, error) {
	if err := target.Verify(); err != nil {
		return nil, err
	}

	depth := &binance.DepthResponse{}
	err := error(nil)
	if market == fintypes.MarketPerp {
		depthPerp, err := ex.inPerp.NewDepthService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		if err != nil {
			return nil, err
		}

		// convert futures.DepthResponse to binance.DepthResponse
		depth.LastUpdateID = depthPerp.LastUpdateID
		for _, v := range depthPerp.Asks {
			depth.Asks = append(depth.Asks, binance.Ask{Price: v.Price, Quantity: v.Quantity})
		}
		for _, v := range depthPerp.Bids {
			depth.Bids = append(depth.Bids, binance.Bid{Price: v.Price, Quantity: v.Quantity})
		}
	} else {
		depth, err = ex.in.NewDepthService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		if err != nil {
			return nil, err
		}
	}

	res := fintypes.Depth{}
	res.Time = ex.property.Clock.Now()
	for _, v := range depth.Bids {
		item := fintypes.OrderBook{}
		item.Price, err = gdecimal.NewFromString(v.Price)
		if err != nil {
			return nil, err
		}
		item.Amount, err = gdecimal.NewFromString(v.Quantity)
		if err != nil {
			return nil, err
		}
		res.Buys = append(res.Buys, item)
	}
	for _, v := range depth.Asks {
		item := fintypes.OrderBook{}
		item.Price, err = gdecimal.NewFromString(v.Price)
		if err != nil {
			return nil, err
		}
		item.Amount, err = gdecimal.NewFromString(v.Quantity)
		if err != nil {
			return nil, err
		}
		res.Sells = append(res.Sells, item)
	}

	return &res, nil
}

func (ex *Client) GetTicks() (map[fintypes.PairM]fintypes.Tick, error) {
	ticks, err := ex.in.NewListPricesService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	ticksPerp, err := ex.inPerp.NewListPricesService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	res := make(map[fintypes.PairM]fintypes.Tick)
	now := ex.property.Clock.Now()

	for _, symbolPrice := range ticks {
		pair, err := fintypes.ParsePairCustom(symbolPrice.Symbol, ex.Property())
		if err != nil {
			return nil, err
		}
		item := fintypes.Tick{}
		item.Time = now
		item.Last, err = gdecimal.NewFromString(symbolPrice.Price)
		if err != nil {
			return nil, err
		}
		res[pair.SetM(fintypes.MarketSpot)] = item
	}

	for _, symbolPrice := range ticksPerp {
		pair, err := fintypes.ParsePairCustom(symbolPrice.Symbol, ex.Property())
		if err != nil {
			return nil, err
		}
		item := fintypes.Tick{}
		item.Time = now
		item.Last, err = gdecimal.NewFromString(symbolPrice.Price)
		if err != nil {
			return nil, err
		}
		res[pair.SetM(fintypes.MarketPerp)] = item
	}

	return res, nil
}

func (ex *Client) GetKline(market fintypes.Market, target fintypes.Pair, period fintypes.Period, since *time.Time) (*fintypes.Kline, error) {
	if err := target.Verify(); err != nil {
		return nil, err
	}
	if market == fintypes.MarketFuture {
		return nil, gerror.Errorf("binance doesn't support future market")
	}

	binancePeriod, err := period.CustomFormat(ex.Property())
	if err != nil {
		return nil, err
	}

	if since == nil {
		now := ex.property.Clock.Now()
		since = &now
	}

	var ks []*binance.Kline
	r := new(fintypes.Kline)
	r.Pair = target.SetI(period).SetM(market).SetP(fintypes.Binance)

	// download kline
	if market == fintypes.MarketSpot {
		ks, err = ex.in.NewKlinesService().Symbol(target.CustomFormat(ex.Property())).
			Interval(binancePeriod).StartTime(gtime.TimeToEpochMillis(*since)).Limit(1000 /*max limit is 1000*/).Do(context.Background())
		if err != nil {
			return nil, err
		}
	} else if market == fintypes.MarketPerp {
		ksPerp, err := ex.inPerp.NewKlinesService().Symbol(target.CustomFormat(ex.Property())).
			Interval(binancePeriod).StartTime(gtime.TimeToEpochMillis(*since)).Limit(1000 /*max limit is 1000*/).Do(context.Background())
		if err != nil {
			return nil, err
		}

		// convert []*futures.K to []*binance.K
		// 实际上定义一模一样，但是在go-binance库里定义在两个不同模块中
		for _, v := range ksPerp {
			item := binance.Kline{
				OpenTime:                 v.OpenTime,
				Open:                     v.Open,
				High:                     v.High,
				Low:                      v.Low,
				Close:                    v.Close,
				Volume:                   v.Volume,
				CloseTime:                v.CloseTime,
				QuoteAssetVolume:         v.QuoteAssetVolume,
				TradeNum:                 v.TradeNum,
				TakerBuyBaseAssetVolume:  v.TakerBuyBaseAssetVolume,
				TakerBuyQuoteAssetVolume: v.TakerBuyQuoteAssetVolume,
			}
			ks = append(ks, &item)
		}
	} else {
		return nil, gerror.Errorf("unsupported market %s", market)
	}

	for _, v := range ks {
		item := fintypes.Bar{}
		// no need to check v.CloseTime, it is always = OpenTime+Period
		item.T = gtime.EpochMillisToTime(v.OpenTime)
		item.H, err = gdecimal.NewFromString(v.High)
		if err != nil {
			return nil, err
		}
		item.L, err = gdecimal.NewFromString(v.Low)
		if err != nil {
			return nil, err
		}
		item.O, err = gdecimal.NewFromString(v.Open)
		if err != nil {
			return nil, err
		}
		item.C, err = gdecimal.NewFromString(v.Close)
		if err != nil {
			return nil, err
		}
		item.V, err = gdecimal.NewFromString(v.Volume)
		if err != nil {
			return nil, err
		}
		r.Items = append(r.Items, item)
	}

	// fix BTC/TUSD TRX/BNB bad first record like this
	/*
		2018-05-22 06:45:00 +0000 UTC 0.001 0.001 0.001 0.001 1
		2019-03-13 04:00:00 +0000 UTC 0.001 0.001 0.001 0.001 0
		2019-03-13 04:01:00 +0000 UTC 0.001 0.001 0.001 0.001 0
		2019-03-13 04:02:00 +0000 UTC 0.001 0.001 0.001 0.001 0
		2019-03-13 04:03:00 +0000 UTC 0.001 0.001 0.001 0.001 0
		2019-03-13 04:04:00 +0000 UTC 0.001 0.001 0.001 0.001 0
		2019-03-13 04:05:00 +0000 UTC 0.001 0.001 0.001 0.001 0
		2019-03-13 04:06:00 +0000 UTC 0.001 0.001 0.001 0.001 0
		2019-03-13 04:07:00 +0000 UTC 0.001 0.001 0.001 0.001 0
		2019-03-13 04:08:00 +0000 UTC 0.001 0.001 0.001 0.001 0
	*/
	if since.Unix() <= ex.Property().TradeBeginTime.Unix() {
		if r.Items[1].T.Sub(r.Items[0].T) != period.ToDuration() {
			r.Items = r.Items[1:]
		}
	}

	r.Sort()
	return r, nil
}

/*
func (ex *Client) MaxFill(pair Pair) (Fill, error) {
	fs, err := ex.GetFills(pair, nil)
	if err != nil {
		return Fill{}, err
	}
	if len(fs) == 0 {
		return Fill{}, errors.Errorf("pair(%s) get max filled error", pair.String())
	}
	return fs[len(fs)-1], nil
}*/
// get 1 hour filled from since time, 1 hour duration max in binance
/*func (ex *Client) GetFilled(pair Pair, since time.Time) ([]Fill, error) {
	return ex.GetAggFills(pair, &FillOption{BeginTime: since, TimeDuration:time.Hour})
}*/

// get fills by fromId (optional)
// limit: 1000 max
func (ex *Client) GetFills(market fintypes.Market, target fintypes.Pair, since fintypes.Since) ([]fintypes.Fill, error) {
	if err := target.Verify(); err != nil {
		return nil, err
	}
	if since.Id != nil && *since.Id < 0 {
		return nil, gerror.Errorf("invalid since Id %d", *since.Id)
	}
	limit := 1000

	// Note: 貌似这俩设置不管用
	//if err := httpz.SetTLSTimeout(ex.in.HTTPClient, time.Second*3); err != nil {
	//	return nil, err
	//}
	//httpz.SetTimeout(ex.in.HTTPClient, time.Second*5)
	err := error(nil)
	var res []fintypes.Fill

	if market == fintypes.MarketSpot {
		var fills []*binance.Trade
		if since.Id != nil {
			fills, err = ex.in.NewHistoricalTradesService().Symbol(target.CustomFormat(ex.Property())).FromID(*since.Id).Limit(limit).Do(context.Background())
		} else {
			fills, err = ex.in.NewHistoricalTradesService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		}
		if err != nil {
			return nil, err
		}
		for _, v := range fills {
			item := fintypes.Fill{}
			item.Id = v.ID
			item.Time = gtime.EpochMillisToTime(v.Time)
			item.Price, err = gdecimal.NewFromString(v.Price)
			if err != nil {
				return nil, err
			}
			item.UnitQty, err = gdecimal.NewFromString(v.Quantity)
			if err != nil {
				return nil, err
			}
			if v.IsBuyerMaker {
				item.Side = "buy"
			} else {
				item.Side = "sell"
			}
			res = append(res, item)
		}
		return res, nil
	} else if market == fintypes.MarketPerp {
		var fillsPerp []*futures.Trade
		if since.Id != nil {
			fillsPerp, err = ex.inPerp.NewHistoricalTradesService().Symbol(target.CustomFormat(ex.Property())).FromID(*since.Id).Limit(limit).Do(context.Background())
		} else {
			fillsPerp, err = ex.inPerp.NewHistoricalTradesService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		}
		if err != nil {
			return nil, err
		}
		for _, v := range fillsPerp {
			item := fintypes.Fill{}
			item.Id = v.ID
			item.Time = gtime.EpochMillisToTime(v.Time)
			item.Price, err = gdecimal.NewFromString(v.Price)
			if err != nil {
				return nil, err
			}
			item.UnitQty, err = gdecimal.NewFromString(v.Quantity)
			if err != nil {
				return nil, err
			}
			if v.IsBuyerMaker {
				item.Side = "buy"
			} else {
				item.Side = "sell"
			}
			res = append(res, item)
		}
		return res, nil
	} else {
		return nil, gerror.Errorf("Binance doesn't support Market(%s)", market)
	}
}

func (ex *Client) GetBorrowable(margin fintypes.Margin, asset string) (gdecimal.Decimal, error) {
	if margin != fintypes.MarginCross {
		return gdecimal.N0, gerror.Errorf("unsupported margin(%s)", margin)
	}
	maxBorrowable, err := ex.in.NewGetMaxBorrowableService().Asset(asset).Do(context.Background())
	if err != nil {
		return gdecimal.Zero, err
	}
	return gdecimal.NewFromString(maxBorrowable.Amount)
}

func (ex *Client) Borrow(margin fintypes.Margin, asset string, amount gdecimal.Decimal) error {
	if margin != fintypes.MarginCross {
		return gerror.Errorf("unsupported margin(%s)", margin)
	}
	_, err := ex.in.NewMarginLoanService().Asset(asset).Amount(amount.String()).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (ex *Client) Repay(margin fintypes.Margin, asset string, amount gdecimal.Decimal) error {
	if margin != fintypes.MarginCross {
		return gerror.Errorf("unsupported margin(%s)", margin)
	}
	_, err := ex.in.NewMarginRepayService().Asset(asset).Amount(amount.String()).Do(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (ex *Client) Transfer(asset string, amount gdecimal.Decimal, from, to fintypes.SubAcc) error {
	saFrom, err := from.Parse()
	if err != nil {
		return err
	}
	saTo, err := to.Parse()
	if err != nil {
		return err
	}
	fromMarket := saFrom.Market
	fromMargin := saFrom.Margin
	toMarket := saTo.Market
	toMargin := saTo.Margin

	if err := toMarket.Verify(); err != nil {
		return err
	}
	if (fromMarket == fintypes.MarketSpot && fromMargin == fintypes.MarginNo && toMarket == fintypes.MarketSpot && toMargin == fintypes.MarginCross) || (fromMarket == fintypes.MarketSpot && fromMargin == fintypes.MarginCross && toMarket == fintypes.MarketSpot && toMargin == fintypes.MarginNo) {
		mts := ex.in.NewMarginTransferService().Asset(asset).Amount(amount.String())
		if toMarket == fintypes.MarketSpot {
			mts = mts.Type(binance.MarginTransferTypeToMain)
		} else {
			mts = mts.Type(binance.MarginTransferTypeToMargin)
		}
		_, err := mts.Do(context.Background())
		return err
	} else if (fromMarket == fintypes.MarketSpot && toMarket == fintypes.MarketFuture) || (fromMarket == fintypes.MarketFuture && toMarket == fintypes.MarketSpot) {
		mts := ex.in.NewFuturesTransferService().Asset(asset).Amount(amount.String())
		if toMarket == fintypes.MarketSpot {
			mts = mts.Type(binance.FuturesTransferTypeToMain)
		} else {
			mts = mts.Type(binance.FuturesTransferTypeToFutures)
		}
		_, err := mts.Do(context.Background())
		return err
	} else {
		return gerror.Errorf("unsupported transfer %s -> %s", fromMarket, toMarket)
	}
}

func (ex *Client) typeSideToBinance(side fintypes.OrderSide, orderType fintypes.OrderType) (binance.SideType, binance.OrderType, error) {
	resSide := binance.SideTypeBuy
	resType := binance.OrderTypeLimit
	if side == fintypes.OrderSideBuyLong {
		resSide = binance.SideTypeBuy
	} else if side == fintypes.OrderSideSellShort {
		resSide = binance.SideTypeSell
	} else {
		return resSide, resType, errors.Errorf("unsupported OrderSide(%s)", side)
	}
	if orderType == fintypes.OrderTypeLimit {
		resType = binance.OrderTypeLimit
	} else if orderType == fintypes.OrderTypeMarket {
		resType = binance.OrderTypeMarket
	} else if orderType == fintypes.OrderTypeStopLimit {
		resType = binance.OrderTypeStopLossLimit
	} else {
		return resSide, resType, errors.Errorf("unsupported OrderType(%s)", orderType)
	}

	return resSide, resType, nil
}

func (ex *Client) typeSideToBinanceContract(side fintypes.OrderSide, orderType fintypes.OrderType) (futures.SideType, futures.OrderType, error) {
	resSide := futures.SideTypeBuy
	resType := futures.OrderTypeLimit
	if side == fintypes.OrderSideBuyLong {
		resSide = futures.SideTypeBuy
	} else if side == fintypes.OrderSideSellShort {
		resSide = futures.SideTypeSell
	} else {
		return resSide, resType, errors.Errorf("unsupported OrderSide(%s)", side)
	}
	if orderType == fintypes.OrderTypeLimit {
		resType = futures.OrderTypeLimit
	} else if orderType == fintypes.OrderTypeMarket {
		resType = futures.OrderTypeMarket
	} else {
		return resSide, resType, errors.Errorf("unsupported OrderType(%s)", orderType)
	}

	return resSide, resType, nil
}

func (ex *Client) Trade(market fintypes.Market, margin fintypes.Margin, leverage int, target fintypes.Pair, side fintypes.OrderSide, orderType fintypes.OrderType, amount, price, stopPrice gdecimal.Decimal) (*fintypes.OrderId, error) {
	if err := target.Verify(); err != nil {
		return nil, err
	}

	// get market info if necessary
	if ex.marketInfoCache.Infos == nil || ex.property.Clock.Now().Sub(ex.marketInfoUpdate) > gtime.Day {
		_, err := ex.GetMarketInfo() // it will get and cache market info
		if err != nil {
			return nil, err
		}
	}
	/*pairMi, ok := ex.marketInfoCache.Infos[target.SetM(market)]
	if !ok {
		return nil, errors.Errorf("market info required for pair(%s)", target.String())
	}*/

	// process spot and margin trade request
	if market == fintypes.MarketSpot {
		// parse side and type
		bncSide, bncOt, err := ex.typeSideToBinance(side, orderType)
		if err != nil {
			return nil, err
		}

		od := &binance.CreateOrderResponse{}
		if margin == fintypes.MarginNo {
			cos := ex.in.NewCreateOrderService().Symbol(target.CustomFormat(ex.Property())).Side(bncSide).Type(bncOt).TimeInForce(binance.TimeInForceTypeGTC).Quantity(amount. /*.Trunc2(pairMi.UnitMin, pairMi.UnitStep.Float64())*/ String())
			if orderType.IsLimit() {
				cos = cos.Price(price. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String())
			}
			if orderType.IsStopLimit() {
				cos = cos.Price(price. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String()).StopPrice(stopPrice. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String())
			}
			od, err = cos.Do(context.Background())
		} else if margin == fintypes.MarginCross {
			cos := ex.in.NewCreateMarginOrderService().Symbol(target.CustomFormat(ex.Property())).Side(bncSide).Type(bncOt).TimeInForce(binance.TimeInForceTypeGTC).Quantity(amount. /*.Trunc2(pairMi.UnitMin, pairMi.UnitStep.Float64())*/ String())
			if orderType.IsLimit() {
				cos = cos.Price(price. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String())
			}
			if orderType.IsStopLimit() {
				cos = cos.Price(price. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String()).StopPrice(stopPrice. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String())
			}
			od, err = cos.Do(context.Background())
		} else {
			return nil, gerror.Errorf("binance doesn't support Market(%s) & Margin(%s)", market, margin)
		}

		if err != nil {
			return nil, err
		}
		res := fintypes.NewOrderId(market, margin, target, gnum.ToString(od.OrderID))
		return &res, nil
	}

	// process perp trade request
	if market == fintypes.MarketPerp {
		// parse side and type
		bncSide, bncOt, err := ex.typeSideToBinanceContract(side, orderType)
		if err != nil {
			return nil, err
		}

		// 修改仓位模式
		marginType := futures.MarginTypeIsolated
		if margin == fintypes.MarginIsolated {
			marginType = futures.MarginTypeIsolated
		} else if margin == fintypes.MarginCross {
			marginType = futures.MarginTypeCrossed
		} else {
			return nil, gerror.Errorf("Margin(%s) not supported in SetPosition", margin)
		}
		if err := ex.inPerp.NewChangeMarginTypeService().Symbol(target.CustomFormat(ex.Property())).MarginType(marginType).Do(context.Background()); err != nil {
			return nil, err
		}

		// 修改杠杆倍数
		_, err = ex.inPerp.NewChangeLeverageService().Symbol(target.CustomFormat(ex.Property())).Leverage(leverage).Do(context.Background())
		if err != nil {
			return nil, err
		}

		// 下单
		cos := ex.inPerp.NewCreateOrderService().Symbol(target.CustomFormat(ex.Property())).Side(bncSide).Type(bncOt).TimeInForce(futures.TimeInForceTypeGTC).Quantity(amount. /*.Trunc2(pairMi.UnitMin, pairMi.UnitStep.Float64())*/ String())
		if orderType.IsLimit() {
			cos = cos.Price(price. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String())
		}
		if orderType.IsStopLimit() {
			cos = cos.Price(price. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String()).StopPrice(stopPrice. /*.Trunc2(pairMi.QuoteStep, pairMi.QuoteStep.Float64())*/ String())
		}
		od, err := cos.Do(context.Background())
		if err != nil {
			return nil, err
		}
		res := fintypes.NewOrderId(market, margin, target, gnum.ToString(od.OrderID))
		return &res, nil
	}

	return nil, gerror.Errorf("invalid Market(%s) in SetPosition", market)
}

func (ex *Client) GetAllOrders(market fintypes.Market, margin fintypes.Margin, target fintypes.Pair) ([]fintypes.Order, error) {
	if err := target.Verify(); err != nil {
		return nil, err
	}

	var r []fintypes.Order

	if market == fintypes.MarketSpot && margin == fintypes.MarginNo {
		spotOpenOrders, err := ex.in.NewListOrdersService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		if err != nil {
			return nil, err
		}
		for _, o := range spotOpenOrders {
			item, err := ex.binanceOrderToApiOrder(fintypes.MarketSpot, margin, o)
			if err != nil {
				return nil, err
			}
			r = append(r, *item)
		}
	} else if market == fintypes.MarketSpot && margin != fintypes.MarginNo {
		marginOpenOrders, err := ex.in.NewListMarginOrdersService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		if err != nil {
			return nil, err
		}
		for _, o := range marginOpenOrders {
			item, err := ex.binanceOrderToApiOrder(fintypes.MarketSpot, margin, ex.binanceMarginAllOrderToOrder(o))
			if err != nil {
				return nil, err
			}
			r = append(r, *item)
		}
	} else if market == fintypes.MarketPerp {
		perpOpenOrders, err := ex.inPerp.NewListOrdersService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		if err != nil {
			return nil, err
		}
		for _, o := range perpOpenOrders {
			item, err := ex.binancePerpOrderToApiOrder(fintypes.MarketPerp, margin, o)
			if err != nil {
				return nil, err
			}
			r = append(r, *item)
		}
	} else {
		return nil, gerror.Errorf("unsupported Market(%s)", market)
	}

	return r, nil
}

func (ex *Client) GetOpenOrders(market *fintypes.Market, margin *fintypes.Margin, target *fintypes.Pair) ([]fintypes.Order, error) {
	if target != nil {
		if err := target.Verify(); err != nil {
			return nil, err
		}
	}

	var r []fintypes.Order

	// 无杠杆现货
	if market == nil || *market == fintypes.MarketSpot {
		if margin == nil || *margin == fintypes.MarginNo {
			svc := ex.in.NewListOpenOrdersService()
			if target != nil {
				svc = svc.Symbol(target.CustomFormat(ex.Property()))
			}
			spotOpenOrders, err := svc.Do(context.Background())
			if err != nil {
				return nil, err
			}
			for _, o := range spotOpenOrders {
				item, err := ex.binanceOrderToApiOrder(fintypes.MarketSpot, fintypes.MarginNo, o) // 无杠杆现货，可以固化用MarginNo
				if err != nil {
					return nil, err
				}
				r = append(r, *item)
			}
		}
	}

	// 带杠杆现货
	if market == nil || *market == fintypes.MarketSpot {
		if margin == nil || *margin != fintypes.MarginNo {
			svc := ex.in.NewListMarginOpenOrdersService()
			if target != nil {
				svc = svc.Symbol(target.CustomFormat(ex.Property()))
			}
			marginOpenOrders, err := svc.Do(context.Background())
			if err != nil {
				return nil, err
			}
			for _, o := range marginOpenOrders {
				item, err := ex.binanceOrderToApiOrder(fintypes.MarketSpot, fintypes.MarginCross, o) // FIXME binance现货杠杆现在全部是全仓，所以用MarginCross
				if err != nil {
					return nil, err
				}
				r = append(r, *item)
			}
		}
	}

	// FIXME 目前的账户获取永续挂单会报错<APIError> code=-2015, msg=Invalid API-key, IP, or permissions for action, request ip: **。**。**。**
	// 永续合约
	/*if market == nil || *market == fintypes.MarketPerp {
		svc := ex.inPerp.NewListOpenOrdersService()
		if target != nil {
			svc = svc.Symbol(target.CustomFormat(ex.Property()))
		}
		perpOpenOrders, err := svc.Do(context.Background())
		if err != nil {
			return nil, err
		}
		for _, o := range perpOpenOrders {
			item, err := ex.binancePerpOrderToApiOrder(fintypes.MarketPerp, fintypes.MarginIsolated, o) // FIXME binance永续现在全部是逐仓，所以用MarginIsolated，这里很可能不对！
			if err != nil {
				return nil, err
			}
			r = append(r, *item)
		}
	}*/
	return r, nil
}

/*
func (ex *Client) GetOpenOrders(market fintypes.Market, margin fintypes.Margin, target fintypes.Pair) ([]fintypes.Order, error) {
	if err := target.Verify(); err != nil {
		return nil, err
	}

	var r []fintypes.Order

	if market == fintypes.MarketSpot && margin == fintypes.MarginNo {
		spotOpenOrders, err := ex.in.NewListOpenOrdersService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		if err != nil {
			return nil, err
		}
		for _, o := range spotOpenOrders {
			item, err := ex.binanceOrderToApiOrder(fintypes.MarketSpot, margin, o)
			if err != nil {
				return nil, err
			}
			r = append(r, *item)
		}
	} else if market == fintypes.MarketSpot && margin != fintypes.MarginNo {
		marginOpenOrders, err := ex.in.NewListMarginOpenOrdersService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		if err != nil {
			return nil, err
		}
		for _, o := range marginOpenOrders {
			item, err := ex.binanceOrderToApiOrder(fintypes.MarketSpot, margin, o)
			if err != nil {
				return nil, err
			}
			r = append(r, *item)
		}
	} else if market == fintypes.MarketPerp {
		perpOpenOrders, err := ex.inPerp.NewListOpenOrdersService().Symbol(target.CustomFormat(ex.Property())).Do(context.Background())
		if err != nil {
			return nil, err
		}
		for _, o := range perpOpenOrders {
			item, err := ex.binancePerpOrderToApiOrder(fintypes.MarketPerp, margin, o)
			if err != nil {
				return nil, err
			}
			r = append(r, *item)
		}
	} else {
		return nil, gerror.Errorf("unsupported Market(%s)", market)
	}
	return r, nil
}*/

func (ex *Client) GetOrder(id fintypes.OrderId) (*fintypes.Order, error) {
	if err := id.Verify(); err != nil {
		return nil, err
	}
	int64Id, err := strconv.ParseInt(id.StrId(), 10, 64)
	if err != nil {
		return nil, err
	}
	market := id.Market()
	if market == fintypes.MarketError {
		return nil, gerror.Errorf("OrderId(%s) in CancelOrder required market member", id.String())
	}
	margin := id.Margin()
	if margin == fintypes.MarginError {
		return nil, gerror.Errorf("OrderId(%s) in CancelOrder required margin member", id.String())
	}

	// perp market
	if market == fintypes.MarketPerp {
		odPerp, err := ex.inPerp.NewGetOrderService().Symbol(id.Pair().CustomFormat(ex.Property())).OrderID(int64Id).Do(context.Background())
		if err != nil {
			return nil, err
		}
		return ex.binancePerpOrderToApiOrder(market, margin, odPerp)
	}

	// spot market
	od := &binance.Order{}
	if market == fintypes.MarketSpot && margin == fintypes.MarginNo {
		od, err = ex.in.NewGetOrderService().Symbol(id.Pair().CustomFormat(ex.Property())).OrderID(int64Id).Do(context.Background())
	} else if market == fintypes.MarketSpot && margin != fintypes.MarginNo {
		od, err = ex.in.NewGetMarginOrderService().Symbol(id.Pair().CustomFormat(ex.Property())).OrderID(int64Id).Do(context.Background())
	} else {
		err = gerror.Errorf("unsupported Market(%s)", market)
	}
	if err != nil {
		return nil, err
	}
	return ex.binanceOrderToApiOrder(market, margin, od)
}

func (ex *Client) CancelOrder(id fintypes.OrderId) error {
	// check input param
	if err := id.Verify(); err != nil {
		return err
	}
	market := id.Market()
	if market == fintypes.MarketError {
		return gerror.Errorf("OrderId(%s) in CancelOrder required market member", id.String())
	}
	margin := id.Margin()
	if margin == fintypes.MarginError {
		return gerror.Errorf("OrderId(%s) in CancelOrder required margin member", id.String())
	}
	int64Id, err := strconv.ParseInt(id.StrId(), 10, 64)
	if err != nil {
		return err
	}

	if market == fintypes.MarketSpot && margin == fintypes.MarginNo {
		_, err = ex.in.NewCancelOrderService().Symbol(id.Pair().CustomFormat(ex.Property())).OrderID(int64Id).Do(context.Background())
	} else if market == fintypes.MarketSpot && margin != fintypes.MarginNo {
		_, err = ex.in.NewCancelMarginOrderService().Symbol(id.Pair().CustomFormat(ex.Property())).OrderID(int64Id).Do(context.Background())
	} else if market == fintypes.MarketPerp {
		_, err = ex.inPerp.NewCancelOrderService().Symbol(id.Pair().CustomFormat(ex.Property())).OrderID(int64Id).Do(context.Background())
	} else {
		err = gerror.Errorf("unsupported Market/Margin(%s,%s)", market, margin)
	}
	if err != nil {
		return err
	}
	return nil
}

/*
type MarketStat struct {
	OpenTime    time.Time `json:"OpenTime"`
	CloseTime   time.Time `json:"CloseTime"`
	O        decimals.Decimal   `json:"O"`
	H        decimals.Decimal   `json:"H"`
	L         decimals.Decimal   `json:"L"`
	Last        decimals.Decimal   `json:"Last"`
	UnitVolume  decimals.Decimal   `json:"UnitVolume"`
	QuoteVolume decimals.Decimal   `json:"QuoteVolume"`
	BidPrice    decimals.Decimal   `json:"b"`
	BidQty      decimals.Decimal   `json:"B"`
	AskPrice    decimals.Decimal   `json:"a"`
	AskQty      decimals.Decimal   `json:"A"`
}

type MarketStats struct {
	T  time.Time
	items map[PairExt]MarketStat
}

func newMarketStats() *MarketStats {
	r := MarketStats{
		items: make(map[PairExt]MarketStat),
	}
	return &r
}

func (ex *Client) SubMarketStat(interval time.Duration) (retC chan *MarketStats, doneC, stopC chan struct{}, errC chan error, err error) {
	retC = make(chan *MarketStats, 1024)
	errC = make(chan error, 1024)
	lastSnapTime := clock.ZeroTime

	// 最近的24小时的"K线"，但是由于起止时间是滑动的，所以只能用作获取实时Tick
	doneC, stopC, err = binance.WsAllMarketsStatServe(
		func(event binance.WsAllMarketsStatEvent) {
			if ex.property.Clock.Now().Sub(lastSnapTime) < interval {
				return
			}

			ms := newMarketStats()
			for _, v := range event {
				pair, err := ParsePairCustom(v.Symbol, ex.Property())
				if err != nil {
					errC <- err
					continue
				}
				open, err := decimals.NewFromString(v.OpenPrice)
				if err != nil {
					errC <- err
					continue
				}
				high, err := decimals.NewFromString(v.HighPrice)
				if err != nil {
					errC <- err
					continue
				}
				low, err := decimals.NewFromString(v.LowPrice)
				if err != nil {
					errC <- err
					continue
				}
				last, err := decimals.NewFromString(v.LastPrice)
				if err != nil {
					errC <- err
					continue
				}
				/*unitAvgPrice, err := strconv.ParseFloat(v.WeightedAvgPrice, 64)
				if err != nil {
					errC <- err
					continue
				}*/ /*
				unitVolume, err := decimals.NewFromString(v.BaseVolume)
				if err != nil {
					errC <- err
					continue
				}
				quoteVolume, err := decimals.NewFromString(v.QuoteVolume)
				if err != nil {
					errC <- err
					continue
				}
				bidPrice, err := decimals.NewFromString(v.BidPrice)
				if err != nil {
					errC <- err
					continue
				}
				bidQty, err := decimals.NewFromString(v.BidQty)
				if err != nil {
					errC <- err
					continue
				}
				askPrice, err := decimals.NewFromString(v.AskPrice)
				if err != nil {
					errC <- err
					continue
				}
				askQty, err := decimals.NewFromString(v.AskQty)
				if err != nil {
					errC <- err
					continue
				}
				item := MarketStat{
					OpenTime:    clock.EpochMillisToTime(v.OpenTime),
					CloseTime:   clock.EpochMillisToTime(v.CloseTime),
					O:        open,
					H:        high,
					L:         low,
					Last:        last,
					UnitVolume:  unitVolume,
					QuoteVolume: quoteVolume,
					BidPrice:    bidPrice,
					BidQty:      bidQty,
					AskPrice:    askPrice,
					AskQty:      askQty,
				}
				ms.T = clock.EpochMillisToTime(v.T)
				ms.items[pair.SetP(Binance).SetM(MarketSpot)] = item
			}

			lastSnapTime = ex.property.Clock.Now()
			retC <- ms
		},
		func(err error) {
			errC <- err
		},
	)

	if err != nil {
		return nil, nil, nil, nil, err
	}
	return retC, doneC, stopC, errC, nil
}

func (ex *Client) SubKline(pair PairExt, period Period) (retC chan *Bar, doneC, stopC chan struct{}, errC chan error, err error) {
	retC = make(chan *Bar, 1024)
	errC = make(chan error, 1024)
	periodStr, err := period.CustomFormat(ex.Property())
	if err != nil {
		return nil, nil, nil, nil, err
	}

	doneC, stopC, err = binance.WsKlineServe(
		pair.Pair().CustomFormat(ex.Property()),
		periodStr,
		func(event *binance.WsKlineEvent) {
			if !event.K.IsFinal {
				return
			}
			open, err := decimals.NewFromString(event.K.O)
			if err != nil {
				errC <- err
				return
			}
			high, err := decimals.NewFromString(event.K.H)
			if err != nil {
				errC <- err
				return
			}
			low, err := decimals.NewFromString(event.K.L)
			if err != nil {
				errC <- err
				return
			}
			_close, err := decimals.NewFromString(event.K.C)
			if err != nil {
				errC <- err
				return
			}
			unitVolume, err := decimals.NewFromString(event.K.V)
			if err != nil {
				errC <- err
				return
			}
			/*quoteVolume, err := strconv.ParseFloat(event.Bar.V, 64)
			if err != nil {
				errC <- err
				return
			}*/ /*
			item := Bar{
				T: clock.EpochMillisToTime(event.K.StartTime),
				//CloseTime:xclock.EpochMillisToTime(v.CloseTime),
				O:   open,
				H:   high,
				L:    low,
				C:  _close,
				V: unitVolume,
				//QuoteVolume:quoteVolume,
			}
			retC <- &item
		},
		func(err error) {
			errC <- err
		},
	)

	if err != nil {
		return nil, nil, nil, nil, err
	}
	return retC, doneC, stopC, errC, nil
}*/

/*
func (ex *Client) SubDepth(pair PairExt, period Period) (retC chan *Depth, doneC, stopC chan struct{}, errC chan error, err error) {
	retC = make(chan *Bar, 1024)
	errC = make(chan error, 1024)
	periodStr, err := period.CustomFormat(ex.Property())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	binance.WsAllMarketsStatServe()

	doneC, stopC, err = binance.WsDepthServe(
		pair.Pair().CustomFormat(ex.Property()),
		func(event *binance.WsDepthEvent) {
			if !event..IsFinal {
				return
			}
			open, err := decimals.NewFromString(event.K.O)
			if err != nil {
				errC <- err
				return
			}
			high, err := decimals.NewFromString(event.K.H)
			if err != nil {
				errC <- err
				return
			}
			low, err := decimals.NewFromString(event.K.L)
			if err != nil {
				errC <- err
				return
			}
			_close, err := decimals.NewFromString(event.K.C)
			if err != nil {
				errC <- err
				return
			}
			unitVolume, err := decimals.NewFromString(event.K.V)
			if err != nil {
				errC <- err
				return
			}
			/*quoteVolume, err := strconv.ParseFloat(event.Bar.V, 64)
			if err != nil {
				errC <- err
				return
			}*/ /*
			item := Bar{
				T: clock.EpochMillisToTime(event.K.StartTime),
				//CloseTime:xclock.EpochMillisToTime(v.CloseTime),
				O:   open,
				H:   high,
				L:    low,
				C:  _close,
				V: unitVolume,
				//QuoteVolume:quoteVolume,
			}
			retC <- &item
		},
		func(err error) {
			errC <- err
		},
	)

	if err != nil {
		return nil, nil, nil, nil, err
	}
	return retC, doneC, stopC, errC, nil
}*/

// get agg fills by option
// API limit: 1 hour duration max, 1000 IdLimit max
func (ex *Client) GetAggFills(pair fintypes.Pair, option *fintypes.FillOption) ([]fintypes.Fill, error) {
	if err := pair.Verify(); err != nil {
		return nil, err
	}
	if option != nil {
		if err := option.VerifyBinance(); err != nil {
			return nil, err
		}
	}

	var fills []*binance.AggTrade
	err := error(nil)

	if option != nil {
		if option.BeginId > 0 {
			fills, err = ex.in.NewAggTradesService().Symbol(pair.CustomFormat(ex.Property())).FromID(option.BeginId).Limit(int(option.IdLimit)).Do(context.Background())
		} else {
			fills, err = ex.in.NewAggTradesService().Symbol(pair.CustomFormat(ex.Property())).StartTime(gtime.TimeToEpochMillis(option.BeginTime)).EndTime(gtime.TimeToEpochMillis(option.BeginTime.Add(option.TimeDuration))).Do(context.Background())
		}
	} else {
		fills, err = ex.in.NewAggTradesService().Symbol(pair.CustomFormat(ex.Property())).Do(context.Background())
	}

	if err != nil {
		return nil, err
	}

	var r []fintypes.Fill
	for _, v := range fills {
		item := fintypes.Fill{}
		item.Id = v.AggTradeID
		item.Time = gtime.EpochMillisToTime(v.Timestamp)
		item.Price, err = gdecimal.NewFromString(v.Price)
		if err != nil {
			return nil, err
		}
		item.UnitQty, err = gdecimal.NewFromString(v.Quantity)
		if err != nil {
			return nil, err
		}
		if v.IsBuyerMaker {
			item.Side = "buy"
		} else {
			item.Side = "sell"
		}
		r = append(r, item)
	}
	return r, nil
}
