package fintypes

import (
	"github.com/shawnwyckoff/gopkg/apputil/gtest"
	"golang.org/x/text/currency"
	"testing"
)

func TestCurrencyCodesHistorical(t *testing.T) {
	codes := CurrencyCodesHistorical()

	hasLTL := false
	for _, v := range codes {
		if v.String() == "LTL" {
			hasLTL = true
		}
	}
	if !hasLTL {
		gtest.PrintlnExit(t, "LTL not found")
	}
}

func TestCurrencyIsFiat(t *testing.T) {
	cl := gtest.NewCaseList()
	cl.New().Input(currency.XAU).Expect(false)
	cl.New().Input(currency.XAG).Expect(false)
	cl.New().Input(currency.XPD).Expect(false)
	cl.New().Input(currency.XPT).Expect(false)
	cl.New().Input(currency.XTS).Expect(false)
	cl.New().Input(currency.XXX).Expect(false)
	cl.New().Input(currency.USD).Expect(true)
	cl.New().Input(currency.CNY).Expect(true)
	cl.New().Input(currency.EUR).Expect(true)

	for _, v := range cl.Get() {
		unit := v.Inputs[0].(currency.Unit)
		got := CurrencyIsFiat(unit)
		expect := v.Expects[0].(bool)
		if got != expect {
			gtest.PrintlnExit(t, "IsFiat(%s) got %t, but %t expect", unit.String(), got, expect)
		}
	}
}
