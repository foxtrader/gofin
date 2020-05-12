package fintypes

var (
	IndexDJI, _  = NewAsset("i.dji")  // Dow Jones Industrial Average (^DJI)
	IndexSP, _   = NewAsset("i.sp")   // S&P 500 (^GSPC)
	IndexIXIC, _ = NewAsset("i.ixic") // NASDAQ Composite (^IXIC)
	IndexNYA, _  = NewAsset("i.nya")  // NYSE Composite (^NYA)
	IndexRUT, _  = NewAsset("i.rut")  // Russell 2000 (^RUT)
	IndexVIX, _  = NewAsset("i.vix")  // CBOE Volatility Index (^VIX)
	IndexSHH, _  = NewAsset("i.shh")  // SSE Composite Index (000001.SS), 上证指数
)

func AllIndexes() []Asset {
	return []Asset{
		IndexDJI,
	}
}

// index is special trade pair
func IndexToPairP(index Asset) PairP {
	_, ok := index.ToIndex()
	if !ok {
		return PairPErr
	}
	switch index {
	case IndexDJI:
		return NewPair("DJI", "USD").SetP(PlatformIndex)
	case IndexSP:
		return NewPair("SP", "USD").SetP(PlatformIndex)
	case IndexIXIC:
		return NewPair("IXIC", "USD").SetP(PlatformIndex)
	case IndexNYA:
		return NewPair("NYA", "USD").SetP(PlatformIndex)
	case IndexRUT:
		return NewPair("RUT", "USD").SetP(PlatformIndex)
	case IndexVIX:
		return NewPair("VIX", "USD").SetP(PlatformIndex)
	case IndexSHH:
		return NewPair("SHH", "CNY").SetP(PlatformIndex)
	}
	return PairPErr
}
