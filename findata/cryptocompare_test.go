package findata

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestCryptoCompareGetAll(t *testing.T) {
	all, err := CCGetAll("", time.Minute)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(all)
	t.Log(len(all))
}

func TestCCGetKline(t *testing.T) {
	r, err := CCGetKline("BTC", nil, "")
	if err != nil {
		t.Error(err)
		return
	}
	buf, err := json.Marshal(r)
	if err != nil {
		t.Error(err)
		return
	}
	if len(r.Items) == 0 {
		t.Errorf("CCGetKline returns no records")
		return
	}
	fmt.Println(string(buf))
}
