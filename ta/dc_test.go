package ta

import (
	"fmt"
	"testing"
)

func TestDC(t *testing.T) {
	inLow := []float64{100, 78, 56, 77, 89, 99, 100, 131, 150, 200, 300}
	inHigh := []float64{105, 83, 61, 82, 93, 104, 105, 136, 155, 205, 305}
	outLow, outHigh := DC(inLow, inHigh, 3)
	fmt.Println(outLow)
	fmt.Println(outHigh)
}
