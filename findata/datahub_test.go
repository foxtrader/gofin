package findata

import (
	"fmt"
	"testing"
)

func TestGetGoldPriceHistory(t *testing.T) {
	ymr, vpr, err := GetGoldKline()
	if err != nil {
		t.Error(err)
		return
	}
	for _, v := range ymr {
		fmt.Println(v.YearMonth, v.Price)
	}
	for _, v := range vpr.Items {
		fmt.Println(v.T, v.O)
	}
}
