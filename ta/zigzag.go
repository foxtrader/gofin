package ta

import (
	"github.com/shawnwyckoff/gopkg/container/gternary"
	"math"
)

// FIXME: error implement
// 获取所有转折点，非转折点填0，-1表示下拐，1表示上扬
// minTrendLen = 2, minShockScale = 0.03: 至少2根线，且累计震幅达到最小要求3%，才能算形成趋势，否则只能算趋势上的震荡
// return dots: -1 down knee point, 0 in the trend of last dot, 1 up knee point
func ZigZagPoints(inOpen, inClose []float64, minTrendLen int, minShockScale float64) []float64 {
	if len(inOpen) != len(inClose) {
		return nil
	}

	var shock []float64 // K线当天的震幅，包含了阴阳信息
	for i := 0; i < len(inOpen); i++ {
		shock = append(shock, (inClose[i]-inOpen[i])/inOpen[i])
	}

	sameSignTotalShock := func(s []float64, reversFrom int) (begin int, totalShock float64) {
		isMinus := s[reversFrom] < 0
		totalShock = s[reversFrom]
		for i := reversFrom - 1; i >= 0; i-- {
			if s[i] < 0 != isMinus {
				return i + 1, totalShock
			}
			totalShock += s[i]
		}
		return 0, totalShock
	}

	r := make([]float64, len(inOpen))
	for i := minTrendLen - 1; i < len(shock); i++ {
		begin, totalShock := sameSignTotalShock(shock, i)
		if i-begin+1 < minTrendLen { // 线根数太少不足以形成趋势
			//for k := begin; k <= i; k ++ {
			//	r[k] = -2
			//}
			continue
		}
		if math.Abs(totalShock) < math.Abs(minShockScale) {
			continue
		}

		r[begin] = gternary.If(totalShock > 0).Float64(1, -1)
		//for j := begin+1; j <= i; j++ {
		//	r[j] = 0
		//}
	}

	// 抹掉一些重复相连的点。之所以出现这样的点，是因为中间有少数符号相反且震幅极小的点。
	lastKneePoint := func(vals []float64, reverseFrom int) float64 {
		for i := reverseFrom - 1; i >= 0; i-- {
			if vals[i] != 0 {
				return vals[i]
			}
		}
		return 0
	}
	for i := 0; i < len(r); i++ {
		if lv := lastKneePoint(r, i); lv == r[i] {
			r[i] = 0
		}
	}

	return r
}

// 获取所有转折点的价格，非转折点填0
func ZIGZAG(inOpen, inClose []float64, minTrendLen int, minShockScale float64) []float64 {
	zigzag := ZigZagPoints(inOpen, inClose, minTrendLen, minShockScale)

	r := make([]float64, len(inOpen))
	for i := 0; i < len(inClose); i++ {
		if zigzag[i] == -1 || zigzag[i] == 1 {
			r[i] = inClose[i]
		}
	}
	return r
}

// get plus/minus price from ZIGZAG
// use negative price to replace -1
func getPmPriceWithZigZag(inPrice []float64, zigzag []float64, lastCount int) []float64 {
	var r []float64

	for i := len(inPrice) - 1; i >= 0; i-- {
		if len(r) >= lastCount {
			break
		}
		if zigzag[i] == -1 {
			r = append(r, 0-inPrice[i])
		}
		if zigzag[i] == 1 {
			r = append(r, inPrice[i])
		}
	}

	reversePut := func(in []float64) []float64 {
		var out []float64
		for idx := len(in) - 1; idx >= 0; idx-- {
			out = append(out, in[idx])
		}
		return out
	}
	return reversePut(r)
}

// 从尾巴上识别zigzag形态
// return M_TOP / W_BOTTOM / UNKNOWN
func ZigZagLastForm(inOpen, inClose []float64, minTrendLen int, minShockScale float64) string {
	zigzag := ZIGZAG(inOpen, inClose, minTrendLen, minShockScale)
	prices := getPmPriceWithZigZag(inClose, zigzag, 4)
	if len(prices) < 4 {
		return "UNKNOWN"
	}
	if prices[0] > 0 && prices[1] < 0 && prices[2] > 0 && prices[3] < 0 && prices[3] < prices[1] {
		return "M_TOP"
	}
	if prices[0] < 0 && prices[1] > 0 && prices[2] < 0 && prices[3] > 0 && math.Abs(prices[3]) > math.Abs(prices[1]) {
		return "W_BOTTOM"
	}
	return "UNKNOWN"
}
