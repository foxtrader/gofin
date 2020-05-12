package fintypes

import "testing"

func TestFixNameSymbol(t *testing.T) {
	type item struct {
		src      string
		expected string
	}

	items := []item{
		{src: "globalboost-y", expected: "globalboost-y"},
		{src: "Alt.Estate token", expected: "alt-estate-token"},
		{src: "EtherDelta Token", expected: "etherdelta-token"},
		{src: "I/O Coin", expected: "i-o-coin"},
		{src: "Carboneum [C8] Token", expected: "carboneum-token"},
		{src: "BLOC.MONEY", expected: "bloc-money"},
		{src: "COMSA [XEM]", expected: "comsa"},
		{src: "  COMSA [XEM]", expected: "comsa"},
		{src: "  COMSA [XEM]  ", expected: "comsa"},
	}

	for _, v := range items {
		if got := FixCoinName(v.src); got != v.expected {
			t.Errorf("src (%s), expected (%s), but get (%s)", v.src, v.expected, got)
			return
		}
	}
}
