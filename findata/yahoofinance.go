package findata

// https://github.com/piquette/finance-go
// 可以参考这个库看看人家怎么取的Volume

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/foxtrader/gofin/fintypes"
	fintypes2 "github.com/foxtrader/gofin/fintypes"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/net/ghttp"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

var apiYahooFinance = "https://query1.finance.yahoo.com/v7/finance/quote"

// APIError represents an error response
type (
	APIError struct {
		Code        string `json:"code"`        // "argument-error"
		Description string `json:"description"` // "Missing value for the \"symbols\" argument"
	}
	// APIMessage represents a standard API response
	APIMessage struct {
		Error *APIError `json:"error,omitempty"`
	}

	// APIEnvelope is the wrapping envelope around a Yahoo Finance API response
	APIEnvelope map[string]APIMessage

	YFAPI struct {
		proxy            string
		minAssetOpenDate gtime.Date
	}

	yfQuote struct {
		Symbol    string             `json:"symbol"`
		Precision int64              `json:"-"`
		Date      []time.Time        `json:"date"`
		Open      []gdecimal.Decimal `json:"open"`
		High      []gdecimal.Decimal `json:"high"`
		Low       []gdecimal.Decimal `json:"low"`
		Close     []gdecimal.Decimal `json:"close"`
		Volume    []gdecimal.Decimal `json:"volume"`
	}
)

func NewYahooFinance(proxy string) (*YFAPI, error) {
	r := new(YFAPI)
	r.minAssetOpenDate = fintypes.NYSEOpenDate
	r.proxy = proxy
	return r, nil
}

func (yf *YFAPI) GetKlineProviderInfo() (*fintypes.KlineProviderInfo, error) {
	r := fintypes.KlineProviderInfo{}
	stocks, indexes, err := StockListAll()
	if err != nil {
		return nil, err
	}
	for _, v := range stocks {
		r.SupportedPairs = append(r.SupportedPairs, v.Pair().SetM(fintypes.MarketSpot) /*.SetI(fintypes.Period1Day).SetP(fintypes.YahooFinance)*/)
	}
	for _, v := range indexes {
		r.SupportedPairs = append(r.SupportedPairs, v.Pair().SetM(fintypes.MarketSpot) /*.SetI(fintypes.Period1Day).SetP(fintypes.YahooFinance)*/)
	}
	r.MinPeriod = fintypes.Period1Day
	r.KlineRequestRateLimit = time.Second * 30       // note: this rate limit is not tested
	r.FirstTrade = fintypes.NYSEOpenDate.ToTimeUTC() // note: this time is NOT accurate
	return &r, nil
}

func (yf *YFAPI) pairPToYahooFinanceSymbol(target fintypes.PairP) (string, error) {
	symbol := target.Pair().Unit()
	ex := target.P()

	if ex == fintypes.PlatformIndex {
		if target == fintypes.IndexToPairP(fintypes.IndexDJI) {
			return "^DJI", nil
		} else if target == fintypes.IndexToPairP(fintypes.IndexSHH) {
			return "000001.SS", nil
		}
	} else if ex == fintypes.Sse {
		return symbol + ".SS", nil
	} else if ex == fintypes.Szse {
		return symbol + ".SZ", nil
	} else if ex == fintypes.Hkex {
		if len(symbol) == 5 {
			symbol = symbol[1:]
			return symbol + ".HK", nil
		}
	} else if ex == fintypes.Nasdaq || ex == fintypes.Nyse || ex == fintypes.Amex {
		return symbol, nil
	}

	return "", errors.Errorf("yahoo finance unsupported PairP(%s)", target.String())
}

// index Pair samples: DJI/USD@open, SHH/CNY@open
func (yf *YFAPI) GetKlineEx(platform fintypes.Platform, market fintypes.Market, target fintypes.Pair, period fintypes.Period, since *time.Time) (*fintypes2.Kline, error) {
	if period != fintypes.Period1Day {
		return nil, errors.Errorf("Yahoo Finance doesn't support Period %s", period.String())
	}

	symbol, err := yf.pairPToYahooFinanceSymbol(target.SetP(platform))
	if err != nil {
		return nil, err
	}
	if since == nil {
		t := yf.minAssetOpenDate.ToTimeUTC()
		since = &t
	}
	sinceDate := gtime.TimeToDate(*since, time.UTC)
	// WARN
	// adjustQuote填true的话，close取的yahoo的 adj close, 存在close小于low的情况，
	// adjustQuote填false的话，也存在少数这种情况
	q, err := newQuoteFromYahoo(strings.ToUpper(symbol), sinceDate.ToTimeUTC(), gtime.Today(time.UTC).ToTimeUTC(), fintypes.Period1Day, false, yf.proxy, time.Minute)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("stock(%s)", symbol))
	}
	// FIXME 这里的Market信息该怎样设定呢
	ks := fintypes2.Kline{Pair: target.SetI(period).SetM(market).SetP(platform)}
	if len(q.Date) != len(q.Close) || len(q.Date) != len(q.Low) ||
		len(q.Date) != len(q.High) || len(q.Date) != len(q.Volume) ||
		len(q.Date) != len(q.Open) {
		return nil, errors.Errorf("stock(%s) has error quote length", symbol)
	}

	for i := range q.Date {
		item := fintypes2.Bar{
			T: q.Date[i],
			O: q.Open[i].WithPrec(2),
			C: q.Close[i].WithPrec(2),
			H: q.High[i].WithPrec(2),
			L: q.Low[i].WithPrec(2),
			V: q.Volume[i].WithPrec(2),
		}

		ks.Items = append(ks.Items, item)
	}
	return &ks, nil
}

