package fincalc

import (
	"fmt"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"github.com/shawnwyckoff/gopkg/container/gnum"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"math"
	"sort"
	"time"
)

type (
	netValue struct {
		time time.Time
		net  float64
	}

	PnlCalc struct {
		dataSourcePeriod fintypes.Period
		items            []netValue
		sorted           bool
	}

	PNL struct {
		TotalReturns      gnum.ElegantFloat // 最后一次交易的回报
		AnnualizedReturns gnum.ElegantFloat // in fiat
		MaxDrawDown       gnum.ElegantFloat // 与此之前最高余额的跌幅
		SharpeRatio       gnum.ElegantFloat
		TradeTimes        map[fintypes.TradeIntent]int
	}
)

func (p PNL) TotalTradeTimes() int {
	total := 0
	for _, v := range p.TradeTimes {
		total += v
	}
	return total
}

func NewCalc(dataSourcePeriod fintypes.Period) *PnlCalc {
	return &PnlCalc{dataSourcePeriod: dataSourcePeriod}
}

func (c *PnlCalc) Len() int {
	return len(c.items)
}

func (c *PnlCalc) Less(i, j int) bool {
	return c.items[i].time.Before(c.items[j].time)
}

func (c *PnlCalc) Swap(i, j int) {
	c.items[i], c.items[j] = c.items[j], c.items[i]
}

func (c *PnlCalc) Add(tm time.Time, net float64) {
	c.items = append(c.items, netValue{time: tm, net: net})
	c.sorted = false
}

func (c *PnlCalc) Sort() {
	if c.sorted == false {
		sort.Sort(c)
		c.sorted = true
	}
}

func (c *PnlCalc) ReturnsRate() []float64 {
	r := make([]float64, c.Len()-1)
	for i := 1; i < c.Len(); i++ {
		r[i-1] = (c.items[i].net / c.items[i-1].net) - 1
	}
	return r
}

// Total returns (总收益率)
func (c *PnlCalc) TotalReturns() float64 {
	c.Sort()
	return (c.items[c.Len()-1].net - c.items[0].net) / c.items[0].net
}

func (c *PnlCalc) maxValueUntil(untilIdx int) float64 {
	maxVal := 0.0
	for i := 0; i <= untilIdx; i++ {
		if c.items[i].net > maxVal {
			maxVal = c.items[i].net
		}
	}
	return maxVal
}

// todo: test required
func (c *PnlCalc) MaxDrawDown() float64 {
	c.Sort()
	maxDrawDown := 0.0
	for i := 1; i < c.Len(); i++ {
		maxVal := c.maxValueUntil(i - 1)
		drawDown := 1.0 - (c.items[i].net / maxVal)
		if drawDown > 0 && drawDown > maxDrawDown {
			maxDrawDown = drawDown
		}
	}
	return maxDrawDown
}

func (c *PnlCalc) Calc(period fintypes.Period, is24hTrade bool) PNL {
	if c.Len() < 2 {
		return PNL{}
	}

	return PNL{
		TotalReturns:      gnum.NewElegantFloat(1, -1), //gnum.NewElegantFloat(c.TotalReturns(), -1),
		AnnualizedReturns: gnum.NewElegantFloat(c.AnnualizedReturns(is24hTrade), -1),
		MaxDrawDown:       gnum.NewElegantFloat(c.MaxDrawDown(), -1),
		//SharpeRatio:       gnum.NewElegantFloat(c.SharpeRatio(period, 0.04, is24hTrade), -1),
		TradeTimes: nil,
	}
}

// 根据周期转换为交易时长
func getPeriodDuration(period fintypes.Period, is24hTrade bool) time.Duration {
	periodDuration := time.Duration(0)
	if is24hTrade {
		periodDuration = period.ToDuration()
	} else {
		switch period {
		case fintypes.Period1Day:
			periodDuration = gtime.Day
		case fintypes.Period1Week:
			periodDuration = gtime.Day * 5
		case fintypes.Period1MonthFUZZY:
			periodDuration = gtime.Day * 21
		case fintypes.Period1YearFUZZY:
			periodDuration = gtime.Day * 252
		}
	}
	return periodDuration
}

