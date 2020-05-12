package fintypes

import (
	"fmt"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

var (
	dt20181231 = gtime.NewDatePanic(2018, 12, 31).ToTimeUTC()
	dt20190101 = gtime.NewDatePanic(2019, 1, 1).ToTimeUTC()
	dt20190102 = gtime.NewDatePanic(2019, 1, 2).ToTimeUTC()
	dt20190103 = gtime.NewDatePanic(2019, 1, 3).ToTimeUTC()
	dt20190104 = gtime.NewDatePanic(2019, 1, 4).ToTimeUTC()
	dt20190105 = gtime.NewDatePanic(2019, 1, 5).ToTimeUTC()

	/*tm20190101_000000 = time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	tm20190101_000100 = time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC)
	tm20190101_000200 = time.Date(2019, 1, 1, 0, 2, 0, 0, time.UTC)
	tm20190101_000300 = time.Date(2019, 1, 1, 0, 3, 0, 0, time.UTC)
	tm20190101_000400 = time.Date(2019, 1, 1, 0, 4, 0, 0, time.UTC)
	tm20190101_000500 = time.Date(2019, 1, 1, 0, 5, 0, 0, time.UTC)
	tm20190101_000600 = time.Date(2019, 1, 1, 0, 6, 0, 0, time.UTC)
	tm20190101_000700 = time.Date(2019, 1, 1, 0, 7, 0, 0, time.UTC)
	tm20190101_000800 = time.Date(2019, 1, 1, 0, 8, 0, 0, time.UTC)
	tm20190101_000900 = time.Date(2019, 1, 1, 0, 9, 0, 0, time.UTC)
	tm20190101_001000 = time.Date(2019, 1, 1, 0, 10, 0, 0, time.UTC)
	tm20190101_001100 = time.Date(2019, 1, 1, 0, 11, 0, 0, time.UTC)
	tm20190101_001200 = time.Date(2019, 1, 1, 0, 12, 0, 0, time.UTC)
	tm20190101_001300 = time.Date(2019, 1, 1, 0, 13, 0, 0, time.UTC)
	tm20190101_001400 = time.Date(2019, 1, 1, 0, 14, 0, 0, time.UTC)
	tm20190101_001500 = time.Date(2019, 1, 1, 0, 15, 0, 0, time.UTC)
	tm20190101_001600 = time.Date(2019, 1, 1, 0, 15, 0, 0, time.UTC)
	tm20190101_001700 = time.Date(2019, 1, 1, 0, 15, 0, 0, time.UTC)
	tm20190101_001800 = time.Date(2019, 1, 1, 0, 15, 0, 0, time.UTC)*/
)

func newTestMinuteKline(begin, end time.Time) *Kline {
	k := &Kline{Pair: PairIMP("BTC/USDT.1min.spot.binance"), sorted: true}
	begin = begin.Round(time.Minute)
	end = end.Round(time.Minute)
	n := int(end.Sub(begin)/time.Minute) + 1
	for i := 0; i < n; i++ {
		dot := Bar{T: begin.Add(time.Minute * time.Duration(i)),
			O: gdecimal.NewFromInt(i),
			L: gdecimal.NewFromInt(i / 2),
			H: gdecimal.NewFromInt(i * 2),
			C: gdecimal.NewFromInt(i),
			V: gdecimal.NewFromInt(i * 10),
		}
		k.UpsertDot(dot)
	}
	return k
}

func newTestMinuteKline2(begin, end time.Time) *Kline {
	k := &Kline{Pair: PairIMP("BTC/USDT.1min.spot.binance"), sorted: true}
	begin = begin.Round(time.Minute)
	end = end.Round(time.Minute)
	n := int(end.Sub(begin)/time.Minute) + 1
	for i := 0; i < n; i++ {
		open := gdecimal.NewFromInt(gtime.TimeToIntYYYYMMDDHHMM(begin.Add(time.Minute * time.Duration(i))))
		dot := Bar{T: begin.Add(time.Minute * time.Duration(i)),
			O: open,
			L: open.SubInt(1),
			H: open.AddInt(1),
			C: open,
			V: gdecimal.One,
		}
		k.UpsertDot(dot)
	}
	return k
}

func newTestMinuteKline3() *Kline {
	k := &Kline{Pair: PairIMP("BTC/USDT.1min.spot.binance"), sorted: true}
	tm1 := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	tm2 := time.Date(2019, 2, 1, 0, 0, 0, 0, time.UTC)

	dot1 := Bar{T: tm1}
	dot2 := Bar{T: tm2}
	k.UpsertDot(dot1)
	k.UpsertDot(dot2)

	return k
}

func newTestKline1() *Kline {
	r := new(Kline)

	date, _ := gtime.NewDate(2019, 1, 1)
	r.Items = append(r.Items, Bar{T: date.ToTimeUTC(), H: gdecimal.NewFromInt(1)})
	date, _ = gtime.NewDate(2019, 1, 2)
	r.Items = append(r.Items, Bar{T: date.ToTimeUTC(), H: gdecimal.NewFromInt(2)})
	date, _ = gtime.NewDate(2019, 1, 3)
	r.Items = append(r.Items, Bar{T: date.ToTimeUTC(), H: gdecimal.NewFromInt(3)})

	r.Sort()
	return r
}

/*
func TestKDot_Scale(t *testing.T) {
	dot := Bar{L:decimals.NewFromInt(10), H:decimals.NewFromInt(20), O:decimals.NewFromInt(18), C:decimals.NewFromInt(12)}
	ant := dot.arLHS(nil)
	if ant.DownShadowScale != -20 || ant.UpShadowScale != -20 {
		t.Errorf("dot.ARAntecedent() error1")
		return
	}

	dot = Bar{L:decimals.NewFromInt(10), H:decimals.NewFromInt(20), O:decimals.NewFromInt(12), C:decimals.NewFromInt(18)}
	ant = dot.ARAntecedent(nil)
	if ant.DownShadowScale != 20 || ant.UpShadowScale != 20 {
		t.Errorf("dot.ARAntecedent() error2")
		return
	}
}*/

func TestSyncXAxis(t *testing.T) {
	ks1 := new(Kline)
	ks2 := new(Kline)

	date, _ := gtime.NewDate(2019, 1, 1)
	ks1.Items = append(ks1.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(1)})
	date, _ = gtime.NewDate(2019, 1, 2)
	ks1.Items = append(ks1.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(2)})
	date, _ = gtime.NewDate(2019, 1, 3)
	ks1.Items = append(ks1.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(3)})

	date, _ = gtime.NewDate(2019, 1, 2)
	ks2.Items = append(ks2.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(2)})
	date, _ = gtime.NewDate(2019, 1, 3)
	ks2.Items = append(ks2.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(3)})
	date, _ = gtime.NewDate(2019, 1, 4)
	ks2.Items = append(ks2.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(4)})

	SyncXAxis(ks1, ks2, Bar{H: gdecimal.NewFromInt(-1)})
	fmt.Println(ks1)
	fmt.Println(ks2)

	if !ks2.Items[0].H.Equal(gdecimal.NewFromInt(-1)) {
		t.Error("SyncXAxis error")
		return
	}
	if !ks1.Items[3].H.Equal(gdecimal.NewFromInt(-1)) {
		t.Error("SyncXAxis error")
		return
	}

}

