package ta

import (
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gnum"
)

func KDJ(inLow, inHigh, inClose []float64, inN1, inN2, inN3 int) (k, d, j []float64, err error) {
	if len(inLow) == 0 {
		return nil, nil, nil, gerror.Errorf("input length is zero")
	}
	if len(inLow) != len(inHigh) || len(inHigh) != len(inClose) {
		return nil, nil, nil, gerror.Errorf("input length not equal")
	}

	sma := func(x []float64, n float64) (r []float64) {
		r = make([]float64, len(x))
		for i := 0; i < len(x); i++ {
			if i == 0 {
				r[i] = x[i]
			} else {
				r[i] = (1.0*x[i] + (n-1.0)*r[i-1]) / n
			}
		}
		return
	}

	ln := len(inLow)
	rsv := make([]float64, ln)
	j = make([]float64, ln)
	rsv[0] = 50.0
	for i := 1; i < ln; i++ {
		m := i + 1 - inN1
		if m < 0 {
			m = 0
		}
		h := gnum.MaxFloat(inHigh[m : i+1]...)
		l := gnum.MinFloat(inLow[m : i+1]...)
		rsv[i] = (inClose[i] - l) * 100.0 / (h - l)
		rsv[i] = rsv[i]
	}
	k = sma(rsv, float64(inN2))
	d = sma(k, float64(inN3))
	for i := 0; i < ln; i++ {
		j[i] = 3.0*k[i] - 2.0*d[i]
	}
	return
}
