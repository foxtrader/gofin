package fintypes

import (
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"time"
)

var (
	// https://www.blockchain.com/btc/block/000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f
	BTCGenesisBlockTime = time.Date(2009, 1, 3, 18, 15, 5, 0, time.UTC)
	BTCGenesisBlockDate = gtime.TimeToDate(BTCGenesisBlockTime, time.UTC)
)
