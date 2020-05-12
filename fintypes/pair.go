package fintypes

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"sort"
	"strings"
)

/*
Pair ISO standard format: Unit/Quote
Unit: Base Currency / Unit Currency (amount is 1 when exchange these two currencies)
Quote: Quote Currency / Pricing Currency / Secondary Currency
Example: USD/CNY = 6.6
means 1USD = 6.6CNY, 以USD为基准或者1个单位，需要6.6CNY的报价，所以USD是基准货币，CNY是报价货币

Notice:
ok、binance、hitbtc... has correct Unit / Quote order, but Bittrex is not, it's order is quote/unit.
*/

// TODO test pairExt
type (
	pairExt string // any format pair: BTC/USD.1min.swap.okex BTCG2020/USD.1min.future.CME BTC/USDT.1min.spot.binance BTC/USDT.1day.sse BTC/USDT.binance
	Pair    string // BTC/USDT
	PairI   string // Pair with Interval
	PairM   string // Pair with Market
	PairP   string // Pair with Platform
	PairIM  string // Pair with Interval & Market
	PairMP  string // Pair with Market & Platform
	PairIP  string // Pair with Interval & Platform
	PairIMP string // Pair with Interval, Market & Platform
)

const (
	PairErr          = Pair("")
	pairExtErr       = pairExt("")
	PairIErr         = PairI("")
	PairMErr         = PairM("")
	PairPErr         = PairP("")
	PairIMErr        = PairIM("")
	PairMPErr        = PairMP("")
	PairIPErr        = PairIP("")
	PairIMPErr       = PairIMP("")
	pairExtDelimiter = "."
)

var (
	peCache             = map[pairExt]pairExt{}
	commonPairDelimiter = []string{"/", "-", "_", "|", "+", ":", "#", ",", ".", "\\", "*"}
)

func parsePairWithOptions(s string, delimiters []string, leftTailDelimiters, rightHeaderDelimiters []string, ISOOrder bool) (pair Pair, unit, quote string, err error) {
	if s == "" {
		return Pair(""), "", "", errors.Errorf("nil input pair")
	}
	if len(delimiters) == 0 {
		delimiters = commonPairDelimiter
	}
	s = strings.Replace(s, " ", "", -1)
	s = strings.ToUpper(s)
	first := ""
	second := ""

	// Like LTC-ETH...
	for _, pd := range delimiters {
		if pd == "" {
			continue
		}
		ss := strings.Split(s, pd)
		if len(ss) != 2 {
			continue
		}
		first = ss[0]
		second = ss[1]
		break
	}

	//specialDelimiters := []string{"BULL"}

	// special delimiters which will append unit, on binance there are pairs like XRPBULL/BUSD ETHBULL/USDT
	// 需要跟在左侧Unit的特殊分隔符，币安中存在XRPBULL/BUSD ETHBULL/USDT等交易对
	for _, sd := range leftTailDelimiters {
		ss := strings.Split(s, sd)
		ss = gstring.RemoveByValue(ss, "") // this is absolutely necessary, if s start with "BULL", it is an empty string at 0 of ss
		if len(ss) != 2 {
			continue
		}
		first = ss[0] + sd
		second = ss[1]
		break
	}

	// special delimiters which will be quote header, no examples for now
	// 需要在右侧Quote头部的特殊分隔符，目前还不需要
	for _, sd := range rightHeaderDelimiters {
		ss := strings.Split(s, sd)
		ss = gstring.RemoveByValue(ss, "") // this is absolutely necessary, if s start with "BULL", it is an empty string at 0 of ss
		if len(ss) != 2 {
			continue
		}
		first = ss[0]
		second = sd + ss[1]
		break
	}

	// examples: "LTCETH", "ETHBTC"
	if first == "" || second == "" {
		for _, v := range AllQuoteAssets() {
			/*quoteSymbol := v.Symbol()
			/*if v.TypeSide() == AssetTypeFiat {
				quoteSymbol = v.Name()
			}*/
			if gstring.StartWith(s, v.TradeSymbol()) {
				first = v.TradeSymbol()
				second = gstring.RemoveHead(s, len(v.TradeSymbol()))
				break
			}

			if gstring.EndWith(s, v.TradeSymbol()) {
				first = gstring.RemoveTail(s, len(v.TradeSymbol()))
				second = v.TradeSymbol()
				break
			}
		}
	}

	if first == "" || second == "" {
		return Pair(""), "", "", errors.Errorf("invalid trade pair %s", s)
	}

	if ISOOrder {
		return Pair(fmt.Sprintf("%s/%s", strings.ToUpper(first), strings.ToUpper(second))), strings.ToUpper(first), strings.ToUpper(second), nil
	} else {
		return Pair(fmt.Sprintf("%s/%s", strings.ToUpper(second), strings.ToUpper(first))), strings.ToUpper(second), strings.ToUpper(first), nil

	}
}

