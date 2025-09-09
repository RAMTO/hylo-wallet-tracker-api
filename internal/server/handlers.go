package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"hylo-wallet-tracker-api/internal/solana"
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
		s.writeValidationError(w, "Wallet address is required", "Address parameter missing from URL path")
		return
	}

	// Parse and validate wallet address
	wallet := solana.Address(addressStr)
	if err := wallet.Validate(); err != nil {
		s.writeValidationError(w, "Invalid wallet address format", err.Error())
		return
	}

	// Fetch wallet balances using token service
	// This implements strict error handling - all tokens must succeed
	balances, err := s.tokenService.GetWalletBalances(r.Context(), wallet)
	if err != nil {
		// Categorize the error appropriately for better error handling
		if isNetworkError(err) {
			s.writeNetworkError(w, err.Error())
		} else if isValidationError(err) {
			s.writeValidationError(w, "Failed to fetch wallet balances", err.Error())
		} else {
			s.writeInternalError(w, err.Error())
		}
		return
	}

	// Return direct WalletBalances JSON response (maintains backward compatibility)
	s.writeJSONSuccess(w, balances)
}
