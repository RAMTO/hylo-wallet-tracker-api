package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	_ "hylo-wallet-tracker-api/internal/price" // Required for swagger type generation
	"hylo-wallet-tracker-api/internal/solana"
	_ "hylo-wallet-tracker-api/internal/tokens" // Required for swagger type generation
	_ "hylo-wallet-tracker-api/internal/trades" // Required for swagger type generation
)

// handleHealth returns basic liveness status
// @Summary Health check endpoint
// @Description Check the health and connectivity of the service and Solana RPC
// @Tags health
// @Produce json
// @Success 200 {object} server.HealthResponse "Service is healthy"
// @Success 503 {object} server.HealthResponse "Service is unhealthy - Solana RPC issues"
// @Router /health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := s.solanaService.Health(r.Context())

	response := HealthResponse{
		Status:    "ok",
		Solana:    status,
		Timestamp: getCurrentTimestamp(),
	}

	statusCode := http.StatusOK
	// Return 503 if Solana connection is unhealthy
	if !status.IsHealthy() {
		response.Status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	// Use helper function for consistency while maintaining current response format
	s.writeJSONSuccessWithCode(w, statusCode, response)
}

// handleWalletBalances returns token balances for a specific wallet
// @Summary Get wallet token balances
// @Description Fetch balances for hyUSD, sHYUSD, and xSOL tokens for a specific wallet address
// @Tags wallet
// @Param address path string true "Wallet address (base58 encoded)"
// @Produce json
// @Success 200 {object} tokens.WalletBalances "Wallet token balances"
// @Failure 400 {object} server.ErrorResponse "Validation error"
// @Failure 500 {object} server.ErrorResponse "Internal server error"
// @Failure 502 {object} server.ErrorResponse "Network connectivity error"
// @Router /wallet/{address}/balances [get]
func (s *Server) handleWalletBalances(w http.ResponseWriter, r *http.Request) {
	// Extract wallet address from URL path
	addressStr := chi.URLParam(r, "address")
	if addressStr == "" {
		s.logger.LogValidationError(r.Context(), "get_wallet_balances", "address", "", fmt.Errorf("address parameter missing from URL path"))
		s.writeValidationError(w, "Wallet address is required", "Address parameter missing from URL path")
		return
	}

	// Parse and validate wallet address
	wallet := solana.Address(addressStr)
	if err := wallet.Validate(); err != nil {
		s.logger.LogValidationError(r.Context(), "get_wallet_balances", "address", addressStr, err)
		s.writeValidationError(w, "Invalid wallet address format", err.Error())
		return
	}

	// Fetch wallet balances using token service
	// This implements strict error handling - all tokens must succeed
	balances, err := s.tokenService.GetWalletBalances(r.Context(), wallet)
	if err != nil {
		// Log error with wallet context
		logger := s.logger.WithWalletAddress(string(wallet))

		// Categorize the error appropriately for better error handling
		if isNetworkError(err) {
			logger.LogExternalAPIError(r.Context(), "token-service", "GetWalletBalances", err, 0)
			s.writeNetworkError(w, err.Error())
		} else if isValidationError(err) {
			logger.LogValidationError(r.Context(), "get_wallet_balances", "wallet_data", wallet, err)
			s.writeValidationError(w, "Failed to fetch wallet balances", err.Error())
		} else {
			logger.LogHandlerError(r.Context(), "get_wallet_balances", err)
			s.writeInternalError(w, err.Error())
		}
		return
	}

	// Return direct WalletBalances JSON response (maintains backward compatibility)
	s.writeJSONSuccess(w, balances)
}

