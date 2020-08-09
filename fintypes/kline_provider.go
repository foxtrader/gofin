package fintypes

import "time"

type (
	KlineProviderInfo struct {
		//SupportedPairs        []PairIMP
		SupportedPairs        []PairM
		MinPeriod             Period
		KlineRequestRateLimit time.Duration
		FirstTrade            time.Time
	}
)
