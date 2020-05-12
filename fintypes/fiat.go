package fintypes

import (
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"golang.org/x/text/currency"
	"strings"
)

// Support ISO currency name and symbol
// 注意，有些常见的Symbol是简化的不完整版本，会在多个法币的完整Symbol中出现，比如¥和$
func ParseFiat(s string) (Asset, error) {
	defErr := errors.Errorf("invalid fiat(%s)", s)

	if unit, err := currency.ParseISO(s); err == nil {
		return newFiat(unit.String())
	}

	s = strings.TrimSpace(s)
	for _, v := range AllFiats() {
		if gstring.EqualLower(s, v.Name()) || gstring.EqualLower(s, v.Symbol()) {
			return v, nil
		}
	}

	return AssetNil, defErr
}
