package price

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DefaultConfig returns a PriceConfig with sensible defaults for production use
func DefaultConfig() *PriceConfig {
	return &PriceConfig{
		// DexScreener API configuration
		DexScreenerURL:     "https://api.dexscreener.com",
		DexScreenerTimeout: 10 * time.Second,

		// Price validation bounds - reasonable SOL price range
		SOLUSDMinPrice: 50.0,   // Minimum reasonable SOL price in USD
		SOLUSDMaxPrice: 1000.0, // Maximum reasonable SOL price in USD

		// Caching disabled for fresh prices - all requests go to API
		CacheTTL:        0, // Caching disabled
		UpdateInterval:  0, // No scheduled updates
		MaxStalenessSec: 0, // No staleness checks

		// Rate limiting configuration - respect API limits
		RequestsPerMinute: 10,              // Conservative rate limit
		RateLimitWindow:   1 * time.Minute, // Rate limit window

		// Retry configuration - resilient error handling
		MaxRetries:        3,                // Maximum retry attempts
		BaseBackoff:       1 * time.Second,  // Initial backoff delay
		MaxBackoff:        10 * time.Second, // Maximum backoff delay
		BackoffMultiplier: 2.0,              // Exponential backoff multiplier
	}
}

// NewConfig creates a PriceConfig loading values from environment variables
// This is the standard constructor used by the server - follows existing patterns
func NewConfig() *PriceConfig {
	return NewConfigFromEnvironment()
}

// NewConfigFromEnvironment creates a PriceConfig loading values from environment variables
// Falls back to defaults for any missing or invalid environment variables
func NewConfigFromEnvironment() *PriceConfig {
	config := DefaultConfig()

	// Load DexScreener configuration
	if url := os.Getenv("DEXSCREENER_API_URL"); url != "" {
		config.DexScreenerURL = strings.TrimSpace(url)
	}

	if timeoutStr := os.Getenv("DEXSCREENER_TIMEOUT_SEC"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			config.DexScreenerTimeout = time.Duration(timeout) * time.Second
		}
	}

	// Load price validation bounds
	if minPriceStr := os.Getenv("SOL_USD_MIN_PRICE"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil && minPrice > 0 {
			config.SOLUSDMinPrice = minPrice
		}
	}

	if maxPriceStr := os.Getenv("SOL_USD_MAX_PRICE"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil && maxPrice > 0 {
			config.SOLUSDMaxPrice = maxPrice
		}
	}

	// Load caching configuration
	if cacheTTLStr := os.Getenv("PRICE_CACHE_TTL_SEC"); cacheTTLStr != "" {
		if cacheTTL, err := strconv.Atoi(cacheTTLStr); err == nil && cacheTTL > 0 {
			config.CacheTTL = time.Duration(cacheTTL) * time.Second
		}
	}

	if updateIntervalStr := os.Getenv("PRICE_UPDATE_INTERVAL_SEC"); updateIntervalStr != "" {
		if updateInterval, err := strconv.Atoi(updateIntervalStr); err == nil && updateInterval > 0 {
			config.UpdateInterval = time.Duration(updateInterval) * time.Second
		}
	}

	if maxStalenessStr := os.Getenv("PRICE_MAX_STALENESS_SEC"); maxStalenessStr != "" {
		if maxStaleness, err := strconv.Atoi(maxStalenessStr); err == nil && maxStaleness > 0 {
			config.MaxStalenessSec = maxStaleness
		}
	}

	// Load rate limiting configuration
	if rpmStr := os.Getenv("PRICE_REQUESTS_PER_MINUTE"); rpmStr != "" {
		if rpm, err := strconv.Atoi(rpmStr); err == nil && rpm > 0 {
			config.RequestsPerMinute = rpm
		}
	}

	// Load retry configuration
	if maxRetriesStr := os.Getenv("PRICE_MAX_RETRIES"); maxRetriesStr != "" {
		if maxRetries, err := strconv.Atoi(maxRetriesStr); err == nil && maxRetries >= 0 {
			config.MaxRetries = maxRetries
		}
	}

	if baseBackoffStr := os.Getenv("PRICE_BASE_BACKOFF_SEC"); baseBackoffStr != "" {
		if baseBackoff, err := strconv.Atoi(baseBackoffStr); err == nil && baseBackoff > 0 {
			config.BaseBackoff = time.Duration(baseBackoff) * time.Second
		}
	}

	if maxBackoffStr := os.Getenv("PRICE_MAX_BACKOFF_SEC"); maxBackoffStr != "" {
		if maxBackoff, err := strconv.Atoi(maxBackoffStr); err == nil && maxBackoff > 0 {
			config.MaxBackoff = time.Duration(maxBackoff) * time.Second
		}
	}

	return config
}

