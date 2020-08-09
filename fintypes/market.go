package fintypes

import "github.com/shawnwyckoff/gopkg/apputil/gerror"

var (
	allMarkets []Market

	MarketError  Market = ""
	MarketSpot          = enrollNewMarket("spot")   // 现货
	MarketFuture        = enrollNewMarket("future") // 交割合约
	MarketPerp          = enrollNewMarket("perp")   // perpetual swap （永续合约）
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

func (m Market) IsContract() bool {
	return m == MarketFuture || m == MarketPerp
}

func ParseMarket(s string) (Market, error) {
	if err := Market(s).Verify(); err != nil {
		return MarketError, err
	}
	return Market(s), nil
}
