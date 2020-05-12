package findata

import (
	"fmt"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/container/gjson"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDetail(t *testing.T) {
	cmc, err := NewAssetDataSource(fintypes.CoinMarketCap, "45c9ef34-1d6b-47af-8ac6-1b82549432dd", "")
	assert.Nil(t, err)

	res, err := cmc.GetDetails()
	assert.Nil(t, err)

	fmt.Println(gjson.MarshalStringDefault(res, true))
}
