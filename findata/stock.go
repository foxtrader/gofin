package findata

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/foxtrader/gofin/fintypes"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gnum"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"github.com/shawnwyckoff/gopkg/net/ghtml"
	"github.com/shawnwyckoff/gopkg/net/ghttp"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

type (
	StockDetail struct {
		Name        string
		Symbol      string
		Exchange    fintypes.Platform
		TotalSupply int
	}
)

var (
	DowJonesOpenDate, _ = gtime.NewDate(1896, 5, 26)
)

/*
func ParseStockExchange(s string) (Platform, error) {
	s = strings.ToLower(s)
	if strings.Contains(s, "nasdaq") {
		return PLTNasdaq, nil
	}
	if strings.Contains(s, "nyse") {
		return PLTNyse, nil
	}
	if strings.Contains(s, "amex") {
		return PLTAmex, nil
	}
	return "", errors.Errorf("unknown stock exchange(%s)", s)
}


func (a Asset) OpenDate() (clock.Date, error) {
	if a.Type() == AssetTypeStock {
		exchange, _, _ := a.ToStock()
		list := AllPlatformInfos()
		for plt, info := range list {
			if plt == exchange {
				return info.OpenDate, nil
			}
		}
		return DowJonesOpenDate, nil
	} else if a.Type() == AssetTypeIndex {
		switch a {
		case IndexDJI:
			return DowJonesOpenDate, nil
		default:
			return DowJonesOpenDate, nil
		}
	} else if a.Type() == AssetTypeCoin {
		return clock.TimeToDate(BTCGenesisBlockTime, time.UTC), nil
	}
	return clock.ZeroDate, errors.Errorf("open date unknown asset(%s)", a.String())
}*/

// NewMarketList - download a list of market symbols to an array of strings
func newStockMarketList(market string) ([]string, error) {
	var symbols []string
	var url string
	switch market {
	case "nasdaq":
		url = "http://www.nasdaq.com/screening/companies-by-name.aspx?letter=0&exchange=nasdaq&render=download"
	case "amex":
		url = "http://www.nasdaq.com/screening/companies-by-name.aspx?letter=0&exchange=amex&render=download"
	case "nyse":
		url = "http://www.nasdaq.com/screening/companies-by-name.aspx?letter=0&exchange=nyse&render=download"
	default:
		return symbols, fmt.Errorf("invalid market")
	}

	resp, err := http.Get(url)
	if err != nil {
		return symbols, err
	}
	defer func() { _ = resp.Body.Close() }()

	var csvdata [][]string
	reader := csv.NewReader(resp.Body)
	csvdata, err = reader.ReadAll()
	if err != nil {
		return symbols, err
	}

	r, _ := regexp.Compile("^[a-z]+$")
	for row := 1; row < len(csvdata); row++ {
		sym := strings.TrimSpace(strings.ToLower(csvdata[row][0]))
		if r.MatchString(sym) {
			symbols = append(symbols, sym)
		}
	}
	sort.Strings(symbols)
	return symbols, nil
}

