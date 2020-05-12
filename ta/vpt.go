package ta

// https://en.wikipedia.org/wiki/Volume%E2%80%93price_trend
func VPT(inClose, inVolume []float64) []float64 {
	if len(inClose) != len(inVolume) {
		return nil
	}
	r := []float64{}
	r = append(r, 0)

	vpt := float64(0)
	for i := 1; i < len(inClose); i++ {
		lastClose := inClose[i-1]
		vpt = vpt + (inVolume[i] * (inClose[i] - lastClose) / lastClose)
		r = append(r, vpt)
	}
	return r
}
