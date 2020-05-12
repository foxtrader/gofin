package findata

import (
	"github.com/shawnwyckoff/gopkg/apputil/gtest"
	"testing"
)

func TestEcb_GetKline(t *testing.T) {
	_, err := NewEcb().GetKline()
	gtest.Assert(t, err)
}