// FIXME
func TestKline_ToPeriod(t *testing.T) {
	/*src = src.SliceTail(600)
	cks, err := src.ToPeriod(comm.Period1Hour, comm.DefaultPeriodRoundConfig)
	fmt.Println(cks.SliceTail(10))
	return*/

	prcUtc := DefaultPeriodRoundConfig
	prcUtc.Location = *time.UTC

	// 测试"交易所发生维护"的Kline转换
	k2 := newTestMinuteKline3()
	k3, err := k2.ToPeriod(Period1Day, prcUtc)
	if err != nil {
		t.Error(err)
		return
	}
	if k3.Len() != 2 {
		t.Errorf("ToPeriod with time-lack K error")
	}

	// 测试非UTC时区下转换后的时间
	prcSh := DefaultPeriodRoundConfig
	prcSh.Location = *gtime.TimeZoneAsiaShanghai
	k4, err := k2.ToPeriod(Period1Day, prcSh)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(k4)
	if k4.Len() != 2 {
		t.Errorf("ToPeriod with Local location K error")
	}

	// 测试其他
	// 2017/1/1是周日，若采用默认转换配置，各个周期都是整数，方便测试
	// WARNING: 如果begin和PeriodRoundConfig不匹配，换言之，begin不是PeriodRoundConfig各个周期的第一天，那测试用例本身就是不对的
	begin := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2017, 12, 31, 23, 59, 0, 0, time.UTC)
	src := newTestMinuteKline2(begin, end)
	//fmt.Println(src)
	//gfs.StringToFile(gjson.MarshalStringDefault(src, true), "/Users/wongkashing/git/go/src/github.com/shawnwyckoff/rusbeta/backtest/abc.json")

	for _, period := range AllPeriods {
		r, err := src.ToPeriod(period, prcUtc)
		if err != nil {
			t.Error(err)
			return
		}

		for i, v := range r.Items {
			//fmt.Println(v.T.Location())
			expectOpen := gtime.TimeToIntYYYYMMDDHHMM(v.T)
			expectLow := expectOpen - 1
			closeTime := gtime.Sub(v.T.Add(period.ToDurationExact(v.T, time.UTC)), time.Minute)
			if i == r.Len()-1 {
				closeTime = end
			}
			expectClose := gtime.TimeToIntYYYYMMDDHHMM(closeTime)
			expectHigh := expectClose + 1
			expectVolume := int(closeTime.Add(time.Minute).Sub(v.T) / time.Minute)

			if !v.O.EqualInt(expectOpen) {
				fmt.Println(r)
				t.Errorf("period(%s) time(%s) open: %s got but %d expected", period, v.T.UTC().String(), v.O.String(), expectOpen)
				return
			}
			if !v.L.EqualInt(expectLow) {
				t.Errorf("period %s time %s: low %s got, but %d expected", period, v.T.String(), v.L.String(), expectLow)
				return
			}
			if !v.H.EqualInt(expectHigh) {
				t.Errorf("period %s time %s: high %s got, but %d expected", period, v.T.String(), v.H.String(), expectHigh)
				return
			}
			if !v.C.EqualInt(expectClose) {
				t.Errorf("period %s time %s: close %s got, but %d expected", period, v.T.String(), v.C.String(), expectClose)
				return
			}
			if !v.V.EqualInt(expectVolume) {
				t.Errorf("period %s time %s: volume %s got, but %d expected", period, v.T.String(), v.V.String(), expectVolume)
				return
			}
		}
	}
}

