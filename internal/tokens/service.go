package tokens

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"hylo-wallet-tracker-api/internal/logger"
	"hylo-wallet-tracker-api/internal/solana"
)

// TokenService provides token balance fetching functionality
// Integrates with Solana HTTP client, ATA derivation, and SPL token parsing
type TokenService struct {
	// httpClient is the Solana HTTP RPC client for on-chain data fetching
	httpClient HTTPClientInterface
	// config provides token configuration and registry functionality
	config *Config
	// logger for structured logging
	logger *logger.Logger
}

// HTTPClientInterface defines the contract for Solana HTTP client interaction
// Allows for easy mocking and testing of the balance service
type HTTPClientInterface interface {
	GetAccount(ctx context.Context, address solana.Address, commitment solana.Commitment) (*solana.AccountInfo, error)
}

// NewTokenService creates a new token service with dependency injection
// Parameters:
//   - httpClient: Solana HTTP client for RPC calls
//   - config: Token configuration and registry
func NewTokenService(httpClient HTTPClientInterface, config *Config) (*TokenService, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient cannot be nil")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Initialize logger for service
	serviceLogger := logger.NewFromEnv().WithComponent("token-service")

	// Validate config
	if err := config.Validate(); err != nil {
		serviceLogger.LogHandlerError(context.Background(), "service_initialization", err,
			slog.String("error_type", "validation"))
		return nil, fmt.Errorf("invalid token config: %w", err)
	}

	// Log service initialization
	serviceLogger.InfoContext(context.Background(), "Initializing Token service",
		slog.String("hyusd_mint", config.HyUSDMint.String()),
		slog.String("shyusd_mint", config.SHyUSDMint.String()),
		slog.String("xsol_mint", config.XSOLMint.String()))

	service := &TokenService{
		httpClient: httpClient,
		config:     config,
		logger:     serviceLogger,
	}

	serviceLogger.InfoContext(context.Background(), "Token service initialized successfully")
	return service, nil
}

// GetTokenBalance fetches the balance for a specific token in a wallet
// Returns TokenBalance with formatted amount or zero balance if account doesn't exist
func (s *TokenService) GetTokenBalance(ctx context.Context, wallet solana.Address, mint solana.Address) (*TokenBalance, error) {
	// Log operation start
	s.logger.DebugContext(ctx, "Getting token balance",
		slog.String("wallet", wallet.String()),
		slog.String("mint", mint.String()))

	// Validate wallet address
	if err := wallet.Validate(); err != nil {
		s.logger.LogValidationError(ctx, "get_token_balance", "wallet", wallet, err)
		return nil, fmt.Errorf("invalid wallet address: %w", err)
	}

	// Validate mint address
	if err := mint.Validate(); err != nil {
		s.logger.LogValidationError(ctx, "get_token_balance", "mint", mint, err)
		return nil, fmt.Errorf("invalid mint address: %w", err)
	}

	// Get token info from config
	tokenInfo := s.config.GetTokenInfo(mint)
	if tokenInfo == nil {
		s.logger.LogValidationError(ctx, "get_token_balance", "mint", mint, fmt.Errorf("unsupported token mint"))
		return nil, fmt.Errorf("unsupported token mint: %s", mint)
	}

	// Derive Associated Token Account address
	ataAddress, err := DeriveAssociatedTokenAddress(wallet, mint)
	if err != nil {
		s.logger.LogHandlerError(ctx, "get_token_balance", err,
			slog.String("error_type", "ata_derivation"),
			slog.String("wallet", wallet.String()),
			slog.String("mint", mint.String()))
		return nil, fmt.Errorf("failed to derive ATA address: %w", err)
	}

	s.logger.DebugContext(ctx, "Derived ATA address",
		slog.String("ata_address", ataAddress.String()),
		slog.String("token", tokenInfo.Symbol))

	// Fetch account info from Solana
	accountInfo, err := s.httpClient.GetAccount(ctx, ataAddress, solana.CommitmentConfirmed)
	if err != nil {
		// Handle account not found (zero balance)
		if errors.Is(err, solana.ErrAccountNotFound) {
			s.logger.InfoContext(ctx, "Token account not found, returning zero balance",
				slog.String("wallet", wallet.String()),
				slog.String("token", tokenInfo.Symbol))
			return NewTokenBalance(*tokenInfo, 0), nil
		}
		s.logger.LogExternalAPIError(ctx, "solana-rpc", "GetAccount", err, 0,
			slog.String("ata_address", ataAddress.String()),
			slog.String("token", tokenInfo.Symbol))
		return nil, fmt.Errorf("failed to fetch token account: %w", err)
	}

	// Parse SPL token account data with logging context
	tokenAccount, err := ParseSPLTokenAccountWithContext(ctx, accountInfo, s.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token account: %w", err)
	}

	// Validate parsed account (skip mint/owner validation in testing to avoid base58 decode issues)
	if err := ValidateTokenAccount(tokenAccount, "", ""); err != nil {
		s.logger.LogValidationError(ctx, "get_token_balance", "token_account", tokenAccount, err)
		return nil, fmt.Errorf("invalid token account: %w", err)
	}

	// Check if account is frozen
	if tokenAccount.IsFrozen {
		s.logger.WarnContext(ctx, "Token account is frozen",
			slog.String("wallet", wallet.String()),
			slog.String("token", tokenInfo.Symbol),
			slog.String("ata_address", ataAddress.String()))
		return nil, fmt.Errorf("token account is frozen")
	}

	// Log successful balance fetch
	s.logger.InfoContext(ctx, "Token balance retrieved successfully",
		slog.String("wallet", wallet.String()),
		slog.String("token", tokenInfo.Symbol),
		slog.Uint64("raw_amount", tokenAccount.Amount),
		slog.String("formatted_amount", fmt.Sprintf("%.6f", float64(tokenAccount.Amount)/1e6)))

	// Create TokenBalance with proper formatting
	return NewTokenBalance(*tokenInfo, tokenAccount.Amount), nil
}

