package findata

// CC相比CMC，多的是ICO信息
// reference: github.com/lucazulian/cryptocomparego

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/foxtrader/gofin/fintypes"
	fintypes2 "github.com/foxtrader/gofin/fintypes"
	"github.com/lucazulian/cryptocomparego"
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gdecimal"
	"github.com/shawnwyckoff/gopkg/net/ghttp"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ccMinBaseURL      = "https://min-api.cryptocompare.com/"
	ccHistodyBasePath = "data/histoday"
)

type (
	cryptoBasicCC struct {
		Name   string
		Symbol string
		ccId   string // CryptoCompare only
	}

	CC struct {
		proxy string
	}

	ccResponse struct {
		*http.Response

		Monitor string
	}

	ccHistoday struct {
		Time       int64            `json:"time"`
		Close      gdecimal.Decimal `json:"close"`
		High       gdecimal.Decimal `json:"high"`
		Low        gdecimal.Decimal `json:"low"`
		Open       gdecimal.Decimal `json:"open"`
		VolumeFrom gdecimal.Decimal `json:"volumefrom"`
		VolumeTo   gdecimal.Decimal `json:"volumeto"`
	}

	ccHistodayRequest struct {
		Fsym          string
		Tsym          string
		E             string
		ExtraParams   string
		Sign          bool
		TryConversion bool
		Aggregate     int // Not Used For Now
		Limit         int
		ToTs          time.Time // Not Used For Now
		AllData       bool
	}

	conversionType struct {
		Type             string `json:"type"`
		ConversionSymbol string `json:"conversionSymbol"`
	}

	ccHistodayResp struct {
		Response          string         `json:"Response"`
		Message           string         `json:"Message"` // Error Message
		Type              int            `json:"TypeSide"`
		Aggregated        bool           `json:"Aggregated"`
		Data              []ccHistoday   `json:"Data"`
		TimeTo            int64          `json:"TimeTo"`
		TimeFrom          int64          `json:"TimeFrom"`
		FirstValueInArray bool           `json:"FirstValueInArray"`
		ConversionType    conversionType `json:"ConversionType"`
	}

	cryptoBasicCCList []cryptoBasicCC
)

func (cl cryptoBasicCCList) Exists(name, symbol string) bool {
	for _, v := range ([]cryptoBasicCC)(cl) {
		if strings.ToLower(v.Name) == strings.ToLower(name) &&
			strings.ToLower(v.Symbol) == strings.ToLower(symbol) {
			return true
		}
	}
	return false
}

func (cl cryptoBasicCCList) findId(name, symbol string) (id string, err error) {
	for _, v := range []cryptoBasicCC(cl) {
		if strings.ToLower(v.Name) == strings.ToLower(name) &&
			strings.ToLower(v.Symbol) == strings.ToLower(symbol) {
			return v.ccId, nil
		}
	}
	return "", errors.Errorf("%s(%s) not found.", strings.ToLower(name), strings.ToUpper(symbol))
}

// 可能有重复的,比如XSPEC（spectrecoin）,transfercoin
func ccGetAll(cli *http.Client) (cryptoBasicCCList, error) {
	ccApi := cryptocomparego.NewClient(cli)

	// Pull from CC
	rawCCList, _, err := ccApi.Coin.List(context.Background())
	if err != nil {
		return nil, err
	}

	// Convert CC
	var r []cryptoBasicCC
	item := cryptoBasicCC{}
	for _, v := range rawCCList {
		item.Name = fintypes.FixCoinName(v.CoinName)
		item.Symbol = fintypes.FixCoinSymbol(v.Name)
		item.ccId = v.Id
		r = append(r, item)
	}

	return r, nil
}

func CCGetAll(proxy string, timeout time.Duration) ([]fintypes.Asset, error) {
	cli := http.DefaultClient
	if err := ghttp.SetProxy(cli, proxy); err != nil {
		return nil, err
	}
	ghttp.SetTimeout(cli, timeout)

	ccApi := cryptocomparego.NewClient(cli)
	rawCCList, _, err := ccApi.Coin.List(context.Background())
	if err != nil {
		return nil, err
	}
	var r []fintypes.Asset
	for _, v := range rawCCList {
		r = append(r, fintypes.NewCoin(fintypes.FixCoinName(v.CoinName), fintypes.FixCoinSymbol(v.Name), fintypes.CryptoCompare))
	}

	return r, nil
}

func ccNewHistodayRequest(fsym string, tsym string, limitDays int, allData bool) *ccHistodayRequest {
	pr := ccHistodayRequest{Fsym: fsym, Tsym: tsym}
	pr.E = "CCCAGG"
	pr.Sign = false
	pr.TryConversion = true
	pr.Aggregate = 1
	if limitDays < 1 {
		limitDays = 1
	}
	if limitDays > 2000 {
		limitDays = 2000
	}
	pr.Limit = limitDays
	pr.AllData = allData
	return &pr
}

