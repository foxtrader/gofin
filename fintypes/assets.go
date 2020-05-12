package fintypes

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gnum"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"golang.org/x/text/currency"
	"strings"
)

/*

Assets Examples:

f.USD
m.XAU
i.DJI
s.AAPL@nasdaq
b.HOT@binance // coin defined by symBol
c.Bitcoin.BTC@cmc
c.Bitcoin.BTC@open
*/

type (
	AssetType string
	Asset     string
)

const (
	AssetTypeUnknown AssetType = "unknown"
	AssetTypeFiat    AssetType = "fiat"
	AssetTypeMetal   AssetType = "metal"
	AssetTypeStock   AssetType = "stock"
	AssetTypeIndex   AssetType = "index"
	AssetTypeCoin    AssetType = "coin"
)

var (
	AssetNil = Asset("")
)

func (a Asset) Type() AssetType {
	s := string(a)
	if len(s) > 2 {
		if gstring.StartWith(s, "f.") {
			return AssetTypeFiat
		} else if gstring.StartWith(s, "b.") {
			return AssetTypeCoin
		} else if gstring.StartWith(s, "c.") {
			return AssetTypeCoin
		} else if gstring.StartWith(s, "m.") {
			return AssetTypeMetal
		} else if gstring.StartWith(s, "s.") {
			return AssetTypeStock
		} else if gstring.StartWith(s, "i.") {
			return AssetTypeIndex
		}
	}
	return AssetTypeUnknown
}

func (a Asset) ToFiatUnit() (currency.Unit, bool) {
	if a.Type() == AssetTypeFiat {
		if unit, err := currency.ParseISO(string(a)[2:]); err != nil {
			return currency.Unit{}, false
		} else {
			return unit, true
		}
	}
	return currency.Unit{}, false
}

func (a Asset) ToFiat() (string, bool) {
	if a.Type() == AssetTypeFiat {
		if unit, err := currency.ParseISO(string(a)[2:]); err != nil {
			return "", false
		} else {
			return unit.String(), true
		}
	}
	return "", false
}

func (a Asset) ToMetal() (string, bool) {
	if a.Type() == AssetTypeMetal {
		return string(a)[2:], true
	} else {
		return "", false
	}
}

func (a Asset) ToIndex() (string, bool) {
	if a.Type() == AssetTypeIndex {
		return string(a)[2:], true
	} else {
		return "", false
	}
}

func (a Asset) ToStock() (exchange Platform, symbol string, ok bool) {
	if a.Type() == AssetTypeStock {
		ss := strings.Split(string(a)[2:], "@")
		if len(ss) == 2 {
			if plt, err := ParsePlatform(ss[1]); err != nil {
				return PlatformUnknown, "", false
			} else {
				return plt, ss[0], true
			}
		}
	}

	return PlatformUnknown, "", false
}

func (a Asset) ToCoin() (platform Platform, name, symbol string, ok bool) {
	if a.Type() == AssetTypeCoin {
		if ss := strings.Split(string(a)[2:], "@"); len(ss) == 2 {

			if gstring.StartWith(string(a), "b.") {
				if plt, err := ParsePlatform(ss[1]); err == nil {
					return plt, "", ss[0], true
				}
			}

			if gstring.StartWith(string(a), "c.") {
				if plt, err := ParsePlatform(ss[1]); err == nil {
					if ns := strings.Split(ss[0], "."); len(ns) == 2 {
						return plt, ns[0], ns[1], true
					}
				}
			}

		}
	}

	return PlatformUnknown, "", "", false
}

func (a Asset) Against(quote Asset) Pair {
	return Pair(a.TradeSymbol() + "/" + quote.TradeSymbol())
}

func newFiat(fiatName string) (Asset, error) {
	if f, err := CurrencyParse(fiatName); err != nil {
		return AssetNil, errors.Errorf(err.Error()+"invalid fiat name(%s)", fiatName)
	} else {
		return Asset("f." + strings.ToUpper(f.String())), nil
	}
}

func newFiatFromUnit(unit currency.Unit) Asset {
	return Asset("f." + strings.ToUpper(unit.String()))
}

func NewCoinWithSymbol(symbol string, plt Platform) Asset {
	if symbol == "" {
		return AssetNil
	}
	return Asset("b." + strings.ToUpper(symbol) + "@" + strings.ToLower(plt.String()))
}

func NewCoin(name, symbol string, plt Platform) Asset {
	if name == "" || symbol == "" {
		return AssetNil
	}
	return Asset("c." + gstring.OnlyFirstLetterUpperCase(name) + "." + strings.ToUpper(symbol) + "@" + strings.ToLower(plt.String()))
}

func newMetal(metalName string) Asset {
	return Asset("m." + strings.ToUpper(metalName))
}

func NewStock(symbol string, exchange Platform) Asset {
	// check symbol
	if strings.ToLower(exchange.String()) == strings.ToLower(Hkex.String()) {
		if len(symbol) == 4 {
			symbol = "0" + symbol
		}
		if len(symbol) != 5 {
			return AssetNil
		}
	}
	if strings.ToLower(exchange.String()) == strings.ToLower(Sse.String()) ||
		strings.ToLower(exchange.String()) == strings.ToLower(Szse.String()) {
		if len(symbol) != 6 {
			return AssetNil
		}
		if !gnum.IsDigit(symbol) {
			return AssetNil
		}
	}

	return Asset("s." + strings.ToUpper(symbol) + "@" + strings.ToLower(exchange.String()))
}

