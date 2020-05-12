package ta

import "github.com/markcheno/go-talib"

/*
https://github.com/haibeicode/indicator/blob/fea53b9ec084568fce8937166acf488e90e74c13/indicator/base.py
https://github.com/haibeicode/indicator/blob/58d9331a87c8046756a6b01a82a30df675915595/indicator/technology/countertrend.py
*/
// FIXME 需要权威的实现或者测试用例
func SKDJ(inLow, inHigh, inClose []float64, N, M int) (skdjRsv, skdjK, skdjD []float64) {
	lowV := LLV(inLow, N)
	highV := HHV(inHigh, N)
	var rsv []float64
	for i := 0; i < len(inClose); i++ {
		if i < N-1 { // when i < N - 1, highV[i] == 0, and lowV[i] == 0
			rsv = append(rsv, 0)
			continue
		}
		v := ((inClose[i] - lowV[i]) / (highV[i] - lowV[i])) * 100
		rsv = append(rsv, v)
	}

	// fixme 从这里往上，和python版的执行结果一致（区别仅仅是rsv头部几个值是0还是NaN），往下Ema执行之后就不一样了

	skdjRsv = talib.Ema(rsv, M)
	skdjK = talib.Ema(skdjRsv, M)
	skdjD = talib.Sma(skdjK, M)
	return skdjRsv, skdjK, skdjD
}
