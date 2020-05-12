package findata

import (
	"github.com/foxtrader/gofin/fintypes"
	"testing"
)

func TestStockExchangeList(t *testing.T) {
	list, err := StockExchangeList(fintypes.Hkex)
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) < 5000 {
		t.Errorf("hkex stocks count must > 5000")
		return
	}
	t.Log(list)
	if list[0] == fintypes.PairPErr {
		t.Errorf("hkex get stock error")
		return
	}

	list, err = StockExchangeList(fintypes.Nasdaq)
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) < 5000 {
		t.Errorf("nasdaq stocks count must > 5000")
		return
	}
	if list[0] == fintypes.PairPErr {
		t.Errorf("hkex get stock error")
		return
	}
}

func TestStockListAll(t *testing.T) {
	stocks, indexes, err := StockListAll()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(stocks))
	t.Log(stocks)
	t.Log(indexes)
}
