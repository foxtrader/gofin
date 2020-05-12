package fintypes

import "github.com/shawnwyckoff/gopkg/apputil/gerror"

var (
	allMarkets []Market

	MarketError  Market = ""
	MarketSpot          = enrollNewMarket("spot")   // 现货
	MarketFuture        = enrollNewMarket("future") // 交割合约
	MarketPerp          = enrollNewMarket("perp")   // perpetual swap （永续合约）

	//MarketMargin                = enrollNewMarket("margin")          // 现货杠杆
	//MarketIsolatedMargin        = enrollNewMarket("isolated-margin") // 逐仓现货杠杆
	//MarketCrossMargin           = enrollNewMarket("cross-margin")    // 全仓现货杠杆
	//MarketIsolatedFuture        = enrollNewMarket("isolated-future") // 逐仓交割合约
	//MarketCrossFuture           = enrollNewMarket("cross-future")    // 全仓交割合约
	//MarketIsolatedPerp          = enrollNewMarket("isolated-perp")   // 逐仓永续合约（perpetual swap）
	//MarketCrossPerp             = enrollNewMarket("cross-perp")      // 全仓永续合约（perpetual swap）
)

type (
	Market string
)

func enrollNewMarket(name string) Market {
	res := Market(name)
	allMarkets = append(allMarkets, res)
	return res
}

func (m Market) Verify() error {
	for _, v := range allMarkets {
		if m == v {
			return nil
		}
	}
	return gerror.Errorf("invalid Market(%s)", m)
}

/*
func (m Market) IsSpot() bool {
	return m == MarketSpot
}

func (m Market) IsMargin() bool {
	return m == MarketMargin || m == MarketIsolatedMargin || m == MarketCrossMargin
}

func (m Market) IsFuture() bool {
	return m == MarketFuture || m == MarketIsolatedFuture || m == MarketCrossFuture
}

func (m Market) IsPerp() bool {
	return m == MarketPerp || m == MarketIsolatedPerp || m == MarketCrossPerp
}

func (m Market) IsIsolated() bool {
	return m == MarketIsolatedMargin || m == MarketIsolatedFuture || m == MarketIsolatedPerp
}

func (m Market) IsCross() bool {
	return m == MarketCrossMargin || m == MarketCrossFuture || m == MarketCrossPerp
}*/

func (m Market) IsContract() bool {
	return m == MarketFuture || m == MarketPerp
}

func ParseMarket(s string) (Market, error) {
	if err := Market(s).Verify(); err != nil {
		return MarketError, err
	}
	return Market(s), nil
}
