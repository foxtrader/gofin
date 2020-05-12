package fintypes

import (
	"encoding/json"
	"fmt"
	"github.com/foxtrader/gofin/ta"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/container/gnum"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"github.com/shawnwyckoff/gopkg/container/gternary"
	"github.com/shawnwyckoff/gopkg/encoding/gcolor"
	"github.com/shawnwyckoff/gopkg/net/ghtml"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"sort"
	"strings"
	"time"
)

/*
如何判断是否收线？
分钟线可以在ToPeriod后删除末尾的点
yahoo finance提供的日线则需要通过系统时间判断
*/

// TODO 删减接口，所有重要接口增加测试用例
// TODO 将ABC/BTC,BTC/USDC这样的交易对转换成虚拟的ABC/USDT

const (
	ExprOpen   = "O"
	ExprHigh   = "H"
	ExprLow    = "L"
	ExprClose  = "C"
	ExprVolume = "V"
)

var (
	KlineDeepCopy  = false // if set it true, will cost about 2.5x time than false
	KInternalExprs = []string{ExprOpen, ExprHigh, ExprLow, ExprClose, ExprVolume}
)

type (
	Dot struct {
		T time.Time
		V float64
	}
	Bar struct {
		T time.Time        `json:"T" bson:"_id" csv:"T"`                             // statistic begin time
		O gdecimal.Decimal `json:"O,omitempty" bson:"O,omitempty" csv:"O,omitempty"` // open price in USD
		L gdecimal.Decimal `json:"L,omitempty" bson:"L,omitempty" csv:"L,omitempty"` // lowest price
		H gdecimal.Decimal `json:"H,omitempty" bson:"H,omitempty" csv:"H,omitempty"` // highest price
		C gdecimal.Decimal `json:"C,omitempty" bson:"C,omitempty" csv:"C,omitempty"` // close price
		V gdecimal.Decimal `json:"V,omitempty" bson:"V,omitempty" csv:"V,omitempty"` // volume in unit asset, always, it is BTC in BTC/USDT pair.
		//VQ         *gdecimal.Decimal `json:"VQ,omitempty" bson:"VQ,omitempty" csv:"VQ,omitempty"` // volume in quote asset
		Indicators map[string]float64
	}

	// Bar items for Global
	Kline struct {
		Pair   PairIMP // 原始K线的周期
		Items  []Bar   // 原始K线数据
		sorted bool    // 原始K线是否排序

		// 用于PT, 缓存转换周期的数据
		//originUpdateTime time.Time         // 原始K线的更新时间, 这个标记是用在PT中的，所以只有在PT下更新时才需要更新次字段
		//convertKs        map[string]*Kline // 根据原始K线转换后的生成K线
	}

	KlineConverter struct {
		base      *Kline
		converted map[Period]*Kline
	}

	doubleVP struct {
		a *Bar
		b *Bar
	}

	doubleVPList []doubleVP
)

// TODO: 需要吗
func (d *Bar) ExprValue(expr string) (gdecimal.Decimal, bool) {
	expr = strings.ToUpper(expr)
	switch expr {
	case ExprOpen:
		return d.O, true
	case ExprLow:
		return d.L, true
	case ExprHigh:
		return d.H, true
	case ExprClose:
		return d.C, true
	case ExprVolume:
		return d.V, true
	}
	val, ok := d.Indicators[expr]
	if !ok {
		return gdecimal.Zero, false
	}
	return gdecimal.NewFromFloat64(val), true
}

func (d *Bar) ExprValueFloat64(expr string) (float64, bool) {
	upperExpr := strings.ToUpper(expr)
	switch upperExpr {
	case ExprOpen:
		return d.O.Float64(), true
	case ExprLow:
		return d.L.Float64(), true
	case ExprHigh:
		return d.H.Float64(), true
	case ExprClose:
		return d.C.Float64(), true
	case ExprVolume:
		return d.V.Float64(), true
	}
	val, ok := d.Indicators[expr]
	return val, ok
}

// TODO: 需要吗
func (d *Bar) IsPositiveLine() bool {
	return d.C.GreaterThan(d.O)
}

// TODO: 需要吗
func (d *Bar) IsNegativeLine() bool {
	return d.C.LessThan(d.O)
}

// TODO: 需要吗
func (d *Bar) IsCrossLine() bool {
	return d.C == d.O && d.H != d.L
}

/*
// feature value: +-[0, 100]
func (d *Bar) lhs() LHS {
	return LHS{Form: d.Form()}
}

func (d *Bar) Form() KForm {
	if !d.O.IsPositive() || !d.H.IsPositive() || !d.L.IsPositive() || !d.C.IsPositive() {
		return KFormUnknown
	}
	total := d.H.Sub(d.L)
	if total.IsZero() {
		return KFormOne
	}

	// 计算实体线比例（涨跌幅）
	entityScale := d.C.Sub(d.O).Div(d.C)

	// 计算下影线比例
	downShadowScale := gdecimal.Min(d.O, d.C).Sub(d.L).Div(d.C)

	// 计算上影线比例
	upperShadowScale := d.H.Sub(gdecimal.Max(d.O, d.C)).Div(d.C)

	// 0:zero, 1:mini, 2:mid, 3:big
	classify := func(v float64) int {
		v = math.Abs(v)
		r := 0
		if v < 0.0003 {
			r = 0
		} else if v < 0.01 {
			r = 1
		} else if v < 0.06 {
			r = 2
		} else {
			r = 3
		}
		return r
	}

	entityClass := classify(entityScale.Float64())
	downShadowClass := classify(downShadowScale.Float64())
	upperShadowClass := classify(upperShadowScale.Float64())
	r := upperShadowClass*100 + entityClass*10 + downShadowClass
	if d.C.LessThan(d.O) {
		r = -r
	}
	return KForm(r)

}*/

