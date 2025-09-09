package hylo

import (
	"fmt"
	"time"
)

// XSOLTrade represents a parsed xSOL trade transaction with all relevant details
type XSOLTrade struct {
	// Transaction identifiers
	Signature string `json:"signature"` // Transaction signature
	Slot      uint64 `json:"slot"`      // Solana slot number
	BlockTime int64  `json:"blockTime"` // Unix timestamp

	// Trade details
	Side          string `json:"side"`          // BUY or SELL (from constants)
	XSOLAmount    string `json:"xsolAmount"`    // Formatted xSOL amount (e.g., "1.5")
	CounterAmount string `json:"counterAmount"` // Formatted counter-asset amount
	CounterAsset  string `json:"counterAsset"`  // "SOL" or "hyUSD"

	// Display fields
	Timestamp   time.Time `json:"timestamp"`   // Parsed timestamp
	ExplorerURL string    `json:"explorerUrl"` // Solscan transaction URL

	// Raw amounts for calculations (optional, for internal use)
	XSOLAmountRaw    uint64 `json:"-"` // Raw xSOL amount (lamports/smallest unit)
	CounterAmountRaw uint64 `json:"-"` // Raw counter-asset amount
}

// TradeParseResult contains the result of transaction parsing
type TradeParseResult struct {
	Trade *XSOLTrade `json:"trade,omitempty"` // Parsed trade, nil if not an xSOL trade
	Error string     `json:"error,omitempty"` // Parse error message, if any
}

// NewXSOLTrade creates a new XSOLTrade with basic transaction info
func NewXSOLTrade(signature string, slot uint64, blockTime int64) *XSOLTrade {
	var timestamp time.Time
	if blockTime > 0 {
		timestamp = time.Unix(blockTime, 0)
	}

	return &XSOLTrade{
		Signature:   signature,
		Slot:        slot,
		BlockTime:   blockTime,
		Timestamp:   timestamp,
		ExplorerURL: generateSolscanURL(signature),
	}
}

// SetTradeDetails sets the trade-specific information
func (t *XSOLTrade) SetTradeDetails(side string, xsolAmount, counterAmount uint64, counterAsset string) {
	t.Side = side
	t.XSOLAmountRaw = xsolAmount
	t.CounterAmountRaw = counterAmount
	t.CounterAsset = counterAsset

	// Format amounts for display (using 6 decimals for xSOL, 9 for SOL, 6 for hyUSD)
	t.XSOLAmount = formatAmount(xsolAmount, 6) // xSOL has 6 decimals

	switch counterAsset {
	case "SOL":
		t.CounterAmount = formatAmount(counterAmount, 9) // SOL has 9 decimals (lamports)
	case "hyUSD":
		t.CounterAmount = formatAmount(counterAmount, 6) // hyUSD has 6 decimals
	default:
		t.CounterAmount = fmt.Sprintf("%d", counterAmount) // Raw amount as fallback
	}
}

// generateSolscanURL creates a Solscan explorer URL for the transaction
func generateSolscanURL(signature string) string {
	return fmt.Sprintf("https://solscan.io/tx/%s", signature)
}

// formatAmount formats raw token amount with proper decimal precision
// This is a simplified version of the tokens package formatting logic
func formatAmount(rawAmount uint64, decimals uint8) string {
	if rawAmount == 0 {
		return "0"
	}

	// Calculate divisor for decimal precision
	divisor := uint64(1)
	for i := uint8(0); i < decimals; i++ {
		divisor *= 10
	}

	integerPart := rawAmount / divisor
	fractionalPart := rawAmount % divisor

	if fractionalPart == 0 {
		return fmt.Sprintf("%d", integerPart)
	}

	// Format with decimals, removing trailing zeros
	fracStr := fmt.Sprintf("%0*d", decimals, fractionalPart)
	// Remove trailing zeros
	for len(fracStr) > 1 && fracStr[len(fracStr)-1] == '0' {
		fracStr = fracStr[:len(fracStr)-1]
	}

	return fmt.Sprintf("%d.%s", integerPart, fracStr)
}

// IsValidTrade checks if the trade has valid data
func (t *XSOLTrade) IsValidTrade() bool {
	return t.Signature != "" &&
		(t.Side == TradeSideBuy || t.Side == TradeSideSell) &&
		t.XSOLAmountRaw > 0 &&
		(t.CounterAsset == "SOL" || t.CounterAsset == "hyUSD")
}
