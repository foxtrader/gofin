package fintypes

import "time"

type (
	KlineProviderInfo struct {
		SupportedPairs        []PairIMP
		MinPeriod             Period
		KlineRequestRateLimit time.Duration
		FirstTrade            time.Time
	}
)
