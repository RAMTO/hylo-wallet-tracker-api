package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"hylo-wallet-tracker-api/internal/solana"
	_ "hylo-wallet-tracker-api/internal/tokens" // Required for swagger type generation

	_ "hylo-wallet-tracker-api/docs/api" // This line is important for swagger to work
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health endpoint
	r.Get("/health", s.handleHealth)

	// Wallet endpoints
	r.Route("/wallet", func(r chi.Router) {
		r.Get("/{address}/balances", s.handleWalletBalances)
	})

	// Swagger documentation endpoint
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // The URL pointing to API definition
	))

	return r
}

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
