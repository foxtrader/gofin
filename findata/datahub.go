package findata

import (
	"github.com/foxtrader/gofin/fintypes"
	"github.com/gocarina/gocsv"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/net/ghttp"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"time"
)

const (
	goldMonthlyPriceCsv = "https://datahub.io/core/gold-prices/r/monthly.csv"
)

type (
	GoldMonthlyPrice struct {
		YearMonth int              `bson:"_id"`
		Price     gdecimal.Decimal `bson:"Price"`
	}

	csvGoldMonthlyPrice struct {
		Date  string           `csv:"Date"`
		Price gdecimal.Decimal `csv:"Price"`
	}
)

func GetGoldKline() ([]GoldMonthlyPrice, *fintypes.Kline, error) {
	s, err := ghttp.GetString(goldMonthlyPriceCsv, "", time.Minute)
	if err != nil {
		return nil, nil, err
	}

	var inItems []csvGoldMonthlyPrice
	var ymOut []GoldMonthlyPrice
	vpOut := new(fintypes.Kline)
	if err := gocsv.UnmarshalString(s, &inItems); err != nil {
		return nil, nil, err
	}
	for _, v := range inItems {
		ym, err := gtime.ParseYearMonthString(v.Date)
		if err != nil {
			return nil, nil, err
		}
		if v.Price.Float64() <= 0 {
			return nil, nil, errors.Errorf("gold in %s has invalid price %f", v.Date, v.Price.Float64())
		}
		ymOut = append(ymOut, GoldMonthlyPrice{Price: v.Price, YearMonth: ym.Int()})
		vpOut.Items = append(vpOut.Items, fintypes.Bar{T: ym.ToTimeDefault(), O: v.Price, L: v.Price, H: v.Price, C: v.Price})

	}
	return ymOut, vpOut, nil
}
