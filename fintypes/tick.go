package fintypes

import (
	"encoding/json"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"strings"
	"time"
)

// In some exchange, it won't return H/L/V, even returns T & Last only, like binance
// They will be filled with -1
type (
	Tick struct {
		Time   time.Time        `json:"time"`
		Last   gdecimal.Decimal `json:"last"` // 最新成交价, binance has Last only
		Buy    gdecimal.Decimal `json:"buy"`  // 买一价
		Sell   gdecimal.Decimal `json:"sell"` // 卖一价
		High   gdecimal.Decimal `json:"high"` // 最高价
		Low    gdecimal.Decimal `json:"low"`  // 最低价
		Volume gdecimal.Decimal `json:"vol"`  // 最近的24小时成交量
	}

	Ticks struct {
		Items map[PairM]gdecimal.Decimal
	}
)

func (t Tick) String() string {
	buf, err := json.Marshal(t)
	if err != nil {
		return ""
	}
	return string(buf)
}

func TicksToPrices(ticks map[PairM]Tick) Ticks {
	r := Ticks{Items: map[PairM]gdecimal.Decimal{}}
	for pair, tick := range ticks {
		r.Items[pair] = tick.Last
	}
	return r
}

// unit: 查询哪个资产的法币价格
func (t Ticks) GetUSDPrice(market Market, unit string) (gdecimal.Decimal, error) {
	// "USDT/USD", USDT itself
	// FIXME 需要完善一下
	if strings.ToUpper(unit) == "USDT" {
		return gdecimal.One, nil
	}

	var usdCoins []string
	for _, usdCoin := range StableCoinsByFiat(USD) {
		usdCoins = append(usdCoins, usdCoin.Symbol())
	}
	usdCoins = append(usdCoins, "USD")

	// "***/USDT"...
	for _, usdCoin := range usdCoins {
		p := NewPair(unit, usdCoin)
		price, ok := t.Items[p.SetM(market)]
		if ok {
			return price, nil
		}
	}

	// "***/BTC"...
	for _, quoteCoin := range AllQuoteCoins() {
		p := NewPair(unit, quoteCoin.Symbol())
		unitQuotePrice, ok := t.Items[p.SetM(market)]
		if ok {
			for _, usdCoin := range usdCoins {
				p2 := NewPair(quoteCoin.Symbol(), usdCoin)
				quoteUsdPrice, ok2 := t.Items[p2.SetM(market)]
				if ok2 {
					return unitQuotePrice.Mul(quoteUsdPrice), nil
				}
			}
		}
	}

	return gdecimal.Zero, gerror.Errorf("can't get USD balance for %s", unit)
}
