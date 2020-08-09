package fintypes

import (
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"time"
)

/*
note: all timestamps are UTC + 0 timezone
*/

type (
	ExApi string
)

const (
	ExApiGetKline ExApi = ExApi("ExApiGetKline")
	ExApiGetFill  ExApi = ExApi("ExApiGetFill")
)

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
		RateLimits             map[ExApi]time.Duration
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
