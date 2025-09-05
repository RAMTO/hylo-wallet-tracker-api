package tokens

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"hylo-wallet-tracker-api/internal/solana"
)

// TokenInfo represents metadata and configuration for a single token
type TokenInfo struct {
	// Mint is the Solana mint address for this token
	Mint solana.Address `json:"mint"`

	// Symbol is the short identifier (e.g., "hyUSD", "xSOL")
	Symbol string `json:"symbol"`

	// Name is the human-readable display name
	Name string `json:"name"`

	// Decimals is the number of decimal places for this token
	Decimals uint8 `json:"decimals"`
}

// Validate checks if the TokenInfo has valid configuration
func (ti *TokenInfo) Validate() error {
	if err := ti.Mint.Validate(); err != nil {
		return fmt.Errorf("invalid mint address: %w", err)
	}

	if ti.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	if ti.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if ti.Decimals > 18 {
		return fmt.Errorf("decimals cannot exceed 18")
	}

	return nil
}

// TokenBalance represents a token balance with formatting utilities
type TokenBalance struct {
	// TokenInfo contains the token metadata (not exported in JSON)
	TokenInfo TokenInfo `json:"-"`

	// RawAmount is the raw token amount as stored on-chain (without decimal adjustment)
	RawAmount uint64 `json:"raw_amount"`

	// DecimalAmount is the human-readable amount adjusted for token decimals
	DecimalAmount string `json:"decimal_amount"`

	// USDValue is the USD value of this token balance (optional, for display)
	USDValue *float64 `json:"usd_value,omitempty"`
}

// NewTokenBalance creates a new TokenBalance from raw amount and token info
func NewTokenBalance(tokenInfo TokenInfo, rawAmount uint64) *TokenBalance {
	balance := &TokenBalance{
		TokenInfo: tokenInfo,
		RawAmount: rawAmount,
	}

	// Format the decimal amount using the token's decimal precision
	balance.DecimalAmount = balance.FormatDecimal()

	return balance
}

// FormatDecimal converts the raw amount to decimal representation with proper precision
func (tb *TokenBalance) FormatDecimal() string {
	if tb.RawAmount == 0 {
		return "0"
	}

	// Convert raw amount to big.Int for precise decimal math
	rawBig := new(big.Int).SetUint64(tb.RawAmount)

	// Calculate the divisor (10^decimals)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tb.TokenInfo.Decimals)), nil)

	// Perform division to get integer part
	integerPart := new(big.Int).Div(rawBig, divisor)

	// Get remainder for fractional part
	remainder := new(big.Int).Mod(rawBig, divisor)

	// If no fractional part, return integer part
	if remainder.Cmp(big.NewInt(0)) == 0 {
		return integerPart.String()
	}

	// Format fractional part with leading zeros if needed
	fracStr := remainder.String()
	if len(fracStr) < int(tb.TokenInfo.Decimals) {
		// Pad with leading zeros
		fracStr = strings.Repeat("0", int(tb.TokenInfo.Decimals)-len(fracStr)) + fracStr
	}

	// Remove trailing zeros from fractional part
	fracStr = strings.TrimRight(fracStr, "0")

	// If no fractional digits remain, return just integer part
	if fracStr == "" {
		return integerPart.String()
	}

	return integerPart.String() + "." + fracStr
}

// IsZero returns true if the token balance is zero
func (tb *TokenBalance) IsZero() bool {
	return tb.RawAmount == 0
}

// SetUSDValue sets the USD value for this token balance
func (tb *TokenBalance) SetUSDValue(usdValue float64) {
	tb.USDValue = &usdValue
}

// WalletBalances represents all token balances for a specific wallet
type WalletBalances struct {
	// Wallet is the wallet address these balances belong to
	Wallet solana.Address `json:"wallet"`

	// Slot is the Solana slot when these balances were fetched
	Slot solana.Slot `json:"slot"`

	// UpdatedAt is the timestamp when these balances were last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Balances is a map of token symbol to token balance
	Balances map[string]*TokenBalance `json:"balances"`

	// TotalUSDValue is the sum of all token balances in USD (optional)
	TotalUSDValue *float64 `json:"total_usd_value,omitempty"`
}

