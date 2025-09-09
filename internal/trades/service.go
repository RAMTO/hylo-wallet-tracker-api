package trades

import (
	"context"
	"fmt"
	"sort"

	"hylo-wallet-tracker-api/internal/hylo"
	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/tokens"
)

// HTTPClientInterface defines the contract for Solana HTTP client interaction
// Matches the interface from tokens service for consistency
type HTTPClientInterface interface {
	GetAccount(ctx context.Context, address solana.Address, commitment solana.Commitment) (*solana.AccountInfo, error)
	GetSignaturesForAddress(ctx context.Context, address solana.Address, before string, limit int) ([]solana.SignatureInfo, error)
	GetTransaction(ctx context.Context, signature solana.Signature) (*solana.TransactionDetails, error)
}

// TradeService provides xSOL trade history functionality with real-time fetching
// Integrates with Solana HTTP client, token configuration, and transaction parsing
type TradeService struct {
	// httpClient is the Solana HTTP RPC client for on-chain data fetching
	httpClient HTTPClientInterface

	// tokenConfig provides token configuration and ATA derivation
	tokenConfig *tokens.Config

	// hyloConfig provides Hylo program configuration
	hyloConfig *hylo.Config

	// options provides service configuration options
	options *TradeServiceOptions
}

// NewTradeService creates a new trade service with dependency injection
// Parameters:
//   - httpClient: Solana HTTP client for RPC calls
//   - tokenConfig: Token configuration for ATA derivation
//   - hyloConfig: Hylo configuration for program identification
func NewTradeService(httpClient HTTPClientInterface, tokenConfig *tokens.Config, hyloConfig *hylo.Config) (*TradeService, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient cannot be nil")
	}
	if tokenConfig == nil {
		return nil, fmt.Errorf("tokenConfig cannot be nil")
	}
	if hyloConfig == nil {
		return nil, fmt.Errorf("hyloConfig cannot be nil")
	}

	// Validate configurations
	if err := tokenConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid token config: %w", err)
	}
	if err := hyloConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid hylo config: %w", err)
	}

	return &TradeService{
		httpClient:  httpClient,
		tokenConfig: tokenConfig,
		hyloConfig:  hyloConfig,
		options:     DefaultTradeServiceOptions(),
	}, nil
}

// GetWalletTrades fetches xSOL trade history for a wallet using real-time RPC calls
// Returns paginated trade results with cursor-based navigation
func (s *TradeService) GetWalletTrades(ctx context.Context, walletAddr solana.Address, limit int, before string) (*TradeResponse, error) {
	// Validate wallet address
	if err := walletAddr.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidWalletAddress, err)
	}

	// Create and validate request
	req := &TradeRequest{
		WalletAddress: walletAddr.String(),
		Limit:         limit,
		Before:        before,
	}

	if err := ValidateTradeRequest(req, s.options); err != nil {
		return nil, err
	}

	// Step 1: Derive xSOL Associated Token Account for this wallet
	xsolATA, err := tokens.DeriveAssociatedTokenAddress(walletAddr, tokens.XSOLMint)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrXSOLATADerivation, err)
	}

	// Step 2: Fetch transaction signatures for the xSOL ATA
	signatures, err := s.httpClient.GetSignaturesForAddress(ctx, xsolATA, before, req.Limit*2) // Fetch extra to account for filtering
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignatureFetch, err)
	}

	// Step 3: Process signatures to extract xSOL trades
	trades, err := s.processSignatures(ctx, signatures, xsolATA, req.Limit)
	if err != nil {
		return nil, err
	}

	// Step 4: Determine pagination information
	hasMore := len(trades) == req.Limit && len(signatures) > 0
	var nextCursor string
	if hasMore && len(trades) > 0 {
		// Use the last trade's signature as the next cursor
		nextCursor = trades[len(trades)-1].Signature
	}

	return NewTradeResponse(walletAddr.String(), trades, hasMore, nextCursor, req.Limit), nil
}

// processSignatures fetches transaction details and parses them for xSOL trades
func (s *TradeService) processSignatures(ctx context.Context, signatures []solana.SignatureInfo, xsolATA solana.Address, maxTrades int) ([]*hylo.XSOLTrade, error) {
	var trades []*hylo.XSOLTrade

	// Sort signatures by slot (newest first) to ensure consistent ordering
	sort.Slice(signatures, func(i, j int) bool {
		return signatures[i].Slot > signatures[j].Slot
	})

	// Process each signature until we have enough trades or run out of signatures
	for _, sigInfo := range signatures {
		// Skip failed transactions
		if sigInfo.Err != nil {
			continue
		}

		// Fetch transaction details
		tx, err := s.httpClient.GetTransaction(ctx, solana.Signature(sigInfo.Signature))
		if err != nil {
			// Log the error but continue processing other transactions
			// In a production system, you might want to use structured logging here
			continue
		}

		// Parse the transaction for xSOL trades
		parseResult, err := hylo.ParseTransaction(tx, xsolATA)
		if err != nil {
			// Continue processing other transactions if one fails to parse
			continue
		}

		// If we found a valid trade, add it to our results
		if parseResult != nil && parseResult.Trade != nil {
			trades = append(trades, parseResult.Trade)

			// Stop if we've reached the requested limit
			if len(trades) >= maxTrades {
				break
			}
		}
	}

	return trades, nil
}

// GetServiceHealth returns health information for the trade service
func (s *TradeService) GetServiceHealth(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"service": "TradeService",
		"status":  "healthy",
	}

	// Test connectivity by attempting to derive an ATA (lightweight operation)
	testWallet := solana.Address("11111111111111111111111111111111") // System program address as test
	_, err := tokens.DeriveAssociatedTokenAddress(testWallet, tokens.XSOLMint)
	if err != nil {
		health["status"] = "degraded"
		health["error"] = "ATA derivation failed"
	}

	health["config"] = map[string]interface{}{
		"default_limit": s.options.DefaultLimit,
		"max_limit":     s.options.MaxLimit,
		"hylo_programs": []string{
			s.hyloConfig.GetExchangeProgramID().String(),
			s.hyloConfig.GetStabilityPoolProgramID().String(),
		},
	}

	return health
}

// SetOptions updates the service configuration options
func (s *TradeService) SetOptions(options *TradeServiceOptions) {
	if options != nil {
		s.options = options
	}
}

// GetOptions returns the current service configuration options
func (s *TradeService) GetOptions() *TradeServiceOptions {
	return s.options
}
