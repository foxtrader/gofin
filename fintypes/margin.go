package fintypes

import "github.com/shawnwyckoff/gopkg/apputil/gerror"

var (
	allMargins []Margin

	MarginError    = Margin("")
	MarginNo       = enrollNewMargin("no")       // 无杠杆
	MarginIsolated = enrollNewMargin("isolated") // 逐仓
	MarginCross    = enrollNewMargin("cross")    // 全仓
)

type (
	Margin string
)

func enrollNewMargin(name string) Margin {
	res := Margin(name)
	allMargins = append(allMargins, res)
	return res
}

func (m Margin) Verify() error {
	for _, v := range allMargins {
		if m == v {
			return nil
		}
	}
	return gerror.Errorf("invalid Margin(%s)", m)
}

func ParseMargin(s string) (Margin, error) {
	if err := Margin(s).Verify(); err != nil {
		return MarginError, err
	}
	return Margin(s), nil
}