// parse ISO standard pair string
func ParsePair(s string) (Pair, error) {
	pair, _, _, err := parsePairWithOptions(s, []string{"/"}, nil, nil, true)
	return pair, err
}

func ParsePairCustom(s string, config *ExProperty) (Pair, error) {
	if config != nil {
		if pair, _, _, err := parsePairWithOptions(s, []string{config.PairDelimiter}, config.PairDelimiterLeftTail, config.PairDelimiterRightHead, config.PairNormalOrder); err != nil {
			return Pair(""), err
		} else {
			return pair, nil
		}
	} else {
		if pair, _, _, err := parsePairWithOptions(s, nil, nil, nil, true); err != nil {
			return Pair(""), err
		} else {
			return pair, nil
		}
	}
}

func NewPair(unit, quote string) Pair {
	if unit == "" || quote == "" {
		return PairErr
	}
	p, err := ParsePair(fmt.Sprintf("%s/%s", strings.ToUpper(unit), strings.ToUpper(quote)))
	if err != nil {
		return PairErr
	}
	return p
}

func (p Pair) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.FormatISO() + `"`), nil
}

func (p Pair) FormatISO() string {
	return p.Format("/", true)
}

func (p Pair) String() string {
	return p.FormatISO()
}

func (p Pair) Format(split string, upper bool) string {
	_, unit, quote, err := parsePairWithOptions(string(p), nil, nil, nil, true)
	if err != nil {
		return ""
	}
	if upper {
		unit = strings.ToUpper(unit)
		quote = strings.ToUpper(quote)
		split = strings.ToUpper(split)
	} else {
		unit = strings.ToLower(unit)
		quote = strings.ToLower(quote)
		split = strings.ToLower(split)
	}
	return fmt.Sprintf("%s%s%s", unit, split, quote)
}

func (p Pair) First() string {
	return p.Unit()
}

func (p Pair) Second() string {
	return p.Quote()
}

// get first asset in ISO trade pair format
func (p Pair) Unit() string {
	_, unit, _, err := parsePairWithOptions(string(p), nil, nil, nil, true)
	if err != nil {
		return ""
	}
	return strings.ToUpper(unit)
}

// get first asset in ISO trade pair format
func (p Pair) Quote() string {
	_, _, quote, err := parsePairWithOptions(string(p), nil, nil, nil, true)
	if err != nil {
		return ""
	}
	return strings.ToUpper(quote)
}

func (p Pair) CustomFormat(config *ExProperty) string {
	delimiter, normalOrder, upperCase := config.PairDelimiter, config.PairNormalOrder, config.PairUpperCase
	first := ""
	second := ""
	if normalOrder {
		if upperCase {
			first = strings.ToUpper(p.Unit())
			second = strings.ToUpper(p.Quote())
			delimiter = strings.ToUpper(delimiter)
		} else {
			first = strings.ToLower(p.Unit())
			second = strings.ToLower(p.Quote())
			delimiter = strings.ToLower(delimiter)
		}
	} else {
		if upperCase {
			first = strings.ToUpper(p.Quote())
			second = strings.ToUpper(p.Unit())
			delimiter = strings.ToUpper(delimiter)
		} else {
			first = strings.ToLower(p.Quote())
			second = strings.ToLower(p.Unit())
			delimiter = strings.ToLower(delimiter)
		}
	}
	s := fmt.Sprintf("%s%s%s", first, delimiter, second)
	if config.PairUpperCase {
		return strings.ToUpper(s)
	} else {
		return strings.ToLower(s)
	}
}

