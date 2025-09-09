package trades

import (
	"fmt"
	"time"

	"hylo-wallet-tracker-api/internal/hylo"
)

// TradeRequest represents the input parameters for fetching wallet trades
type TradeRequest struct {
	// WalletAddress is the wallet to fetch trades for
	WalletAddress string `json:"walletAddress" validate:"required"`

	// Limit is the maximum number of trades to return (1-50)
	Limit int `json:"limit" validate:"min=1,max=50"`

	// Before is the signature cursor for pagination (optional)
	// When provided, returns trades before this signature
	Before string `json:"before,omitempty"`
}

// TradeResponse represents the response structure for wallet trades
type TradeResponse struct {
	// Trades is the array of xSOL trades for this wallet
	Trades []*hylo.XSOLTrade `json:"trades"`

	// Pagination metadata for frontend navigation
	Pagination PaginationInfo `json:"pagination"`

	// Request metadata
	WalletAddress string    `json:"walletAddress"`
	RequestedAt   time.Time `json:"requestedAt"`
	Count         int       `json:"count"` // Number of trades returned
}

// PaginationInfo provides cursor-based pagination metadata
type PaginationInfo struct {
	// HasMore indicates if there are more trades available
	HasMore bool `json:"hasMore"`

	// NextCursor is the signature cursor for the next page
	// Only present when HasMore is true
	NextCursor string `json:"nextCursor,omitempty"`

	// Limit is the maximum number of items per page
	Limit int `json:"limit"`

	// Count is the number of items in the current response
	Count int `json:"count"`
}

// TradeServiceOptions provides configuration options for the trade service
type TradeServiceOptions struct {
	// DefaultLimit is the default number of trades to fetch when no limit is specified
	DefaultLimit int

	// MaxLimit is the maximum allowed limit per request
	MaxLimit int

	// EnableValidation enables request validation
	EnableValidation bool
}

// DefaultTradeServiceOptions returns sensible defaults for the trade service
func DefaultTradeServiceOptions() *TradeServiceOptions {
	return &TradeServiceOptions{
		DefaultLimit:     10, // Default to last 10 trades
		MaxLimit:         50, // Maximum 50 trades per request to prevent abuse
		EnableValidation: true,
	}
}

// ValidateTradeRequest validates the trade request parameters
func ValidateTradeRequest(req *TradeRequest, options *TradeServiceOptions) error {
	if req.WalletAddress == "" {
		return ErrInvalidWalletAddress
	}

	// Apply default limit if not specified
	if req.Limit <= 0 {
		req.Limit = options.DefaultLimit
	}

	// Enforce maximum limit
	if req.Limit > options.MaxLimit {
		req.Limit = options.MaxLimit
	}

	return nil
}

// Trade service errors
var (
	ErrInvalidWalletAddress = fmt.Errorf("wallet address is required and must be valid")
	ErrInvalidLimit         = fmt.Errorf("limit must be between 1 and 50")
	ErrServiceNotReady      = fmt.Errorf("trade service is not properly initialized")
	ErrXSOLATADerivation    = fmt.Errorf("failed to derive xSOL Associated Token Account")
	ErrSignatureFetch       = fmt.Errorf("failed to fetch transaction signatures")
	ErrTransactionFetch     = fmt.Errorf("failed to fetch transaction details")
	ErrTradeParsing         = fmt.Errorf("failed to parse transaction for trade details")
)

// NewTradeResponse creates a new trade response with proper initialization
func NewTradeResponse(walletAddress string, trades []*hylo.XSOLTrade, hasMore bool, nextCursor string, limit int) *TradeResponse {
	return &TradeResponse{
		Trades:        trades,
		WalletAddress: walletAddress,
		RequestedAt:   time.Now(),
		Count:         len(trades),
		Pagination: PaginationInfo{
			HasMore:    hasMore,
			NextCursor: nextCursor,
			Limit:      limit,
			Count:      len(trades),
		},
	}
}
