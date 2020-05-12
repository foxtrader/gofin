package findata

import (
	"github.com/foxtrader/gofin/fintypes"
	"github.com/foxtrader/gofin/testdata"
	"github.com/gocarina/gocsv"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
)

func NewTestKlineMaotai() (*fintypes.Kline, error) {
	r := new(fintypes.Kline)
	type tmpDot struct {
		Date   string  `csv:"Date"`
		Open   float64 `csv:"O"`
		Low    float64 `csv:"L"`
		High   float64 `csv:"H"`
		Close  float64 `csv:"C"`
		Volume float64 `csv:"V"`
	}
	var dots []tmpDot
	if err := gocsv.UnmarshalString(testdata.MaoTaiKline, &dots); err != nil {
		return nil, err
	}
	for _, v := range dots {
		dt, err := gtime.ParseDateString(v.Date, true)
		if err != nil {
			return nil, err
		}
		kdot := fintypes.Bar{
			T: dt.ToTimeUTC(),
			O: gdecimal.NewFromFloat64(v.Open),
			L: gdecimal.NewFromFloat64(v.Low),
			H: gdecimal.NewFromFloat64(v.High),
			C: gdecimal.NewFromFloat64(v.Close),
			V: gdecimal.NewFromFloat64(v.Volume),
		}
		r.UpsertDot(kdot)
	}

	return r, nil
}
