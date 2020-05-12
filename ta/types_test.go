package ta

import (
	"fmt"
	"testing"
)

func TestIndicator_ParamCount(t *testing.T) {
	fmt.Println(IndicatorEMA.ParamCount())
}

func TestIndExpr_Parse(t *testing.T) {
	ind, params, err := IndExpr("EMA(10)").Parse()
	if err != nil || ind != IndicatorEMA || params[0].String() != "10" {
		t.Errorf("expr EMA(10) parse error")
		return
	}

	ind, params, err = IndExpr("").Parse()
	if err != nil || ind != IndicatorError {
		t.Errorf("expr null parse error")
		return
	}
}