// http://eoddata.com/stocklist/AMEX.htm
// https://stackoverflow.com/questions/25338608/download-all-stock-symbol-list-of-a-market
func StockExchangeList(exchange fintypes.Platform) ([]fintypes.PairP, error) {
	var r []fintypes.PairP

	if exchange == fintypes.Nasdaq || exchange == fintypes.Nyse || exchange == fintypes.Amex {
		ss, err := newStockMarketList(strings.ToLower(exchange.String()))
		if err != nil {
			return nil, err
		}
		for _, symbol := range ss {
			if symbol == "" {
				continue
			}
			//a := NewStock(symbol, exchange)
			r = append(r, fintypes.NewPair(symbol, "USD").SetP(exchange) /*NewPair2(a, exchange.Info().DefaultQuoteFiat)*/)
		}
		return r, nil
	} else if exchange == fintypes.Sse {
		// backup data source:
		// http://quote.eastmoney.com/stocklist.html
		uri := "http://www.sse.com.cn/js/common/ssesuggestdata.js"
		s, err := ghttp.GetString(uri, "", time.Minute)
		if err != nil {
			return nil, err
		}
		ss := strings.Split(s, "\n")
		for _, v := range ss {
			if !strings.Contains(v, `{val:"`) {
				continue
			}
			if len(v) < 10 {
				continue
			}
			symbol, err := gstring.SubstrBetweenUTF8(v, `{val:"`, `",`, true, true, false, false)
			if err != nil {
				continue
			}
			if len(symbol) != 6 {
				continue
			}

			r = append(r, fintypes.NewPair(symbol, "CNY").SetP(exchange) /*NewPair2(NewStock(symbol, PLTSse), exchange.Info().DefaultQuoteFiat)*/)
		}
		return r, nil
	} else if exchange == fintypes.Szse {
		// backup data source:
		// http://quote.eastmoney.com/stocklist.html

		// 下面这段代码因为excel库升级接口不兼容暂时不可用
		/*
		uris := []string{
			"http://www.szse.cn/api/report/ShowReport?SHOWTYPE=xlsx&CATALOGID=1110&TABKEY=tab1",
			"http://www.szse.cn/api/report/ShowReport?SHOWTYPE=xlsx&CATALOGID=1110&TABKEY=tab2",
			"http://www.szse.cn/api/report/ShowReport?SHOWTYPE=xlsx&CATALOGID=1110&TABKEY=tab3",
		}

		for _, uri := range uris {
			b, err := ghttp.GetBytes(uri, "", time.Minute)
			if err != nil {
				return nil, err
			}
			f, err := xlsx.OpenBinary(b)
			if err != nil {
				return nil, err
			}
			if len(f.Sheets) < 1 || len(f.Sheets[0].Rows) < 2 {
				return nil, errors.Errorf("nil sheets/rows")
			}
			symbolColIdx := -1
			for i, cell := range f.Sheets[0].Rows[0].Cells {
				if cell.String() == "A股代码" {
					symbolColIdx = i
				}
			}
			if symbolColIdx < 0 {
				return nil, errors.Errorf("can't find symbol coll")
			}
			for _, row := range f.Sheets[0].Rows[1:] {
				symbol := row.Cells[symbolColIdx].String()
				if symbol == "" {
					continue
				}
				r = append(r, fintypes.NewPair(symbol, "CNY").SetP(exchange) /*NewPair2(NewStock(symbol, PLTSzse), exchange.Info().DefaultQuoteFiat)*//*)
			}
		}*/
		return r, nil
	} else if exchange == fintypes.Hkex {
		//uri := "https://www.hkex.com.hk/-/media/HKEX-Market/Services/Trading/Securities/Securities-Lists/Securities-Using-Standard-Transfer-Form-(including-GEM)-By-English-Stock-Short-Name-Order/englishstk_c.xls"
		uri := "http://quote.eastmoney.com/hk/HStock_list.html"
		s, err := ghttp.GetString(uri, "", time.Minute)
		if err != nil {
			return nil, err
		}
		doc, err := ghtml.NewDocFromHtmlSrc(&s)
		if err != nil {
			return nil, err
		}
		doc.Find(".hklists").Find("li").Each(func(i int, sel *goquery.Selection) {
			if gstring.StartWith(sel.Text(), "(") {
				symbol, err := gstring.SubstrBetweenUTF8(sel.Text(), "(", ")", true, true, false, false)
				if symbol == "" {
					return
				}
				if err == nil && gnum.IsDigit(symbol) {
					r = append(r, fintypes.NewPair(symbol, "HKD").SetP(exchange) /*NewPair2(NewStock(symbol, PLTHkex), exchange.Info().DefaultQuoteFiat)*/)
				}
			}
		})
		return r, nil
	}
	return nil, errors.Errorf("unsupported exchange(%s)", exchange.String())
}

func StockListAll() (stocks []fintypes.PairP, indexes []fintypes.PairP, err error) {
	exs := fintypes.AllStockExchanges()
	for _, ex := range exs {
		list, err := StockExchangeList(ex)
		if err != nil {
			return nil, nil, err
		}
		for _, v := range list {
			if v != fintypes.PairPErr {
				stocks = append(stocks, v)
			}
		}
	}

	allIndexes := fintypes.AllIndexes()
	for _, index := range allIndexes {
		indexPairExt := fintypes.IndexToPairP(index)
		if indexPairExt != fintypes.PairPErr {
			indexes = append(indexes, indexPairExt)
		}
	}
	return stocks, indexes, nil
}
