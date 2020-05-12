package ta

/*
func newIndicatorTestKline() *frame.Kline {
	beginTime := gtime.Sub(gtime.Today(time.UTC).ToTimeUTC(), gtime.Day*30)
	date := gtime.TimeToDate(beginTime, time.UTC)
	open := 1.1
	low := 0.7
	high := 1.5
	close := 1.3
	vol := 100.0

	r := new(frame.Kline)

	for i := 0; i < 30; i++ {
		date = date.NextDay()
		item := frame.Bar{
			T: date.ToTimeUTC(),
			O: gdecimal.NewFromFloat64(open + float64(i)),
			L: gdecimal.NewFromFloat64(low + float64(i)),
			H: gdecimal.NewFromFloat64(high + float64(i)),
			C: gdecimal.NewFromFloat64(close + float64(i)),
			V: gdecimal.NewFromFloat64(vol + float64(i)),
		}
		r.Items = append(r.Items, item)
	}
	return r
}*/

/*
func TestLLV(t *testing.T) {
	k, err := data.NewTestKlineMaotai()
	if err != nil {
		t.Error(err)
		return
	}
	llv := LLV(k.LowValues(), 9)
	if llv[8] != 337.53 ||
		llv[9] != 339.968 ||
		llv[10] != 343.435 ||
		llv[11] != 343.435 {
		t.Errorf("LLV failed")
		return
	}
}

func TestHHV(t *testing.T) {
	k, err := data.NewTestKlineMaotai()
	if err != nil {
		t.Error(err)
		return
	}
	hhv := HHV(k.HighValues(), 9)
	if hhv[8] != 356.004 ||
		hhv[9] != 356.004 ||
		hhv[10] != 356.004 ||
		hhv[11] != 356.004 ||
		hhv[12] != 358.296 {
		t.Errorf("HHV failed")
		return
	}
}



func TestBOLL(t *testing.T) {
	vps := newIndicatorTestKline()

	mid1, up1, low1, err := BOLL(vps.CloseValues(), 20, 2)
	if err != nil {
		t.Error(err)
		return
	}

	var ks []indicator.Kline
	for _, v := range vps.Items {
		item := indicator.Kline{
			Time:  v.T,
			Open:  v.O.Float64(),
			Low:   v.L.Float64(),
			High:  v.H.Float64(),
			Close: v.C.Float64(),
			Vol:   v.V.Float64(),
		}
		ks = append(ks, item)
	}
	boll := indicator.NewBoll(20, 2)
	mid2, up2, low2 := boll.Boll(ks)

	if !gnum.FloatsEqual(mid1, mid2) {
		t.Errorf("Mid doesn't match")
		return
	}
	if !gnum.FloatsEqual(up1, up2) {
		t.Errorf("Up doesn't match")
		return
	}
	if !gnum.FloatsEqual(low1, low2) {
		t.Errorf("L doesn't match")
		return
	}
}
*/
