package fintypes

// technical analysis

import (
	"fmt"
	"github.com/foxtrader/gofin/ta"
	"github.com/markcheno/go-talib"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gjson"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"github.com/shawnwyckoff/gopkg/container/gternary"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	TADirectionDown       Dir = "down"
	TADirectionUp         Dir = "up"
	TADirectionNotSureYet Dir = "not_sure_yet"
)

const (
	KFormUnknown KForm = 0
	KFormOne     KForm = 0
)

type (
	KForm int // ABC 3位数字表示，分别代表上影线震幅，实体线震幅，下影线震幅，正负代表是阴线还是阳线

	Dir       string
	DoubleDir string

	TACross struct {
		Time      time.Time
		Direction Dir
	}

	KTA Kline

	TACrosses []TACross

	// associated rules antecedent
	// 目前只使用了K线形态
	LHS struct {
		Form KForm
	}

	// associated rules consequent
	RHS struct {
		Shock map[int]int // map[PeriodSize]ShockScale[-3,3]
	}

	// K Association Rule
	AR struct {
		LHS
		RHS
	}

	// K Association Rules
	ARs struct {
		Items []AR
	}

	PatternLHS struct {
		Forms [3]KForm
	}

	PatternRHS struct {
		LongTimes  int
		ShockTimes int
		ShortTimes int

		/*MiniLongTimes int
		MidLongTimes int
		BigLongTimes int
		MiniShortTimes int
		MidShortTimes int
		BigShortTimes int
		ShockTimes int*/
	}

	Pattern struct {
		PatternLHS
		PatternRHS
	}

	Patterns struct {
		Items map[PatternLHS]PatternRHS
	}

	// direction with time
	DirWT struct {
		Time time.Time
		Dir  Dir
	}

	dirWTs []DirWT
)

func (d dirWTs) Len() int {
	return len([]DirWT(d))
}

func (d dirWTs) Less(i, j int) bool {
	return []DirWT(d)[i].Time.Before([]DirWT(d)[j].Time)
}

func (d dirWTs) Swap(i, j int) {
	[]DirWT(d)[i], []DirWT(d)[j] = []DirWT(d)[j], []DirWT(d)[i]
}