func (hr *ccHistodayRequest) FormattedQueryString(baseUrl string) string {
	values := url.Values{}

	if len(hr.Fsym) > 0 {
		values.Add("fsym", hr.Fsym)
	}

	if len(hr.Tsym) > 0 {
		values.Add("tsym", hr.Tsym)
	}

	if len(hr.E) > 0 {
		values.Add("e", hr.E)
	}

	if len(hr.ExtraParams) > 0 {
		values.Add("extraParams", hr.ExtraParams)
	}

	values.Add("sign", strconv.FormatBool(hr.Sign))
	values.Add("tryConversion", strconv.FormatBool(hr.TryConversion))
	values.Add("limit", strconv.FormatInt(int64(hr.Limit), 10))
	values.Add("allData", strconv.FormatBool(hr.AllData))

	return fmt.Sprintf("%s?%s", baseUrl, values.Encode())
}

func ccGet(client *http.Client, histodayRequest *ccHistodayRequest) ([]ccHistoday, *ccResponse, error) {

	path := ccHistodyBasePath

	if histodayRequest != nil {
		path = histodayRequest.FormattedQueryString(ccHistodyBasePath)
	}

	minURL, _ := url.Parse(ccMinBaseURL)

	reqUrl := fmt.Sprintf("%s%s", minURL.String(), path)
	fmt.Println(reqUrl)
	resp, err := client.Get(reqUrl)
	res := ccResponse{}
	res.Response = resp
	if err != nil {
		return nil, &res, err
	}
	defer func() { resp.Body.Close() }()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &res, err
	}
	if len(buf) <= 0 {
		return nil, &res, errors.New("Empty response")
	}

	hr := ccHistodayResp{}
	err = json.Unmarshal(buf, &hr)
	if err != nil {
		return nil, &res, errors.Wrap(err, fmt.Sprintf("JSON Unmarshal error, raw string is '%s'", string(buf)))
	}
	if hr.Response == "Error" {
		return nil, &res, errors.New(hr.Message)
	}

	return hr.Data, &res, nil
}

func CCGetKline(symbol string, since *time.Time, proxy string) (*fintypes2.Kline, error) {
	cli := http.DefaultClient
	if proxy != "" {
		if err := ghttp.SetProxy(cli, proxy); err != nil {
			return nil, err
		}
	}

	days := 3000
	if since != nil {
		_, days, _ = gtime.DaysBetween(time.Now(), *since)
	}
	hr := ccNewHistodayRequest(symbol, "USD", days, true)
	data, _, err := ccGet(cli, hr)
	if err != nil {
		return nil, err
	}

	r := new(fintypes2.Kline)
	r.Pair = fintypes.NewPair(symbol, "usd").SetI(fintypes.Period1Day).SetM(fintypes.MarketSpot).SetP(fintypes.CryptoCompare) // FIXME 这里的Market信息应该怎样填写
	for _, v := range data {
		item := fintypes2.Bar{}
		if v.High.IsZero() && v.Low.IsZero() {
			continue
		}
		item.T = gtime.EpochSecToTime(v.Time)
		item.H = v.High
		item.L = v.Low
		item.O = v.Open
		item.C = v.Close
		//item.VolumeSelf = v.VolumeFrom // 以自己为单位的成交量，只有CryptoCompare提供了这个数据
		item.V = v.VolumeTo // volume in quote, like USD(s)/BTC/ETH, tested
		r.Items = append(r.Items, item)
	}
	return r, nil
}

func NewCryptoCompare(proxy string) (*CC, error) {
	r := new(CC)
	r.proxy = proxy
	return r, nil
}

func (cc *CC) GetKlineProviderInfo() (*fintypes.KlineProviderInfo, error) {
	r := fintypes.KlineProviderInfo{}
	allAssets, err := CCGetAll(cc.proxy, time.Minute*3)
	if err != nil {
		return nil, err
	}
	for _, v := range allAssets {
		r.SupportedPairs = append(r.SupportedPairs, fintypes.NewPair(v.Name(), "USD").SetM(fintypes.MarketSpot) /*.SetI(fintypes.Period1Day).SetP(fintypes.CryptoCompare)*/)
	}
	r.MinPeriod = fintypes.Period1Day
	r.KlineRequestRateLimit = time.Second / 100 * 127
	return &r, nil
}

func (cc *CC) GetKlineEx(platform fintypes.Platform, market fintypes.Market, target fintypes.Pair, period fintypes.Period, since *time.Time) (*fintypes2.Kline, error) {
	if platform != fintypes.CryptoCompare {
		return nil, errors.Errorf("PLTCC doesn't support kline of pair(%s)", target.String())
	}
	if target.Quote() != "USD" {
		return nil, errors.Errorf("PLTCC doesn't support kline of pair(%s)", target.String())
	}
	return CCGetKline(target.Unit(), since, cc.proxy)
}
