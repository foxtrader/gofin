package fintypes

/*
一个交易有多个属性，包括：

OrderType 市价单或者限价单（必要）
market/limit

OrderSide 做空还是做多（必要）
short/long

TradeIntent这个交易的主观目的（注解）
open/reduce/add/close

TradeIncome这个交易（reduce或者close时）是否盈利（注解）
stop-loss/stop-profit
*/

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gjson"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"strings"
	"time"
)

type (
	OrderId string

	OrderStatus string

	OrderSide string

	OrderType string

	TradeIntent string // 交易意图

	TradeIncome string

	Order struct {
		Id         OrderId
		Time       time.Time
		Market     Market
		Margin     Margin
		Leverage   int
		Pair       Pair
		Side       OrderSide
		Type       OrderType
		Status     OrderStatus
		Price      gdecimal.Decimal
		Amount     gdecimal.Decimal // initial total amount in unit, unit always
		AvgPrice   gdecimal.Decimal // Binance貌似不提供AvgPrice
		DealAmount gdecimal.Decimal // filled amount in unit, NOT quote,PaperEx在撮合的时候是这么理解的，如果以后要改，也要修正paperEx
		Fee        gdecimal.Decimal // Binance貌似不提供Fee
	}
)

const (
	OrderIdDelimiter = ":"

	OrderStatusError           OrderStatus = ""
	OrderStatusNew             OrderStatus = "new"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCanceled        OrderStatus = "canceled"
	OrderStatusCanceling       OrderStatus = "canceling"
	OrderStatusRejected        OrderStatus = "rejected"
	OrderStatusExpired         OrderStatus = "expired"

	OrderSideError     OrderSide = ""
	OrderSideBuyLong   OrderSide = "buy"
	OrderSideSellShort OrderSide = "sell"

	OrderTypeError  OrderType = ""
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"

	TradeIntentError  TradeIntent = ""
	TradeIntentOpen   TradeIntent = "open"   // 开仓进场
	TradeIntentReduce TradeIntent = "reduce" // 减仓
	TradeIntentAdd    TradeIntent = "add"    // 加仓
	TradeIntentClose  TradeIntent = "close"  // 平仓离场

	TradeIncomeError  TradeIncome = ""
	TradeIncomeLoss   TradeIncome = "loss"   // 止损
	TradeIncomeProfit TradeIncome = "profit" // 止盈
)

func (ts OrderStatus) String() string {
	return string(ts)
}

func (ts OrderStatus) End() bool {
	return ts == OrderStatusFilled || ts == OrderStatusCanceled || ts == OrderStatusRejected || ts == OrderStatusExpired
}

func NewOrderId(market Market, margin Margin, pair Pair, strId string) OrderId {
	return OrderId(fmt.Sprintf("%s%s%s%s%s%s%s", market, OrderIdDelimiter, margin, OrderIdDelimiter, pair.String(), OrderIdDelimiter, strId))
}

func (id OrderId) Market() Market {
	ss := strings.Split(string(id), OrderIdDelimiter)
	if len(ss) != 4 {
		return MarketError
	}
	accType := Market(ss[0])
	if err := accType.Verify(); err != nil {
		return MarketError
	}
	return accType
}

func (id OrderId) Margin() Margin {
	ss := strings.Split(string(id), OrderIdDelimiter)
	if len(ss) != 4 {
		return MarginError
	}
	accType := Margin(ss[1])
	if err := accType.Verify(); err != nil {
		return MarginError
	}
	return accType
}

func (id OrderId) Pair() Pair {
	ss := strings.Split(string(id), OrderIdDelimiter)
	if len(ss) != 4 {
		return PairErr
	}
	p, err := ParsePair(ss[2])
	if err != nil {
		return PairErr
	}
	return p
}

func (id OrderId) StrId() string {
	ss := strings.Split(string(id), OrderIdDelimiter)
	if len(ss) != 4 {
		return ""
	}
	return ss[3]
}

func (id OrderId) Verify() error {
	errInvalidOrderId := errors.Errorf(`invalid OrderId(%s)`, string(id))
	if id.Market() == MarketError {
		return errInvalidOrderId
	}
	if id.Pair() == PairErr {
		return errInvalidOrderId
	}
	if id.StrId() == "" {
		return errInvalidOrderId
	}
	return nil
}

func (id OrderId) String() string {
	return string(id)
}

func (id OrderId) MarshalJSON() ([]byte, error) {
	return []byte(`"` + id.String() + `"`), nil
}

func (id *OrderId) UnmarshalJSON(b []byte) error {
	errInvalidOrderId := errors.Errorf(`invalid OrderId(%s)`, string(b))

	s := string(b)
	s = gstring.RemoveHead(s, 1)
	s = gstring.RemoveTail(s, 1)

	oi := OrderId(s)
	if oi.Market() == MarketError {
		return errInvalidOrderId
	}
	if oi.Pair() == PairErr {
		return errInvalidOrderId
	}
	if oi.StrId() == "" {
		return errInvalidOrderId
	}

	*id = oi
	return nil
}

func (od Order) String() string {
	return gjson.MarshalStringDefault(od, false)
}

func (tt OrderSide) String() string {
	return string(tt)
}

func (tt OrderSide) IsBuy() bool {
	return tt == OrderSideBuyLong
}

func (tt OrderSide) IsSell() bool {
	return tt == OrderSideSellShort
}

func (tt OrderSide) CustomFormat(config *ExProperty) string {
	for k, v := range config.OrderSides {
		if k == tt {
			return v
		}
	}
	return tt.String()
}

func (tt OrderSide) Verify() error {
	if tt != OrderSideBuyLong && tt != OrderSideSellShort {
		return errors.Errorf("invalid Side(%s)", string(tt))
	}
	return nil
}

func (tt OrderType) String() string {
	return string(tt)
}

func (tt OrderType) IsLimit() bool {
	return tt == OrderTypeLimit
}

func (tt OrderType) IsMarket() bool {
	return tt == OrderTypeMarket
}

func (tt OrderType) CustomFormat(config *ExProperty) string {
	for k, v := range config.OrderTypes {
		if k == tt {
			return v
		}
	}
	return tt.String()
}

func (tt OrderType) Verify() error {
	if tt != OrderTypeLimit && tt != OrderTypeMarket {
		return errors.Errorf("invalid Side(%s)", string(tt))
	}
	return nil
}