func TestKlineSet_IsTimeOverlappingAreaEqual(t *testing.T) {
	ks1 := new(Kline)
	ks2 := new(Kline)

	date, _ := gtime.NewDate(2019, 1, 1)
	ks1.Items = append(ks1.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(1)})
	date, _ = gtime.NewDate(2019, 1, 2)
	ks1.Items = append(ks1.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(2)})
	date, _ = gtime.NewDate(2019, 1, 3)
	ks1.Items = append(ks1.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(3)})

	date, _ = gtime.NewDate(2019, 1, 4)
	ks2.Items = append(ks2.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(4)})
	date, _ = gtime.NewDate(2019, 1, 1)
	ks2.Items = append(ks2.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(1)})
	date, _ = gtime.NewDate(2019, 1, 2)
	ks2.Items = append(ks2.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromFloat64(2.1)})
	date, _ = gtime.NewDate(2019, 1, 3)
	ks2.Items = append(ks2.Items, Bar{T: date.ToTime(0, 0, 0, 0, time.UTC), H: gdecimal.NewFromInt(3)})

	equal := ks1.IsTimeOverlappingAreaEqual(ks2)
	if equal != false {
		t.Errorf("IsTimeOverlappingAreaEqual 1")
		return
	}

	date, _ = gtime.NewDate(2019, 1, 2)
	ks2.Update(date.ToTime(0, 0, 0, 0, time.UTC), "H", gdecimal.NewFromInt(2))
	equal = ks1.IsTimeOverlappingAreaEqual(ks2)
	if equal != true {
		t.Errorf("IsTimeOverlappingAreaEqual 2")
		return
	}
}

