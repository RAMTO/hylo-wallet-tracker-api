package price

import (
	"time"
)

// SOLUSDPrice represents a SOL/USD price point with metadata
type SOLUSDPrice struct {
	// Price is the SOL price in USD (e.g., 182.35)
	Price float64 `json:"price"`

	// Timestamp indicates when this price was fetched/updated
	Timestamp time.Time `json:"timestamp"`

	// Source identifies the price data provider (e.g., "dexscreener")
	Source string `json:"source"`

	// Pair identifies the trading pair used (e.g., "SOL/USDC", "SOL/USDT")
	Pair string `json:"pair"`

	// Liquidity represents the liquidity amount for this pair (if available)
	Liquidity float64 `json:"liquidity,omitempty"`

	// Volume24h represents 24-hour trading volume (if available)
	Volume24h float64 `json:"volume_24h,omitempty"`
}

// XSOLPrice represents xSOL price data in both SOL and USD terms
type XSOLPrice struct {
	// PriceInSOL is the xSOL price denominated in SOL (e.g., 2.7149)
	PriceInSOL float64 `json:"price_sol"`

	// PriceInUSD is the xSOL price denominated in USD (e.g., 494.04)
	PriceInUSD float64 `json:"price_usd"`

	// Timestamp indicates when this price was calculated
	Timestamp time.Time `json:"timestamp"`

	// CollateralRatio represents the protocol health metric
	CollateralRatio float64 `json:"collateral_ratio,omitempty"`

	// EffectiveLeverage represents the current leverage multiplier
	EffectiveLeverage float64 `json:"effective_leverage,omitempty"`
}

// CombinedPriceResponse represents the complete price response for the API
// This matches the PRD specification for the /price endpoint response format
type CombinedPriceResponse struct {
	// SOLUSD is the current SOL price in USD
	SOLUSD float64 `json:"sol_usd"`

	// XSOLInSOL is the current xSOL price in SOL terms
	XSOLInSOL float64 `json:"xsol_sol"`

	// XSOLInUSD is the current xSOL price in USD terms
	XSOLInUSD float64 `json:"xsol_usd"`

	// UpdatedAt indicates the timestamp of the most recent price update
	UpdatedAt time.Time `json:"updated_at"`
}

// PriceConfig holds configuration for price service operations
type PriceConfig struct {
	// DexScreener API configuration
	DexScreenerURL     string        `json:"dexscreener_url"`
	DexScreenerTimeout time.Duration `json:"dexscreener_timeout"`

	// Price validation bounds
	SOLUSDMinPrice float64 `json:"sol_usd_min_price"`
	SOLUSDMaxPrice float64 `json:"sol_usd_max_price"`

	// Caching configuration
	CacheTTL        time.Duration `json:"cache_ttl"`
	UpdateInterval  time.Duration `json:"update_interval"`
	MaxStalenessSec int           `json:"max_staleness_sec"`

	// Rate limiting configuration
	RequestsPerMinute int           `json:"requests_per_minute"`
	RateLimitWindow   time.Duration `json:"rate_limit_window"`

	// Retry configuration
	MaxRetries        int           `json:"max_retries"`
	BaseBackoff       time.Duration `json:"base_backoff"`
	MaxBackoff        time.Duration `json:"max_backoff"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
}

// DexScreenerPair represents a trading pair from DexScreener API response
type DexScreenerPair struct {
	ChainID     string      `json:"chainId"`
	DexID       string      `json:"dexId"`
	URL         string      `json:"url"`
	PairAddr    string      `json:"pairAddress"`
	BaseToken   Token       `json:"baseToken"`
	QuoteToken  Token       `json:"quoteToken"`
	PriceNative string      `json:"priceNative"`
	PriceUSD    string      `json:"priceUsd"`
	Txns        Txns        `json:"txns"`
	Volume      Volume      `json:"volume"`
	PriceChange PriceChange `json:"priceChange"`
	Liquidity   Liquidity   `json:"liquidity"`
	FDV         float64     `json:"fdv"`
	MarketCap   float64     `json:"marketCap"`
}

// Token represents token information in DexScreener response
type Token struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

// Txns represents transaction counts in DexScreener response
type Txns struct {
	M5  TxnCount `json:"m5"`
	H1  TxnCount `json:"h1"`
	H6  TxnCount `json:"h6"`
	H24 TxnCount `json:"h24"`
}

// TxnCount represents buy/sell transaction counts
type TxnCount struct {
	Buys  int `json:"buys"`
	Sells int `json:"sells"`
}

// Volume represents trading volume data
type Volume struct {
	H24 float64 `json:"h24"`
	H6  float64 `json:"h6"`
	H1  float64 `json:"h1"`
	M5  float64 `json:"m5"`
}

// PriceChange represents price change percentages
type PriceChange struct {
	M5  float64 `json:"m5"`
	H1  float64 `json:"h1"`
	H6  float64 `json:"h6"`
	H24 float64 `json:"h24"`
}

// Liquidity represents liquidity information
type Liquidity struct {
	USD   float64 `json:"usd"`
	Base  float64 `json:"base"`
	Quote float64 `json:"quote"`
}

// DexScreenerResponse represents the complete API response from DexScreener
type DexScreenerResponse struct {
	SchemaVersion string            `json:"schemaVersion"`
	Pairs         []DexScreenerPair `json:"pairs"`
}

// IsValid checks if the SOL/USD price is within acceptable bounds
func (p *SOLUSDPrice) IsValid(minPrice, maxPrice float64) bool {
	return p.Price >= minPrice && p.Price <= maxPrice && p.Price > 0
}

// IsStale checks if the price data is older than the specified max age
func (p *SOLUSDPrice) IsStale(maxAge time.Duration) bool {
	return time.Since(p.Timestamp) > maxAge
}

// IsValid checks if the xSOL price data is reasonable
func (x *XSOLPrice) IsValid() bool {
	return x.PriceInSOL > 0 && x.PriceInUSD > 0
}

// IsStale checks if the xSOL price data is older than the specified max age
func (x *XSOLPrice) IsStale(maxAge time.Duration) bool {
	return time.Since(x.Timestamp) > maxAge
}
