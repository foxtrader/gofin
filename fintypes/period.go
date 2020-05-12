package fintypes

import (
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"github.com/shawnwyckoff/gopkg/container/gternary"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"time"
)

// Period types

const (
	PeriodError       Period = ""
	Period1Min        Period = "1min"
	Period3Min        Period = "3min"
	Period5Min        Period = "5min"
	Period15Min       Period = "15min"
	Period30Min       Period = "30min"
	Period1Hour       Period = "1hour"
	Period2Hour       Period = "2hour"
	Period4Hour       Period = "4hour"
	Period6Hour       Period = "6hour"
	Period8Hour       Period = "8hour"
	Period12Hour      Period = "12hour"
	Period1Day        Period = "1day"
	Period1Week       Period = "1week"
	Period1MonthFUZZY Period = "1mon"
	Period1YearFUZZY  Period = "1year"
)

var (
	AllPeriods = []Period{
		Period1Min,
		Period3Min,
		Period5Min,
		Period15Min,
		Period30Min,
		Period1Hour,
		Period2Hour,
		Period4Hour,
		Period6Hour,
		Period8Hour,
		Period12Hour,
		Period1Day,
		Period1Week,
		Period1MonthFUZZY,
		Period1YearFUZZY,
	}

	DefaultPeriodRoundConfig = PeriodRoundConfig{Location: *time.UTC, WeekBegin: time.Sunday, UseLocalZeroOClockAsDayBeginning: false}
)

type Period string

func (p Period) String() string {
	return string(p)
}

func (p Period) ToSeconds() int64 {
	switch p {
	case PeriodError:
		return 0
	case Period1Min:
		return int64(time.Minute / time.Second)
	case Period3Min:
		return int64((time.Minute / time.Second) * 3)
	case Period5Min:
		return int64((time.Minute / time.Second) * 5)
	case Period15Min:
		return int64((time.Minute / time.Second) * 15)
	case Period30Min:
		return int64((time.Minute / time.Second) * 30)
	case Period1Hour:
		return int64(time.Hour / time.Second)
	case Period2Hour:
		return int64((time.Hour / time.Second) * 2)
	case Period4Hour:
		return int64((time.Hour / time.Second) * 4)
	case Period6Hour:
		return int64((time.Hour / time.Second) * 6)
	case Period8Hour:
		return int64((time.Hour / time.Second) * 8)
	case Period12Hour:
		return int64((time.Hour / time.Second) * 12)
	case Period1Day:
		return int64((time.Hour / time.Second) * 24)
	//case Period3Day:
	//	return int64((time.Hour / time.Second) * 3 * 24)
	case Period1Week:
		return int64((time.Hour / time.Second) * 7 * 24)
	case Period1MonthFUZZY:
		return int64((time.Hour / time.Second) * 30 * 24)
	case Period1YearFUZZY:
		return int64((time.Hour / time.Second) * 365 * 24)
	default:
		return -1
	}
}

func (p Period) ToDuration() time.Duration {
	sec := p.ToSeconds()
	return gtime.MulDuration(sec, time.Second)
}

// 对于Period1MonthFUZZY和Period1YearFUZZY也能精确计算
func (p Period) ToDurationExact(t time.Time, tz *time.Location) time.Duration {
	if tz == nil {
		tz = time.UTC
	}
	if p == Period1MonthFUZZY {
		return gtime.GetMonthDuration(t.In(tz).Year(), int(t.In(tz).Month()))
	} else if p == Period1YearFUZZY {
		return gtime.GetYearDuration(t.In(tz).Year())
	} else {
		sec := p.ToSeconds()
		return gtime.MulDuration(sec, time.Second)
	}
}

func (p Period) CustomFormat(config *ExProperty) (string, error) {
	if s, ok := config.Periods[p]; ok {
		return s, nil
	} else {
		return "", errors.Errorf("exchange %s doesn't support %s period", config.Name.String(), p.String())
	}
}

func (p Period) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Period) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = gstring.RemoveHead(s, 1)
	s = gstring.RemoveTail(s, 1)
	if sec := Period(s).ToSeconds(); sec == -1 {
		return errors.Errorf("unknown period(%s)", s)
	}
	*p = Period(s)
	return nil
}

/*
func (p *Period) Of(pt PairAt) PeriodPairAt {
	return PeriodPairAt(p.String() + delimiterBetweenPeriodAndPairAt + pt.String())
}*/