func (p Pair) Verify() error {
	_, _, _, err := parsePairWithOptions(string(p), nil, nil, nil, true)
	return err
}

func (p Pair) SetI(period Period) PairI {
	return PairI(fmt.Sprintf("%s%s%s", p.String(), pairExtDelimiter, period.String()))
}

func (p Pair) SetP(platform Platform) PairP {
	return PairP(fmt.Sprintf("%s%s%s", p.String(), pairExtDelimiter, platform.String()))
}

func (p Pair) SetM(market Market) PairM {
	return PairM(fmt.Sprintf("%s%s%s", p.String(), pairExtDelimiter, market))
}

// TODO 也许需要提高性能
func ParsePairExtString(s string) (Pair, *Period, *Market, *Platform, error) {
	defErr := gerror.Errorf("invalid pairExt(%s)", s)

	ss := strings.Split(s, pairExtDelimiter)
	if len(ss) <= 0 || len(ss) >= 5 {
		return PairErr, nil, nil, nil, defErr
	}
	if err := Pair(ss[0]).Verify(); err != nil {
		return PairErr, nil, nil, nil, defErr
	}
	parsedSecCount := 0 // 最终解析出几个段落
	parsedSecCount++

	var resPeriod *Period = nil
	var resMarket *Market = nil
	var resPlatform *Platform = nil
	for i := 1; i < len(ss); i++ {
		period, err := ParsePeriod(ss[i])
		if err == nil {
			if resPeriod != nil { // 重复出现了，这是异常
				return PairErr, nil, nil, nil, defErr
			} else {
				resPeriod = &period
				parsedSecCount++
			}
		}

		market, err := ParseMarket(ss[i])
		if err == nil {
			if resMarket != nil { // 重复出现了，这是异常
				return PairErr, nil, nil, nil, defErr
			} else {
				resMarket = &market
				parsedSecCount++
			}
		}

		platform, err := ParsePlatform(ss[i])
		if err == nil {
			if resPlatform != nil { // 重复出现了，这是异常
				return PairErr, nil, nil, nil, defErr
			} else {
				resPlatform = &platform
				parsedSecCount++
			}
		}
	}

	// 最终解析的段落个数不正确
	if parsedSecCount != len(ss) {
		return PairErr, nil, nil, nil, defErr
	}

	return Pair(ss[0]), resPeriod, resMarket, resPlatform, nil
}

func parsePairExt(s string) (pairExt, error) {
	pair, period, market, platform, err := ParsePairExtString(s)
	if err != nil {
		return pairExtErr, err
	}
	return NewPairExt(pair, period, market, platform), nil
}

func ParsePairIMP(s string) (PairIMP, error) {
	pair, period, market, platform, err := ParsePairExtString(s)
	if err != nil {
		return PairIMPErr, err
	}
	res := NewPairExt(pair, period, market, platform).PairIMP()
	if err := res.Verify(); err != nil {
		return PairIMPErr, err
	}
	return res, nil
}

func NewPairExt(p Pair, period *Period, market *Market, platform *Platform) pairExt {
	if period == nil && market == nil && platform == nil {
		return pairExtErr
	}
	// 虽然都是有效指针，但指向的内容全部是错误代码，也是不允许的
	if (period != nil && *period == PeriodError) && (market != nil && *market == MarketError) && (platform != nil && *platform == PlatformUnknown) {
		return pairExtErr
	}
	s := p.String()
	if period != nil {
		s += pairExtDelimiter + period.String()
	}
	if market != nil {
		s += pairExtDelimiter + string(*market)
	}
	if platform != nil {
		s += pairExtDelimiter + platform.String()
	}
	return pairExt(s)
}

