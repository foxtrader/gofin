package fintypes

import (
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gstring"
)

type (
	// 币安的永续合约说明，可以辅助完善PairInfo
	// https://binance.zendesk.com/hc/zh-cn/articles/360033161972-%E5%90%88%E7%BA%A6%E7%BB%86%E5%88%99
	//MinNotional decimals.Decimal // notional = price * amount
	PairInfo struct {
		// spot only
		Enabled        bool
		UnitPrecision  int              // precision of unit amount? NO, seems useless
		QuotePrecision int              // precision of quote price? NO, seems useless
		UnitMin        gdecimal.Decimal // unit min trade amount, mostly it is the same as LostStep
		UnitStep       gdecimal.Decimal // unit min step amount, min trade amount in unit, same as amount in SetPosition()
		QuoteStep      gdecimal.Decimal // quote min movement

		// shared
		//MinLeverage    int // 最小杠杆倍数
		//MaxLeverage    int // 最大杠杆倍数
		MarginIsolatedEnabled bool
		MarginCrossEnabled    bool

		// contract only
		MaintMarginPercent    gdecimal.Decimal
		RequiredMarginPercent gdecimal.Decimal
	}

	MarketInfo struct {
		Infos map[PairM]PairInfo
	}
)

// all trading pairs, if some pair exist in different markets(spot,margin...), only one be kept
func (mi *MarketInfo) Pairs() []Pair {
	var res []Pair
	resMap := map[Pair]bool{}
	for pe := range mi.Infos {
		resMap[pe.Pair()] = true
	}
	for k := range resMap {
		res = append(res, k)
	}
	return res
}

func (mi *MarketInfo) PairMs() []PairM {
	var res []PairM
	for pe := range mi.Infos {
		res = append(res, pe)
	}

	return res
}

func (mi *MarketInfo) PairIMPs(period Period, platform Platform) []PairIMP {
	var res []PairIMP
	for pm := range mi.Infos {
		res = append(res, pm.SetI(period).SetP(platform))
	}
	return res
}

func (mi *MarketInfo) PairsIncludeFilter(includeSymbols []string, excludeSymbols []string) []Pair {
	includeSymbols = gstring.ToUpper(includeSymbols)
	excludeSymbols = gstring.ToUpper(excludeSymbols)

	var all []Pair
	for _, p := range mi.Pairs() {
		all = append(all, p)
	}

	var res []Pair
	for _, p := range all {
		if len(includeSymbols) > 0 {
			if !gstring.Contains(includeSymbols, p.Quote()) && !gstring.Contains(includeSymbols, p.Unit()) {
				continue
			}
		}

		if len(excludeSymbols) > 0 {
			if gstring.Contains(excludeSymbols, p.Quote()) || gstring.Contains(excludeSymbols, p.Unit()) {
				continue
			}
		}

		res = append(res, p)
	}

	return res
}

func (mi *MarketInfo) PairsAllowedFilter(allowedSymbols []string) []Pair {
	allowedSymbols = gstring.ToUpper(allowedSymbols)
	if len(allowedSymbols) == 0 {
		return nil
	}

	var all []Pair
	for _, p := range mi.Pairs() {
		all = append(all, p)
	}

	var res []Pair
	for _, p := range all {
		if gstring.Contains(allowedSymbols, p.Quote()) && gstring.Contains(allowedSymbols, p.Unit()) {
			res = append(res, p)
		}
	}

	return res
}

func (mi *MarketInfo) AvailableSymbols() []string {
	var res []string
	tmp := map[string]bool{}
	for _, pe := range mi.PairMs() {
		tmp[pe.Pair().Unit()] = true
		tmp[pe.Pair().Quote()] = true
	}
	for symbol := range tmp {
		res = append(res, symbol)
	}
	return res
}

func (mi *MarketInfo) SupportMargin(pair Pair) bool {
	info, ok := mi.Infos[pair.SetM(MarketSpot)]
	if !ok {
		return false
	}
	return ok && info.Enabled && (info.MarginIsolatedEnabled || info.MarginCrossEnabled)
}

func (mi *MarketInfo) Verify() error {
	for pe := range mi.Infos {
		if err := pe.Verify(); err != nil {
			return gerror.Errorf("invalid PairM(%s)", pe.String())
		}
	}
	return nil
}
