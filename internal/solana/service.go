package solana

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Service manages Solana connectivity and provides a clean interface
// for HTTP client access with health monitoring and lifecycle management
type Service struct {
	config        *Config
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

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create HTTP client
	httpClient, err := NewHTTPClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create health tracker
	healthTracker := NewHealthTracker()

	service := &Service{
		config:        config,
		httpClient:    httpClient,
		healthTracker: healthTracker,
	}

	// Perform initial health check to populate baseline
	go service.performInitialHealthCheck()

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
		s.healthTracker.RecordError(err)
		return err
	}

	// Use the existing Health method from HTTP client
	startTime := time.Now()
	err := client.Health(ctx)
	responseTime := time.Since(startTime)

	if err != nil {
		s.healthTracker.RecordError(err)
		return err
	}

	s.healthTracker.RecordSuccess(responseTime)
	return nil
}

// performInitialHealthCheck runs an initial health check in background
func (s *Service) performInitialHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Don't fail service creation if initial health check fails
	_ = s.performHealthCheck(ctx)
}

// Close gracefully shuts down the service
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil // Already closed
	}

	s.closed = true

	// Close HTTP client
	if s.httpClient != nil {
		if err := s.httpClient.Close(); err != nil {
			return fmt.Errorf("failed to close HTTP client: %w", err)
		}
	}

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