func NewKline(pair PairIMP, dots []Bar) *Kline {
	k := &Kline{
		Pair: pair,
	}

	if dots != nil && len(dots) > 0 {
		k.Items = dots
		k.Sort()
	}
	return k
}

func NewAndCopyBasicInfo(src *Kline) *Kline {
	return &Kline{
		Pair: src.Pair,
	}
}

// implement TimeSeries and Series(in github.com/iwat) interface
func (k *Kline) Time(i int) time.Time { return k.Items[i].T }
func (k *Kline) High(i int) float64   { return k.Items[i].H.Float64() }
func (k *Kline) Open(i int) float64   { return k.Items[i].O.Float64() }
func (k *Kline) Close(i int) float64  { return k.Items[i].C.Float64() }
func (k *Kline) Low(i int) float64    { return k.Items[i].L.Float64() }
func (k *Kline) Volume(i int) float64 { return k.Items[i].V.Float64() }

func (k *Kline) HighRaw(i int) gdecimal.Decimal   { return k.Items[i].H }
func (k *Kline) OpenRaw(i int) gdecimal.Decimal   { return k.Items[i].O }
func (k *Kline) CloseRaw(i int) gdecimal.Decimal  { return k.Items[i].C }
func (k *Kline) LowRaw(i int) gdecimal.Decimal    { return k.Items[i].L }
func (k *Kline) VolumeRaw(i int) gdecimal.Decimal { return k.Items[i].V }

func (k *Kline) Len() int           { return len(k.Items) }
func (k *Kline) Less(i, j int) bool { return k.Items[i].T.Before(k.Items[j].T) }
func (k *Kline) Swap(i, j int)      { k.Items[i], k.Items[j] = k.Items[j], k.Items[i] }

type klineLowSlice Kline

func (ks *klineLowSlice) Len() int { return len((*Kline)(ks).Items) }
func (ks *klineLowSlice) Less(i, j int) bool {
	return (*Kline)(ks).Items[i].L.LessThan((*Kline)(ks).Items[j].L)
}
func (ks *klineLowSlice) Swap(i, j int) {
	(*Kline)(ks).Items[i], (*Kline)(ks).Items[j] = (*Kline)(ks).Items[j], (*Kline)(ks).Items[i]
}

type klineHighSlice Kline

func (ks *klineHighSlice) Len() int { return len((*Kline)(ks).Items) }
func (ks *klineHighSlice) Less(i, j int) bool {
	return (*Kline)(ks).Items[i].H.LessThan((*Kline)(ks).Items[j].H)
}
func (ks *klineHighSlice) Swap(i, j int) {
	(*Kline)(ks).Items[i], (*Kline)(ks).Items[j] = (*Kline)(ks).Items[j], (*Kline)(ks).Items[i]
}

type klineOpenSlice Kline

func (ks *klineOpenSlice) Len() int { return len((*Kline)(ks).Items) }
func (ks *klineOpenSlice) Less(i, j int) bool {
	return (*Kline)(ks).Items[i].O.LessThan((*Kline)(ks).Items[j].O)
}
func (ks *klineOpenSlice) Swap(i, j int) {
	(*Kline)(ks).Items[i], (*Kline)(ks).Items[j] = (*Kline)(ks).Items[j], (*Kline)(ks).Items[i]
}

type klineCloseSlice Kline

func (ks *klineCloseSlice) Len() int { return len((*Kline)(ks).Items) }
func (ks *klineCloseSlice) Less(i, j int) bool {
	return (*Kline)(ks).Items[i].C.LessThan((*Kline)(ks).Items[j].C)
}
func (ks *klineCloseSlice) Swap(i, j int) {
	(*Kline)(ks).Items[i], (*Kline)(ks).Items[j] = (*Kline)(ks).Items[j], (*Kline)(ks).Items[i]
}

func (k *Kline) MaxIndex() int { return k.Len() - 1 }

