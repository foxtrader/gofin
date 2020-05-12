package findata

import (
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
)

type (
	AssetDataSource interface {
		GetDetails() ([]AssetDetail, error)
	}
)

func NewAssetDataSource(platform fintypes.Platform, apiKey, proxy string) (AssetDataSource, error) {
	switch platform {
	case fintypes.CoinGecko:
		return nil, nil
	case fintypes.CoinMarketCap:
		return &CmcClient{proxy: proxy, apiKey: apiKey}, nil
	default:
		return nil, gerror.Errorf("unsupported Platform %s", platform.String())
	}
}
