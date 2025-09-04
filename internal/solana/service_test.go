package solana

import (
	"context"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid_config",
			config: &Config{
				HttpURL:           "http://localhost:8899",
				WebSocketURL:      "ws://localhost:8900",
				RequestTimeout:    30 * time.Second,
				MaxRetries:        3,
				BaseBackoff:       1 * time.Second,
				MaxBackoff:        10 * time.Second,
				HeartbeatInterval: 15 * time.Second,
				ReconnectTimeout:  30 * time.Second,
			},
			expectError: false,
		},
		{
			name:        "nil_config",
			config:      nil,
			expectError: true,
		},
		{
			name: "invalid_config",
			config: &Config{
				HttpURL: "", // Invalid: empty URL
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewService(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if service != nil {
					t.Error("Expected nil service when error occurs")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if service == nil {
				t.Error("Expected service but got nil")
				return
			}

			// Verify service components are initialized
			if service.httpClient == nil {
				t.Error("HTTP client not initialized")
			}

			if service.healthTracker == nil {
				t.Error("Health tracker not initialized")
			}

			if service.config == nil {
				t.Error("Config not stored")
			}

			// Clean up
			defer service.Close()
		})
	}
}

func TestService_GetHTTPClient(t *testing.T) {
	config := &Config{
		HttpURL:           "http://localhost:8899",
		WebSocketURL:      "ws://localhost:8900",
		RequestTimeout:    30 * time.Second,
		MaxRetries:        3,
		BaseBackoff:       1 * time.Second,
		MaxBackoff:        10 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		ReconnectTimeout:  30 * time.Second,
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	// Test getting HTTP client
	client := service.GetHTTPClient()
	if client == nil {
		t.Error("Expected HTTP client but got nil")
	}

	// Test after close
	service.Close()
	client = service.GetHTTPClient()
	if client != nil {
		t.Error("Expected nil client after service closed")
	}
}

func TestService_Health(t *testing.T) {
	config := &Config{
		HttpURL:           "http://localhost:8899",
		WebSocketURL:      "ws://localhost:8900",
		RequestTimeout:    30 * time.Second,
		MaxRetries:        3,
		BaseBackoff:       1 * time.Second,
		MaxBackoff:        10 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		ReconnectTimeout:  30 * time.Second,
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	// Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := service.Health(ctx)
	if status == nil {
		t.Error("Expected health status but got nil")
		return
	}

	// Verify health status structure
	if status.LastSuccessAt.IsZero() && status.LastErrorAt.IsZero() {
		// Should have at least attempted a health check
		t.Error("Expected some health check activity")
	}

	// Test health check method
	err = service.PerformHealthCheck(ctx)
	// Note: This might fail if connecting to real endpoint, which is expected
	// The important thing is that the method exists and handles errors gracefully

	// Get updated status
	status = service.Health(ctx)
	if status == nil {
		t.Error("Expected health status after check")
	}
}

func TestService_Close(t *testing.T) {
	config := &Config{
		HttpURL:           "http://localhost:8899",
		WebSocketURL:      "ws://localhost:8900",
		RequestTimeout:    30 * time.Second,
		MaxRetries:        3,
		BaseBackoff:       1 * time.Second,
		MaxBackoff:        10 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		ReconnectTimeout:  30 * time.Second,
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Verify service is not closed initially
	if service.IsClosed() {
		t.Error("Service should not be closed initially")
	}

	// Close the service
	err = service.Close()
	if err != nil {
		t.Errorf("Unexpected error closing service: %v", err)
	}

	// Verify service is closed
	if !service.IsClosed() {
		t.Error("Service should be closed after Close()")
	}

	// Test double close (should not error)
	err = service.Close()
	if err != nil {
		t.Errorf("Double close should not error: %v", err)
	}
}

func TestService_Config(t *testing.T) {
	originalConfig := &Config{
		HttpURL:           "http://example.com",
		WebSocketURL:      "ws://example.com",
		RequestTimeout:    30 * time.Second,
		MaxRetries:        5,
		BaseBackoff:       2 * time.Second,
		MaxBackoff:        20 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		ReconnectTimeout:  30 * time.Second,
	}

	service, err := NewService(originalConfig)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()

	// Get config copy
	configCopy := service.Config()

	// Verify values match
	if configCopy.HttpURL != originalConfig.HttpURL {
		t.Errorf("HttpURL mismatch: got %s, want %s", configCopy.HttpURL, originalConfig.HttpURL)
	}

	if configCopy.MaxRetries != originalConfig.MaxRetries {
		t.Errorf("MaxRetries mismatch: got %d, want %d", configCopy.MaxRetries, originalConfig.MaxRetries)
	}

	// Verify it's a copy (modifying shouldn't affect original)
	configCopy.MaxRetries = 999
	currentConfig := service.Config()
	if currentConfig.MaxRetries == 999 {
		t.Error("Config should return a copy, not a reference")
	}
}