func (k *Kline) String() string {
	buf, err := json.Marshal(k)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (k *Kline) Sort() {
	if k.sorted {
		return
	}

	sort.Sort(k)
	k.sorted = true
}

func (k *Kline) SortLow() {
	sort.Sort((*klineLowSlice)(k))
	k.sorted = false
}

func (k *Kline) SortLowReverse() {
	sort.Sort(sort.Reverse((*klineLowSlice)(k)))
	k.sorted = false
}

func (k *Kline) SortHigh() {
	sort.Sort((*klineHighSlice)(k))
	k.sorted = false
}

func (k *Kline) SortHighReverse() {
	sort.Sort(sort.Reverse((*klineHighSlice)(k)))
	k.sorted = false
}

func (k *Kline) SortOpen() {
	sort.Sort((*klineOpenSlice)(k))
	k.sorted = false
}

func (k *Kline) SortOpenReverse() {
	sort.Sort(sort.Reverse((*klineOpenSlice)(k)))
	k.sorted = false
}

func (k *Kline) SortClose() {
	sort.Sort((*klineCloseSlice)(k))
	k.sorted = false
}

func (k *Kline) SortCloseReverse() {
	sort.Sort(sort.Reverse((*klineCloseSlice)(k)))
	k.sorted = false
}

// TODO 用sort.Search改造？
// 找出在指定时间tm之前，且与tm时间间隔最小的Kline条目
func (k *Kline) BeforeEqualClosest(tm time.Time) (item *Bar, exists bool) {
	index := -1
	for i, v := range k.Items {
		if v.T.Before(tm) || v.T.Equal(tm) {
			index = i
		} else {
			break
		}
	}

	if index < 0 {
		return nil, false
	}
	return &k.Items[index], true
}

// search min index whose T is >= t, index whose T == t is included
// if result <= k.Len() - 1 && k.items[result].T == t, there is a dot whose time is t
// otherwise not found, but result still means which index is mostly close to t
func (k *Kline) MinIndexGTE(t time.Time) int {
	k.Sort()
	idx := sort.Search(k.Len(), func(i int) bool {
		return gtime.AfterEqual(k.Items[i].T, t)
	})
	if idx > k.Len()-1 {
		return -1
	}
	return idx
}

// search min index whose T > t, index whose T == t is NOT included
func (k *Kline) MinIndexGT(t time.Time) int {
	k.Sort()
	idx := sort.Search(k.Len(), func(i int) bool {
		return k.Items[i].T.After(t)
	})
	if idx > k.Len()-1 {
		return -1
	}
	return idx
}

// TODO test required
// search max index whose T <= t, index whose T == t is included
func (k *Kline) MaxIndexLTE(t time.Time) int {
	idx := k.MinIndexGTE(t)
	if idx == -1 {
		return k.MaxIndex()
	} else {
		if k.Items[idx].T.Equal(t) {
			return idx
		} else {
			return idx - 1
		}
	}
}

// search max index whose T < t, index whose T == t is NOT included
func (k *Kline) MaxIndexLT(t time.Time) int {
	idx := k.MinIndexGTE(t)
	if idx == -1 {
		return k.MaxIndex()
	} else {
		return idx - 1
	}
}

// search index whose T == t
func (k *Kline) IndexEqual(t time.Time) (int, Bar, bool) {
	i := k.MinIndexGTE(t)
	if i >= 0 && i <= k.Len()-1 && k.Items[i].T.Equal(t) {
		return i, k.Items[i], true
	}
	return -1, Bar{}, false
}

func (k *Kline) HasTime(tm time.Time) bool {
	_, _, has := k.IndexEqual(tm)
	return has
}

// TODO test required
// fixme 有问题
func (k *Kline) SliceBetween(gt, lt time.Time) *Kline {
	k.Sort()
	r := NewAndCopyBasicInfo(k)
	r.sorted = true // performance related

	gtIdx := k.MinIndexGT(gt)
	if gtIdx < 0 || gtIdx >= k.Len() {
		return r
	}
	ltIdx := k.MaxIndexLT(lt)
	if ltIdx < 0 || ltIdx >= k.Len() {
		return r
	}
	if ltIdx < gtIdx {
		return r
	}
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[gtIdx:ltIdx+1]...) // deep copy
	} else {
		r.Items = k.Items[gtIdx : ltIdx+1]
	}
	return r
}

// TODO test required
func (k *Kline) SliceBetweenEqual(gte, lte time.Time) *Kline {
	k.Sort()
	r := NewAndCopyBasicInfo(k)
	r.sorted = true // performance related

	gteIdx := k.MinIndexGTE(gte)
	if gteIdx < 0 || gteIdx >= k.Len() {
		return r
	}
	lteIdx := k.MaxIndexLTE(lte)
	if lteIdx < 0 || lteIdx >= k.Len() {
		return r
	}
	if lteIdx < gteIdx {
		return r
	}
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[gteIdx:lteIdx+1]...) // deep copy
	} else {
		r.Items = k.Items[gteIdx : lteIdx+1]
	}
	return r
}

// TODO: test required
// select item which time before or equal 'lte'
func (k *Kline) SliceBefore(lt time.Time) *Kline {
	k.Sort()
	r := NewAndCopyBasicInfo(k)
	r.sorted = true // performance related

	milt := k.MaxIndexLT(lt)
	if milt < 0 {
		return r
	}
	if milt >= k.Len()-1 {
		r = k.Clone()
		return r
	}
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[:milt+1]...) // deep copy
	} else {
		r.Items = k.Items[:milt+1]
	}
	return r
}

// TODO: test required
// select item which time before or equal 'lte'
func (k *Kline) SliceBeforeEqual(lte time.Time) *Kline {
	k.Sort()
	r := NewAndCopyBasicInfo(k)
	r.sorted = true // performance related

	milte := k.MaxIndexLTE(lte)
	if milte < 0 {
		return r
	}
	if milte >= k.Len()-1 {
		r = k.Clone()
		return r
	}
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[:milte+1]...) // deep copy
	} else {
		r.Items = k.Items[:milte+1]
	}
	return r
}

// select item which time before 'lte'
func (k *Kline) SliceAfter(gt time.Time) *Kline {
	k.Sort()
	r := NewAndCopyBasicInfo(k)
	r.sorted = true // performance related

	migt := k.MinIndexGT(gt)
	if migt < 0 || migt >= k.Len() {
		return r
	}
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[migt:]...) // deep copy
	} else {
		r.Items = k.Items[migt:]
	}
	return r
}

// select item which time before or equal 'lte'
func (k *Kline) SliceAfterEqual(gte time.Time) *Kline {
	k.Sort()
	r := NewAndCopyBasicInfo(k)
	r.sorted = true // performance related

	migte := k.MinIndexGTE(gte)
	if migte < 0 || migte >= k.Len() {
		return r
	}
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[migte:]...) // deep copy
	} else {
		r.Items = k.Items[migte:]
	}
	return r
}

func (k *Kline) SliceHead(n int) *Kline {
	if n >= k.Len() {
		return k.Clone()
	}

	r := NewAndCopyBasicInfo(k)
	r.sorted = k.sorted // performance related
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[0:n-1]...) // deep copy
	} else {
		r.Items = k.Items[0 : n-1]
	}
	return r
}

