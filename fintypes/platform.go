package fintypes

import (
	"github.com/pkg/errors"
	"github.com/shawnwyckoff/gopkg/container/gstring"
	"github.com/shawnwyckoff/gopkg/sys/gtime"
	"strings"
)

type (
	Platform string

	PlatformInfo struct {
		Support  []AssetType
		OpenDate gtime.Date
	}
)

var (
	NASDAQOpenDate, _ = gtime.NewDate(1971, 2, 4)
	NYSEOpenDate, _   = gtime.NewDate(1817, 3, 8)
	AMEXOpenDate, _   = gtime.NewDate(1971, 2, 8) // https://www.loc.gov/rr/business/amex/amex.html
	SSEOpenDate, _    = gtime.NewDate(1990, 12, 19)
	SZSEOpenDate, _   = gtime.NewDate(1991, 7, 3)
	HKEXOpenDate, _   = gtime.NewDate(1891, 2, 3) // https://www.ximalaya.com/shangye/22958651/178058751
	TwOpenDate, _     = gtime.NewDate(1962, 2, 9)

	PlatformUnknown = Platform("")
	PlatformOpen    = enrollPlatform("open", PlatformInfo{Support: []AssetType{AssetTypeCoin, AssetTypeStock, AssetTypeIndex, AssetTypeMetal}, OpenDate: 0})  // fake platform for internet finance data like gold price
	PlatformIndex   = enrollPlatform("index", PlatformInfo{Support: []AssetType{AssetTypeCoin, AssetTypeStock, AssetTypeIndex, AssetTypeMetal}, OpenDate: 0}) // fake platform for all indexes

	// 第三方金融数据平台
	CryptoCompare = enrollPlatform("cryptocompare", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})
	DataHub       = enrollPlatform("datahub", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})
	YahooFinance  = enrollPlatform("yahoofinance", PlatformInfo{Support: []AssetType{AssetTypeCoin, AssetTypeStock}, OpenDate: 0})
	GoogleFinance = enrollPlatform("googlefinance", PlatformInfo{Support: []AssetType{AssetTypeCoin, AssetTypeStock}, OpenDate: 0})
	CoinMarketCap = enrollPlatform("coinmarketcap", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})
	CoinGecko     = enrollPlatform("coingecko", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})

	// 国内加密货币市场
	Binance = enrollPlatform("binance", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})
	Huobi   = enrollPlatform("huobi", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})
	Gate    = enrollPlatform("gate", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})

	// 美国证券市场
	Nasdaq = enrollPlatform("nasdaq", PlatformInfo{Support: []AssetType{AssetTypeStock}, OpenDate: NASDAQOpenDate})
	Nyse   = enrollPlatform("nyse", PlatformInfo{Support: []AssetType{AssetTypeStock}, OpenDate: NYSEOpenDate})
	Amex   = enrollPlatform("amex", PlatformInfo{Support: []AssetType{AssetTypeStock}, OpenDate: AMEXOpenDate}) // belongs to NYSE now

	// 中国期货市场
	Shfe = enrollPlatform("shfe", PlatformInfo{Support: []AssetType{AssetTypeUnknown}, OpenDate: gtime.NewDatePanic(1990, 11, 26)}) // 上海期货交易所
	Czce = enrollPlatform("czce", PlatformInfo{Support: []AssetType{AssetTypeUnknown}, OpenDate: gtime.NewDatePanic(1990, 10, 12)}) // 郑州商品交易所
	Dce  = enrollPlatform("dce", PlatformInfo{Support: []AssetType{AssetTypeUnknown}, OpenDate: gtime.NewDatePanic(1993, 2, 28)})   // 大连商品交易所

	// 国外加密货币市场
	Kraken  = enrollPlatform("kraken", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})
	Deribit = enrollPlatform("deribit", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})
	Ftx     = enrollPlatform("ftx", PlatformInfo{Support: []AssetType{AssetTypeCoin}, OpenDate: 0})

	// 中国证券市场
	Szse = enrollPlatform("szse", PlatformInfo{Support: []AssetType{AssetTypeStock}, OpenDate: SZSEOpenDate}) // Shen Zhen Stock Exchange
	Sse  = enrollPlatform("sse", PlatformInfo{Support: []AssetType{AssetTypeStock}, OpenDate: SSEOpenDate})   // Shanghai Stock Exchange
	Hkex = enrollPlatform("hkex", PlatformInfo{Support: []AssetType{AssetTypeStock}, OpenDate: HKEXOpenDate}) // Hong Kong Exchange

	allPlatformInfos = map[Platform]PlatformInfo{}
)

func enrollPlatform(name string, info PlatformInfo) Platform {
	allPlatformInfos[Platform(name)] = info
	return Platform(name)
}

func (p *Platform) Info() PlatformInfo {
	info, ok := allPlatformInfos[*p]
	if !ok {
		return PlatformInfo{}
	}
	return info
}

func (p Platform) String() string {
	return string(p)
}

// Check whether supported exchange or n ot.
func (p Platform) IsSupported() bool {
	for _, v := range AllSupportedExs {
		if v.String() == strings.ToLower(p.String()) {
			return true
		}
	}
	return false
}

func (p Platform) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *Platform) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = gstring.RemoveHead(str, 1)
	str = gstring.RemoveTail(str, 1)
	plt, err := ParsePlatform(str)
	if err != nil {
		return err
	}
	*p = plt
	return nil
}

func ParsePlatform(name string) (Platform, error) {
	for plt := range allPlatformInfos {
		if strings.ToLower(name) == strings.ToLower(plt.String()) {
			return plt, nil
		}
	}

	return PlatformUnknown, errors.Errorf("PlatformUnknown platform %s", name)
}

func RemoveDuplicatePlatforms(platforms []Platform) []Platform {
	var ss []string
	for _, v := range platforms {
		ss = append(ss, v.String())
	}
	ss = gstring.RemoveDuplicate(ss)
	var res []Platform
	for _, v := range ss {
		res = append(res, Platform(v))
	}
	return res
}

func AllStockExchanges() []Platform {
	var res []Platform
	for plt, info := range allPlatformInfos {
		for _, v := range info.Support {
			if v == AssetTypeStock {
				res = append(res, plt)
			}
		}
	}
	return res
}
