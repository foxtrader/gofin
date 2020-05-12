package fintypes

import (
	"fmt"
	"testing"
)

func TestMarketInfo_PairIMPs(t *testing.T) {
	mi := MarketInfo{Infos: map[PairM]PairInfo{}}
	mi.Infos[PairM("ETH/USDT.spot")] = PairInfo{}
	pes := mi.PairIMPs(Period1Min, Binance)
	fmt.Println(pes)
}