func (k *Kline) SliceTail(n int) *Kline {
	if n >= k.Len() {
		return k.Clone()
	}

	r := NewAndCopyBasicInfo(k)
	r.sorted = k.sorted // performance related
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[k.Len()-n:]...) // deep copy
	} else {
		r.Items = k.Items[k.Len()-n:]
	}
	return r
}

func (k *Kline) SliceBetweenId(gte, lte int) *Kline {
	if gte > lte || gte > (k.Len()-1) {
		return NewAndCopyBasicInfo(k)
	}

	// fix lte if necessary
	if lte > k.Len()-1 {
		lte = k.Len() - 1
	}

	k.Sort()
	r := NewAndCopyBasicInfo(k)
	r.sorted = true // performance related
	if KlineDeepCopy {
		r.Items = append(r.Items, k.Items[gte:lte+1]...) // deep copy
	} else {
		r.Items = k.Items[gte : lte+1]
	}
	return r
}

func (k *Kline) HasDuplicatedKeys() ([]time.Time, bool) {
	k.Sort()
	// FIXME 不可以用map[time.Time]
	dup := make(map[time.Time]int)
	for i := 0; i < k.Len()-1; i++ {
		if k.Items[i].T.Equal(k.Items[i+1].T) {
			dup[k.Items[i].T] = dup[k.Items[i].T] + 1
		}
	}
	var r []time.Time
	for k := range dup {
		r = append(r, k)
	}
	return r, len(r) > 0
}

func (k *Kline) RemoveTail(n int) *Kline {
	if k.Len() == 0 {
		return k
	}
	k.Items = k.Items[0 : k.Len()-1]
	return k
}

// 计算最后收线时间分界点
// 比如 分钟线[14:59 15:00 15:01]，返回值 15:00
func (k *Kline) LastClosedTime(targetPeriod Period, config PeriodRoundConfig) time.Time {
	if k.Len() == 0 {
		return gtime.ZeroTime
	}
	if targetPeriod.ToDuration() < k.DetectPeriod().ToDuration() {
		return gtime.ZeroTime
	}

	lastTime := k.Items[k.Len()-1].T
	periodBegin := RoundPeriodEarlier(lastTime, targetPeriod, config)
	if periodBegin.Before(k.Items[0].T) || periodBegin.After(k.Items[k.Len()-1].T) {
		return gtime.ZeroTime
	}
	return periodBegin
}

func (k *Kline) Update(time time.Time, itemName string, value gdecimal.Decimal) (updatedCount int) {
	updated := 0
	for i := range k.Items {
		if k.Items[i].T.Equal(time) {
			switch itemName {
			case "O":
				k.Items[i].O = value
				updated++
			case "C":
				k.Items[i].C = value
				updated++
			case "H":
				k.Items[i].H = value
				updated++
			case "L":
				k.Items[i].L = value
				updated++
			case "V":
				k.Items[i].V = value
				updated++
			}
		}
	}
	return updated
}

func (k *Kline) Set(newKs *Kline) {
	*k = *newKs
}

func (k *Kline) Upsert(toUpsert *Kline) {
	var toAdd []Bar
	for _, newDot := range toUpsert.Items {
		if i, _, has := k.IndexEqual(newDot.T); has {
			k.Items[i] = newDot
		} else {
			toAdd = append(toAdd, newDot)
		}
	}

	// append and sort
	if len(toAdd) > 0 {
		k.Items = append(k.Items, toAdd...)
		k.Sort()
	}
}

func (k *Kline) UpsertDot(dot Bar) {
	if i, _, has := k.IndexEqual(dot.T); has {
		k.Items[i] = dot
	} else {
		k.Items = append(k.Items, dot)
		k.Sort()
	}
}

func (k *Kline) FirstDay(tz *time.Location) gtime.Date {
	if tz == nil {
		tz = time.UTC
	}
	return gtime.TimeToDate(k.Items[0].T, tz)
}

func (k *Kline) LastDay(tz *time.Location) gtime.Date {
	if tz == nil {
		tz = time.UTC
	}
	return gtime.TimeToDate(k.Items[k.Len()-1].T, time.UTC)
}

func (k *Kline) AcrossDays() float64 {
	return float64(k.Items[k.Len()-1].T.Sub(k.Items[0].T)) / float64(gtime.Day)
}

// Min time in records
func (k *Kline) FirstTimeEx() (last *time.Time, exists bool) {
	if len(k.Items) == 0 {
		return nil, false
	}

	k.Sort()
	t := k.Items[0].T
	return &t, true
}

// Max time in records
func (k *Kline) LastTimeEx() (last *time.Time, exists bool) {
	if len(k.Items) == 0 {
		return nil, false
	}

	k.Sort()
	t := k.Items[k.Len()-1].T
	return &t, true
}

// First T
func (k *Kline) FirstTime(defIfEmpty time.Time) time.Time {
	if len(k.Items) == 0 {
		return defIfEmpty
	}

	k.Sort()
	return k.Items[0].T
}

// Last T
func (k *Kline) LastTime(defIfEmpty time.Time) time.Time {
	if len(k.Items) == 0 {
		return defIfEmpty
	}

	k.Sort()
	return k.Items[k.Len()-1].T
}

func (k *Kline) First() (Bar, bool) {
	k.Sort()
	if len(k.Items) == 0 {
		return Bar{}, false
	}
	return k.Items[0], true
}

func (k *Kline) Last() (Bar, bool) {
	k.Sort()
	if len(k.Items) == 0 {
		return Bar{}, false
	}
	return k.Items[len(k.Items)-1], true
}

func (k *Kline) LastLow(defIfEmpty gdecimal.Decimal) gdecimal.Decimal {
	k.Sort()
	if len(k.Items) == 0 {
		return defIfEmpty
	}
	return k.Items[len(k.Items)-1].L
}

