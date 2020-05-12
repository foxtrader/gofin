package ta

import (
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"math"
)

func OSC(inOpen, inLow, inHigh, inClose []float64) ([]float64, error) {
	if len(inOpen) != len(inLow) || len(inLow) != len(inHigh) || len(inHigh) != len(inClose) {
		return nil, gerror.Errorf("input length not equal")
	}

	var r []float64
	idx := float64(0)

	for i := 0; i < len(inOpen); i++ {
		idx = (inHigh[i] - inLow[i]) / math.Abs(inOpen[i]-inClose[i])
		if i > 0 && inHigh[i] < inHigh[i-1] {
			idx = -idx
		}
		r = append(r, idx)
	}
	return r, nil
}