func TestKline_SliceBetweenId(t *testing.T) {
	k1 := newTestKline1()
	ks1 := k1.SliceBetweenId(0, 1)
	if ks1.Len() != 2 {
		t.Errorf("Kline_SliceById failed 1")
		return
	}
	ks2 := k1.SliceBetweenId(1, 2)
	if ks2.Len() != 2 {
		t.Errorf("Kline_SliceById failed 2")
		return
	}
	ks3 := k1.SliceBetweenId(0, 2)
	if ks3.Len() != 3 {
		t.Errorf("Kline_SliceById failed 3")
		return
	}
	ks4 := k1.SliceBetweenId(0, 3)
	if ks4.Len() != 3 {
		t.Errorf("Kline_SliceById failed 4")
		return
	}
}

func TestKline_IndexEqual(t *testing.T) {
	k1 := newTestKline1()

	idx1, _, ok := k1.IndexEqual(dt20190101)
	if !ok || idx1 != 0 {
		t.Errorf("IndexEqual failed 1")
		return
	}

	idx2, _, ok := k1.IndexEqual(dt20190102)
	if !ok || idx2 != 1 {
		t.Errorf("IndexEqual failed 2")
		return
	}

	idx3, _, ok := k1.IndexEqual(dt20190103)
	if !ok || idx3 != 2 {
		t.Errorf("IndexEqual failed 3")
		return
	}

	idx4, _, ok := k1.IndexEqual(dt20190104)
	if ok || idx4 != -1 {
		t.Errorf("IndexEqual failed 4")
		return
	}
}

func TestKline_MaxIndexLT(t *testing.T) {
	k1 := newTestKline1()

	idx1 := k1.MaxIndexLT(dt20190101)
	if idx1 != -1 {
		t.Errorf("MaxIndexLT failed 1")
		return
	}

	idx2 := k1.MaxIndexLT(dt20190102)
	if idx2 != 0 {
		t.Errorf("MaxIndexLT failed 2")
		return
	}

	idx3 := k1.MaxIndexLT(dt20190103)
	if idx3 != 1 {
		t.Errorf("MaxIndexLT failed 3")
		return
	}

	idx4 := k1.MaxIndexLT(dt20190104)
	if idx4 != 2 {
		t.Errorf("MaxIndexLT failed 4")
		return
	}

	idx5 := k1.MaxIndexLT(dt20190105)
	if idx5 != 2 {
		t.Errorf("MaxIndexLT failed 4")
		return
	}
}

func TestKline_MinIndexGT(t *testing.T) {
	k1 := newTestKline1()

	idx0 := k1.MinIndexGT(dt20181231)
	if idx0 != 0 {
		t.Errorf("MinIndexGT failed 0")
		return
	}

	idx1 := k1.MinIndexGT(dt20190101)
	if idx1 != 1 {
		t.Errorf("MinIndexGT failed 1")
		return
	}

	idx2 := k1.MinIndexGT(dt20190102)
	if idx2 != 2 {
		t.Errorf("MinIndexGT failed 2")
		return
	}

	idx3 := k1.MinIndexGT(dt20190103)
	if idx3 != -1 {
		t.Errorf("MinIndexGT failed 3")
		return
	}

	idx4 := k1.MinIndexGT(dt20190104)
	if idx4 != -1 {
		t.Errorf("MinIndexGT failed 4")
		return
	}
}