// NewQuote - new empty Quote struct
func newQuote(symbol string, bars int) yfQuote {
	return yfQuote{
		Symbol: symbol,
		Date:   make([]time.Time, bars),
		Open:   make([]gdecimal.Decimal, bars),
		High:   make([]gdecimal.Decimal, bars),
		Low:    make([]gdecimal.Decimal, bars),
		Close:  make([]gdecimal.Decimal, bars),
		Volume: make([]gdecimal.Decimal, bars),
	}
}

// NewQuoteFromYahoo - Yahoo historical prices for a symbol
func newQuoteFromYahoo(symbol string, from, to time.Time, period fintypes.Period, adjustQuote bool, proxy string, timeout time.Duration) (*yfQuote, error) {
	if timeout == 0 {
		timeout = time.Minute
	}
	if period != fintypes.Period1Day {
		return nil, errors.New("Yahoo support 1 day period only")
	}

	// Get crumb
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: timeout,
		Jar:     jar,
	}

	if proxy != "" {
		if err := ghttp.SetProxy(client, proxy); err != nil {
			return nil, err
		}
	}

	initReq, err := http.NewRequest("GET", "https://finance.yahoo.com", nil)
	if err != nil {
		return nil, err
	}
	initReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; U; Linux i686) Gecko/20071127 Firefox/2.0.0.11")
	resp, err := client.Do(initReq)
	if err != nil {
		return nil, err
	}

	crumbReq, err := http.NewRequest("GET", "https://query1.finance.yahoo.com/v1/test/getcrumb", nil)
	if err != nil {
		return nil, err
	}
	crumbReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; U; Linux i686) Gecko/20071127 Firefox/2.0.0.11")
	resp, err = client.Do(crumbReq)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(resp.Body)
	crumb, err := reader.Read()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error getting crumb for '%s'\n", symbol))
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v7/finance/download/%s?period1=%d&period2=%d&interval=1d&events=history&crumb=%s",
		symbol,
		from.Unix(),
		to.Unix(),
		crumb[0])
	resp, err = client.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("symbol '%s' not found\n", symbol))
	}

	//defer resp.Body.C()

	type ErrMsg struct {
		Chart struct {
			Error struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			}
		} `json:"chart"`

		Finance struct {
			Error struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			}
		} `json:"finance"`
	}

	b, err := ghttp.ReadBodyBytes(resp)
	if err != nil {
		return nil, err
	}

	errMsg := ErrMsg{}
	if err := json.Unmarshal(b, &errMsg); err == nil {
		if errMsg.Chart.Error.Code != "" {
			return nil, errors.New(errMsg.Chart.Error.Code + ": " + errMsg.Chart.Error.Description)
		}
		if errMsg.Finance.Error.Code != "" {
			return nil, errors.New(errMsg.Finance.Error.Code + ": " + errMsg.Finance.Error.Description)
		}
	}

	var csvdata [][]string
	reader = csv.NewReader(bytes.NewReader(b))
	csvdata, err = reader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("bad data for symbol '%s'\n", symbol))
	}
	if len(csvdata) == 0 {
		return nil, errors.New("empty csvdata")
	}

	numrows := len(csvdata) - 1
	quote := newQuote(symbol, numrows)

	for row := 1; row < len(csvdata); row++ {

		// Parse row of data
		d, err := time.Parse("2006-01-02", csvdata[row][0])
		if err != nil {
			return nil, err
		}
		// note:
		// all stock price/volume has prec 2, BTC/ETH has prec 8
		// but from yahoo finance we download stock data only
		// so we set prec into 2
		o, err := gdecimal.NewFromString(csvdata[row][1])
		if err != nil {
			return nil, err
		}
		h, err := gdecimal.NewFromString(csvdata[row][2])
		if err != nil {
			return nil, err
		}
		l, err := gdecimal.NewFromString(csvdata[row][3])
		if err != nil {
			return nil, err
		}
		c, err := gdecimal.NewFromString(csvdata[row][4])
		if err != nil {
			return nil, err
		}
		a, err := gdecimal.NewFromString(csvdata[row][5])
		if err != nil {
			return nil, err
		}
		v, err := gdecimal.NewFromString(csvdata[row][6])
		if err != nil {
			return nil, err
		}

		quote.Date[row-1] = d
		// Adjustment ratio
		if adjustQuote {
			quote.Open[row-1] = o
			quote.High[row-1] = h
			quote.Low[row-1] = l
			quote.Close[row-1] = a
		} else {
			ratio := c.Div(a)
			quote.Open[row-1] = o.Mul(ratio)
			quote.High[row-1] = h.Mul(ratio)
			quote.Low[row-1] = l.Mul(ratio)
			quote.Close[row-1] = c
		}

		quote.Volume[row-1] = v

	}

	return &quote, nil
}