// Validate checks the configuration for consistency and reasonable values
func (c *PriceConfig) Validate() error {
	// Validate DexScreener configuration
	if c.DexScreenerURL == "" {
		return fmt.Errorf("DexScreener URL cannot be empty")
	}
	if c.DexScreenerTimeout <= 0 {
		return fmt.Errorf("DexScreener timeout must be positive, got %v", c.DexScreenerTimeout)
	}

	// Validate price bounds
	if c.SOLUSDMinPrice <= 0 {
		return fmt.Errorf("SOL USD minimum price must be positive, got %v", c.SOLUSDMinPrice)
	}
	if c.SOLUSDMaxPrice <= c.SOLUSDMinPrice {
		return fmt.Errorf("SOL USD maximum price (%v) must be greater than minimum price (%v)",
			c.SOLUSDMaxPrice, c.SOLUSDMinPrice)
	}

	// Caching is disabled for fresh prices - skip cache validation
	// Cache TTL of 0 means no caching, which is the desired behavior

	// No cache-related warnings needed when caching is disabled

	// Validate rate limiting
	if c.RequestsPerMinute <= 0 {
		return fmt.Errorf("requests per minute must be positive, got %v", c.RequestsPerMinute)
	}
	if c.RateLimitWindow <= 0 {
		return fmt.Errorf("rate limit window must be positive, got %v", c.RateLimitWindow)
	}

	// Validate retry configuration
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative, got %v", c.MaxRetries)
	}
	if c.BaseBackoff <= 0 {
		return fmt.Errorf("base backoff must be positive, got %v", c.BaseBackoff)
	}
	if c.MaxBackoff <= 0 {
		return fmt.Errorf("max backoff must be positive, got %v", c.MaxBackoff)
	}
	if c.BackoffMultiplier <= 1.0 {
		return fmt.Errorf("backoff multiplier must be greater than 1.0, got %v", c.BackoffMultiplier)
	}
	if c.MaxBackoff < c.BaseBackoff {
		return fmt.Errorf("max backoff (%v) must be >= base backoff (%v)",
			c.MaxBackoff, c.BaseBackoff)
	}

	return nil
}

// GetMaxStaleness returns the maximum staleness duration
func (c *PriceConfig) GetMaxStaleness() time.Duration {
	return time.Duration(c.MaxStalenessSec) * time.Second
}

// GetRateLimitDelay calculates the minimum delay between requests based on rate limiting
func (c *PriceConfig) GetRateLimitDelay() time.Duration {
	if c.RequestsPerMinute <= 0 {
		return 0
	}
	return c.RateLimitWindow / time.Duration(c.RequestsPerMinute)
}

// CalculateBackoff calculates exponential backoff delay for retry attempt
func (c *PriceConfig) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.BaseBackoff
	}

	// Calculate exponential backoff: base * multiplier^attempt
	backoff := c.BaseBackoff
	for i := 0; i < attempt; i++ {
		backoff = time.Duration(float64(backoff) * c.BackoffMultiplier)
		if backoff > c.MaxBackoff {
			return c.MaxBackoff
		}
	}

	return backoff
}

// IsValidSOLPrice checks if a price is within configured bounds
func (c *PriceConfig) IsValidSOLPrice(price float64) bool {
	return price >= c.SOLUSDMinPrice && price <= c.SOLUSDMaxPrice && price > 0
}

// ShouldCache determines if the current configuration supports caching
// Always returns false now since we want fresh prices
func (c *PriceConfig) ShouldCache() bool {
	return false // Caching disabled for fresh prices
}

// String returns a string representation of the config (without sensitive data)
func (c *PriceConfig) String() string {
	return fmt.Sprintf("PriceConfig{URL:%s, CacheTTL:%v, UpdateInterval:%v, RPM:%d, MaxRetries:%d}",
		c.DexScreenerURL, c.CacheTTL, c.UpdateInterval, c.RequestsPerMinute, c.MaxRetries)
}