func (k *Kline) LastClose(defIfEmpty gdecimal.Decimal) gdecimal.Decimal {
	k.Sort()
	if len(k.Items) == 0 {
		return defIfEmpty
	}
	return k.Items[len(k.Items)-1].C
}

func (k *Kline) Times() []time.Time {
	var r []time.Time
	for _, v := range k.Items {
		r = append(r, v.T)
	}
	return r
}

func (k *Kline) Highs() []gdecimal.Decimal {
	var r []gdecimal.Decimal
	for i := range k.Items {
		r = append(r, k.Items[i].H)
	}
	return r
}

func (k *Kline) Lows() []gdecimal.Decimal {
	var r []gdecimal.Decimal
	for i := range k.Items {
		r = append(r, k.Items[i].L)
	}
	return r
}

func (k *Kline) Opens() []gdecimal.Decimal {
	var r []gdecimal.Decimal
	for i := range k.Items {
		r = append(r, k.Items[i].O)
	}
	return r
}

func (k *Kline) Closes() []gdecimal.Decimal {
	var r []gdecimal.Decimal
	for i := range k.Items {
		r = append(r, k.Items[i].C)
	}
	return r
}

func (k *Kline) Volumes() []gdecimal.Decimal {
	var r []gdecimal.Decimal
	for i := range k.Items {
		r = append(r, k.Items[i].V)
	}
	return r
}

func (k *Kline) HighValues() []float64 {
	return gdecimal.ToFloat64s(k.Highs())
}

func (k *Kline) LowValues() []float64 {
	return gdecimal.ToFloat64s(k.Lows())
}

func (k *Kline) OpenValues() []float64 {
	return gdecimal.ToFloat64s(k.Opens())
}

func (k *Kline) CloseValues() []float64 {
	return gdecimal.ToFloat64s(k.Closes())
}

func (k *Kline) VolumeValues() []float64 {
	return gdecimal.ToFloat64s(k.Volumes())
}

func (k *Kline) ToBar() Bar {
	res := Bar{}
	res.T = k.Items[0].T
	res.O = k.Items[0].O
	res.C = k.Items[k.Len()-1].C
	var minL *gdecimal.Decimal = nil
	var maxH *gdecimal.Decimal = nil
	for _, v := range k.Items {
		if minL == nil || v.L.LessThan(*minL) {
			minL = &v.L
		}
		if maxH == nil || v.H.GreaterThan(*maxH) {
			maxH = &v.H
		}
	}
	res.L = *minL
	res.H = *maxH
	return res
}

func (k *Kline) Merge() Bar {
	if k.Len() == 0 {
		return Bar{}
	}
	k.Sort()
	r := Bar{}
	r.T = k.Items[0].T
	r.O = k.Items[0].O
	r.L = k.Items[0].L
	r.H = k.Items[0].H
	r.C = k.Items[k.Len()-1].C
	for i := 0; i < k.Len(); i++ {
		r.L = gdecimal.Min(r.L, k.Items[i].L)
		r.H = gdecimal.Max(r.H, k.Items[i].H)
		r.V = r.V.Add(k.Items[i].V)
	}

	return r
}

// todo: test required
// note: 使用 nk := *k; return nk  这样是无法深度拷贝k的
func (k *Kline) Clone() *Kline {
	if KlineDeepCopy {
		r := NewAndCopyBasicInfo(k) // performance related
		r.sorted = k.sorted
		r.Items = append(r.Items, k.Items...) // deep copy
		return r
	} else {
		r := new(Kline)
		*r = *k
		return r
	}
}

// 转换周期，但不缓存转换之后的数据
// NOTE: 如果交易所中间维护，中间空缺的数据会被忽略，而不是填充空数据
// FIXME: 如果最后一个周期数据尚未Close，那最后一根线会画出来吗？
func (k *Kline) ToPeriod(newPeriod Period, config PeriodRoundConfig) (*Kline, error) {
	if k.Len() == 0 {
		newK := NewAndCopyBasicInfo(k)
		//newK.convertKs = nil
		return newK, nil
	}

	if err := k.Pair.Verify(); err != nil {
		return nil, err
	}
	if k.Pair.I() == newPeriod {
		return k, nil
	}
	if newPeriod.ToSeconds() < k.Pair.I().ToSeconds() {
		return nil, errors.Errorf("can't convert period of K from %s to %s", k.Pair.I().String(), newPeriod.String())
	}

	k.Sort()

	r := NewAndCopyBasicInfo(k)
	r.Pair = r.Pair.SetI(newPeriod) // note: don't forget this
	r.sorted = true
	//r.convertKs = nil

	// transfer to cache
	// WARNING: map[time.Time] is dangerous, DON'T use it, use map[int64] instead
	data := make(map[int64]*Kline)
	for _, v := range k.Items {
		newPeriodOpen := RoundPeriodEarlier(v.T, newPeriod, config)
		if data[newPeriodOpen.UnixNano()] == nil {
			data[newPeriodOpen.UnixNano()] = new(Kline)
		}
		data[newPeriodOpen.UnixNano()].UpsertDot(v)
	}

	// sort times
	var times []time.Time
	for k := range data {
		times = append(times, gtime.UnixNanoToTime(k, &config.Location))
	}
	times = gtime.SortTimes(times)

	// merge dots
	for _, tm := range times {
		newDot := data[tm.UnixNano()].Merge()
		newDot.T = tm
		r.Items = append(r.Items, newDot) // append
	}

	return r, nil
}