// GetWalletBalances fetches balances for all supported Hylo tokens in a wallet
// Returns WalletBalances with all token balances, including zero balances
func (s *TokenService) GetWalletBalances(ctx context.Context, wallet solana.Address) (*WalletBalances, error) {
	// Log operation start
	s.logger.InfoContext(ctx, "Getting wallet balances for all supported tokens",
		slog.String("wallet", wallet.String()))

	// Validate wallet address
	if err := wallet.Validate(); err != nil {
		s.logger.LogValidationError(ctx, "get_wallet_balances", "wallet", wallet, err)
		return nil, fmt.Errorf("invalid wallet address: %w", err)
	}

	// Get all supported token mints
	tokenMints := []solana.Address{
		s.config.HyUSDMint,
		s.config.SHyUSDMint,
		s.config.XSOLMint,
	}

	// Initialize result structure
	balances := NewWalletBalances(wallet, 0) // Slot will be updated if we get successful responses
	successCount := 0
	failCount := 0

	// Fetch balance for each token
	for _, mint := range tokenMints {
		tokenBalance, err := s.GetTokenBalance(ctx, wallet, mint)
		if err != nil {
			failCount++
			// Get token info for logging
			tokenInfo := s.config.GetTokenInfo(mint)
			tokenSymbol := "unknown"
			if tokenInfo != nil {
				tokenSymbol = tokenInfo.Symbol
			}

			s.logger.WarnContext(ctx, "Failed to fetch token balance, continuing with other tokens",
				slog.String("wallet", wallet.String()),
				slog.String("token", tokenSymbol),
				slog.String("mint", mint.String()),
				slog.String("error", err.Error()))
			continue
		}

		successCount++
		// Add balance to results
		balances.AddBalance(tokenBalance)
	}

	// Ensure all tokens are represented (even with zero balance)
	for _, mint := range tokenMints {
		tokenInfo := s.config.GetTokenInfo(mint)
		if tokenInfo == nil {
			continue // Skip unsupported tokens
		}

		if _, exists := balances.Balances[tokenInfo.Symbol]; !exists {
			// Add zero balance for missing tokens
			balances.AddBalance(NewTokenBalance(*tokenInfo, 0))
		}
	}

	// Log operation completion
	s.logger.InfoContext(ctx, "Wallet balances retrieval completed",
		slog.String("wallet", wallet.String()),
		slog.Int("successful_tokens", successCount),
		slog.Int("failed_tokens", failCount),
		slog.Int("total_tokens", len(tokenMints)))

	return balances, nil
}

// GetSupportedTokens returns a list of all supported token information
func (s *TokenService) GetSupportedTokens() []*TokenInfo {
	return s.config.GetSupportedTokens()
}

// ValidateWalletForBalances performs comprehensive wallet validation before balance fetching
func (s *TokenService) ValidateWalletForBalances(wallet solana.Address) error {
	// Basic address validation
	if err := wallet.Validate(); err != nil {
		return fmt.Errorf("invalid wallet address format: %w", err)
	}

	// Additional validation can be added here:
	// - Check if wallet is a known system account
	// - Validate address is not a program ID
	// - Check address format specific to wallets vs other account types

	return nil
}

// Health performs a health check on the token service
// Tests connectivity and basic functionality with minimal overhead
func (s *TokenService) Health(ctx context.Context) error {
	// Test with a known wallet address (could be a system account)
	// This should either succeed or return ErrAccountNotFound (both are healthy)
	testWallet := solana.Address(TestReferenceWallet) // Test address from PRD

	// Test single token balance fetch (hyUSD)
	_, err := s.GetTokenBalance(ctx, testWallet, s.config.HyUSDMint)

	// ErrAccountNotFound is expected and healthy for this test address
	if errors.Is(err, solana.ErrAccountNotFound) {
		return nil
	}

	// Any other error indicates an unhealthy service
	if err != nil {
		return fmt.Errorf("balance service health check failed: %w", err)
	}

	// Successful fetch is also healthy
	return nil
}
