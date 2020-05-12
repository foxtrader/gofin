package ta

import (
	"github.com/markcheno/go-talib"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gnum"
	"math"
)

// OBV计算方法：
// 主公式：当日OBV=前一日OBV+今日成交量
// 1.基期OBV值为0，即该股上市的第一天，OBV值为0
// 2.若当日收盘价＞上日收盘价，则当日OBV=前一日OBV＋今日成交量
// 3.若当日收盘价＜上日收盘价，则当日OBV=前一日OBV－今日成交量
// 4.若当日收盘价＝上日收盘价，则当日OBV=前一日OBV

/*
SOBV基本同OBV一样，不同的地方是决定当天成交量，是属于多方的能量还是属于空方能量不是依据收盘价，而是以当天的K线的阴阳决定。
（1）如果今天是阳线，则今天的成交量是属于多方的能量。相当于今天收盘价≥昨天
收盘价。
（2）如果今天是阴线，则今天的成交量属于空方的能量。相当于今收盘<昨收盘。
*/

/*
https://www.investopedia.com/terms/o/onbalancevolume.asp

Calculating OBV
On-balance volume provides a running total of an asset's trading volume and indicates whether this volume is flowing in or out of a given security or currency pair. The OBV is a cumulative total of volume (positive and negative). There are three rules implemented when calculating the OBV. They are:

1. If today's closing price is higher than yesterday's closing price, then: Current OBV = Previous OBV + today's volume

2. If today's closing price is lower than yesterday's closing price, then: Current OBV = Previous OBV - today's volume

3. If today's closing price equals yesterday's closing price, then: Current OBV = Previous OBV
*/

// Lowest value over a period
func LLV(inReal []float64, inTimePeriod int) []float64 {
	return talib.Min(inReal, inTimePeriod)
}

// Highest value over a period
func HHV(inReal []float64, inTimePeriod int) []float64 {
	return talib.Max(inReal, inTimePeriod)
}

func boll_sma(inClose []float64) float64 {
	s := len(inClose)
	var sum float64 = 0
	for i := 0; i < s; i++ {
		sum += float64(inClose[i])
	}
	return sum / float64(s)
}
func boll_dma(inClose []float64, ma float64, inN int) float64 {
	s := len(inClose)
	var sum float64 = 0
	for i := 0; i < s; i++ {
		sum += (inClose[i] - ma) * (inClose[i] - ma)
	}
	return math.Sqrt(sum / float64(inN-1))
}

func BOLL(inClose []float64, inN, inK int) (basis []float64, upper []float64, lower []float64, err error) {
	length := len(inClose)
	if length < inN {
		return nil, nil, nil, errors.Errorf("length is smaller than n")
	}

	basis = make([]float64, length)
	upper = make([]float64, length)
	lower = make([]float64, length)

	for i := length - 1; i > inN-1; i-- {
		ps := inClose[(i - inN + 1) : i+1]
		basis[i] = boll_sma(ps)
		dm := float64(inK) * boll_dma(ps, basis[i], inN)
		upper[i] = basis[i] + dm
		lower[i] = basis[i] - dm
	}
	return
}

func SMA(inReal []float64, inTimePeriod int) ([]float64, error) {
	if len(inReal) == 0 || inTimePeriod <= 0 || len(inReal) < inTimePeriod {
		return nil, errors.Errorf("invalid inReal length(%d) or inTimePeriod(%d)", len(inReal), inTimePeriod)
	}
	return talib.Sma(inReal, inTimePeriod), nil
}

func EMA(inReal []float64, inTimePeriod int) ([]float64, error) {
	if len(inReal) == 0 || inTimePeriod <= 0 || len(inReal) < inTimePeriod {
		return nil, errors.Errorf("invalid inReal length(%d) or inTimePeriod(%d)", len(inReal), inTimePeriod)
	}
	return talib.Ema(inReal, inTimePeriod), nil
}

func ATRK(inHigh []float64, inLow []float64, inClose []float64, inTimePeriod int, k float64) []float64 {
	atr := talib.Atr(inHigh, inLow, inClose, inTimePeriod)
	return gnum.MulSV(atr, k)
}

// https://www.zhihu.com/question/56068713/answer/903382369
// Average Daily Range (ADR)
func ADR() {
}

// 把成交量的影响加权方式算进去
func MAOCV(inOpen, inClose, inVolume []float64, inTimePeriod int) []float64 {
	return nil
}