// handleWalletTrades returns xSOL trade history for a specific wallet
// @Summary Get wallet xSOL trade history
// @Description Fetch paginated xSOL trade history for a specific wallet address with real-time RPC data
// @Tags wallet
// @Param address path string true "Wallet address (base58 encoded)"
// @Param limit query int false "Maximum number of trades to return (1-50, default 10)"
// @Param before query string false "Cursor for pagination - signature to fetch trades before"
// @Produce json
// @Success 200 {object} trades.TradeResponse "Wallet xSOL trade history"
// @Failure 400 {object} server.ErrorResponse "Validation error"
// @Failure 500 {object} server.ErrorResponse "Internal server error"
// @Failure 502 {object} server.ErrorResponse "Network connectivity error"
// @Router /wallet/{address}/trades [get]
func (s *Server) handleWalletTrades(w http.ResponseWriter, r *http.Request) {
	// Extract wallet address from URL path
	addressStr := chi.URLParam(r, "address")
	if addressStr == "" {
		s.logger.LogValidationError(r.Context(), "get_wallet_trades", "address", "", fmt.Errorf("address parameter missing from URL path"))
		s.writeValidationError(w, "Wallet address is required", "Address parameter missing from URL path")
		return
	}

	// Parse and validate wallet address
	wallet := solana.Address(addressStr)
	if err := wallet.Validate(); err != nil {
		s.logger.LogValidationError(r.Context(), "get_wallet_trades", "address", addressStr, err)
		s.writeValidationError(w, "Invalid wallet address format", err.Error())
		return
	}

	// Parse query parameters with defaults
	limit := 10 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		} else {
			s.logger.LogParsingError(r.Context(), "get_wallet_trades", "limit_parameter", err, slog.String("invalid_value", limitStr))
			s.writeValidationError(w, "Invalid limit parameter", "Limit must be a valid integer")
			return
		}
	}

	// Validate limit range
	if limit < 1 || limit > 50 {
		s.logger.LogValidationError(r.Context(), "get_wallet_trades", "limit", limit, fmt.Errorf("limit must be between 1 and 50"))
		s.writeValidationError(w, "Invalid limit parameter", "Limit must be between 1 and 50")
		return
	}

	// Extract before cursor for pagination (optional)
	before := r.URL.Query().Get("before")

	// Fetch wallet trades using trade service
	trades, err := s.tradeService.GetWalletTrades(r.Context(), wallet, limit, before)
	if err != nil {
		// Log error with wallet context
		logger := s.logger.WithWalletAddress(string(wallet))

		// Categorize the error appropriately for better error handling
		if isNetworkError(err) {
			logger.LogExternalAPIError(r.Context(), "trade-service", "GetWalletTrades", err, 0)
			s.writeNetworkError(w, err.Error())
		} else if isValidationError(err) {
			logger.LogValidationError(r.Context(), "get_wallet_trades", "wallet_data", wallet, err)
			s.writeValidationError(w, "Failed to fetch wallet trades", err.Error())
		} else {
			logger.LogHandlerError(r.Context(), "get_wallet_trades", err)
			s.writeInternalError(w, err.Error())
		}
		return
	}

	// Return TradeResponse JSON response (follows existing patterns)
	s.writeJSONSuccess(w, trades)
}

// handlePrice returns current price data for all supported assets
// @Summary Get current asset prices
// @Description Fetch current prices for SOL/USD, xSOL/SOL, and xSOL/USD with caching
// @Tags price
// @Produce json
// @Success 200 {object} price.CombinedPriceResponse "Current asset prices"
// @Failure 500 {object} server.ErrorResponse "Internal server error"
// @Failure 502 {object} server.ErrorResponse "Network connectivity error"
// @Router /price [get]
func (s *Server) handlePrice(w http.ResponseWriter, r *http.Request) {
	// Always fetch fresh prices - no caching for maximum freshness
	prices, err := s.priceService.GetCombinedPriceResponse(r.Context())
	if err != nil {
		// Log error
		logger := s.logger.WithOperation("get_price")

		// Return error - no fallback cache to rely on
		if isNetworkError(err) {
			logger.LogExternalAPIError(r.Context(), "price-service", "GetCombinedPriceResponse", err, 0)
			s.writeNetworkError(w, err.Error())
		} else {
			logger.LogHandlerError(r.Context(), "get_price", err)
			s.writeInternalError(w, err.Error())
		}
		return
	}

	// Return fresh CombinedPriceResponse JSON (matches PRD specification)
	s.writeJSONSuccess(w, prices)
}

// handlePriceDebug returns detailed price calculation information for debugging
// This is a temporary endpoint to help debug the xSOL price calculation
func (s *Server) handlePriceDebug(w http.ResponseWriter, r *http.Request) {
	// Get calculation details
	details, err := s.priceService.GetPriceCalculationDetails(r.Context())
	if err != nil {
		s.logger.LogHandlerError(r.Context(), "get_price_debug", err)
		s.writeInternalError(w, fmt.Sprintf("Failed to get calculation details: %v", err))
		return
	}

	s.writeJSONSuccess(w, details)
}
