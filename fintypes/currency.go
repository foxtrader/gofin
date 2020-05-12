package fintypes

import (
	"fmt"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"golang.org/x/text/currency"
	"strings"
	"time"
	"unicode"
)

func CurrencyCodesHistorical() []currency.Unit {
	var r []currency.Unit
	iter := currency.Query(currency.Historical) // NOTE: if currency.Historical used, there are some historical currencies, they are NOT valid now.
	for iter.Next() {
		r = append(r, iter.Unit())
	}
	return r
}

func CurrencyCodesNow() []currency.Unit {
	var r []currency.Unit
	iter := currency.Query(currency.Date(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)))
	for iter.Next() {
		r = append(r, iter.Unit())
	}
	return r
}

func CurrencyIsMetal(unit currency.Unit) bool {
	switch unit {
	case currency.XAU:
		return true
	case currency.XAG:
		return true
	case currency.XPD:
		return true
	case currency.XPT:
		return true
	default:
		return false
	}
}

func CurrencyIsTest(unit currency.Unit) bool {
	return unit == currency.XTS || unit == currency.XXX
}

func CurrencyIsFiat(unit currency.Unit) bool {
	return !CurrencyIsMetal(unit) && !CurrencyIsTest(unit)
}

// Returns $ like UTF8 string
func CurrencyGetSymbol(unit currency.Unit) string {
	s := fmt.Sprintf("%s", currency.Symbol(unit.Amount(1.2)))
	splitPos := gstring.LenUTF8(s)
	for i, v := range []rune(s) {
		if v == ' ' || unicode.IsDigit(v) {
			splitPos = i
			break
		}
	}
	return string([]rune(s)[0:splitPos])
}

// Support ISO currency name and symbol
// 注意，有些常见的Symbol是简化的不完整版本，会在多个法币的完整Symbol中出现，比如¥和$
func CurrencyParse(s string) (currency.Unit, error) {
	s = strings.TrimSpace(s)
	for _, v := range CurrencyCodesHistorical() {
		if strings.ToLower(s) == strings.ToLower(v.String()) ||
			strings.ToLower(s) == strings.ToLower(CurrencyGetSymbol(v)) {
			return v, nil
		}
	}
	return currency.Unit{}, gerror.Errorf("PlatformUnknown currency unit %s", s)
}
