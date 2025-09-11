package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "hylo-wallet-tracker-api/docs/api" // This line is important for swagger to work
	"hylo-wallet-tracker-api/internal/logger"
	_ "hylo-wallet-tracker-api/internal/tokens" // Required for swagger type generation
	_ "hylo-wallet-tracker-api/internal/trades" // Required for swagger type generation
)

// RegisterRoutes configures all HTTP routes and middleware for the server
func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	// Middleware configuration
	r.Use(logger.RequestIDMiddleware) // Add request ID to all requests
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API Routes
	r.Get("/health", s.handleHealth)

	// Price endpoint
	r.Get("/price", s.handlePrice)
	r.Get("/price/debug", s.handlePriceDebug)

	// Wallet endpoints
	r.Route("/wallet", func(r chi.Router) {
		r.Get("/{address}/balances", s.handleWalletBalances)
		r.Get("/{address}/trades", s.handleWalletTrades)
	})

	// Documentation endpoint
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	return r
}
