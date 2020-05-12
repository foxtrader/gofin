package fintypes

import (
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gjson"
)

type (
	HighlyRelevant struct {
		Items []PairMP
	}

	ModeratelyRelevant struct {
		Highs []HighlyRelevant
	}

	LowlyRelevant struct {
		Mids []ModeratelyRelevant
	}

	Irrelevant struct {
		Lows []LowlyRelevant
	}

	Portfolio struct {
		RelevantSet *Irrelevant `json:"RelevantSet,omitempty"`
		DirectSet   []PairMP    `json:"DirectSet,omitempty"`
	}
)

func (p Portfolio) Verify() error {
	if len(p.RelevantSet.Lows) == 0 && len(p.DirectSet) == 0 {
		return gerror.Errorf("empty Portfolio")
	}

	if len(p.RelevantSet.Lows) > 0 {
		rMap := map[PairMP]bool{}
		total := 0
		for _, LrSet := range p.RelevantSet.Lows {
			for _, MidSet := range LrSet.Mids {
				for _, HighSet := range MidSet.Highs {
					for _, target := range HighSet.Items {
						rMap[target] = true
						total++
					}
				}
			}
		}
		if total != len(rMap) {
			return gerror.Errorf("these is duplicated targets")
		}
	}

	return nil
}

func (p Portfolio) Targets() []PairMP {
	if len(p.DirectSet) > 0 {
		return p.DirectSet
	}

	if len(p.RelevantSet.Lows) > 0 {
		rMap := map[PairMP]bool{}
		total := 0
		for _, LrSet := range p.RelevantSet.Lows {
			for _, MidSet := range LrSet.Mids {
				for _, HighSet := range MidSet.Highs {
					for _, target := range HighSet.Items {
						rMap[target] = true
						total++
					}
				}
			}
		}

		var r []PairMP
		for k := range rMap {
			r = append(r, k)
		}
		return r
	}
	return nil
}

func (p Portfolio) String() string {
	return gjson.MarshalStringDefault(p, false)
}
