package fintypes

import (
	"fmt"
	"github.com/shawnwyckoff/gopkg/apputil/gtest"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gjson"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"testing"
	"time"
)

func TestNewTestAccount(t *testing.T) {
	ticks := Ticks{Items: map[PairM]gdecimal.Decimal{}}
	ticks.Items[PairM("BTC/USDT.spot")] = gdecimal.NewFromInt(5000)
	ticks.Items[PairM("ETH/USDT.spot")] = gdecimal.NewFromInt(500)
	ticks.Items[PairM("IOTX/ETH.spot")] = gdecimal.NewFromFloat64(0.005)
	acc, err := NewTestAccount([]string{"BTC", "USDT", "ETH", "IOTX"}, gdecimal.NewFromInt(10000), ticks)
	if err != nil {
		t.Error(err)
		return
	}
	if !acc.GetAmountByName("BTC").Free.Equal(gdecimal.NewFromFloat64(0.5)) {
		t.Errorf("NewTestAccount error1")
		return
	}
	if !acc.GetAmountByName("ETH").Free.Equal(gdecimal.NewFromInt(5)) {
		t.Errorf("NewTestAccount error2")
		return
	}
	if !acc.GetAmountByName("IOTX").Free.Equal(gdecimal.NewFromInt(1000)) {
		t.Errorf("NewTestAccount error3")
		return
	}
	if !acc.GetAmountByName("USDT").Free.Equal(gdecimal.NewFromInt(2500)) {
		t.Errorf("NewTestAccount error4")
		return
	}
}

func TestAccount_TotalInUSD(t *testing.T) {
	ticks := Ticks{Items: map[PairM]gdecimal.Decimal{}}
	ticks.Items[PairM("BTC/USDT.spot")] = gdecimal.NewFromInt(5000)
	ticks.Items[PairM("ETH/USDT.spot")] = gdecimal.NewFromInt(500)
	ticks.Items[PairM("IOTX/ETH.spot")] = gdecimal.NewFromFloat64(0.005)
	acc, err := NewTestAccount([]string{"BTC", "USDT", "ETH", "IOTX"}, gdecimal.NewFromInt(10000), ticks)
	if err != nil {
		t.Error(err)
		return
	}
	total, err := acc.ExchangeToUSD(ticks, false)
	if err != nil {
		t.Error(err)
		return
	}
	if total == nil || !total.Free.EqualInt(10000) {
		t.Errorf("TotalInUSD error")
		return
	}
}

func TestBalance_Add(t *testing.T) {
	blc := Balance{AssetAmount: AssetAmount{Free: gdecimal.NewFromInt(1000)}}
	blc.Add(blc)
	fmt.Println(blc)
}

func TestAccount_Add(t *testing.T) {
	acc := NewEmptyAccount()
	toAdd := NewEmptyAccount()
	toAdd.SetAmount(AssetProperty{MarketSpot, MarginNo, "", "USDT"}, AssetAmount{Free: gdecimal.NewFromInt(10000)})
	acc.AddAccount(*toAdd)
	fmt.Println(gjson.MarshalStringDefault(acc, true))
}

func TestAccount_Repay(t *testing.T) {
	acc := NewEmptyAccount()
	acc.AddFree(AssetProperty{MarketSpot, MarginCross, "", "BTC"}, gdecimal.NewFromInt(1))
	err := acc.Borrow(gtime.Sub(time.Now(), time.Hour), MarginCross, "BTC", gdecimal.NewFromFloat64(0.5))
	gtest.Assert(t, err)

	acc.Repay(time.Now(), gdecimal.N0, MarginCross, "BTC", gdecimal.NewFromFloat64(0.25))
	gtest.Assert(t, err)
	borrowed := acc.GetAmountByName("BTC").Borrowed
	if borrowed.String() != "0.25" {
		gtest.PrintlnExit(t, "borrowed expect 0.25 but got %s", borrowed.String())
	}
}
