package trades

import (
	"context"
	"fmt"
	"hylo-wallet-tracker-api/internal/hylo"
	"hylo-wallet-tracker-api/internal/logger"
	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/tokens"
	"log/slog"
	"sort"
	"time"
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

	// logger for structured logging
	logger *logger.Logger

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

	// Initialize logger for service
	serviceLogger := logger.NewFromEnv().WithComponent("trade-service")

	// Validate configurations
	if err := tokenConfig.Validate(); err != nil {
		serviceLogger.LogHandlerError(context.Background(), "service_initialization", err,
			slog.String("error_type", "token_config_validation"))
		return nil, fmt.Errorf("invalid token config: %w", err)
	}
	if err := hyloConfig.Validate(); err != nil {
		serviceLogger.LogHandlerError(context.Background(), "service_initialization", err,
			slog.String("error_type", "hylo_config_validation"))
		return nil, fmt.Errorf("invalid hylo config: %w", err)
	}

	serviceLogger.InfoContext(context.Background(), "Initializing Trade service")

	service := &TradeService{
		httpClient:  httpClient,
		tokenConfig: tokenConfig,
		hyloConfig:  hyloConfig,
		logger:      serviceLogger,
		options:     DefaultTradeServiceOptions(),
	}

	serviceLogger.InfoContext(context.Background(), "Trade service initialized successfully")
	return service, nil
}

// GetWalletTrades fetches xSOL trade history for a wallet using real-time RPC calls
// Returns paginated trade results with cursor-based navigation
func (s *TradeService) GetWalletTrades(ctx context.Context, walletAddr solana.Address, limit int, before string) (*TradeResponse, error) {
	startTime := time.Now()

	s.logger.InfoContext(ctx, "Getting wallet trades",
		slog.String("wallet", walletAddr.String()),
		slog.Int("limit", limit),
		slog.String("before", before))

	// Validate wallet address
	if err := walletAddr.Validate(); err != nil {
		s.logger.LogValidationError(ctx, "get_wallet_trades", "wallet", walletAddr, err)
		return nil, fmt.Errorf("%w: %v", ErrInvalidWalletAddress, err)
	}

	// Create and validate request
	req := &TradeRequest{
		WalletAddress: walletAddr.String(),
		Limit:         limit,
		Before:        before,
	}

	if err := ValidateTradeRequest(req, s.options); err != nil {
		s.logger.LogValidationError(ctx, "get_wallet_trades", "request", req, err)
		return nil, err
	}

	// Step 1: Derive xSOL Associated Token Account for this wallet
	xsolATA, err := tokens.DeriveAssociatedTokenAddress(walletAddr, tokens.XSOLMint)
	if err != nil {
		s.logger.LogHandlerError(ctx, "get_wallet_trades", err,
			slog.String("error_type", "ata_derivation"),
			slog.String("wallet", walletAddr.String()))
		return nil, fmt.Errorf("%w: %v", ErrXSOLATADerivation, err)
	}

	s.logger.DebugContext(ctx, "Derived xSOL ATA address",
		slog.String("ata_address", xsolATA.String()))

	// Step 2: Fetch transaction signatures for the xSOL ATA
	signatures, err := s.httpClient.GetSignaturesForAddress(ctx, xsolATA, before, req.Limit*2) // Fetch extra to account for filtering
	if err != nil {
		s.logger.LogExternalAPIError(ctx, "solana-rpc", "GetSignaturesForAddress", err, 0,
			slog.String("ata_address", xsolATA.String()))
		return nil, fmt.Errorf("%w: %v", ErrSignatureFetch, err)
	}

	s.logger.InfoContext(ctx, "Fetched signatures for xSOL ATA",
		slog.Int("signature_count", len(signatures)),
		slog.String("ata_address", xsolATA.String()))

	// Step 3: Process signatures to extract xSOL trades
	trades, err := s.processSignatures(ctx, signatures, xsolATA, req.Limit)
	if err != nil {
		s.logger.LogHandlerError(ctx, "get_wallet_trades", err,
			slog.String("error_type", "signature_processing"))
		return nil, err
	}

	// Step 4: Determine pagination information
	hasMore := len(trades) == req.Limit && len(signatures) > 0
	var nextCursor string
	if hasMore && len(trades) > 0 {
		// Use the last trade's signature as the next cursor
		nextCursor = trades[len(trades)-1].Signature
	}

	// Log operation completion
	s.logger.InfoContext(ctx, "Wallet trades retrieval completed",
		slog.String("wallet", walletAddr.String()),
		slog.Int("trades_found", len(trades)),
		slog.Bool("has_more", hasMore),
		slog.Duration("elapsed", time.Since(startTime)))

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
			s.logger.WarnContext(ctx, "Failed to fetch transaction details, continuing with others",
				slog.String("signature", sigInfo.Signature),
				slog.String("error", err.Error()))
			continue
		}

		// Parse the transaction for xSOL trades with logging context
		parseResult, err := hylo.ParseTransactionWithContext(ctx, tx, xsolATA, s.logger)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to parse transaction, continuing with others",
				slog.String("signature", sigInfo.Signature),
				slog.String("error", err.Error()))
			continue
		}

		// If we found a valid trade, add it to our results
		if parseResult != nil && parseResult.Trade != nil {
			trades = append(trades, parseResult.Trade)

			s.logger.DebugContext(ctx, "Successfully parsed and added trade",
				slog.String("signature", sigInfo.Signature),
				slog.String("side", parseResult.Trade.Side),
				slog.String("xsol_amount", parseResult.Trade.XSOLAmount))

			// Stop if we've reached the requested limit
			if len(trades) >= maxTrades {
				s.logger.DebugContext(ctx, "Reached trade limit, stopping processing",
					slog.Int("trades_found", len(trades)),
					slog.Int("limit", maxTrades))
				break
			}
		} else if parseResult != nil && parseResult.Error != "" {
			s.logger.DebugContext(ctx, "Transaction parsing returned error",
				slog.String("signature", sigInfo.Signature),
				slog.String("parse_error", parseResult.Error))
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