// 局部周期转换，只转换lastConverted的最后一个时间（包含）往后的数据
// lastConverted is input and output config
// TODO: test required
func (k *Kline) ToPeriodPartly(lastConverted *Kline, newPeriod Period, config PeriodRoundConfig) error {
	newK, err := k.SliceAfterEqual(lastConverted.Items[len(lastConverted.Items)-1].T).ToPeriod(newPeriod, config)
	if err != nil {
		return err
	}
	//fmt.Println("partly upsert", newK.Len(), k.LastTime(gtime.ZeroTime), newK.LastTime(gtime.ZeroTime))
	lastConverted.Upsert(newK)
	return nil
}

// just convert last tail dots in newPeriod
func (k *Kline) ToPeriodWithTail(newPeriod Period, config PeriodRoundConfig, tail int) (*Kline, error) {
	if k.Len() == 0 {
		newK := NewAndCopyBasicInfo(k)
		return newK, nil
	}

	if err := k.Pair.Verify(); err != nil {
		return nil, err
	}
	if k.Pair.I() == newPeriod {
		return k.SliceTail(tail), nil
	}
	if newPeriod.ToSeconds() < k.Pair.I().ToSeconds() {
		return nil, errors.Errorf("can't convert period of K from %s to %s", k.Pair.I().String(), newPeriod.String())
	}

	k.Sort()

	r := NewAndCopyBasicInfo(k)
	r.Pair = r.Pair.SetI(newPeriod) // note: don't forget this
	r.sorted = true

	// transfer to cache
	// FIXME 不可以用map[time.Time]
	data := make(map[time.Time]*Kline)
	for i := len(k.Items) - 1; i >= 0; i-- {
		newPeriodOpen := RoundPeriodEarlier(k.Items[i].T, newPeriod, config)
		if data[newPeriodOpen] == nil {
			data[newPeriodOpen] = new(Kline)
		}
		if len(data) == tail {
			if _, exist := data[newPeriodOpen]; !exist { // dot before tail
				break
			}
		}
		data[newPeriodOpen].UpsertDot(k.Items[i])
	}

	// sort times
	var times []time.Time
	for k := range data {
		times = append(times, k)
	}
	times = gtime.SortTimes(times)

	// merge dots
	for _, t := range times {
		newDot := data[t].Merge()
		newDot.T = t
		r.Items = append(r.Items, newDot) // append
	}

	return r, nil
}

// compare data of same timestamp items, time different items between them are ignored
func (k *Kline) IsTimeOverlappingAreaEqual(cmp *Kline) (equal bool) {
	k.Sort()
	cmp.Sort()

	exist, begin, end := k.TimeOverlappingArea(cmp)
	if !exist {
		return true
	}

	ks_s := k.SliceBetweenEqual(begin, end)
	cmp_s := cmp.SliceBetweenEqual(begin, end)
	if ks_s.Len() != cmp_s.Len() {
		fmt.Printf("len not equal")
		return false
	}

	for i := range ks_s.Items {
		if !ks_s.Items[i].T.Equal(cmp_s.Items[i].T) ||
			!ks_s.Items[i].O.Equal(cmp_s.Items[i].O) ||
			!ks_s.Items[i].H.Equal(cmp_s.Items[i].H) ||
			!ks_s.Items[i].L.Equal(cmp_s.Items[i].L) ||
			!ks_s.Items[i].C.Equal(cmp_s.Items[i].C) ||
			!ks_s.Items[i].V.Equal(cmp_s.Items[i].V) { /* ||
			!ks_s.items[i].MarketCap.Equal(cmp_s.items[i].MarketCap)*/
			//fmt.Println(i, ks_s.Items[i], cmp_s.Items[i])
			//fmt.Println(ks_s.Items[i].T.UTC().String(), "value not equal")
			return false
		}
	}

	return true
}

func (k *Kline) CleanHMS(loc *time.Location) *Kline {
	if loc == nil {
		loc = time.UTC
	}
	r := NewAndCopyBasicInfo(k)
	r.sorted = k.sorted // performance related
	for _, v := range k.Items {
		v.T = gtime.TimeToDate(v.T, loc).ToTimeLocation(loc)
		r.Items = append(r.Items, v)
	}
	/*
		for i := 0; i < len(k.items); i++ {
			k.items[i].T = xclock.TimeToDate(k.items[i].T, time.UTC).ToTime(0, 0, 0, 0, time.UTC)
		}*/
	return r
}

func (k *Kline) IsConsequentDayLine() bool {
	for i := range k.Items {
		if i > 0 {
			if k.Items[i].T.Sub(k.Items[i-1].T).Hours() != 24 {
				return false
			}
		}
	}
	return true
}

func (k *Kline) DetectPeriod() Period {
	if k.Len() <= 1 {
		return PeriodError
	}
	return DurationToPeriod(k.Items[1].T.Sub(k.Items[0].T))
}

func (k *Kline) TimeOverlappingArea(cmp *Kline) (exist bool, begin, end time.Time) {
	if k.Len() == 0 ||
		cmp.Len() == 0 ||
		k.FirstTime(gtime.ZeroTime).After(cmp.LastTime(gtime.ZeroTime)) ||
		k.LastTime(gtime.ZeroTime).Before(cmp.FirstTime(gtime.ZeroTime)) {
		return false, gtime.ZeroTime, gtime.ZeroTime
	}

	begin = gtime.MaxTime(k.FirstTime(gtime.ZeroTime), cmp.FirstTime(gtime.ZeroTime))
	end = gtime.MinTime(k.LastTime(gtime.ZeroTime), cmp.LastTime(gtime.ZeroTime))
	return true, begin, end
}

