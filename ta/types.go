package ta

import (
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"strings"
)

const (
	IndSeriesUnknown = ""
	IndSeriesPrice   = "price"
	IndSeriesVolume  = "volume"
	IndSeriesUnique  = "unique"

	IndStyleUnknown = ""
	IndStyleLine    = "line"
	IndStyleBar     = "bar"
)

type (
	IndSeries string

	IndStyle string

	Indicator struct {
		name string
		mask string // OLHCV(TP)[PVUBLZ], O L H C V (T Property) [Price V Unique Bar Line Zigzag]
	}

	// dataSource.Indicator(params...).NextIndicator(params...)
	// EMA(10).MA(30)
	// MFI(default,5).EMA(30)
	// OBV(default).MA(30)
	IndExpr string
)

func registerIndicator(name, mask string) Indicator {
	ind := Indicator{name: name, mask: mask}
	AllIndicators = append(AllIndicators, ind)
	return ind
}

func ParseIndicator(s string) Indicator {
	s = strings.ToUpper(s)
	for _, v := range AllIndicators {
		if v.Name() == s {
			return v
		}
	}
	return IndicatorError
}

var (
	AllIndicators []Indicator

	IndicatorError = registerIndicator("", "")

	// price indicator
	IndicatorMA         = registerIndicator("MA", "C(T)[PL]") // MA is SMA in fact
	IndicatorEMA        = registerIndicator("EMA", "C(T)[PL]")
	IndicatorMACD       = registerIndicator("MACD", "C(TTT)[UL]{Hist:B}") // Hist成员是特殊情况，是Bar，其余都是L（Line）
	IndicatorKDJ        = registerIndicator("KDJ", "LHC(TPP)[UL]")        // mask可能不严谨
	IndicatorSKDJ       = registerIndicator("SKDJ", "LHC(TP)[UL]")
	IndicatorBOLL       = registerIndicator("BOLL", "C(TP)[PL]")
	IndicatorDUALTHRUST = registerIndicator("DUALTHRUST", "OHLC(TPP)[PL]")
	IndicatorATR        = registerIndicator("ATR", "HLC(T)[UL]")
	IndicatorATRK       = registerIndicator("ATRK", "HLC(TP)[UL]")
	IndicatorDC         = registerIndicator("DC", "LH(T)[PL]")

	// volume price indicator
	IndicatorAD    = registerIndicator("AD", "HLCV()[UL]") // Chaikin A/D Line
	IndicatorADOSC = registerIndicator("ADOSC", "HLCV(TT)[UL]")
	IndicatorEMV   = registerIndicator("EMV", "HLV(T)[UL]")
	IndicatorMFI   = registerIndicator("MFI", "HLCV(T)[UL]") // 1in 1out
	IndicatorOBV   = registerIndicator("OBV", "CV()[UL]")    // OBV使用的是Close价。很多工具中的OBV有1个参数，这个参数其实就是对OBV做的MA需要的参数
	IndicatorVPT   = registerIndicator("VPT", "CV()[UL]")

	IndicatorVAR      = registerIndicator("VAR", "C(T)[UL]")
	IndicatorSTOCKRSI = registerIndicator("StockRSI", "C(TTT)[UL]") // 最后一个参数是给MA的，出2条线
	IndicatorRSI      = registerIndicator("RSI", "C(TTT)[UL]")      // 3个参数出3条线，一起用才行
	IndicatorSAR      = registerIndicator("SAR", "HL(PP)[UL]")
	IndicatorTRIX     = registerIndicator("TRIX", "C(T)[UL]") // 2个参数，1个给TRIX，1个给MA，出2条线(TRIX, TRIX+MA)
	IndicatorWR       = registerIndicator("WR", "HLC(T)[UL]") // 用2个参数画出2条线，一起用
	IndicatorROC      = registerIndicator("ROC", "C(T)[UL]")
	IndicatorDMI      = registerIndicator("DMI", "HLC(T)[UL]") // DMI指标包含4条线：PDI、MDI、ADX和ADXR

	// IndicatorLLT // https://zhuanlan.zhihu.com/p/34097449

	IndicatorZIGZAG = registerIndicator("ZIGZAG", "OC(PP)[PZ]")
	IndicatorK      = registerIndicator("Id", "(P)[]")
)

func (i *Indicator) Name() string {
	return strings.ToUpper(i.name)
}

func (i *Indicator) ParamCount() int {
	paramDefine, err := gstring.SubstrBetween(i.mask, "(", ")", true, true, false, false)
	if err != nil {
		return -1
	}
	return len(paramDefine)
}

