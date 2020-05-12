package findata

import (
	"github.com/foxtrader/gofin/fintypes"
	fintypes2 "github.com/foxtrader/gofin/fintypes"
	"github.com/openprovider/ecbrates"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"math"
	"strconv"
)

/**
forex 外汇
// https://fixer.io 更齐全, https://github.com/LordotU/go-fixerio/
*/

type (
	Ecb struct{}
)

func NewEcb() *Ecb {
	return &Ecb{}
}

func rate2tick(data ecbrates.Rates) (gtime.Date, map[fintypes.Pair]gdecimal.Decimal, error) {
	rTick := map[fintypes.Pair]gdecimal.Decimal{}
	rDate, err := gtime.ParseDateString(data.Date, true)
	if err != nil {
		return gtime.ZeroDate, nil, err
	}

	for k, v := range data.Rate {
		unit, err := fintypes.CurrencyParse(string(k))
		if err != nil {
			return gtime.ZeroDate, nil, err
		}
		price, err := gdecimal.NewFromString(v.(string))
		if err != nil {
			return gtime.ZeroDate, nil, err
		}
		rTick[fintypes.NewPair("EUR", unit.String())] = price
	}

	return rDate, rTick, nil
}

func (ecb *Ecb) GetCurrentTicks() (map[fintypes.Pair]gdecimal.Decimal, error) {
	data, err := ecbrates.New()
	if err != nil {
		return nil, err
	}
	_, rTick, err := rate2tick(*data)
	return rTick, err
}

func (ecb *Ecb) GetKline() (map[fintypes.Pair]fintypes2.Kline, error) {
	//rates, err := ecbrates.Load() // 90 days history
	rates, err := ecbrates.LoadAll() // ALL history
	if err != nil {
		return nil, err
	}

	r := map[fintypes.Pair]fintypes2.Kline{}
	for _, v := range rates {
		date, tick, err := rate2tick(v)
		if err != nil {
			return nil, err
		}
		for k, v := range tick {
			origin := r[k]
			origin.UpsertDot(fintypes2.Bar{T: date.ToTimeUTC(), O: v, H: v, L: v, C: v})
			r[k] = origin
		}
	}

	for k, v := range r {
		v.Sort()
		r[k] = v
	}
	return r, nil
}

func round64(x float64, prec int) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}

	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)

	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow * sign
}

// Convert a value "from" one Currency -> "to" other Currency
func Convert(r *ecbrates.Rates, value float64, from, to ecbrates.Currency) (float64, error) {
	if r.Rate[to] == nil || r.Rate[from] == nil {
		return 0, gerror.New("Perhaps one of the values ​​of currencies does not exist")
	}
	errorMessage := "Perhaps one of the values ​​of currencies could not parsed correctly"
	strFrom, okFrom := r.Rate[from].(string)
	strTo, okTo := r.Rate[to].(string)
	if !okFrom || !okTo {
		return 0, gerror.New(errorMessage)
	}
	vFrom, err := strconv.ParseFloat(strFrom, 32)
	if err != nil {
		return 0, gerror.New(errorMessage)
	}
	vTo, err := strconv.ParseFloat(strTo, 32)
	if err != nil {
		return 0, gerror.New(errorMessage)
	}
	return round64(value*round64(vTo, 4)/round64(vFrom, 4), 4), nil
}
