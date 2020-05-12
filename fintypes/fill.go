package fintypes

import (
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"time"
)

type (
	//QuoteQty   decimals.Decimal `json:"QuoteQty" bson:"QuoteQty"`
	//BuyerMaker    bool      `json:"BuyerMaker" bson:"BuyerMaker"` // binance use it only
	Fill struct {
		Id      int64            `json:"Id" bson:"_id"`
		Time    time.Time        `json:"T" bson:"T"`
		Price   gdecimal.Decimal `json:"Price" bson:"Price"`
		UnitQty gdecimal.Decimal `json:"UnitQty" bson:"UnitQty"`
		Side    string           `json:"Side" bson:"Side"` // "buy", "sell", "auction"...
	}

	FillOption struct {
		BeginTime    time.Time
		TimeDuration time.Duration

		BeginId int64
		IdLimit int64
	}
)

func (fo FillOption) VerifyBinance() error {
	if gtime.BeforeEqual(fo.BeginTime, gtime.ZeroTime) &&
		fo.BeginId <= 0 {
		return gerror.Errorf("both beginTime(%s) and beginId[%d] are invalid", fo.BeginTime.String(), fo.BeginId)
	}

	if fo.BeginTime.After(gtime.ZeroTime) {
		if gtime.BeforeEqual(fo.BeginTime, BTCGenesisBlockTime) {
			return gerror.Errorf("invalid begin time(%s)", fo.BeginTime)
		}
		if fo.TimeDuration <= 0 || fo.TimeDuration > time.Hour {
			return gerror.Errorf("invalid time duration(%s), [1s, 1hour] limited by API", fo.TimeDuration)
		}
	}

	if fo.BeginId > 0 {
		if fo.IdLimit > 1000 {
			return gerror.Errorf("invalid IdLimit[%d], [1, 1000] limited by API", fo.IdLimit)
		}
	}
	return nil
}
