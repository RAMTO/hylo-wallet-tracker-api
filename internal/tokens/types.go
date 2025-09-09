package tokens

import (
	"fmt"
	"time"

	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/utils"
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

	// Decimals is the number of decimal places for this token
	Decimals uint8 `json:"decimals"`

	// FormattedAmount is the human-readable amount with proper decimal adjustment
	FormattedAmount string `json:"formatted_amount"`

	// USDValue is the USD value of this token balance (optional, for display)
	USDValue *float64 `json:"usd_value,omitempty"`
}

// NewTokenBalance creates a new TokenBalance from raw amount and token info
func NewTokenBalance(tokenInfo TokenInfo, rawAmount uint64) *TokenBalance {
	balance := &TokenBalance{
		TokenInfo: tokenInfo,
		RawAmount: rawAmount,
		Decimals:  tokenInfo.Decimals,
	}

	// Format the amount as human-readable
	balance.FormattedAmount = utils.FormatTokenAmount(balance.RawAmount, balance.TokenInfo.Decimals)

	return balance
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
