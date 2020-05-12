package ta

/*
func TestKDJ(t *testing.T) {
	vps := newIndicatorTestKline()

	k1, d1, j1, err := KDJ(vps.LowValues(), vps.HighValues(), vps.CloseValues(), 10, 10, 20)
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
	kdj := indicator.NewKdj(10, 10, 20)
	k2, d2, j2 := kdj.Kdj(ks)

	if !gnum.FloatsEqual(k1, k2) {
		t.Errorf("Id doesn't match")
		return
	}
	if !gnum.FloatsEqual(d1, d2) {
		t.Errorf("D doesn't match")
		return
	}
	if !gnum.FloatsEqual(j1, j2) {
		t.Errorf("J doesn't match")
		return
	}
}
*/
