package fintypes

import (
	"golang.org/x/text/currency"
	"testing"
)

func TestParseUnit(t *testing.T) {
	type item struct {
		s        string
		expected currency.Unit
	}
	items := []item{
		{s: "US$ ", expected: currency.USD},
		{s: "R$", expected: currency.BRL},
		{s: "JP¥", expected: currency.JPY},
		{s: "CN¥", expected: currency.CNY},
		{s: "€", expected: currency.EUR},
	}

	for _, v := range items {
		res, err := CurrencyParse(v.s)
		if err != nil {
			t.Error(err)
			return
		}
		if res != v.expected {
			t.Errorf("Expected %s, but get %s", v.expected.String(), res.String())
		}
	}
}
