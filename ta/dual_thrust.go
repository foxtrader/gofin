package ta

import "github.com/shawnwyckoff/gopkg/container/gnum"

/*
N日High的最高价HH
N日Close的最低价LC
N日Close的最高价HC
N日Low的最低价LL
Range = Max(HH-LC,HC-LL)
BuyLine = O + K1×Range
SellLine = O + K2×Range
*/
func DualThrust(inOpen, inHigh, inLow, inClose []float64, n int, k1, k2 float64) (buyLine, sellLine []float64) {
	HH := gnum.MaxSV(inHigh, n)
	LC := gnum.MinSV(inClose, n)
	HC := gnum.MaxSV(inClose, n)
	LL := gnum.MinSV(inLow, n)
	Range := gnum.MaxSS(gnum.SubSS(HH, LC), gnum.SubSS(HC, LL))
	K1Range := gnum.MulSV(Range, k1)
	K2Range := gnum.MulSV(Range, k2)
	buyLine = gnum.AddSS(inOpen, K1Range)
	sellLine = gnum.AddSS(inOpen, K2Range)
	gnum.SetHead0(buyLine, n-1)
	gnum.SetHead0(sellLine, n-1)
	return buyLine, sellLine
}
