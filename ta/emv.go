package ta

import "github.com/markcheno/go-talib"

func EMV(inHigh, inLow, inVolume []float64, inTimePeriod int) []float64 {
	if len(inHigh) != len(inLow) || len(inHigh) != len(inVolume) {
		return nil
	}

	EMs := []float64{}
	EMs = append(EMs, 0)
	for i := 1; i < len(inHigh); i++ {
		A := (inHigh[i] + inLow[i]) / 2
		B := (inHigh[i-1] + inLow[i-1]) / 2
		C := inHigh[i] - inLow[i]
		EM := (A - B) * C / inVolume[i]
		EMs = append(EMs, EM)
	}
	return talib.Sum(EMs, inTimePeriod)
}