func (k *Kline) SetIndicatorValue(indExpr ta.IndExpr, subItemName string, values []float64) error {
	s := string(indExpr)
	if subItemName != "" {
		s = s + "." + subItemName
	}

	if len(values) != k.Len() {
		return errors.Errorf("indicator %s length %d != kline length %d", s, len(values), k.Len())
	}
	for i := 0; i < k.Len(); i++ {
		if k.Items[i].Indicators == nil {
			k.Items[i].Indicators = make(map[string]float64)
		}
		k.Items[i].Indicators[s] = values[i]

		/*if i == k.Len()-1 {
			fmt.Println("setIndVal", k.Pair.String(), k.FirstTime(gtime.ZeroTime), k.LastTime(gtime.ZeroTime), i, k.Items[i].T.String(), values[i])
		}*/
	}
	return nil
}

func (k *Kline) IndicatorExps() []string {
	if k.Len() == 0 {
		return nil
	}
	if len(k.Items[0].Indicators) == 0 {
		return nil
	}
	var r []string
	for key := range k.Items[0].Indicators {
		r = append(r, key)
	}
	return r
}

func (k *Kline) ExprValues(expr string) ([]float64, error) {
	// get internal raw data expression values
	expr = strings.ToUpper(expr)
	switch expr {
	case ExprOpen:
		return k.OpenValues(), nil
	case ExprLow:
		return k.LowValues(), nil
	case ExprHigh:
		return k.HighValues(), nil
	case ExprClose:
		return k.CloseValues(), nil
	case ExprVolume:
		return k.VolumeValues(), nil
	}

	// get indicator expression values
	allExps := k.IndicatorExps()
	if !gstring.Contains(allExps, expr) {
		return nil, errors.Errorf("%s indicator expr not found", expr)
	}
	var r []float64
	for i := 0; i < k.Len(); i++ {
		if k.Items[i].Indicators == nil {
			return nil, errors.Errorf("indicators of Bar index(%d) T(%s) is null, total size %d, max time %s", i, k.Items[i].T.UTC().String(), k.Len(), k.LastTime(gtime.ZeroTime))
		}
		val, ok := k.Items[i].Indicators[expr]
		if !ok {
			return nil, errors.Errorf(`indicators "%s" of Bar index(%d) T(%s) not exist`, expr, i, k.Items[i].T.UTC().String())
		}
		r = append(r, val)
	}
	return r, nil
}

func (k *Kline) Exprs() []string {
	if k.Len() == 0 {
		//panic(errors.Errorf("empty k"))
		return nil
	}

	r := KInternalExprs
	if len(k.Items[0].Indicators) == 0 {
		//panic(errors.Errorf("empty indicators"))
		return r
	}
	for key := range k.Items[0].Indicators {
		if key == "" {
			continue
		}
		r = append(r, key)
	}
	return r
}

func (k *Kline) LastExprs() []string {
	if k.Len() == 0 {
		//panic(errors.Errorf("empty k"))
		return nil
	}

	r := KInternalExprs
	if len(k.Items[k.Len()-1].Indicators) == 0 {
		//panic(errors.Errorf("empty indicators"))
		return r
	}
	for key := range k.Items[k.Len()-1].Indicators {
		if key == "" {
			continue
		}
		r = append(r, key)
	}
	return r
}

// TODO: test required
// 转换为chart模板，方便绘图
func (k *Kline) ToChartTemplate(indExprs ...string) (*ghtml.ChartTemplate, error) {
	tpl := &ghtml.ChartTemplate{}

	times := k.Times()
	layout := gtime.DetectBestLayoutRaw(times)
	eTimes := gtime.NewElegantTimeArray(times, layout)

	tpl.Title = k.Pair.String()
	tpl.Times = eTimes

	// K线绘图区
	candleSeries := ghtml.Series{}
	candleSeries.CandleStick = &ghtml.CandleStick{Name: "OHLC"}
	candleSeries.CandleStick.Ohlc[0] = gdecimal.ToElegantFloat64s(k.Opens())
	candleSeries.CandleStick.Ohlc[1] = gdecimal.ToElegantFloat64s(k.Highs())
	candleSeries.CandleStick.Ohlc[2] = gdecimal.ToElegantFloat64s(k.Lows())
	candleSeries.CandleStick.Ohlc[3] = gdecimal.ToElegantFloat64s(k.Closes())
	candleSeries.CandleStick.UpColor = gcolor.Green
	candleSeries.CandleStick.DownColor = gcolor.Red

	// 成交量绘图区
	volumeSeries := ghtml.Series{}
	volumeSeries.Name = "   " + "V"
	volumeSeries.BarStick = &ghtml.Bar{Name: "V"}
	volumeSeries.BarStick.Data = gdecimal.ToElegantFloat64s(k.Volumes())
	volumeSeries.BarStick.Colors = make([]gcolor.Color, k.Len())
	for i := 0; i < k.Len(); i++ {
		if candleSeries.CandleStick.Ohlc[3][i].Raw() > candleSeries.CandleStick.Ohlc[0][i].Raw() { // 阳线
			volumeSeries.BarStick.Colors[i] = gcolor.Green
		} else { // 阴线
			volumeSeries.BarStick.Colors[i] = gcolor.Red
		}
	}

	// 独立绘图区
	uniqueSeries := map[ta.IndExpr]*ghtml.Series{}

	// 其他指标绘图区
	maxPrec := gdecimal.MaxPrec(k.Opens()) // 指数的精度是不明确的，使用价格的精度再增加2位，足够了
	maxPrec += 2
	for _, ie := range indExprs {
		ind, _, err := ta.IndExpr(ie).Parse()
		if err != nil {
			return nil, err
		}
		series := ind.Series()
		if series == ta.IndSeriesUnknown {
			return nil, errors.Errorf("can't find series for IndExpr(%s)", ie)
		}

		eoie := k.ExprsOfIndExpr(ta.IndExpr(ie))
		if len(eoie) == 0 {
			continue
		}

		for _, valExpr := range eoie {
			vals, err := k.ExprValues(valExpr)
			if err != nil {
				return nil, err
			}

			style, err := ta.ValExprStyle(valExpr)
			if err != nil {
				return nil, err
			}
			if style == ta.IndStyleUnknown {
				return nil, errors.Errorf("can't find style for IndExpr(%s)", ie)
			}

			if style == ta.IndStyleLine {
				line := ghtml.Line{Name: valExpr}
				line.Data = gnum.NewElegantFloatPtrArray3(vals, maxPrec, 0) // 0值不画线
				line.Color = gcolor.RandomColor()

				if series == ta.IndSeriesPrice {
					candleSeries.Lines = append(candleSeries.Lines, line)
				} else if series == ta.IndSeriesVolume {
					volumeSeries.Lines = append(volumeSeries.Lines, line)
				} else if series == ta.IndSeriesUnique {
					us, ok := uniqueSeries[ta.IndExpr(ie)]
					if !ok {
						us = &ghtml.Series{}
						us.Name = "   " + ie
					}
					us.Lines = append(us.Lines, line)
					if !ok {
						uniqueSeries[ta.IndExpr(ie)] = us
					}
				}

			} else if style == ta.IndStyleBar {
				bar := &ghtml.Bar{Name: valExpr}
				bar.Data = gnum.NewElegantFloatArray(vals, maxPrec)
				bar.Colors = make([]gcolor.Color, k.Len())
				for i := 0; i < k.Len(); i++ {
					if vals[i] > 0 {

					}
					bar.Colors[i] = gternary.If(vals[i] > 0).Interface(gcolor.Green, gcolor.Red).(gcolor.Color) //color.RandomColor()
				}

				if series == ta.IndSeriesPrice {
					candleSeries.BarStick = bar
				} else if series == ta.IndSeriesVolume {
					volumeSeries.BarStick = bar
				} else if series == ta.IndSeriesUnique {
					us, ok := uniqueSeries[ta.IndExpr(ie)]
					if !ok {
						us = &ghtml.Series{}
						us.Name = "   " + ie
					}
					us.BarStick = bar
					if !ok {
						uniqueSeries[ta.IndExpr(ie)] = us
					}
				}
			}
		}
	}

	tpl.Series = append(tpl.Series, candleSeries)
	tpl.Series = append(tpl.Series, volumeSeries)
	for _, v := range uniqueSeries {
		tpl.Series = append(tpl.Series, *v)
	}
	return tpl, nil
}

