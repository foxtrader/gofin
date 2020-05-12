package fintypes

import (
	"github.com/shawnwyckoff/gopkg/apputil/gerror"
	"golang.org/x/text/currency"
	"strings"
)

type (
	AssetSetting struct {
		AssetType   AssetType
		Name        string
		Symbol      string
		AnchorFiat  string
		UsedAsQuote bool
	}
)

var (
	innerAssetsSettings = map[Asset]AssetSetting{}
)

func enrollAsset(setting AssetSetting) (Asset, error) {
	if setting.AssetType == AssetTypeCoin {
		asset := NewCoin(setting.Name, setting.Symbol, PlatformOpen)
		innerAssetsSettings[asset] = setting
		return asset, nil
	}

	if setting.AssetType == AssetTypeFiat {
		asset, err := ParseFiat(setting.Name)
		if err != nil {
			return AssetNil, err
		}
		innerAssetsSettings[asset] = setting
		return asset, nil
	}

	if setting.AssetType == AssetTypeMetal {
		asset := AssetNil
		switch strings.ToUpper(setting.Name) {
		case "XAU":
			asset = newMetal("XAU") // gold
		case "XAG":
			asset = newMetal("XAG") // silver
		case "XPT":
			asset = newMetal("XPT") // platinum
		case "XPD":
			asset = newMetal("XPD") // Palladium
		default:
			return AssetNil, gerror.Errorf("unknown metal %s", setting.Name)
		}
		innerAssetsSettings[asset] = setting
		return asset, nil
	}

	if setting.AssetType == AssetTypeIndex {
	}

	if setting.AssetType == AssetTypeStock {
	}

	return AssetNil, gerror.Errorf("unknown asset type %s", setting.AssetType)
}

func mustEnrollAsset(setting AssetSetting) Asset {
	asset, err := enrollAsset(setting)
	if err != nil {
		panic(err)
	}
	return asset
}

