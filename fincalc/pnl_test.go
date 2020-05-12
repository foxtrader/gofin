package fincalc

import (
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/apputil/gtest"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"strconv"
	"testing"
	"time"
)

func TestPNL_SharpRatio(t *testing.T) {
	nv := NewCalc(fintypes.Period1Day)
	// Account equity for 3 consecutive days, from https://www.zhihu.com/question/27264526
	nv.Add(time.Date(2019, 9, 2, 10, 0, 0, 0, gtime.TimeZoneAsiaShanghai), 1000000)
	nv.Add(time.Date(2019, 9, 3, 10, 0, 0, 0, gtime.TimeZoneAsiaShanghai), 1)
	nv.Add(time.Date(2019, 9, 4, 10, 0, 0, 0, gtime.TimeZoneAsiaShanghai), 2)
	nv.Add(time.Date(2019, 9, 5, 10, 0, 0, 0, gtime.TimeZoneAsiaShanghai), 3)
	daySharpe := nv.SharpeRatio(fintypes.Period1Day, 0.04, false)
	if s := strconv.FormatFloat(daySharpe, 'f', 7, 64); s != "0.1599778" {
		gtest.PrintlnExit(t, "should be 0.1599778, but %s got", s)
	}
}