func ParsePeriod(s string) (Period, error) {
	p := Period(s)
	for _, v := range AllPeriods {
		if p == v {
			return p, nil
		}
	}
	return PeriodError, errors.Errorf("invalid Period %s", s)
}

type PeriodRoundConfig struct {
	Location  time.Location // FIXME 按下面的说法是不是不需要Location？
	WeekBegin time.Weekday
	// 本身time.Round是和时区无关的，都是依据UTC进行Round的，但是金融市场中根据Period进行RoundPeriod时，某些Period（周）就需要参考时区了，同一时刻不同时区可能属于不同的Weekday
	// 当然，如果偏要以当地时间的0点作为Round的起点，那么就要把UseLocalZeroOClockAsDayBeginning设置为true
	// 以2019-09-01 09:00:00 +0800 CST 为例，Round24小时，标准结果应该等于2019-09-01 08:00:00 +0800 CST，这个是Go标准库的执行结果
	// 要遵循Go标准库time.Round的思路，就需要把UseLocalZeroOClockAsDayBeginning设置为false
	UseLocalZeroOClockAsDayBeginning bool // false: use UTC zero clock
}

func (cc PeriodRoundConfig) String() string {
	return cc.Location.String() + "," + cc.WeekBegin.String() + "," + gternary.If(cc.UseLocalZeroOClockAsDayBeginning).String("true", "false")
}

// 根据给定时间和周期值，计算它归属于哪个OpenTime的周期
// get kline dot open time by dotTime
func RoundPeriodEarlier(dotTime time.Time, period Period, prc PeriodRoundConfig) time.Time {
	//tz := dotTime.Location()
	tz := &prc.Location
	dotTime = dotTime.In(time.UTC)

	if period == Period1YearFUZZY {
		if prc.UseLocalZeroOClockAsDayBeginning {
			return time.Date(dotTime.In(tz).Year(), 1, 1, 0, 0, 0, 0, tz)
		} else {
			return time.Date(dotTime.Year(), 1, 1, 0, 0, 0, 0, time.UTC).In(tz)
		}
	} else if period == Period1MonthFUZZY {
		if prc.UseLocalZeroOClockAsDayBeginning {
			return time.Date(dotTime.In(tz).Year(), dotTime.In(tz).Month(), 1, 0, 0, 0, 0, tz)
		} else {
			return time.Date(dotTime.Year(), dotTime.Month(), 1, 0, 0, 0, 0, time.UTC).In(tz)
		}
	} else if period == Period1Week {
		if prc.UseLocalZeroOClockAsDayBeginning {
			dotTime = time.Date(dotTime.In(tz).Year(), dotTime.In(tz).Month(), dotTime.In(tz).Day(), 0, 0, 0, 0, tz)
		} else {
			dotTime = time.Date(dotTime.Year(), dotTime.Month(), dotTime.Day(), 0, 0, 0, 0, time.UTC).In(tz)
		}
		// Weekday()以周日开始，但是Period以第一个工作日周一开始
		days := int(dotTime.Weekday() - prc.WeekBegin)
		if days == -1 {
			days = 6
		}
		return gtime.Sub(dotTime, gtime.Day*time.Duration(days))
	} else if period == Period1Day {
		if prc.UseLocalZeroOClockAsDayBeginning {
			return time.Date(dotTime.In(tz).Year(), dotTime.In(tz).Month(), dotTime.In(tz).Day(), 0, 0, 0, 0, tz)
		} else {
			return time.Date(dotTime.Year(), dotTime.Month(), dotTime.Day(), 0, 0, 0, 0, time.UTC).In(tz)
			// 等同于return clock.RoundEarlier(dotTime, period.ToDuration()).In(tz)
		}
	} else {
		return gtime.RoundEarlier(dotTime, period.ToDuration()).In(tz)
	}
}

func DurationToPeriod(d time.Duration) Period {
	for _, p := range AllPeriods {
		if p.ToDuration() == d {
			return p
		}
	}

	minMon := time.Hour * 24 * 28
	maxMon := time.Hour * 24 * 31
	if minMon <= d && d <= maxMon {
		return Period1MonthFUZZY
	}

	minYear := time.Hour * 24 * 365
	maxYear := time.Hour * 24 * 366
	if minYear <= d && d <= maxYear {
		return Period1YearFUZZY
	}

	return PeriodError
}