// NewWalletBalances creates a new WalletBalances instance
func NewWalletBalances(wallet solana.Address, slot solana.Slot) *WalletBalances {
	return &WalletBalances{
		Wallet:    wallet,
		Slot:      slot,
		UpdatedAt: time.Now(),
		Balances:  make(map[string]*TokenBalance),
	}
}

// AddBalance adds a token balance to the wallet balances
func (wb *WalletBalances) AddBalance(balance *TokenBalance) {
	wb.Balances[balance.TokenInfo.Symbol] = balance
	wb.UpdatedAt = time.Now()
}

// GetBalance returns the balance for a specific token symbol
func (wb *WalletBalances) GetBalance(symbol string) (*TokenBalance, bool) {
	balance, exists := wb.Balances[symbol]
	return balance, exists
}

// HasAnyBalance returns true if the wallet has any non-zero token balances
func (wb *WalletBalances) HasAnyBalance() bool {
	for _, balance := range wb.Balances {
		if !balance.IsZero() {
			return true
		}
	}
	return false
}

// CalculateTotalUSDValue calculates and sets the total USD value of all balances
func (wb *WalletBalances) CalculateTotalUSDValue() {
	var total float64
	hasAnyUSDValue := false

	for _, balance := range wb.Balances {
		if balance.USDValue != nil {
			total += *balance.USDValue
			hasAnyUSDValue = true
		}
	}

	if hasAnyUSDValue {
		wb.TotalUSDValue = &total
	}
}

// GetHyUSDBalance is a convenience method to get hyUSD balance
func (wb *WalletBalances) GetHyUSDBalance() (*TokenBalance, bool) {
	return wb.GetBalance(HyUSDSymbol)
}

// GetSHyUSDBalance is a convenience method to get sHYUSD balance
func (wb *WalletBalances) GetSHyUSDBalance() (*TokenBalance, bool) {
	return wb.GetBalance(SHyUSDSymbol)
}

// GetXSOLBalance is a convenience method to get xSOL balance
func (wb *WalletBalances) GetXSOLBalance() (*TokenBalance, bool) {
	return wb.GetBalance(XSOLSymbol)
}

// Validate checks if the WalletBalances has valid data
func (wb *WalletBalances) Validate() error {
	if err := wb.Wallet.Validate(); err != nil {
		return fmt.Errorf("invalid wallet address: %w", err)
	}

	if wb.Balances == nil {
		return fmt.Errorf("balances map cannot be nil")
	}

	// Validate each token balance
	for symbol, balance := range wb.Balances {
		if balance == nil {
			return fmt.Errorf("balance for %s cannot be nil", symbol)
		}

		if err := balance.TokenInfo.Validate(); err != nil {
			return fmt.Errorf("invalid token info for %s: %w", symbol, err)
		}

		if balance.TokenInfo.Symbol != symbol {
			return fmt.Errorf("balance symbol mismatch: expected %s, got %s", symbol, balance.TokenInfo.Symbol)
		}
	}

	return nil
}

// ParseDecimalAmount parses a decimal string into raw token amount
// This is the reverse of FormatDecimal() - useful for testing and input parsing
func ParseDecimalAmount(decimalStr string, decimals uint8) (uint64, error) {
	if decimalStr == "" || decimalStr == "0" {
		return 0, nil
	}

	// Split on decimal point
	parts := strings.Split(decimalStr, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid decimal format: %s", decimalStr)
	}

	integerPart := parts[0]
	fractionalPart := ""

	if len(parts) == 2 {
		fractionalPart = parts[1]
	}

	// Validate fractional part doesn't exceed token decimals
	if len(fractionalPart) > int(decimals) {
		return 0, fmt.Errorf("fractional part exceeds token decimals (%d): %s", decimals, decimalStr)
	}

	// Pad fractional part with trailing zeros to match token decimals
	for len(fractionalPart) < int(decimals) {
		fractionalPart += "0"
	}

	// Combine integer and fractional parts
	combinedStr := integerPart + fractionalPart

	// Parse as uint64
	rawAmount, err := strconv.ParseUint(combinedStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse decimal amount: %w", err)
	}

	return rawAmount, nil
}