func (pe pairExt) Verify() error {
	_, _, _, _, err := ParsePairExtString(string(pe))
	return err
}

func (pe pairExt) verifyEx(intervalRequired, marketRequired, platformRequired bool) error {
	_, i, m, p, err := ParsePairExtString(string(pe))
	if err != nil {
		return err
	}
	if intervalRequired && i == nil {
		return gerror.Errorf("Pair(%s) requires interval", pe)
	}
	if marketRequired && m == nil {
		return gerror.Errorf("Pair(%s) requires market", pe)
	}
	if platformRequired && p == nil {
		return gerror.Errorf("Pair(%s) requires platform", pe)
	}
	return nil
}

func (pe pairExt) HasPeriod() bool {
	_, period, _, _, err := ParsePairExtString(string(pe))
	return err == nil && period != nil
}

func (pe pairExt) HasMarket() bool {
	_, _, market, _, err := ParsePairExtString(string(pe))
	return err == nil && market != nil
}

func (pe pairExt) HasPlatform() bool {
	_, _, _, platform, err := ParsePairExtString(string(pe))
	return err == nil && platform != nil
}

func (pe pairExt) setI(newPeriod Period) pairExt {
	pair, period, market, platform, err := ParsePairExtString(string(pe))
	if err != nil {
		return pairExtErr
	}
	period = &newPeriod
	return NewPairExt(pair, period, market, platform)
}

func (pe pairExt) setP(newPlatform Platform) pairExt {
	pair, period, market, platform, err := ParsePairExtString(string(pe))
	if err != nil {
		return pairExtErr
	}
	platform = &newPlatform
	return NewPairExt(pair, period, market, platform)
}

func (pe pairExt) setM(newMarket Market) pairExt {
	pair, period, market, platform, err := ParsePairExtString(string(pe))
	if err != nil {
		return pairExtErr
	}
	market = &newMarket
	return NewPairExt(pair, period, market, platform)
}

func (pe pairExt) Pair() Pair {
	pair, _, _, _, err := ParsePairExtString(string(pe))
	if err != nil {
		return PairErr
	}
	return pair
}

func (pe pairExt) Period() Period {
	_, period, _, _, err := ParsePairExtString(string(pe))
	if err != nil || period == nil {
		return PeriodError
	}
	return *period
}

func (pe pairExt) Platform() Platform {
	_, _, _, platform, err := ParsePairExtString(string(pe))
	if err != nil || platform == nil {
		return PlatformUnknown
	}
	return *platform
}

func (pe pairExt) Market() Market {
	_, _, market, _, err := ParsePairExtString(string(pe))
	if err != nil || market == nil {
		return MarketError
	}
	return *market
}

func (pe pairExt) Filter(keepPeriod, keepMarket, keepPlatform bool) pairExt {
	pair, period, market, platform, err := ParsePairExtString(string(pe))
	if err != nil {
		return pairExtErr
	}
	if !keepPeriod {
		period = nil
	}
	if !keepMarket {
		market = nil
	}
	if !keepPlatform {
		platform = nil
	}

	return NewPairExt(pair, period, market, platform)
}

func (pe pairExt) String() string {
	return string(pe)
}

func (pe pairExt) PairI() PairI {
	return PairI(pe.Filter(true, false, false))
}

func (pe pairExt) PairM() PairM {
	return PairM(pe.Filter(false, true, false))
}

func (pe pairExt) PairP() PairP {
	return PairP(pe.Filter(false, false, true))
}

func (pe pairExt) PairIM() PairIM {
	return PairIM(pe.Filter(true, true, false))
}

func (pe pairExt) PairMP() PairMP {
	return PairMP(pe.Filter(false, true, true))
}

func (pe pairExt) PairIP() PairIP {
	return PairIP(pe.Filter(true, false, true))
}

func (pe pairExt) PairIMP() PairIMP {
	return PairIMP(pe)
}