func newIndex(symbol string) Asset {
	return Asset("i." + strings.ToUpper(symbol))
}

// sdn: self desc name
func NewAsset(sdn string) (Asset, error) {
	defErr := errors.Errorf("invalid asset(%s)", sdn)

	if len(sdn) <= 2 {
		return AssetNil, defErr
	}

	if gstring.StartWith(sdn, "f.") {
		return newFiat(sdn[2:])
	} else if gstring.StartWith(sdn, "b.") {
		if ss := strings.Split(sdn[2:], "@"); len(ss) == 2 {
			plt, err := ParsePlatform(ss[1])
			if err != nil {
				return AssetNil, err
			}
			return NewCoinWithSymbol(ss[0], plt), nil
		}
	} else if gstring.StartWith(sdn, "c.") {
		if ss := strings.Split(sdn[2:], "@"); len(ss) == 2 {
			plt, err := ParsePlatform(ss[1])
			if err != nil {
				return AssetNil, err
			}
			if nameSymbol := strings.Split(ss[0], "."); len(nameSymbol) == 2 {
				return NewCoin(nameSymbol[0], nameSymbol[1], plt), nil
			}
		}
	} else if gstring.StartWith(sdn, "m.") {
		return newMetal(sdn[2:]), nil
	} else if gstring.StartWith(sdn, "s.") {
		ss := strings.Split(sdn[2:], "@")
		if len(ss) == 2 {
			ex, err := ParsePlatform(ss[1])
			if err != nil {
				return AssetNil, err
			}
			return NewStock(ss[0], ex), nil
		}
	} else if gstring.StartWith(sdn, "i.") {
		return newIndex(sdn[2:]), nil
	}
	return AssetNil, defErr
}

// can't find name from stock type Asset
func (a Asset) Name() string {
	if a.Type() == AssetTypeFiat {
		if name, ok := a.ToFiat(); ok {
			return strings.ToLower(name)
		}
	} else if a.Type() == AssetTypeCoin {
		if _, name, _, ok := a.ToCoin(); ok {
			return strings.ToLower(name)
		}
	} else if a.Type() == AssetTypeMetal {
		if name, ok := a.ToMetal(); ok {
			return strings.ToLower(name)
		}
	}
	return ""
}

func (a Asset) Symbol() string {
	if a.Type() == AssetTypeFiat {
		if c, ok := a.ToFiatUnit(); ok {
			return CurrencyGetSymbol(c)
		}
	} else if a.Type() == AssetTypeCoin {
		if _, _, symbol, ok := a.ToCoin(); ok {
			return strings.ToUpper(symbol)
		}
	} else if a.Type() == AssetTypeMetal {
		if a.TradeSymbol() == "XAU" {
			return CurrencyGetSymbol(currency.XAU)
		} else if a.TradeSymbol() == "XAG" {
			return CurrencyGetSymbol(currency.XAG)
		} else if a.TradeSymbol() == "XPT" {
			return CurrencyGetSymbol(currency.XPT)
		} else if a.TradeSymbol() == "XPD" {
			return CurrencyGetSymbol(currency.XPD)
		}
	} else if a.Type() == AssetTypeStock {
		if _, sym, ok := a.ToStock(); ok {
			return sym
		}
	} else if a.Type() == AssetTypeIndex {
		if sym, ok := a.ToIndex(); ok {
			return sym
		}
	}
	return ""
}

// Trade symbol of assets, in fiat Trade Symbol is Name but not symbol
func (a Asset) TradeSymbol() string {
	if a.Type() == AssetTypeFiat {
		return strings.ToUpper(a.Name())
	}
	return a.Symbol()
}

func (a Asset) String() string {
	return string(a)
}

func (a Asset) Equal(cmp Asset) bool {
	return string(a) == string(cmp)
}

func (a Asset) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", a.String())), nil
}

func (a *Asset) UnmarshalJSON(b []byte) error {
	// Init
	*a = AssetNil

	// Remove '"'
	s := string(b)
	defErr := errors.Errorf("invalid json asset '%s'", s)
	if len(s) <= 1 {
		return defErr
	}
	if s[0] != '"' || s[len(s)-1] != '"' {
		return defErr
	}
	s = gstring.RemoveHead(s, 1)
	s = gstring.RemoveTail(s, 1)

	// Parse
	if asset, err := NewAsset(s); err != nil {
		return err
	} else {
		*a = asset
		return nil
	}
}

/*
func (a *Asset) FiatInDefaultPair() (Asset, error) {
	at := a.Type()
	ex := PLTUnknown
	ok := true
	defErr := errors.Errorf("can't get fiat in default pair of Asset(%s)", a.String())

	if at == AssetTypeStock {
		ex, _, ok = a.ToStock()
		if !ok {
			return AssetNil, defErr
		}
	} else if at == AssetTypeCoin {
		ex, _, _, ok = a.ToCoin()
		if !ok {
			return AssetNil, defErr
		}
	} else {
		return AssetNil, defErr
	}

	fiat := ex.Info().DefaultQuoteFiat
	if fiat != AssetNil {
		return fiat, nil
	}
	return AssetNil, defErr
}
*/
