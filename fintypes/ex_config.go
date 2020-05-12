package fintypes

import (
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"time"
)

/*
note: all timestamps are UTC + 0 timezone
*/

var (
	ErrFunctionNotSupported = errors.Errorf("function not supported")
	AllSupportedExs         = []Platform{Binance}
)

type (
	ExProperty struct {
		Name                   Platform
		Email                  string // optional
		MaxDepth               int
		MaxFills               int
		PairDelimiter          string // pair separator
		PairDelimiterLeftTail  []string
		PairDelimiterRightHead []string
		PairNormalOrder        bool // whether is ISO order —— unit first quote second
		PairUpperCase          bool
		PairsSeparator         string // separator between multiple pairs in api request
		Periods                map[Period]string
		OrderStatus            map[OrderStatus]string
		OrderTypes             map[OrderType]string // FIXME 这里用OrderSide还是OrderType
		OrderSides             map[OrderSide]string
		RateLimit              time.Duration
		FillRateLimit          time.Duration
		KlineRateLimit         time.Duration
		MakerFee               gdecimal.Decimal // fee of depth maker, 挂单等吃费率
		TakerFee               gdecimal.Decimal // fee of depth taker, 主动吃单费率
		WithdrawalFees         map[string]gdecimal.Decimal
		MarketEnabled          map[Market]bool
		Clock                  gtime.Clock
		IsBackTestEx           bool
		TradeBeginTime         time.Time
	}
)

func (cc ExProperty) SupportedPeriods() []Period {
	var r []Period
	for p := range cc.Periods {
		r = append(r, p)
	}
	return r
}

func (cc ExProperty) MinPeriod() Period {
	minPeriod := PeriodError

	for k := range cc.Periods {
		if minPeriod == PeriodError {
			minPeriod = k
		} else {
			if k.ToSeconds() < minPeriod.ToSeconds() {
				minPeriod = k
			}
		}
	}
	return minPeriod
}