func PairExtsSort(src []pairExt) {
	var ss []string
	for _, v := range src {
		ss = append(ss, v.String())
	}
	sort.Strings(ss)

	src = nil
	for _, v := range ss {
		src = append(src, pairExt(v))
	}
}

func PairMPSort(src []PairMP) {
	var ss []string
	for _, v := range src {
		ss = append(ss, v.String())
	}
	sort.Strings(ss)

	src = nil
	for _, v := range ss {
		src = append(src, PairMP(v))
	}
}

func PairExtsInclude(src []pairExt, find pairExt) bool {
	for _, v := range src {
		if v == find {
			return true
		}
	}
	return false
}

func PairMPInclude(src []PairMP, find PairMP) bool {
	for _, v := range src {
		if v == find {
			return true
		}
	}
	return false
}

func PairExtsEqual(a []pairExt, b []pairExt) bool {
	var sa []string
	var sb []string
	for _, v := range a {
		sa = append(sa, v.String())
	}
	for _, v := range b {
		sb = append(sb, v.String())
	}
	sort.Strings(sa)
	sort.Strings(sb)
	return strings.Join(sa, ",") == strings.Join(sb, ",")
}

func PairMPEqual(a []PairMP, b []PairMP) bool {
	var sa []string
	var sb []string
	for _, v := range a {
		sa = append(sa, v.String())
	}
	for _, v := range b {
		sb = append(sb, v.String())
	}
	sort.Strings(sa)
	sort.Strings(sb)
	return strings.Join(sa, ",") == strings.Join(sb, ",")
}

// find same Pairs between exchanges
func FindSamePairs(pairs map[Platform][]Pair) map[Pair][]Platform {
	result := make(map[Pair][]Platform)
	for scanEx, scanExPairs := range pairs {
		for _, scanPair := range scanExPairs {
			existedExs := result[scanPair]
			existedExs = append(existedExs, scanEx)
			result[scanPair] = existedExs
		}
	}

	for pair, platforms := range result {
		if len(platforms) <= 1 {
			delete(result, pair)
		}
	}
	return result
}

/* PairI */

func (pi PairI) Verify() error {
	return pi.ext().verifyEx(true, false, false)
}

func (pi PairI) String() string {
	return string(pi)
}

func (pi PairI) ext() pairExt {
	return pairExt(pi)
}

func (pi PairI) Pair() Pair {
	return pi.ext().Pair()
}

func (pi PairI) I() Period {
	return pi.ext().Period()
}

func (pi PairI) SetI(i Period) PairI {
	return pi.ext().setI(i).PairI()
}

func (pi PairI) SetM(m Market) PairIM {
	return pi.ext().setM(m).PairIM()
}

func (pi PairI) SetP(p Platform) PairIP {
	return pi.ext().setP(p).PairIP()
}

/* PairM */

func (pm PairM) Verify() error {
	return pm.ext().verifyEx(false, true, false)
}

func (pm PairM) String() string {
	return string(pm)
}

func (pm PairM) ext() pairExt {
	return pairExt(pm)
}

func (pm PairM) Pair() Pair {
	return pm.ext().Pair()
}

func (pm PairM) M() Market {
	return pm.ext().Market()
}

func (pm PairM) SetI(i Period) PairIM {
	return pm.ext().setI(i).PairIM()
}

func (pm PairM) SetM(m Market) PairM {
	return pm.ext().setM(m).PairM()
}

func (pm PairM) SetP(p Platform) PairMP {
	return pm.ext().setP(p).PairMP()
}

/* PairP */

func (pp PairP) Verify() error {
	return pp.ext().verifyEx(false, false, true)
}

func (pp PairP) String() string {
	return string(pp)
}

func (pp PairP) ext() pairExt {
	return pairExt(pp)
}

func (pp PairP) Pair() Pair {
	return pp.ext().Pair()
}

func (pp PairP) P() Platform {
	return pp.ext().Platform()
}

func (pp PairP) SetI(i Period) PairIP {
	return pp.ext().setI(i).PairIP()
}

