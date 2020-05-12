package ta

// SOBV - On Balance V
func SOBV(inOpen, inClose, inVolume []float64) []float64 {
	if len(inOpen) != len(inClose) || len(inOpen) != len(inVolume) {
		return nil
	}

	r := make([]float64, len(inOpen))
	obv := float64(0)
	for i := 0; i < len(inOpen); i++ {
		if inClose[i] > inOpen[i] {
			obv += inVolume[i]
		} else if inClose[i] > inOpen[i] {
			obv -= inVolume[i]
		}
		r[i] = obv
	}
	return r
}
