package types

type Asset struct {
	Coin   string
	Free   float64
	Frozen float64
	Total  float64
	EqUsd  float64
}

type Assets struct {
	Assets        map[string]Asset
	UniMMR        float64
	TotalUsdEq    float64
	FreeUsdEq     float64
	AccountMargin float64
	Borrowed      float64
}