// famous assets
var (

	// famous quote coins
	BTC  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Bitcoin", "BTC", "", true})
	LTC  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Litecoin", "LTC", "", true})
	ETH  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Ethereum", "ETH", "", true})
	EOS  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "EOS", "EOS", "", true})
	ZRX  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "0x", "ZRX", "", true})
	BNB  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "BinanceCoin", "BNB", "", true})
	ETC  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Ethereum-Classic", "ETC", "", true})
	TRX  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Tron", "TRX", "", true})
	DOGE = mustEnrollAsset(AssetSetting{AssetTypeCoin, "DogeCoin", "DOGE", "", true})
	XLM  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Stellar", "XLM", "", true})
	XRP  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Ripple", "XRP", "", true})

	// famous common coins
	AION = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Aion", "AION", "", false})
	WAN  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "WanCoin", "WAN", "", false})
	FSN  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Fusion", "FSN", "", false})
	HOLO = mustEnrollAsset(AssetSetting{AssetTypeCoin, "HoloToken", "HOT", "", false})
	DCR  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Decred", "DCR", "", false})
	POA  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "POA-Network", "POA", "", false})
	DFN  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Dfinity", "DFN", "", false})
	AE   = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Aeternity", "AE", "", false})
	DOT  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Polkadot", "DOT", "", false})
	ADA  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Cardano", "ADA", "", false})
	RHOC = mustEnrollAsset(AssetSetting{AssetTypeCoin, "RChain", "RHOC", "", false})
	ICX  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "ICON", "ICX", "", false})
	NANO = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Nano", "NANO", "", false})
	ZEC  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "ZenCash", "ZEC", "", false})
	GRIN = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Grin", "GRIN", "", false})

	// stable coins
	USDC = mustEnrollAsset(AssetSetting{AssetTypeCoin, "USD-Coin", "USDC", "USD", true})
	USDT = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Tether", "USDT", "USD", true})
	USDS = mustEnrollAsset(AssetSetting{AssetTypeCoin, "StableUSD", "USDS", "USD", true})
	TUSD = mustEnrollAsset(AssetSetting{AssetTypeCoin, "TrueUSD", "TUSD", "USD", true})
	PAX  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "Paxos-Standard-Token", "PAX", "USD", true})
	DAI  = mustEnrollAsset(AssetSetting{AssetTypeCoin, "DAI", "DAI", "USD", true})
	BUSD = mustEnrollAsset(AssetSetting{AssetTypeCoin, "BinanceUSD", "BUSD", "USD", true})

	// G10 currencies
	USD = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.USD.String(), "", "", true}) // US Dollar ($)
	EUR = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.EUR.String(), "", "", true}) // Euro (€)
	JPY = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.JPY.String(), "", "", true}) // Japanese Yen (¥)
	GBP = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.GBP.String(), "", "", true}) // British Pound Sterling (£)
	CHF = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.CHF.String(), "", "", true}) // Swiss Franc (CHF)
	AUD = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.AUD.String(), "", "", true}) // Australian Dollar, AUD (A$)
	NZD = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.NZD.String(), "", "", true}) // New Zealand Dollar (NZ$)
	CAD = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.CAD.String(), "", "", true}) // Canadian Dollar, CAD (C$)
	SEK = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.SEK.String(), "", "", true}) // Swedish Krona (SEK)
	NOK = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.NOK.String(), "", "", true}) // Norwegian Krone (NOK)

	//  non-famous quote currencies
	NGN = mustEnrollAsset(AssetSetting{AssetTypeFiat, "NGN", "", "", true}) // Nigerian Naira (NGN)

	// Additional common currencies
	BRL = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.BRL.String(), "", "", false}) // Brazilian Real (R$)
	CNY = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.CNY.String(), "", "", false}) // Chinese Yuan (CN¥)
	DKK = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.DKK.String(), "", "", false}) // Danish Krone (DKK)
	INR = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.INR.String(), "", "", false}) // Indian Rupee (Rs.)
	RUB = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.RUB.String(), "", "", false}) // Russian Ruble (RUB)
	HKD = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.HKD.String(), "", "", false}) // Hong Kong Dollar (HK$)
	IDR = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.IDR.String(), "", "", false}) // Indonesian Rupiah (IDR)
	KRW = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.KRW.String(), "", "", false}) // South Korean Won (₩)
	MXN = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.MXN.String(), "", "", false}) // Mexican Peso (MX$)
	PLN = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.PLN.String(), "", "", false}) // Polish Zloty (PLN)
	SAR = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.SAR.String(), "", "", false}) // Saudi Riyal (SAR)
	THB = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.THB.String(), "", "", false}) // Thai Baht (฿)
	TRY = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.TRY.String(), "", "", false}) // Turkish Lira, TRY(₺)
	TWD = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.TWD.String(), "", "", false}) // New Taiwan dollar, TWD (NT$)
	ZAR = mustEnrollAsset(AssetSetting{AssetTypeFiat, currency.ZAR.String(), "", "", false}) // South African Rand (ZAR)

	// precious metal
	XAU = mustEnrollAsset(AssetSetting{AssetTypeMetal, "xau", "", "", false}) // gold
	XAG = mustEnrollAsset(AssetSetting{AssetTypeMetal, "xag", "", "", false}) // silver
	XPT = mustEnrollAsset(AssetSetting{AssetTypeMetal, "xpt", "", "", false}) // platinum
	XPD = mustEnrollAsset(AssetSetting{AssetTypeMetal, "xpd", "", "", false}) // Palladium
)

func StableCoinsByFiat(anchorFiat Asset) []Asset {
	var r []Asset
	for k, v := range innerAssetsSettings {
		if v.AssetType == AssetTypeCoin && v.AnchorFiat == anchorFiat.TradeSymbol() {
			r = append(r, k)
		}
	}
	return r
}

func AllStableCoins() []Asset {
	var r []Asset
	for k, v := range innerAssetsSettings {
		if v.AssetType == AssetTypeCoin && v.AnchorFiat != "" {
			r = append(r, k)
		}
	}
	return r
}

func AllMetals() []Asset {
	return []Asset{XAU, XAG, XPD, XPT}
}

func AllFiats() []Asset {
	var r []Asset
	for _, unit := range CurrencyCodesHistorical() {
		if CurrencyIsFiat(unit) {
			r = append(r, newFiatFromUnit(unit))
		}
	}
	return r
}

func AllQuoteAssets() []Asset {
	var r []Asset
	for k, v := range innerAssetsSettings {
		if v.UsedAsQuote {
			r = append(r, k)
		}
	}
	return r
}

func AllQuoteCoins() []Asset {
	var r []Asset
	for k, v := range innerAssetsSettings {
		if v.AssetType == AssetTypeCoin && v.UsedAsQuote {
			r = append(r, k)
		}
	}
	return r
}
