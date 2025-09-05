package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

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