func TestKline_SliceBefore(t *testing.T) {
	k1 := newTestKline1()

	ks1 := k1.SliceBefore(dt20190101)
	if ks1.Len() != 0 {
		fmt.Println(ks1)
		t.Errorf("SliceBefore failed 1")
		return
	}

	ks2 := k1.SliceBefore(dt20190102)
	if ks2.Len() != 1 {
		t.Errorf("SliceBefore failed 2")
		return
	}

	ks3 := k1.SliceBefore(dt20190103)
	if ks3.Len() != 2 {
		t.Errorf("SliceBefore failed 3")
		return
	}

	ks4 := k1.SliceBefore(dt20190104)
	if ks4.Len() != 3 {
		t.Errorf("SliceBefore failed 4")
		return
	}

	/*
		fmt.Println(ks1)
		fmt.Println(ks2)
		fmt.Println(ks3)
		fmt.Println(ks4)*/
}

func TestKline_SliceAfter(t *testing.T) {
	k1 := newTestKline1()

	ks0 := k1.SliceAfter(dt20181231)
	if ks0.Len() != 3 {
		fmt.Println(ks0)
		t.Errorf("SliceAfter failed 0")
		return
	}

	ks1 := k1.SliceAfter(dt20190101)
	if ks1.Len() != 2 {
		fmt.Println(ks1)
		t.Errorf("SliceAfter failed 1")
		return
	}

	ks2 := k1.SliceAfter(dt20190102)
	if ks2.Len() != 1 {
		t.Errorf("SliceAfter failed 2")
		return
	}

	ks3 := k1.SliceAfter(dt20190103)
	if ks3.Len() != 0 {
		t.Errorf("SliceAfter failed 3")
		return
	}

	ks4 := k1.SliceAfter(dt20190104)
	if ks4.Len() != 0 {
		t.Errorf("SliceAfter failed 4")
		return
	}

	/*
		fmt.Println(ks0)
		fmt.Println(ks1)
			fmt.Println(ks2)
			fmt.Println(ks3)
			fmt.Println(ks4)*/
}

func TestKline_SliceBeforeEqual(t *testing.T) {
	k1 := newTestKline1()

	ks0 := k1.SliceBeforeEqual(dt20181231)
	if ks0.Len() != 0 {
		fmt.Println(ks0)
		t.Errorf("SliceBeforeEqual failed 0")
		return
	}

	ks1 := k1.SliceBeforeEqual(dt20190101)
	if ks1.Len() != 1 {
		fmt.Println(ks1)
		t.Errorf("SliceBeforeEqual failed 1")
		return
	}

	ks2 := k1.SliceBeforeEqual(dt20190102)
	if ks2.Len() != 2 {
		t.Errorf("SliceBeforeEqual failed 2")
		return
	}

	ks3 := k1.SliceBeforeEqual(dt20190103)
	if ks3.Len() != 3 {
		t.Errorf("SliceBeforeEqual failed 3")
		return
	}

	ks4 := k1.SliceBeforeEqual(dt20190104)
	if ks4.Len() != 3 {
		t.Errorf("SliceBeforeEqual failed 4")
		return
	}
}

func TestKDot_MarshalBSON(t *testing.T) {
	dot := Bar{}
	_, err := bson.Marshal(dot)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestKline_LastClosedTime(t *testing.T) {
	begin := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2019, 2, 28, 23, 59, 0, 0, time.UTC)
	k := newTestMinuteKline(begin, end)
	fmt.Println(k.LastTime(gtime.ZeroTime))
	fmt.Println(k.LastClosedTime(Period1Min, DefaultPeriodRoundConfig))
	fmt.Println(k.LastClosedTime(Period5Min, DefaultPeriodRoundConfig))
	fmt.Println(k.LastClosedTime(Period1Hour, DefaultPeriodRoundConfig))
	fmt.Println(k.LastClosedTime(Period1MonthFUZZY, DefaultPeriodRoundConfig))
}

func TestKline_SliceBetween(t *testing.T) {
	begin := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2019, 1, 1, 1, 0, 0, 0, time.UTC)
	k := newTestMinuteKline(begin, end)
	sk := k.SliceBetween(begin, end)
	fmt.Println(sk)
}