// 获取交易时长，最小密度为天，也就是说股市交易的一天（一般4小时）按24小时计算
func getTradeDuration(begin, end time.Time, is24hTrade bool) time.Duration {
	min := gtime.MinTime(begin, end)
	max := gtime.MaxTime(begin, end)

	tradeDuration := time.Duration(0)
	if is24hTrade {
		tradeDuration = max.Sub(min)
	} else {
		tradeDuration = time.Duration(gtime.CountWorkDays(min, max)) * gtime.Day
	}
	return tradeDuration
}

func (c *PnlCalc) Net() []float64 {
	var r []float64
	for _, v := range c.items {
		r = append(r, v.net)
	}
	return r
}

// 特定周期的（平均）收益率，这个周期必须大于等于1天，否则太难计算，因为全球金融市场每个交易日的交易时长不统一
// https://bigquant.com/community/t/topic/257
func (c *PnlCalc) PeriodReturns(period fintypes.Period, is24hTrade bool) float64 {
	if c.Len() <= 1 {
		return 0
	}
	c.Sort()
	if period.ToDuration() < fintypes.Period1Day.ToDuration() {
		panic(gerror.Errorf("Period(%s) too small", period.String()))
	}

	if period == c.dataSourcePeriod { // 数据源的周期和目标周期相同，那么只需要做个平均就行
		returnsRate := make([]float64, c.Len()-1) // rate of returns (收益率)
		for i := 1; i < c.Len(); i++ {
			returnsRate[i-1] = (c.items[i].net / c.items[i-1].net) - 1
		}
		mean := gnum.Mean(returnsRate)
		if math.IsNaN(mean) {
			fmt.Println(returnsRate)
			panic(gerror.Errorf("mean is NaN"))
		} else {
			return mean
		}
	} else { // 数据源的周期和目标周期不同，就要上指数进行计算了
		periodDuration := getPeriodDuration(period, is24hTrade)
		tradeDuration := getTradeDuration(c.items[0].time, c.items[c.Len()-1].time, is24hTrade)
		return math.Pow(1+c.TotalReturns(), float64(periodDuration)/float64(tradeDuration)) - 1
	}
}

// 年化收益率
// https://bigquant.com/community/t/topic/257
func (c *PnlCalc) AnnualizedReturns(is24hTrade bool) float64 {
	return c.PeriodReturns(fintypes.Period1YearFUZZY, is24hTrade)
}

// Annualized SharpeRatio = (年化收益率 - 无风险年化利率) / 年化波动率
// 注意，三个元素必须在同一个周期上
func (c *PnlCalc) SharpeRatio(period fintypes.Period, annualizedRiskFreeRateOfReturn float64, is24hTrade bool) float64 {
	periodDuration := getPeriodDuration(period, is24hTrade)
	yearTradeDuration := time.Duration(0)
	if is24hTrade {
		yearTradeDuration = gtime.Year365
	} else {
		yearTradeDuration = gtime.Day * 250
	}

	returnMean := c.PeriodReturns(period, is24hTrade)                                                                     // 特定周期的收益率
	returnBenchmark := math.Pow(1+annualizedRiskFreeRateOfReturn, float64(periodDuration)/float64(yearTradeDuration)) - 1 // 特定周期的无风险利率（储蓄），也就是Benchmark（基准）利率
	returnStd := gnum.Std(c.ReturnsRate(), 1)                                                                             // 特定周期的波动率(volatility)

	// 如果连续的变化率一直是0，也就是账户法币余额没有发生改变，那么returnStd也等于0，此时是特殊情况
	if returnStd == 0 {
		return 0
	}

	return (returnMean - returnBenchmark) / returnStd
}

// reference: https://github.com/stephenlyu/tds
func CalcAPR(y []float64, scale float64) float64 {
	if len(y) == 0 {
		return 0
	}

	ret := y[len(y)-1] / y[0]

	apr := math.Pow(ret, scale/float64(len(y))) - 1.0
	if math.IsNaN(apr) {
		apr = -100.0
	}
	return apr
}