func (k *Kline) HasExpr(expr string) bool {
	expr = strings.ToUpper(expr)
	if gstring.Contains(KInternalExprs, expr) {
		return true
	}

	lastDot, ok := k.Last()
	if !ok {
		return false
	}
	_, ok = lastDot.Indicators[expr]
	return ok
}

func (k *Kline) TA() *KTA {
	return (*KTA)(k)
}

/*
func (k *Kline) PT() *goquantcore.PT {
	res := &goquantcore.PT{}
	res.originPeriod = k.Pair.I()
	res.originKs = k
	if res.convertKs == nil {
		res.convertKs = map[string]*Kline{}
		res.convertKsVer = map[string]time.Time{}
	}
	return res
}*/

// 是否有指标表达式的值，比如Kline中有MACD(12,9,3).Mid，ie=MACD(12,9,3)，那么就返回true
func (k *Kline) ExprsOfIndExpr(ie ta.IndExpr) []string {
	allExprs := k.Exprs()
	var r []string
	for _, expr := range allExprs {
		if gstring.StartWith(strings.ToUpper(expr), strings.ToUpper(ie.String())) {
			r = append(r, expr)
		}
	}
	return r
}

func (dkl doubleVPList) Len() int {
	return len([]doubleVP(dkl))
}

func (dkl doubleVPList) Less(i, j int) bool {
	return []doubleVP(dkl)[i].a.T.Before([]doubleVP(dkl)[j].a.T)
}

func (dkl doubleVPList) Swap(i, j int) {
	[]doubleVP(dkl)[i], []doubleVP(dkl)[j] = []doubleVP(dkl)[j], []doubleVP(dkl)[i]
}

func SyncXAxis(ks1, ks2 *Kline, fillIfNotExist Bar) {
	// WARNING: don't use map[time.Time]string, because time.Time is a struct, one member(time.Location) is a pointer
	tmpmap := make(map[int64][2]*Bar)
	for i := range ks1.Items {
		item := tmpmap[ks1.Items[i].T.Unix()]
		item[0] = &ks1.Items[i]
		tmpmap[ks1.Items[i].T.Unix()] = item
	}
	for i := range ks2.Items {
		item := tmpmap[ks2.Items[i].T.Unix()]
		item[1] = &ks2.Items[i]
		tmpmap[ks2.Items[i].T.Unix()] = item
	}

	for k := range tmpmap {
		if tmpmap[k][0] == nil {
			tofill := fillIfNotExist
			tofill.T = gtime.EpochSecToTime(k)
			item := tmpmap[k]
			item[0] = &tofill
			tmpmap[k] = item
		}
		if tmpmap[k][1] == nil {
			tofill := fillIfNotExist
			tofill.T = gtime.EpochSecToTime(k)
			item := tmpmap[k]
			item[1] = &tofill
			tmpmap[k] = item
		}
	}

	dkl := doubleVPList{}
	for _, v := range tmpmap {
		dkl = append(dkl, doubleVP{a: v[0], b: v[1]})
	}
	sort.Sort(dkl)

	ks1.Items = nil
	ks2.Items = nil
	for i := range []doubleVP(dkl) {
		ks1.Items = append(ks1.Items, *[]doubleVP(dkl)[i].a)
		ks2.Items = append(ks2.Items, *[]doubleVP(dkl)[i].b)
	}
}
