package solana

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"hylo-wallet-tracker-api/internal/logger"
)

// Service manages Solana connectivity and provides a clean interface
// for HTTP client access with health monitoring and lifecycle management
type Service struct {
	config        *Config
	logger        *logger.Logger
	httpClient    *HTTPClient
	healthTracker *HealthTracker
	mu            sync.RWMutex
	closed        bool
}

// NewService creates and initializes a new Solana service
func NewService(config *Config) (*Service, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Initialize logger for service
	serviceLogger := logger.NewFromEnv().WithComponent("solana-service")

	// Validate configuration
	if err := config.Validate(); err != nil {
		serviceLogger.LogHandlerError(context.Background(), "service_initialization", err,
			slog.String("error_type", "validation"),
			slog.String("config_url", config.HttpURL))
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Log service initialization
	serviceLogger.InfoContext(context.Background(), "Initializing Solana service",
		slog.String("rpc_url", config.HttpURL),
		slog.Duration("timeout", config.RequestTimeout),
		slog.Int("max_retries", config.MaxRetries))

	// Create HTTP client
	httpClient, err := NewHTTPClient(config, serviceLogger)
	if err != nil {
		serviceLogger.LogHandlerError(context.Background(), "service_initialization", err,
			slog.String("error_type", "http_client_creation"))
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create health tracker
	healthTracker := NewHealthTracker()

	service := &Service{
		config:        config,
		logger:        serviceLogger,
		httpClient:    httpClient,
		healthTracker: healthTracker,
	}

	// Perform initial health check to populate baseline
	go service.performInitialHealthCheck()

	serviceLogger.InfoContext(context.Background(), "Solana service initialized successfully")
	return service, nil
}

// GetHTTPClient returns the managed HTTP client for making RPC calls
// This is the primary interface for accessing Solana functionality
func (s *Service) GetHTTPClient() *HTTPClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return nil
	}

	return s.httpClient
}

// Health returns the current health status of the Solana connection
func (s *Service) Health(ctx context.Context) *HealthStatus {
	// Get current status from health tracker
	status := s.healthTracker.GetStatus()

	// If we haven't had a recent success, try a quick health check
	if status.Age() > 30*time.Second {
		s.performHealthCheck(ctx)
		status = s.healthTracker.GetStatus()
	}

	return status
}

// PerformHealthCheck executes a health check and records the result
// This method is exposed for external health monitoring
func (s *Service) PerformHealthCheck(ctx context.Context) error {
	return s.performHealthCheck(ctx)
}

// performHealthCheck executes a health check against Helios
func (s *Service) performHealthCheck(ctx context.Context) error {
	s.mu.RLock()
	client := s.httpClient
	s.mu.RUnlock()

	if client == nil {
		err := fmt.Errorf("HTTP client not available")
		s.logger.LogHandlerError(ctx, "health_check", err,
			slog.String("error_type", "client_unavailable"))
		s.healthTracker.RecordError(err)
		return err
	}

	// Use the existing Health method from HTTP client
	startTime := time.Now()
	err := client.Health(ctx)
	responseTime := time.Since(startTime)

	if err != nil {
		s.logger.LogExternalAPIError(ctx, "solana-rpc", "health", err, 0,
			slog.Duration("response_time", responseTime))
		s.healthTracker.RecordError(err)
		return err
	}

	// Log successful health check
	if responseTime > 5*time.Second {
		// Warn on slow health checks
		s.logger.WarnContext(ctx, "Slow Solana health check",
			slog.Duration("response_time", responseTime),
			slog.String("operation", "health_check"))
	} else {
		s.logger.DebugContext(ctx, "Solana health check successful",
			slog.Duration("response_time", responseTime))
	}

	s.healthTracker.RecordSuccess(responseTime)
	return nil
}

// performInitialHealthCheck runs an initial health check in background
func (s *Service) performInitialHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.logger.InfoContext(ctx, "Starting initial Solana health check")

	// Don't fail service creation if initial health check fails
	err := s.performHealthCheck(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "Initial Solana health check failed, service still available",
			slog.String("error", err.Error()))
	} else {
		s.logger.InfoContext(ctx, "Initial Solana health check completed successfully")
	}
}

// Close gracefully shuts down the service
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil // Already closed
	}

	s.logger.InfoContext(context.Background(), "Closing Solana service")
	s.closed = true

	// Close HTTP client
	if s.httpClient != nil {
		if err := s.httpClient.Close(); err != nil {
			s.logger.LogHandlerError(context.Background(), "service_close", err,
				slog.String("error_type", "http_client_close"))
			return fmt.Errorf("failed to close HTTP client: %w", err)
		}
	}

	s.logger.InfoContext(context.Background(), "Solana service closed successfully")
	return nil
}

// IsClosed returns whether the service has been closed
func (s *Service) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// Config returns a copy of the service configuration
func (s *Service) Config() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.config
}
