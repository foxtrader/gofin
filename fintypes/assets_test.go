package fintypes

import (
	"fmt"
	"golang.org/x/text/currency"
	"testing"
)

func TestEqual(t *testing.T) {
	if BTC == USD {
		t.Errorf("BTC != USD")
		return
	}
}

func TestAsset_String(t *testing.T) {
	fmt.Println(currency.USD.String())
	fmt.Println(USD.Symbol())
}

func TestAsset_ToCoin(t *testing.T) {
	type testItem struct {
		s                Asset
		expectedPlatform Platform
		expectedName     string
		expectedSymbol   string
		expectedRet      bool
	}
	var testItems []testItem
	testItems = append(testItems,
		testItem{"c.Bitcoin.BTC@open", PlatformOpen, "Bitcoin", "BTC", true},
		testItem{"b.BTC@binance", Binance, "", "BTC", true},
		testItem{"c.Bitcoin.BTC@huobi", Huobi, "Bitcoin", "BTC", true},
		testItem{"c.Force-protocol.FOR@huobi", Huobi, "Force-protocol", "FOR", true},
		testItem{"binance.XRP/EUR", PlatformUnknown, "", "", false},
	)

	for _, v := range testItems {
		plt, name, symbol, ok := Asset(v.s).ToCoin()
		if plt != v.expectedPlatform || name != v.expectedName || symbol != v.expectedSymbol || ok != v.expectedRet {
			t.Errorf("ToCoin(%s) error", v.s)
			fmt.Println(fmt.Sprintf("get output:%s,%s,%s,%t.", plt, name, symbol, ok))
			return
		}
		fmt.Println(plt, name, symbol, ok)
	}
}

func TestAsset_ToStock(t *testing.T) {
	type testItem struct {
		s                Asset
		expectedPlatform Platform
		expectedSymbol   string
		expectedRet      bool
	}
	var testItems []testItem
	testItems = append(testItems,
		testItem{"s.AAPL@nasdaq", Nasdaq, "AAPL", true},
		testItem{"s.T@amex", Amex, "T", true},
		testItem{"s.anything", PlatformUnknown, "", false},
	)

	for _, v := range testItems {
		plt, symbol, ok := Asset(v.s).ToStock()
		if plt != v.expectedPlatform || symbol != v.expectedSymbol || ok != v.expectedRet {
			t.Errorf("ToStock(%s) error oputput(%s, %s, %t)", v.s, plt.String(), symbol, ok)
			return
		}
		fmt.Println(plt, symbol, ok)
	}
}

func TestNewAsset(t *testing.T) {
	type testItem struct {
		s             string
		expectedAsset Asset
		expectedRet   bool
	}
	var testItems []testItem
	testItems = append(testItems,
		testItem{"s.AAPL@nasdaq", NewStock("AAPL", "nasdaq"), true},
		testItem{"s.T@amex", NewStock("T", Amex), true},
		testItem{"c.Bitcoin.BTC@open", NewCoin("Bitcoin", "BTC", PlatformOpen), true},
	)

	for _, v := range testItems {
		asset, err := NewAsset(v.s)
		if asset != v.expectedAsset {
			t.Errorf("NewAsset(%s) error oputput(%s, %s)", v.s, asset.String(), err)
			return
		}
		fmt.Println(asset.String(), err)
	}
}

/*
func TestAsset_FiatInDefaultPair(t *testing.T) {
	type testItem struct {
		s            Asset
		expectedFiat Asset
	}
	var testItems []testItem
	testItems = append(testItems,
		testItem{"s.AAPL@nasdaq", USD},
		testItem{"s.T@amex", USD},
		testItem{"c.Bitcoin.BTC@cc", USD},
	)

	for _, v := range testItems {
		fiat, err := v.s.FiatInDefaultPair()
		if fiat != v.expectedFiat {
			t.Errorf("FiatInDefaultPair(%s) error oputput(%s, %s)", v.s, fiat.String(), err)
			return
		}
		fmt.Println(fiat.String(), err)
	}
}
*/