func (pp PairP) SetM(m Market) PairMP {
	return pp.ext().setM(m).PairMP()
}

func (pp PairP) SetP(p Platform) PairP {
	return pp.ext().setP(p).PairP()
}

/* PairIM */

func (pim PairIM) Verify() error {
	return pim.ext().verifyEx(true, true, false)
}

func (pim PairIM) ext() pairExt {
	return pairExt(pim)
}

func (pim PairIM) String() string {
	return string(pim)
}

func (pim PairIM) Pair() Pair {
	return pim.ext().Pair()
}

func (pim PairIM) I() Period {
	return pim.ext().Period()
}

func (pim PairIM) M() Market {
	return pim.ext().Market()
}

func (pim PairIM) SetI(i Period) PairIM {
	return pim.ext().setI(i).PairIM()
}

func (pim PairIM) SetM(m Market) PairIM {
	return pim.ext().setM(m).PairIM()
}

func (pim PairIM) SetP(p Platform) PairIMP {
	return pim.ext().setP(p).PairIMP()
}

/* PairMP */

func (pmp PairMP) Verify() error {
	return pmp.ext().verifyEx(false, true, true)
}

func (pmp PairMP) ext() pairExt {
	return pairExt(pmp)
}

func (pmp PairMP) String() string {
	return string(pmp)
}

func (pmp PairMP) Pair() Pair {
	return pmp.ext().Pair()
}

func (pmp PairMP) M() Market {
	return pmp.ext().Market()
}

func (pmp PairMP) P() Platform {
	return pmp.ext().Platform()
}

func (pmp PairMP) PairM() PairM {
	return pmp.ext().PairM()
}

func (pmp PairMP) PairP() PairP {
	return pmp.ext().PairP()
}

func (pmp PairMP) SetI(i Period) PairIMP {
	return pmp.ext().setI(i).PairIMP()
}

func (pmp PairMP) SetM(m Market) PairMP {
	return pmp.ext().setM(m).PairMP()
}

func (pmp PairMP) SetP(p Platform) PairMP {
	return pmp.ext().setP(p).PairMP()
}

/* PairIP */

func (pip PairIP) Verify() error {
	return pip.ext().verifyEx(true, false, true)
}

func (pip PairIP) ext() pairExt {
	return pairExt(pip)
}

func (pip PairIP) String() string {
	return string(pip)
}

func (pip PairIP) Pair() Pair {
	return pip.ext().Pair()
}

func (pip PairIP) I() Period {
	return pip.ext().Period()
}

func (pip PairIP) P() Platform {
	return pip.ext().Platform()
}

func (pip PairIP) SetI(i Period) PairIP {
	return pip.ext().setI(i).PairIP()
}

func (pip PairIP) SetM(m Market) PairIMP {
	return pip.ext().setM(m).PairIMP()
}

func (pip PairIP) SetP(p Platform) PairIP {
	return pip.ext().setP(p).PairIP()
}

/* PairIMP */

func (pimp PairIMP) Verify() error {
	return pimp.ext().verifyEx(true, true, true)
}

func (pimp PairIMP) ext() pairExt {
	return pairExt(pimp)
}

func (pimp PairIMP) String() string {
	return string(pimp)
}

func (pimp PairIMP) Pair() Pair {
	return pimp.ext().Pair()
}

func (pimp PairIMP) I() Period {
	return pimp.ext().Period()
}

func (pimp PairIMP) M() Market {
	return pimp.ext().Market()
}

func (pimp PairIMP) P() Platform {
	return pimp.ext().Platform()
}

func (pimp PairIMP) PairMP() PairMP {
	return pimp.ext().PairMP()
}

func (pimp PairIMP) SetI(i Period) PairIMP {
	return pimp.ext().setI(i).PairIMP()
}

func (pimp PairIMP) SetM(m Market) PairIMP {
	return pimp.ext().setM(m).PairIMP()
}

func (pimp PairIMP) SetP(p Platform) PairIMP {
	return pimp.ext().setP(p).PairIMP()
}
