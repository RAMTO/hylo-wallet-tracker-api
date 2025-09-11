package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"hylo-wallet-tracker-api/internal/hylo"
	"hylo-wallet-tracker-api/internal/logger"
	"hylo-wallet-tracker-api/internal/price"
	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/tokens"
	"hylo-wallet-tracker-api/internal/trades"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port          int
	logger        *logger.Logger
	solanaService *solana.Service
	tokenService  *tokens.TokenService
	tradeService  *trades.TradeService
	priceService  *hylo.PriceService
	// Note: Price caching removed for fresh prices - all requests fetch live data
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Bootstrap Solana service with environment configuration
	solanaConfig := &solana.Config{
		HttpURL:           os.Getenv("RPC_HTTP_URL"),
		WebSocketURL:      os.Getenv("RPC_WS_URL"),
		RequestTimeout:    30 * time.Second,
		MaxRetries:        3,
		BaseBackoff:       1 * time.Second,
		MaxBackoff:        10 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		ReconnectTimeout:  30 * time.Second,
	}

	solanaService, err := solana.NewService(solanaConfig)
	if err != nil {
		log.Fatalf("Failed to create Solana service: %v", err)
	}

	fmt.Println("✅ Solana service created successfully")

	// Bootstrap Token service with Solana HTTP client and environment configuration
	tokenConfig := tokens.NewConfig()
	tokenService, err := tokens.NewTokenService(solanaService.GetHTTPClient(), tokenConfig)
	if err != nil {
		log.Fatalf("Failed to create Token service: %v", err)
	}

	fmt.Println("✅ Token service created successfully")

	// Bootstrap Trade service with Solana HTTP client, token config, and hylo config
	hyloConfig := hylo.NewConfig()
	tradeService, err := trades.NewTradeService(solanaService.GetHTTPClient(), tokenConfig, hyloConfig)
	if err != nil {
		log.Fatalf("Failed to create Trade service: %v", err)
	}

	fmt.Println("✅ Trade service created successfully")

	// Bootstrap Price service with all required dependencies
	priceConfig := price.NewConfig()
	priceService := hylo.NewPriceService(solanaService.GetHTTPClient(), hyloConfig, priceConfig)

	fmt.Println("✅ Price service created successfully")

	// Bootstrap Logger from environment
	appLogger := logger.NewFromEnv()
	fmt.Println("✅ Logger service created successfully")

	newServer := &Server{
		port:          port,
		logger:        appLogger,
		solanaService: solanaService,
		tokenService:  tokenService,
		tradeService:  tradeService,
		priceService:  priceService,
		// Cache TTL removed - fresh prices always fetched
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
