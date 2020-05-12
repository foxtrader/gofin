package ta

import (
	"github.com/shawnwyckoff/gopkg/container/gnum"
	"math"
)

// TODO test required
// Donchian Channel
func DC(inLow, inHigh []float64, nPeriod int) (outLow, outHigh []float64) {
	//fmt.Println("DC", nPeriod)
	if len(inLow) != len(inHigh) {
		return nil, nil
	}
	if len(inLow) == 0 {
		return nil, nil
	}

	outLow = make([]float64, len(inLow))
	outHigh = make([]float64, len(inHigh))
	for i := 0; i < len(inLow); i++ {
		if i < nPeriod-1 {
			outLow[i] = math.NaN()
			outHigh[i] = math.NaN()
		} else {
			outLow[i] = gnum.MinFloat(inLow[i-nPeriod+1 : i+1]...)
			outHigh[i] = gnum.MaxFloat(inHigh[i-nPeriod+1 : i+1]...)
		}
	}
	//fmt.Println(inLow)
	//fmt.Println(outLow)
	//fmt.Println(inHigh)
	//fmt.Println(outHigh)
	return outLow, outHigh
}
