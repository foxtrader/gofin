package fintypes

import (
	"fmt"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"testing"
	"time"
)

func newTestKlineTA1() *Kline {
	r := new(Kline)

	date, _ := gtime.NewDate(2019, 1, 1)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 2)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 3)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 4)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 5)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 6)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 7)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 8)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 9)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})
	date, _ = gtime.NewDate(2019, 1, 10)
	r.Items = append(r.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), Indicators: make(map[string]float64)})

	/*

		6 *         -
		5   *     -   -     *
		4     * -       - *
		3     - *       * -
		2   -     *   *     -
		1 -         *
		  1 2 3 4 5 6 7 8 9 10

		-  expr1
		*  expr2

	*/
	r.Items[0].Indicators["expr1"] = 1
	r.Items[1].Indicators["expr1"] = 2
	r.Items[2].Indicators["expr1"] = 3
	r.Items[3].Indicators["expr1"] = 4
	r.Items[4].Indicators["expr1"] = 5
	r.Items[5].Indicators["expr1"] = 6
	r.Items[6].Indicators["expr1"] = 5
	r.Items[7].Indicators["expr1"] = 4
	r.Items[8].Indicators["expr1"] = 3
	r.Items[9].Indicators["expr1"] = 2

	r.Items[0].Indicators["expr2"] = 6
	r.Items[1].Indicators["expr2"] = 5
	r.Items[2].Indicators["expr2"] = 4
	r.Items[3].Indicators["expr2"] = 3
	r.Items[4].Indicators["expr2"] = 2
	r.Items[5].Indicators["expr2"] = 1
	r.Items[6].Indicators["expr2"] = 2
	r.Items[7].Indicators["expr2"] = 3
	r.Items[8].Indicators["expr2"] = 4
	r.Items[9].Indicators["expr2"] = 5

	return r
}

func TestTA_Crosses(t *testing.T) {
	r := newTestKlineTA1()
	forms, err := r.TA().Crosses("expr1", "expr2", -1)
	if err != nil {
		t.Error(err)
		return
	}
	if len(forms) != 2 {
		t.Errorf("Crosses test error1")
		return
	}
	if forms[0].Time.UTC().String() != "2019-01-04 00:00:00 +0000 UTC" || forms[0].Direction != TADirectionDown {
		t.Errorf("Crosses test error2")
		return
	}
	if forms[1].Time.UTC().String() != "2019-01-09 00:00:00 +0000 UTC" || forms[1].Direction != TADirectionUp {
		t.Errorf("Crosses test error3")
		return
	}
}

func TestTA_ReverseCrosses(t *testing.T) {
	r := newTestKlineTA1()
	forms, err := r.TA().ReverseCrosses("expr1", "expr2", -1)
	if err != nil {
		t.Error(err)
		return
	}
	if len(forms) != 2 {
		t.Errorf("ReverseCrosses test error1")
		return
	}
	if forms[0].Time.UTC().String() != "2019-01-04 00:00:00 +0000 UTC" || forms[0].Direction != TADirectionDown {
		t.Errorf("ReverseCrosses test error2")
		return
	}
	if forms[1].Time.UTC().String() != "2019-01-09 00:00:00 +0000 UTC" || forms[1].Direction != TADirectionUp {
		t.Errorf("ReverseCrosses test error3")
		return
	}
}

/*
func TestTA_IndicatorsLastCrossAtTail(t *testing.T) {
	r := newTestKlineTA1()
	cross, found, err := r.KTA().IndicatorsLastCrossAtTail("expr1", "expr2", 4)
	if err != nil {
		t.Error(err)
		return
	}
	if !found || cross == nil {
		t.Errorf("IndicatorsLastCrossAtTail test error1")
		return
	}
	if cross.T.UTC().String() != "2019-01-09 00:00:00 +0000 UTC" || cross.Direction != TADirectionUp {
		t.Errorf("IndicatorsLastCrossAtTail test error2")
		return
	}

	r, _ = newTestKlineMaotai()
	r = r.SliceTail(20)
	err = r.UpdateIndicators("EMA(4)", "EMA(3)")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r.IndicatorValues("EMA(4)"))
	fmt.Println(r.IndicatorValues("EMA(3)"))
	crosses, err := r.KTA().Crosses("EMA(4)", "EMA(3)", nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(crosses)
}*/

func TestKDot_Form(t *testing.T) {
	dot := Bar{O: gdecimal.NewFromFloat64(7154.59), L: gdecimal.NewFromFloat64(7150), H: gdecimal.NewFromFloat64(7485), C: gdecimal.NewFromFloat64(7389)}
	fmt.Println(dot.Form())

	dot = Bar{O: gdecimal.NewFromFloat64(8502.87), L: gdecimal.NewFromFloat64(8060), H: gdecimal.NewFromFloat64(8503.52), C: gdecimal.NewFromFloat64(8187.17)}
	fmt.Println(dot.Form())

	dot = Bar{O: gdecimal.NewFromFloat64(8490.74), L: gdecimal.NewFromFloat64(8350.68), H: gdecimal.NewFromFloat64(8635), C: gdecimal.NewFromFloat64(8502.4)}
	fmt.Println(dot.Form())

	dot = Bar{L: gdecimal.NewFromInt(10), H: gdecimal.NewFromInt(20), O: gdecimal.NewFromInt(18), C: gdecimal.NewFromInt(12)}
	fmt.Println(dot.Form())

	dot = Bar{L: gdecimal.NewFromInt(10), H: gdecimal.NewFromInt(20), O: gdecimal.NewFromInt(12), C: gdecimal.NewFromInt(18)}
	fmt.Println(dot.Form())
}