func (i *Indicator) Series() IndSeries {
	ssDefine, err := gstring.SubstrBetween(i.mask, "[", "]", true, true, false, false)
	if err != nil {
		return IndSeriesUnknown
	}
	if strings.Contains(ssDefine, "P") {
		return IndSeriesPrice
	}
	if strings.Contains(ssDefine, "Content") {
		return IndSeriesVolume
	}
	if strings.Contains(ssDefine, "U") {
		return IndSeriesUnique
	}
	return IndSeriesUnknown
}

// MACD(12,9,3.5).Hist => Bar
func ValExprStyle(ve string) (IndStyle, error) {
	ss := strings.Split(ve, ").") // 在参数中可能出现"."，所以单独的点不能作为间隔依据
	if len(ss) == 2 {
		ss[0] += ")"
	}
	ind, _, err := IndExpr(ss[0]).Parse()
	if err != nil {
		return IndStyleUnknown, err
	}

	styleDefine := ""
	if len(ss) == 2 && strings.Contains(ind.mask, "{") {
		specialDefine, err := gstring.SubstrBetween(ind.mask, "{", "}", true, true, false, false)
		if err != nil {
			return IndStyleUnknown, err
		}
		defines := strings.Split(specialDefine, ",")
		for _, define := range defines {
			info := strings.Split(define, ":")
			key := info[0]
			if strings.ToUpper(ss[1]) == strings.ToUpper(key) {
				styleDefine = info[1]
				break
			}
		}
	}

	if styleDefine == "" {
		ssDefine, err := gstring.SubstrBetween(ind.mask, "[", "]", true, true, false, false)
		if err != nil {
			return IndStyleUnknown, err
		}
		styleDefine = ssDefine
	}

	if strings.Contains(styleDefine, "B") {
		return IndStyleBar, nil
	}
	if strings.Contains(styleDefine, "L") {
		return IndStyleLine, nil
	}
	if strings.Contains(styleDefine, "Z") {
		return IndStyleLine, nil
	}
	return IndStyleUnknown, errors.Errorf("can't find style fot val expr(%s)", ve)
}

/*
func (i *Indicator) Style(member string) IndStyle {


	ssDefine, err := strings2.SubstrBetween(i.mask, "[", "]", true, true, false, false)
	if err != nil {
		return IndStyleUnknown
	}
	if strings.Contains(ssDefine, "B") {
		return IndStyleBar
	}
	if strings.Contains(ssDefine, "L") {
		return IndStyleLine
	}
	if strings.Contains(ssDefine, "Z") {
		return IndStyleLine
	}
	return IndStyleUnknown
}*/

func (i *Indicator) TimePeriodParamIndex() []int {
	paramDefine, err := gstring.SubstrBetween(i.mask, "(", ")", true, true, false, false)
	if err != nil {
		return nil
	}

	var r []int
	for i, v := range []rune(paramDefine) {
		if v == 'T' {
			r = append(r, i)
		}
	}
	return r
}

func (ie IndExpr) String() string {
	return string(ie)
}

// EMA(10) -> EMA, 10
// MACD(6, 30, 3) -> MACD, [6, 30, 3]
func (ie IndExpr) Parse() (indicator Indicator, params []gdecimal.Decimal, err error) {
	if string(ie) == "" {
		return IndicatorError, nil, nil
	}

	expr := string(ie)

	// parse params
	s, err := gstring.SubstrBetween(expr, "(", ")", true, false, false, false)
	if err == nil && s != "" {
		ss := strings.Split(s, ",")
		for _, v := range ss {
			v = strings.TrimSpace(v)
			dec, err := gdecimal.NewFromString(v)
			if err != nil {
				return Indicator{}, nil, errors.Errorf("invalid indicator expression (%s)", expr)
			}
			params = append(params, dec)
		}
	}

	// parse indicator
	name := strings.Split(expr, "(")[0]
	r := IndicatorError

	for _, ind := range AllIndicators {
		if gstring.EqualUpper(ind.Name(), name) {
			r = ind
			break
		}
	}
	if r == IndicatorError {
		return IndicatorError, nil, errors.Errorf("unsupported indicator (%s), it's not in AllIndicators list", name)
	}

	// check config count
	if r.ParamCount() != len(params) {
		return Indicator{}, nil, errors.Errorf("IndExpr(%s) want %d params, but got %d", ie, r.ParamCount(), len(params))
	}
	return r, params, nil
}

// if returns nil, means time period is infinite or not sure
func (ie IndExpr) MaxTimePeriod() *int {
	ind, params, err := ie.Parse()
	if err != nil {
		return nil
	}
	if ind == IndicatorZIGZAG {
		return nil
	}

	tppis := ind.TimePeriodParamIndex()
	if len(tppis) == 0 {
		return nil
	}
	max := -1
	for _, v := range tppis {
		if params[v].IntPart() > max {
			max = params[v].IntPart()
		}
	}
	return &max
}
