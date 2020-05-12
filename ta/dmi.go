package ta

import "github.com/markcheno/go-talib"

// DMI指标包含4条线：PDI、MDI、ADX和ADXR
func DMI(inHigh, inLow, inClose []float64, inTimePeriod int) (PDI, MDI, ADX, ADXR []float64) {
	plusDI := talib.PlusDI(inHigh, inLow, inClose, inTimePeriod)
	minusDI := talib.MinusDI(inHigh, inLow, inClose, inTimePeriod)
	adx := talib.Adx(inHigh, inLow, inClose, inTimePeriod)
	adxr := talib.AdxR(inHigh, inLow, inClose, inTimePeriod)
	return plusDI, minusDI, adx, adxr
}
