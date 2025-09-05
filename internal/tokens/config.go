package tokens

import (
	"fmt"
	"os"
	"strings"

	"hylo-wallet-tracker-api/internal/solana"
)

// Config holds token configuration and provides token registry functionality
type Config struct {
	// HyUSDMint can be overridden via HYUSD_MINT environment variable
	HyUSDMint solana.Address

	// SHyUSDMint can be overridden via SHYUSD_MINT environment variable
	SHyUSDMint solana.Address

	// XSOLMint can be overridden via XSOL_MINT environment variable
	XSOLMint solana.Address

	// tokenRegistry is an internal map for fast token lookups
	tokenRegistry map[solana.Address]*TokenInfo
}

// NewConfig creates a new token configuration with default mainnet addresses
// Environment variables can override default mint addresses for different environments
func NewConfig() *Config {
	config := &Config{
		// Default to mainnet addresses from constants
		HyUSDMint:  HyUSDMint,
		SHyUSDMint: SHyUSDMint,
		XSOLMint:   XSOLMint,
	}

	// Load configuration from environment variables
	config.loadFromEnvironment()

	// Build the token registry for fast lookups
	config.buildTokenRegistry()

	return config
}

// loadFromEnvironment loads token mint addresses from environment variables
// This allows different environments (testnet, devnet) to use different addresses
func (c *Config) loadFromEnvironment() {
	// Load hyUSD mint address if provided
	if hyusdMint := os.Getenv("HYUSD_MINT"); hyusdMint != "" {
		c.HyUSDMint = solana.Address(strings.TrimSpace(hyusdMint))
	}

	// Load sHYUSD mint address if provided
	if shyusdMint := os.Getenv("SHYUSD_MINT"); shyusdMint != "" {
		c.SHyUSDMint = solana.Address(strings.TrimSpace(shyusdMint))
	}

	// Load xSOL mint address if provided
	if xsolMint := os.Getenv("XSOL_MINT"); xsolMint != "" {
		c.XSOLMint = solana.Address(strings.TrimSpace(xsolMint))
	}
}

// buildTokenRegistry constructs the internal token registry for fast lookups
func (c *Config) buildTokenRegistry() {
	c.tokenRegistry = make(map[solana.Address]*TokenInfo)

	// hyUSD token info
	c.tokenRegistry[c.HyUSDMint] = &TokenInfo{
		Mint:     c.HyUSDMint,
		Symbol:   HyUSDSymbol,
		Name:     HyUSDName,
		Decimals: HyUSDDecimals,
	}

	// sHYUSD token info
	c.tokenRegistry[c.SHyUSDMint] = &TokenInfo{
		Mint:     c.SHyUSDMint,
		Symbol:   SHyUSDSymbol,
		Name:     SHyUSDName,
		Decimals: SHyUSDDecimals,
	}

	// xSOL token info
	c.tokenRegistry[c.XSOLMint] = &TokenInfo{
		Mint:     c.XSOLMint,
		Symbol:   XSOLSymbol,
		Name:     XSOLName,
		Decimals: XSOLDecimals,
	}
}

// Validate checks if the token configuration is valid
func (c *Config) Validate() error {
	// Validate each mint address
	if err := c.HyUSDMint.Validate(); err != nil {
		return fmt.Errorf("invalid hyUSD mint address: %w", err)
	}

	if err := c.SHyUSDMint.Validate(); err != nil {
		return fmt.Errorf("invalid sHYUSD mint address: %w", err)
	}

	if err := c.XSOLMint.Validate(); err != nil {
		return fmt.Errorf("invalid xSOL mint address: %w", err)
	}

	// Ensure no duplicate mint addresses
	mints := []solana.Address{c.HyUSDMint, c.SHyUSDMint, c.XSOLMint}
	for i, mint1 := range mints {
		for j, mint2 := range mints {
			if i != j && mint1 == mint2 {
				return fmt.Errorf("duplicate mint address detected: %s", mint1)
			}
		}
	}

	// Validate each token info in registry
	for mint, tokenInfo := range c.tokenRegistry {
		if err := tokenInfo.Validate(); err != nil {
			return fmt.Errorf("invalid token info for %s: %w", mint, err)
		}

		// Ensure token info mint matches registry key
		if tokenInfo.Mint != mint {
			return fmt.Errorf("token registry mint mismatch for %s", mint)
		}
	}

	return nil
}

// GetTokenInfo returns token information for a given mint address
// Returns nil if the mint is not supported
func (c *Config) GetTokenInfo(mint solana.Address) *TokenInfo {
	return c.tokenRegistry[mint]
}

