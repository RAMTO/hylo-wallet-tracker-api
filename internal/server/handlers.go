package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"hylo-wallet-tracker-api/internal/solana"
)

// handleHealth returns basic liveness status
// @Summary Health check endpoint
// @Description Check the health and connectivity of the service and Solana RPC
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Service is healthy"
// @Success 503 {object} map[string]interface{} "Service is unhealthy - Solana RPC issues"
// @Router /health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := s.solanaService.Health(r.Context())

	response := map[string]interface{}{
		"status": "ok",
		"solana": status,
	}

	// Return 503 if Solana connection is unhealthy
	if !status.IsHealthy() {
		w.WriteHeader(http.StatusServiceUnavailable)
		response["status"] = "unhealthy"
	}

	json.NewEncoder(w).Encode(response)
}

// handleWalletBalances returns token balances for a specific wallet
// @Summary Get wallet token balances
// @Description Fetch balances for hyUSD, sHYUSD, and xSOL tokens for a specific wallet address
// @Tags wallet
// @Param address path string true "Wallet address (base58 encoded)"
// @Produce json
// @Success 200 {object} tokens.WalletBalances "Wallet token balances"
// @Failure 400 {object} map[string]interface{} "Invalid wallet address"
// @Failure 500 {object} map[string]interface{} "Failed to fetch balances"
// @Router /wallet/{address}/balances [get]
func (s *Server) handleWalletBalances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract wallet address from URL path
	addressStr := chi.URLParam(r, "address")
	if addressStr == "" {
		http.Error(w, `{"error": "wallet address is required"}`, http.StatusBadRequest)
		return
	}

	// Parse and validate wallet address
	wallet := solana.Address(addressStr)
	if err := wallet.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Invalid wallet address format",
			"details": err.Error(),
		})
		return
	}

	// Fetch wallet balances using token service
	// This implements strict error handling - all tokens must succeed
	balances, err := s.tokenService.GetWalletBalances(r.Context(), wallet)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Failed to fetch wallet balances",
			"details": err.Error(),
		})
		return
	}

	// Return direct WalletBalances JSON response (Option A format)
	json.NewEncoder(w).Encode(balances)
}

// Helper functions for handlers

// writeJSONError writes a JSON error response with the specified status code
func writeJSONError(w http.ResponseWriter, statusCode int, message string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error": message,
	}

	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}

	json.NewEncoder(w).Encode(response)
}

// writeJSONResponse writes a JSON response with 200 status code
func writeJSONResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}
