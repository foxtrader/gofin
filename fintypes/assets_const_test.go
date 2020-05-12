package fintypes

import (
	"fmt"
	"github.com/shawnwyckoff/gopkg/apputil/gtest"
	"testing"
)

func TestAllFiats(t *testing.T) {
	list := AllFiats()
	fmt.Println(len(list))
	fmt.Println(list)
}

func TestStableCoinsByFiat(t *testing.T) {
	coins := StableCoinsByFiat(USD)
	if len(coins) < 3 {
		gtest.PrintlnExit(t, "StableCoinsByFiat error")
	}
}