// GetSupportedTokens returns all supported token information
func (c *Config) GetSupportedTokens() []*TokenInfo {
	tokens := make([]*TokenInfo, 0, len(c.tokenRegistry))

	// Return in a predictable order: hyUSD, sHYUSD, xSOL
	orderedMints := []solana.Address{c.HyUSDMint, c.SHyUSDMint, c.XSOLMint}

	for _, mint := range orderedMints {
		if tokenInfo, exists := c.tokenRegistry[mint]; exists {
			tokens = append(tokens, tokenInfo)
		}
	}

	return tokens
}

// IsTokenSupported checks if a given mint address is supported
func (c *Config) IsTokenSupported(mint solana.Address) bool {
	_, exists := c.tokenRegistry[mint]
	return exists
}

// GetTokenBySymbol returns token information for a given symbol (e.g., "hyUSD")
// Returns nil if the symbol is not found
func (c *Config) GetTokenBySymbol(symbol string) *TokenInfo {
	for _, tokenInfo := range c.tokenRegistry {
		if tokenInfo.Symbol == symbol {
			return tokenInfo
		}
	}
	return nil
}

// GetSupportedMints returns all supported mint addresses
func (c *Config) GetSupportedMints() []solana.Address {
	mints := make([]solana.Address, 0, len(c.tokenRegistry))

	// Return in a predictable order: hyUSD, sHYUSD, xSOL
	orderedMints := []solana.Address{c.HyUSDMint, c.SHyUSDMint, c.XSOLMint}

	for _, mint := range orderedMints {
		if _, exists := c.tokenRegistry[mint]; exists {
			mints = append(mints, mint)
		}
	}

	return mints
}

// GetMintBySymbol returns the mint address for a given token symbol
// Returns empty address if the symbol is not found
func (c *Config) GetMintBySymbol(symbol string) solana.Address {
	if tokenInfo := c.GetTokenBySymbol(symbol); tokenInfo != nil {
		return tokenInfo.Mint
	}
	return solana.Address("")
}

// NewTokenBalance creates a new TokenBalance for a given mint and raw amount
// Returns nil if the mint is not supported
func (c *Config) NewTokenBalance(mint solana.Address, rawAmount uint64) *TokenBalance {
	tokenInfo := c.GetTokenInfo(mint)
	if tokenInfo == nil {
		return nil
	}

	return NewTokenBalance(*tokenInfo, rawAmount)
}

// NewTokenBalanceBySymbol creates a new TokenBalance for a given symbol and raw amount
// Returns nil if the symbol is not supported
func (c *Config) NewTokenBalanceBySymbol(symbol string, rawAmount uint64) *TokenBalance {
	tokenInfo := c.GetTokenBySymbol(symbol)
	if tokenInfo == nil {
		return nil
	}

	return NewTokenBalance(*tokenInfo, rawAmount)
}

// ValidateTokenBalances validates a map of token balances against supported tokens
func (c *Config) ValidateTokenBalances(balances map[string]*TokenBalance) error {
	for symbol, balance := range balances {
		// Check if symbol is supported
		expectedTokenInfo := c.GetTokenBySymbol(symbol)
		if expectedTokenInfo == nil {
			return fmt.Errorf("unsupported token symbol: %s", symbol)
		}

		// Validate the token balance
		if err := balance.TokenInfo.Validate(); err != nil {
			return fmt.Errorf("invalid token info for %s: %w", symbol, err)
		}

		// Ensure balance token info matches expected info
		if balance.TokenInfo.Mint != expectedTokenInfo.Mint {
			return fmt.Errorf("token balance mint mismatch for %s: expected %s, got %s",
				symbol, expectedTokenInfo.Mint, balance.TokenInfo.Mint)
		}

		if balance.TokenInfo.Symbol != expectedTokenInfo.Symbol {
			return fmt.Errorf("token balance symbol mismatch: expected %s, got %s",
				expectedTokenInfo.Symbol, balance.TokenInfo.Symbol)
		}

		if balance.TokenInfo.Decimals != expectedTokenInfo.Decimals {
			return fmt.Errorf("token balance decimals mismatch for %s: expected %d, got %d",
				symbol, expectedTokenInfo.Decimals, balance.TokenInfo.Decimals)
		}
	}

	return nil
}

// FormatAmount formats a raw token amount as a decimal string for a given mint
// Returns "0" for unsupported mints
func (c *Config) FormatAmount(mint solana.Address, rawAmount uint64) string {
	tokenInfo := c.GetTokenInfo(mint)
	if tokenInfo == nil {
		return "0"
	}

	balance := NewTokenBalance(*tokenInfo, rawAmount)
	return balance.FormatDecimal()
}

// ParseAmount parses a decimal amount string into raw amount for a given mint
// Returns error for unsupported mints or invalid decimal formats
func (c *Config) ParseAmount(mint solana.Address, decimalAmount string) (uint64, error) {
	tokenInfo := c.GetTokenInfo(mint)
	if tokenInfo == nil {
		return 0, fmt.Errorf("unsupported token mint: %s", mint)
	}

	return ParseDecimalAmount(decimalAmount, tokenInfo.Decimals)
}