// NOTE:
// 能否把已经计算过的缓存起来？很难，因为有些指数是和之前的值相关联的，你无法核实之前的值是否改动过
func (kta *KTA) UpdateIndicators(indicatorExpr ...string) error {
	k := kta.K()
	if k.Len() == 0 {
		return nil
	}

	indicatorExpr = gstring.RemoveDuplicate(gstring.ToUpper(indicatorExpr))

	for _, expString := range indicatorExpr {
		exp := ta.IndExpr(expString)
		indicator, params, err := exp.Parse()
		if err != nil {
			return err
		}

		switch indicator {
		case ta.IndicatorError:
			continue
		case ta.IndicatorMA:
			// 查资料显示，MA默认采用close价
			out, err := ta.SMA(k.CloseValues(), params[0].IntPart())
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "", out)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorEMA:
			// close price is used by default
			// reference: https://www.investopedia.com/terms/e/ema.asp
			out, err := ta.EMA(k.CloseValues(), params[0].IntPart())
			if err != nil {
				return err
			}
			//fmt.Println(out)
			err = k.SetIndicatorValue(exp, "", out)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorMACD:
			a, b, c := talib.Macd(k.CloseValues(), params[0].IntPart(), params[1].IntPart(), params[2].IntPart())
			err := k.SetIndicatorValue(exp, "", a)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "SIGNAL", b)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "HIST", c)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorKDJ:
			kdjK, kdjD, kdjJ, err := ta.KDJ(k.LowValues(), k.HighValues(), k.CloseValues(), params[0].IntPart(), params[1].IntPart(), params[2].IntPart())
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "Id", kdjK)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "D", kdjD)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "J", kdjJ)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorSKDJ:
			rsv, skdjK, skdjD := ta.SKDJ(k.LowValues(), k.HighValues(), k.CloseValues(), params[0].IntPart(), params[1].IntPart())
			err := k.SetIndicatorValue(exp, "RSV", rsv)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "Id", skdjK)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "D", skdjD)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorBOLL:
			basis, upper, lower, err := ta.BOLL(k.CloseValues(), params[0].IntPart(), params[1].IntPart())
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "BASIS", basis)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "UPPER", upper)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "LOWER", lower)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorDUALTHRUST:
			buy, sell := ta.DualThrust(k.OpenValues(), k.HighValues(), k.LowValues(), k.CloseValues(), params[0].IntPart(), params[1].Float64(), params[2].Float64())
			if err := k.SetIndicatorValue(exp, "BUY", buy); err != nil {
				return err
			}
			if err := k.SetIndicatorValue(exp, "SELL", sell); err != nil {
				return err
			}
		case ta.IndicatorATR:
			err := k.SetIndicatorValue(exp, "", talib.Atr(k.HighValues(), k.LowValues(), k.CloseValues(), params[0].IntPart()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorATRK:
			err := k.SetIndicatorValue(exp, "", ta.ATRK(k.HighValues(), k.LowValues(), k.CloseValues(), params[0].IntPart(), params[1].Float64()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorDC:
			lower, upper := ta.DC(k.LowValues(), k.HighValues(), params[0].IntPart())
			err := k.SetIndicatorValue(exp, "UPPER", upper)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "LOWER", lower)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorAD:
			err := k.SetIndicatorValue(exp, "", talib.Ad(k.HighValues(), k.LowValues(), k.CloseValues(), k.VolumeValues()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorADOSC:
			err := k.SetIndicatorValue(exp, "", talib.AdOsc(k.HighValues(), k.LowValues(), k.CloseValues(), k.VolumeValues(), params[0].IntPart(), params[1].IntPart()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorEMV:
			err := k.SetIndicatorValue(exp, "", ta.EMV(k.HighValues(), k.LowValues(), k.VolumeValues(), params[0].IntPart()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorMFI:
			err := k.SetIndicatorValue(exp, "", talib.Mfi(k.HighValues(), k.LowValues(), k.CloseValues(), k.VolumeValues(), params[0].IntPart()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorOBV:
			err := k.SetIndicatorValue(exp, "", talib.Obv(k.CloseValues(), k.VolumeValues()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorVPT:
			err := k.SetIndicatorValue(exp, "", ta.VPT(k.CloseValues(), k.VolumeValues()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorVAR:
			err := k.SetIndicatorValue(exp, "", talib.Var(k.CloseValues(), params[1].IntPart()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorSTOCKRSI:
			a, b := talib.StochRsi(k.CloseValues(), params[0].IntPart(), params[1].IntPart(), params[2].IntPart(), talib.SMA)
			err := k.SetIndicatorValue(exp, "A", a)
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "B", b)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorRSI:
			err := k.SetIndicatorValue(exp, "P0", talib.Rsi(k.CloseValues(), params[0].IntPart()))
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "P1", talib.Rsi(k.CloseValues(), params[1].IntPart()))
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "P2", talib.Rsi(k.CloseValues(), params[2].IntPart()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorSAR:
			err := k.SetIndicatorValue(exp, "", talib.Sar(k.HighValues(), k.LowValues(), params[0].Float64(), params[1].Float64()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorTRIX:
			err := k.SetIndicatorValue(exp, "", talib.Trix(k.CloseValues(), params[0].IntPart()))
			if err != nil {
				return err
			}
			err = k.SetIndicatorValue(exp, "SMA", talib.Ma(talib.Trix(k.CloseValues(), params[0].IntPart()), params[1].IntPart(), talib.SMA))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorWR:
			err := k.SetIndicatorValue(exp, "", talib.WillR(k.HighValues(), k.LowValues(), k.CloseValues(), params[0].IntPart()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorROC:
			err := k.SetIndicatorValue(exp, "", talib.Roc(k.CloseValues(), params[0].IntPart()))
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorZIGZAG:
			zz := ta.ZIGZAG(k.OpenValues(), k.CloseValues(), params[0].IntPart(), params[1].Float64())
			err := k.SetIndicatorValue(exp, "", zz)
			if err != nil {
				return err
			}
			continue
		case ta.IndicatorK:
			// do nothing, Id is a what in what out 辅助"指数"，以便这些非真正意义指数的参数也可以用于IndRangeExpr的Exhaust，方便使用
		default:
			return errors.Errorf("unsupported indicator %s in UpdateIndicators", indicator.Name())
		}

	}
	return nil
}

func (kta *KTA) CleanupIndicators() {
	k := kta.K()
	for i := 0; i < k.Len(); i++ {
		k.Items[i].Indicators = nil
	}
}

func (ccs TACrosses) Len() int {
	return len(([]TACross)(ccs))
}

func (ccs TACrosses) Less(i, j int) bool {
	return ([]TACross)(ccs)[i].Time.Before(([]TACross)(ccs)[j].Time)
}

func (ccs TACrosses) Swap(i, j int) {
	([]TACross)(ccs)[i], ([]TACross)(ccs)[j] = ([]TACross)(ccs)[j], ([]TACross)(ccs)[i]
}

func TACrossesSort(src []TACross) []TACross {
	sort.Sort((TACrosses)(src))
	return src
}

func (kta *KTA) K() *Kline {
	return (*Kline)(kta)
}

// 从前往后找放量上涨
// pricePercent 价格上涨幅度
// volumeCmpDays 成交量大于之前多少天
// lowGt low价必须大于这个值，相当于给出一条最低线，忽略低于这条线的大阳线
func (kta *KTA) FirstHeavyVolumePriceRose(gt time.Time, pricePercent float64, volumeCmpDays int, lowGt *float64) (Bar, bool) {
	// 第一根Bar被忽略，所以减2
	for k := volumeCmpDays; k < kta.K().Len(); k++ {
		v := kta.K().Items[k]
		if v.T.Before(gt) || v.T.Equal(gt) {
			continue
		}

		// 必须是阳线
		if v.C.LessThanOrEqual(v.O) {
			continue
		}

		// 如果lowGt存在，则需要检测这个条件
		if lowGt != nil && v.L.Float64() <= *lowGt {
			continue
		}

		// 检查价格
		upPercent := (v.C.Float64() - v.O.Float64()) / v.O.Float64()
		if upPercent < pricePercent {
			continue
		}

		// 检查成交量
		volToday := v.V.Float64()
		volLastNDays := 0.0
		for n := k - 1; n > (k-1-volumeCmpDays) && n >= 0; n-- {
			volLastNDays += kta.Items[n].V.Float64()
		}
		if volToday <= volLastNDays {
			continue
		}

		return v, true
	}

	return Bar{}, false
}

// 从后往前找放量上涨
// pricePercent 价格上涨幅度
// volumeCmpDays 成交量大于之前多少天
func (kta *KTA) LastHeavyVolumePriceRose(gt time.Time, pricePercent float64, volumeCmpDays int) (Bar, bool) {
	// 最后一根Bar被忽略，所以减2
	for k := kta.K().Len() - 2; k >= 0; k-- {
		v := kta.K().Items[k]
		if v.T.Before(gt) {
			break
		}

		// 必须是阳线
		if v.C.LessThanOrEqual(v.O) {
			continue
		}

		// 检查价格
		upPercent := (v.C.Float64() - v.O.Float64()) / v.O.Float64()
		if upPercent < pricePercent {
			continue
		}

		// 检查成交量
		volToday := v.V.Float64()
		volLastNDays := 0.0
		for n := k - 1; n > (k-1-volumeCmpDays) && n >= 0; n-- {
			volLastNDays += kta.Items[n].V.Float64()
		}
		if volToday <= volLastNDays {
			continue
		}

		return v, true
	}

	return Bar{}, false
}

func (kta *KTA) Break(breakUp bool, beginInclude bool, begin time.Time, endInclude bool, end time.Time, priceInclude bool, price gdecimal.Decimal) (time.Time, bool) {
	subK := kta.K().SliceBetweenEqual(begin, end)
	if !beginInclude {
		subK = subK.SliceAfter(begin)
	}
	if !endInclude {
		subK = subK.SliceBefore(end)
	}
	for k := range subK.Items {
		if breakUp { // 检查向上突破
			if priceInclude {
				if subK.Items[k].H.GreaterThanOrEqual(price) {
					return subK.Items[k].T, true
				}
			} else {
				if subK.Items[k].H.GreaterThan(price) {
					return subK.Items[k].T, true
				}
			}
		} else { // 检查向下突破
			if priceInclude {
				if subK.Items[k].L.LessThanOrEqual(price) {
					return subK.Items[k].T, true
				}
			} else {
				if subK.Items[k].L.LessThan(price) {
					return subK.Items[k].T, true
				}
			}
		}
	}
	return gtime.ZeroTime, false
}

// exp1: slow expression
// exp2: fast expression
func (kta *KTA) Crosses(exp1, exp2 string, n int) ([]TACross, error) {
	k := (*Kline)(kta)
	k.Sort()
	if k.Len() == 0 {
		return nil, nil
	}
	if !k.HasExpr(exp1) {
		return nil, errors.Errorf("expr(%s) not exist, (%s) in kline", exp1, strings.Join(kta.K().Exprs(), ","))
	}
	if !k.HasExpr(exp2) {
		return nil, errors.Errorf("expr(%s) not exist, (%s) in kline", exp2, strings.Join(kta.K().Exprs(), ","))
	}

	var r []TACross
	for i := 1; i < k.Len(); i++ {
		lastIdx := i - 1
		currIdx := i
		lastSlow, okLastSlow := k.Items[lastIdx].ExprValueFloat64(exp1)
		lastFast, okLastFast := k.Items[lastIdx].ExprValueFloat64(exp2)
		currSlow, okCurrSlow := k.Items[currIdx].ExprValueFloat64(exp1)
		currFast, okCurrFast := k.Items[currIdx].ExprValueFloat64(exp2)
		if okLastSlow && okLastFast && okCurrSlow && okCurrFast {
			if lastFast < lastSlow && currFast > currSlow {
				item := TACross{
					Time:      k.Items[currIdx].T,
					Direction: TADirectionUp,
				}
				r = append(r, item)
			}
			if lastFast > lastSlow && currFast < currSlow {
				item := TACross{
					Time:      k.Items[currIdx].T,
					Direction: TADirectionDown,
				}
				r = append(r, item)
			}
			if n > 0 && len(r) >= n {
				break
			}
		}
	}
	return r, nil
}

// 从后往前
// exp1: slow expression
// exp2: fast expression
// n: how much cross do you want, n <= 0: ignore it
func (kta *KTA) ReverseCrosses(exp1, exp2 string, n int) ([]TACross, error) {
	k := (*Kline)(kta)
	k.Sort()
	if k.Len() <= 1 {
		return nil, nil
	}
	if !k.HasExpr(exp1) {
		return nil, errors.Errorf("expr(%s) not exist, (%s) in kline %s", exp1, strings.Join(kta.K().Exprs(), ","), kta.Pair.String())
	}
	if !k.HasExpr(exp2) {
		return nil, errors.Errorf("expr(%s) not exist, (%s) in kline %s", exp2, strings.Join(kta.K().Exprs(), ","), kta.Pair.String())
	}

	var r []TACross
	for i := k.Len() - 1; i >= 1; i-- {
		lastIdx := i - 1
		currIdx := i
		lastSlow, okLastSlow := k.Items[lastIdx].ExprValueFloat64(exp1)
		lastFast, okLastFast := k.Items[lastIdx].ExprValueFloat64(exp2)
		currSlow, okCurrSlow := k.Items[currIdx].ExprValueFloat64(exp1)
		currFast, okCurrFast := k.Items[currIdx].ExprValueFloat64(exp2)
		if okLastSlow && okLastFast && okCurrSlow && okCurrFast {
			if lastFast < lastSlow && currFast > currSlow {
				item := TACross{
					Time:      k.Items[currIdx].T,
					Direction: TADirectionUp,
				}
				r = append(r, item)
			}
			if lastFast > lastSlow && currFast < currSlow {
				item := TACross{
					Time:      k.Items[currIdx].T,
					Direction: TADirectionDown,
				}
				r = append(r, item)
			}
			if n > 0 && len(r) >= n {
				break
			}
		}
	}

	return TACrossesSort(r), nil
}

func (kta *KTA) ReverseCross1(exp1, exp2 string) (*TACross, error) {
	crosses, err := kta.ReverseCrosses(exp1, exp2, 1)
	if err != nil {
		return nil, err
	}
	if len(crosses) == 0 {
		return nil, nil
	}
	return &crosses[0], nil
}

// 某表达式(比如DonChian Channel)最后的涨跌趋势，持平的点忽略不计
func (kta *KTA) LastDirs(exp string, directionSize int, gt *time.Time) ([]DirWT, error) {
	if directionSize <= 0 {
		return nil, gerror.Errorf("invalid direction size %d", directionSize)
	}
	//fmt.Println("kta.K.LastTime", kta.K().LastTime(gtime.ZeroTime), len(kta.K().Items[kta.K().Len()-1].indicators))
	tmp := kta.K()
	//fmt.Println("1tmp.LastTime", tmp.LastTime(gtime.ZeroTime), tmp.Exprs())
	if gt != nil && gt.After(kta.K().FirstTime(gtime.ZeroTime)) {
		tmp = kta.K().SliceAfter(*gt)
		//fmt.Println("gt", gt.String())
	}
	//fmt.Println("2tmp.LastTime", tmp.LastTime(gtime.ZeroTime), tmp.Exprs())

	evs, err := tmp.ExprValues(exp)
	if err != nil {
		return nil, err
	}
	times := tmp.Times()
	if len(times) != len(evs) {
		return nil, gerror.Errorf("ExprValues length %d != Times length %d", len(evs), len(times))
	}

	var res []DirWT
	for i := len(evs) - 1; i > 0; i-- {
		if evs[i] < evs[i-1] {
			//fmt.Println(times[i].String(), exp, evs[i], "<", evs[i-1])
			res = append(res, DirWT{
				Time: times[i],
				Dir:  TADirectionDown,
			})
		} else if evs[i] > evs[i-1] {
			//fmt.Println(times[i].String(), exp, evs[i], ">", evs[i-1])
			//fmt.Println(tmp.Items[i].T, tmp.Items[i-1].T)
			res = append(res, DirWT{
				Time: times[i],
				Dir:  TADirectionUp,
			})
		}
		// 满额，退出
		if len(res) == directionSize {
			break
		}
	}

	oldLen := len(res)
	sort.Sort(dirWTs(res))
	if len(res) != oldLen {
		panic(fmt.Sprintf("%d,%d", oldLen, len(res)))
	}
	return res, nil
}

// 最后一个方向反转处的某表达式的值
// 适用于Donchian Channel
func (kta *KTA) LastDirReversalExprValue(exp string, gt *time.Time) (float64, error) {
	//fmt.Println("kta.K.LastTime", kta.K().LastTime(gtime.ZeroTime), len(kta.K().Items[kta.K().Len()-1].indicators))
	tmp := kta.K()
	if gt != nil && gt.After(kta.K().FirstTime(gtime.ZeroTime)) {
		tmp = kta.K().SliceAfter(*gt)
	}
	//fmt.Println("tmp.LastTime", tmp.LastTime(gtime.ZeroTime), len(tmp.Items[tmp.Len()-1].indicators))

	evs, err := tmp.ExprValues(exp)
	if err != nil {
		return 0.0, err
	}

	var reverseFirstDir *Dir = nil
	for i := len(evs) - 1; i > 0; i-- {
		if evs[i] == evs[i-1] {
			continue
		}
		currDir := gternary.If(evs[i] < evs[i-1]).Interface(TADirectionDown, TADirectionUp).(Dir)
		if reverseFirstDir == nil {
			reverseFirstDir = &currDir
		} else if currDir != *reverseFirstDir {
			return evs[i], nil
		}
	}

	return 0.0, gerror.Errorf("LastDirReversalExprValue not found")
}

func getShockScale(oldDot, newDot Bar) int {
	shock := newDot.C.Sub(oldDot.C).Div(oldDot.C).Float64()
	negative := shock < 0
	shock = math.Abs(shock)
	r := 0

	if shock < 0.0003 {
		r = 0
	} else if shock < 0.01 {
		r = 1
	} else if shock < 0.06 {
		r = 2
	} else {
		r = 3
	}
	if negative {
		r = -r
	}
	return r
}

/*
func (kta *KTA) ARs(minPeriodCount, maxPeriodCount int) *ARs {
	if minPeriodCount < 1 || maxPeriodCount >= 30 || maxPeriodCount < minPeriodCount {
		return nil
	}
	k := kta.K()

	// init
	r := &ARs{}
	r.Items = make([]AR, k.Len())
	for i := 0; i < k.Len(); i++ {
		r.Items[i].Shock = map[int]int{}
	}

	// calc AR LHS
	for i := 0; i < k.Len(); i++ {
		r.Items[i].LHS = k.Items[i].lhs()
	}

	// calc AR RHS
	for periodCount := minPeriodCount; periodCount <= maxPeriodCount; periodCount++ {
		for i := 0; i < k.Len()-periodCount; i++ {
			r.Items[i].Shock[periodCount] = getShockScale(k.Items[i], k.Items[i+periodCount])
		}
	}

	return r
}*/

func (ar *AR) PatternRHS() PatternRHS {
	r := PatternRHS{}
	if ar.IsLong() {
		r.LongTimes = 1
		return r
	}
	if ar.IsShort() {
		r.ShortTimes = 1
		return r
	}
	r.ShockTimes = 1
	return r
}

func (ar *AR) IsLong() bool {
	return ar.Shock[1]+ar.Shock[2]+ar.Shock[3] >= 2
}

func (ar *AR) IsShort() bool {
	return ar.Shock[1]+ar.Shock[2]+ar.Shock[3] <= -2
}

func (ars *ARs) Parse() *Patterns {
	getLHS := func(taars [3]AR) PatternLHS {
		r := PatternLHS{}
		for i := 0; i < 3; i++ {
			r.Forms[i] = taars[i].Form
		}
		return r
	}

	r := &Patterns{Items: map[PatternLHS]PatternRHS{}}
	for i := 2; i < len(ars.Items); i++ {
		ts3 := [3]AR{ars.Items[i-2], ars.Items[i-1], ars.Items[i]}
		lhs := getLHS(ts3)
		newRHS := ars.Items[i].PatternRHS()
		oldRHS := r.Items[lhs]
		newRHS.LongTimes += oldRHS.LongTimes
		newRHS.ShortTimes += oldRHS.ShortTimes
		newRHS.ShockTimes += oldRHS.ShockTimes
		r.Items[lhs] = newRHS
	}
	return r
}

func NewPatterns() *Patterns {
	return &Patterns{Items: map[PatternLHS]PatternRHS{}}
}

func (ps *Patterns) MaxShortTimes() int {
	max := 0
	maxLHS := PatternLHS{}
	maxRHS := PatternRHS{}
	for lhs, rhs := range ps.Items {
		if rhs.ShortTimes > max {
			max = rhs.ShortTimes
			maxLHS = lhs
			maxRHS = rhs
		}
	}
	fmt.Println(gjson.MarshalStringDefault(maxLHS, false), gjson.MarshalStringDefault(maxRHS, false))
	return max
}

func (ps *Patterns) MaxShockTimes() int {
	max := 0
	maxLHS := PatternLHS{}
	maxRHS := PatternRHS{}
	for lhs, rhs := range ps.Items {
		if rhs.ShockTimes > max {
			max = rhs.ShockTimes
			maxLHS = lhs
			maxRHS = rhs
		}
	}
	fmt.Println(gjson.MarshalStringDefault(maxLHS, false), gjson.MarshalStringDefault(maxRHS, false))
	return max
}

func (ps *Patterns) MaxLongTimes() int {
	max := 0
	maxLHS := PatternLHS{}
	maxRHS := PatternRHS{}
	for lhs, rhs := range ps.Items {
		if rhs.LongTimes > max {
			max = rhs.LongTimes
			maxLHS = lhs
			maxRHS = rhs
		}
	}
	fmt.Println(gjson.MarshalStringDefault(maxLHS, false), gjson.MarshalStringDefault(maxRHS, false))
	return max
}

func (ps *Patterns) Join(newPattern *Patterns) *Patterns {
	for k, v := range newPattern.Items {
		if origin, ok := ps.Items[k]; ok {
			origin.ShortTimes += v.ShortTimes
			origin.ShockTimes += v.ShockTimes
			origin.LongTimes += v.LongTimes
			ps.Items[k] = origin
		} else {
			ps.Items[k] = v
		}
	}

	return ps
}

func (ps *Patterns) Best() {
	for lhs, rhs := range ps.Items {
		if rhs.LongTimes > 30 && rhs.ShockTimes > 30 {
			if float64(rhs.LongTimes)/float64(rhs.ShortTimes) > 1.5 || float64(rhs.ShortTimes)/float64(rhs.LongTimes) > 1.5 {
				fmt.Println(lhs, rhs)
			}
		}
	}
}

/*
func lhsClassifyWithRanges(ranges []int, val int) [2]int {
	r := [2]int{}

	if val < ranges[0] {
		r[0] = -1000
		r[1] = ranges[0]
		return r
	}

	if val >= ranges[len(ranges)-1] {
		r[0] = ranges[len(ranges)-1]
		r[1] = 1000
		return r
	}

	for i := 1; i < len(ranges); i++ {
		if val >= ranges[i-1] && val < ranges[i] {
			r[0] = ranges[i-1]
			r[1] = ranges[i]
			break
		}
	}
	return r
}

func lhsClassifyCmpLast(val int) [2]int {
	ranges := []int{-10, -6, 0, 6, 10}
	return lhsClassifyWithRanges(ranges, val)
}

func lhsClassifyCommon(val int) [2]int {
	//ranges := []int{-90, -50, -33, -20, -6, 0, 6, 20, 33, 50, 90}
	ranges := []int{-34, -21, 0, 21, 34}
	return lhsClassifyWithRanges(ranges, val)
}
*/

/*func (taar *AR) PatternLHS() PatternLHS {
	r := PatternLHS{}
	r.CmpLastScale = lhsClassify(taar.CmpLastScale)
	r.UpShadowScale = lhsClassify(taar.UpShadowScale)
	r.DownShadowScale = lhsClassify(taar.DownShadowScale)
	return r
}
*/

/*
func (cdl *KTA) IndicatorsLastCrossAtTail(exp1, exp2 string, tail int) (*TACross, bool, error) {
	if tail <= 1 {
		return nil, false, nil
	}
	k := (*K)(cdl)
	fromIdx := k.Len() - tail
	if fromIdx < 0 {
		fromIdx = 0
	}
	fromTime := k.items[fromIdx].T
	return cdl.IndicatorsLastCross(exp1, exp2, &fromTime, nil)
}

// find first 2 dots pair which has almost equal L value
func (k *K) FindAlmostEqualLow(equalFactor float64) (a, b int, found bool) {
	tmpK := k.Clone()
	tmpK.SortLow()

	found = false
	aTime := time.Time{}
	bTime := time.Time{}
	for i := 1; i < tmpK.Len(); i++ {
		min := decimals.Min(tmpK.Balances[i-1].L, tmpK.Balances[i].L)
		max := decimals.Max(tmpK.Balances[i-1].L, tmpK.Balances[i].L)
		diffFactor := max.Sub(min).Div(min)
		if diffFactor.LessThanOrEqualFloat64(equalFactor) {
			found = true
			aTime = tmpK.Balances[i-1].T
			bTime = tmpK.Balances[i].T
			break
		}
	}
	if found {
		a, _, _ = k.IndexEqual(aTime)
		b, _, _ = k.IndexEqual(bTime)
	}
	return a, b, found
}

// find first 2 dots pair which has almost equal H value
func (k *K) FindAlmostEqualHigh(equalFactor float64) (a, b int, found bool) {
	tmpK := k.Clone()
	tmpK.SortHighReverse()

	found = false
	aTime := time.Time{}
	bTime := time.Time{}
	for i := 1; i < tmpK.Len(); i++ {
		min := decimals.Min(tmpK.Balances[i-1].H, tmpK.Balances[i].H)
		max := decimals.Max(tmpK.Balances[i-1].H, tmpK.Balances[i].H)
		diffFactor := max.Sub(min).Div(min)
		if diffFactor.LessThanOrEqualFloat64(equalFactor) {
			found = true
			aTime = tmpK.Balances[i-1].T
			bTime = tmpK.Balances[i].T
			break
		}
	}
	if found {
		a, _, _ = k.IndexEqual(aTime)
		b, _, _ = k.IndexEqual(bTime)
	}
	return a, b, found
}

func (k *K) FindFirstBreakthroughLow(from int, cmpValue, breakthroughFactor float64) (int, bool) {
	for i := from; i < k.Len(); i++ {
		if k.Balances[i].L.Float64()/cmpValue > 1+breakthroughFactor {
			return i, true
		}
	}
	return -1, false
}

func (k *K) FindFirstBreakthroughHigh(from int, cmpValue, breakthroughFactor float64) (int, bool) {
	for i := from; i < k.Len(); i++ {
		if k.Balances[i].H.Float64()/cmpValue > 1+breakthroughFactor {
			return i, true
		}
	}
	return -1, false
}

func (k *K) ReverseFindFirstBreakthroughLow(from int, cmpValue, breakthroughFactor float64) (int, bool) {
	if from > k.Len()-1 {
		from = k.Len() - 1
	}
	for i := from; i >= 0; i-- {
		if k.Balances[i].L.Float64()/cmpValue > 1+breakthroughFactor {
			return i, true
		}
	}
	return -1, false
}

func (k *K) ReverseFirstBreakthroughHigh(from int, cmpValue, breakthroughFactor float64) (int, bool) {
	if from > k.Len()-1 {
		from = k.Len() - 1
	}
	for i := from; i >= 0; i-- {
		if k.Balances[i].H.Float64()/cmpValue > 1+breakthroughFactor {
			return i, true
		}
	}
	return -1, false
}*/

// 这里有很多形态
// https://cn.investing.com/technical/candlestick-patterns
// https://www.feedroll.com/candlestick-patterns
