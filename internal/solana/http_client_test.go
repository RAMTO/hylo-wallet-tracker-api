package solana

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hylo-wallet-tracker-api/internal/logger"
)

func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      NewConfig("http://localhost:8899", "ws://localhost:8900"),
			expectError: false,
		},
		{
			name:        "missing HTTP URL",
			config:      &Config{WebSocketURL: "ws://localhost:8900"},
			expectError: true,
		},
		{
			name:        "missing WebSocket URL",
			config:      &Config{HttpURL: "http://localhost:8899"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewHTTPClient(tt.config, logger.NewFromEnv())
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				if client != nil {
					t.Error("expected nil client on error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if client == nil {
					t.Error("expected non-nil client")
				}
			}
		})
	}
}

func TestHTTPClient_GetAccount(t *testing.T) {
	// Load test data
	successResp := loadTestData(t, "get_account_response.json")
	notFoundResp := loadTestData(t, "get_account_not_found.json")
	errorResp := loadTestData(t, "rpc_error_response.json")

	tests := []struct {
		name        string
		address     Address
		commitment  Commitment
		serverResp  string
		statusCode  int
		expectError bool
		expectNil   bool
	}{
		{
			name:        "successful account fetch",
			address:     "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // 44 chars
			commitment:  CommitmentConfirmed,
			serverResp:  successResp,
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "account not found",
			address:     "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // Reference wallet
			commitment:  CommitmentFinalized,
			serverResp:  notFoundResp,
			statusCode:  200,
			expectError: true, // Should return ErrAccountNotFound
		},
		{
			name:        "RPC error response",
			address:     "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // 44 chars
			commitment:  CommitmentConfirmed,
			serverResp:  errorResp,
			statusCode:  200,
			expectError: true,
		},
		{
			name:        "invalid address",
			address:     "invalid",
			commitment:  CommitmentConfirmed,
			serverResp:  successResp,
			statusCode:  200,
			expectError: true,
		},
		{
			name:        "invalid commitment",
			address:     "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
			commitment:  "invalid",
			serverResp:  successResp,
			statusCode:  200,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			// Create client
			config := NewConfig(server.URL, "ws://unused")
			client, err := NewHTTPClient(config, logger.NewFromEnv())
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			// Test GetAccount
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			account, err := client.GetAccount(ctx, tt.address, tt.commitment)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if account == nil {
					t.Error("expected non-nil account")
				} else {
					// Validate account data
					if account.Lamports != 1000000000 {
						t.Errorf("expected lamports 1000000000, got %d", account.Lamports)
					}
					if account.Owner != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
						t.Errorf("unexpected owner: %s", account.Owner)
					}
				}
			}
		})
	}
}

func TestHTTPClient_GetSignaturesForAddress(t *testing.T) {
	successResp := loadTestData(t, "get_signatures_response.json")

	tests := []struct {
		name        string
		address     Address
		before      string
		limit       int
		expectError bool
		expectCount int
	}{
		{
			name:        "successful signatures fetch",
			address:     "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // 44 chars
			before:      "",
			limit:       10,
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "with before parameter",
			address:     "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // 44 chars
			before:      "5VWF2BTGZGS9c8uJ8ZmKGZwxAAaG5Wnr4drcmA8zbHEKkArDhYBm2HjRN1QAK1EzQ8sKvQw9GsJJ6sJ4x7q3LQg9",
			limit:       5,
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "invalid address",
			address:     "invalid",
			limit:       10,
			expectError: true,
		},
		{
			name:        "invalid limit (zero)",
			address:     "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // 44 chars
			limit:       0,
			expectError: true,
		},
		{
			name:        "invalid limit (too high)",
			address:     "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // 44 chars
			limit:       1001,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(successResp))
			}))
			defer server.Close()

			// Create client
			config := NewConfig(server.URL, "ws://unused")
			client, err := NewHTTPClient(config, logger.NewFromEnv())
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			// Test GetSignaturesForAddress
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			signatures, err := client.GetSignaturesForAddress(ctx, tt.address, tt.before, tt.limit)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(signatures) != tt.expectCount {
					t.Errorf("expected %d signatures, got %d", tt.expectCount, len(signatures))
				}

				// Validate first signature
				if len(signatures) > 0 {
					sig := signatures[0]
					if sig.Signature != "5VWF2BTGZGS9c8uJ8ZmKGZwxAAaG5Wnr4drcmA8zbHEKkArDhYBm2HjRN1QAK1EzQ8sKvQw9GsJJ6sJ4x7q3LQg9" {
						t.Errorf("unexpected signature: %s", sig.Signature)
					}
					if sig.Slot != 294112233 {
						t.Errorf("expected slot 294112233, got %d", sig.Slot)
					}
				}
			}
		})
	}
}

func TestHTTPClient_RetryLogic(t *testing.T) {
	attempts := 0

	// Server that fails first 2 requests, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if attempts <= 2 {
			// Simulate server error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "server error"}`))
			return
		}

		// Success on 3rd attempt
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(loadTestData(t, "get_account_response.json")))
	}))
	defer server.Close()

	// Create client with fast retries for testing
	config := NewConfig(server.URL, "ws://unused")
	config.BaseBackoff = 10 * time.Millisecond
	config.MaxBackoff = 50 * time.Millisecond
	config.MaxRetries = 3

	client, err := NewHTTPClient(config, logger.NewFromEnv())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test that retry logic works
	ctx := context.Background()
	account, err := client.GetAccount(ctx, "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", CommitmentConfirmed)

	if err != nil {
		t.Errorf("expected success after retries, got error: %v", err)
	}

	if account == nil {
		t.Error("expected non-nil account after retries")
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestHTTPClient_GetTransaction(t *testing.T) {
	successResp := loadTestData(t, "get_transaction_response.json")

	tests := []struct {
		name        string
		signature   Signature
		serverResp  string
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful transaction fetch",
			signature:   "5VWF2BTGZGS9c8uJ8ZmKGZwxAAaG5Wnr4drcmA8zbHEKkArDhYBm2HjRN1QAK1EzQ8sKvQw9GsJJ6sJ4x7q3LQg9",
			serverResp:  successResp,
			statusCode:  200,
			expectError: false,
		},
		{
			name:        "invalid signature",
			signature:   "invalid",
			serverResp:  successResp,
			statusCode:  200,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			config := NewConfig(server.URL, "ws://unused")
			client, err := NewHTTPClient(config, logger.NewFromEnv())
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			tx, err := client.GetTransaction(ctx, tt.signature)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tx == nil {
					t.Error("expected non-nil transaction")
				} else {
					// Validate transaction data
					if tx.Slot != 294112233 {
						t.Errorf("expected slot 294112233, got %d", tx.Slot)
					}
					if tx.BlockTime == nil || *tx.BlockTime != 1694019123 {
						t.Errorf("unexpected block time: %v", tx.BlockTime)
					}
				}
			}
		})
	}
}

func TestHTTPClient_MaxRetriesExceeded(t *testing.T) {
	// Server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "persistent server error"}`))
	}))
	defer server.Close()

	// Create client with fast retries
	config := NewConfig(server.URL, "ws://unused")
	config.BaseBackoff = 1 * time.Millisecond
	config.MaxBackoff = 5 * time.Millisecond
	config.MaxRetries = 2

	client, err := NewHTTPClient(config, logger.NewFromEnv())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test that max retries is respected
	ctx := context.Background()
	_, err = client.GetAccount(ctx, "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", CommitmentConfirmed)

	if err == nil {
		t.Error("expected error after max retries exceeded")
	}

	var netErr *NetworkError
	if !errors.As(err, &netErr) {
		t.Errorf("expected NetworkError, got %T", err)
	}
}

func TestHTTPClient_ContextTimeout(t *testing.T) {
	// Server with artificial delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(loadTestData(t, "get_account_response.json")))
	}))
	defer server.Close()

	// Create client
	config := NewConfig(server.URL, "ws://unused")
	client, err := NewHTTPClient(config, logger.NewFromEnv())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.GetAccount(ctx, "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", CommitmentConfirmed)

	if err == nil {
		t.Error("expected timeout error")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") &&
		!strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	// Server with long delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(loadTestData(t, "get_account_response.json")))
	}))
	defer server.Close()

	// Create client
	config := NewConfig(server.URL, "ws://unused")
	client, err := NewHTTPClient(config, logger.NewFromEnv())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err = client.GetAccount(ctx, "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", CommitmentConfirmed)

	if err == nil {
		t.Error("expected cancellation error")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got: %v", err)
	}
}

func TestHTTPClient_Health(t *testing.T) {
	tests := []struct {
		name          string
		serverResp    string
		statusCode    int
		expectHealthy bool
	}{
		{
			name:          "healthy - account not found is OK",
			serverResp:    loadTestData(t, "get_account_not_found.json"),
			statusCode:    200,
			expectHealthy: true,
		},
		{
			name:          "unhealthy - RPC error",
			serverResp:    loadTestData(t, "rpc_error_response.json"),
			statusCode:    200,
			expectHealthy: false,
		},
		{
			name:          "unhealthy - HTTP error",
			serverResp:    `{"error": "server error"}`,
			statusCode:    500,
			expectHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			config := NewConfig(server.URL, "ws://unused")
			client, err := NewHTTPClient(config, logger.NewFromEnv())
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			ctx := context.Background()
			err = client.Health(ctx)

			if tt.expectHealthy && err != nil {
				t.Errorf("expected healthy client, got error: %v", err)
			}

			if !tt.expectHealthy && err == nil {
				t.Error("expected unhealthy client, got nil error")
			}
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      NewConfig("http://localhost:8899", "ws://localhost:8900"),
			expectError: false,
		},
		{
			name: "zero timeout",
			config: &Config{
				HttpURL:        "http://localhost:8899",
				WebSocketURL:   "ws://localhost:8900",
				RequestTimeout: 0,
			},
			expectError: true,
		},
		{
			name: "negative retries",
			config: &Config{
				HttpURL:        "http://localhost:8899",
				WebSocketURL:   "ws://localhost:8900",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     -1,
			},
			expectError: true,
		},
		{
			name: "max backoff less than base",
			config: &Config{
				HttpURL:        "http://localhost:8899",
				WebSocketURL:   "ws://localhost:8900",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				BaseBackoff:    10 * time.Second,
				MaxBackoff:     5 * time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Error("expected validation error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestTypes_Validation(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func() error
		expectError bool
	}{
		{
			name:        "valid address",
			testFunc:    func() error { return Address("A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g").Validate() },
			expectError: false,
		},
		{
			name:        "invalid address length",
			testFunc:    func() error { return Address("invalid").Validate() },
			expectError: true,
		},
		{
			name: "valid signature",
			testFunc: func() error {
				return Signature("5VWF2BTGZGS9c8uJ8ZmKGZwxAAaG5Wnr4drcmA8zbHEKkArDhYBm2HjRN1QAK1EzQ8sKvQw9GsJJ6sJ4x7q3LQg9").Validate()
			},
			expectError: false,
		},
		{
			name:        "invalid signature length",
			testFunc:    func() error { return Signature("short").Validate() },
			expectError: true,
		},
		{
			name:        "valid commitment",
			testFunc:    func() error { return CommitmentConfirmed.Validate() },
			expectError: false,
		},
		{
			name:        "invalid commitment",
			testFunc:    func() error { return Commitment("invalid").Validate() },
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()
			if tt.expectError && err == nil {
				t.Error("expected validation error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestBackoffCalculation(t *testing.T) {
	config := NewConfig("http://localhost", "ws://localhost")
	config.BaseBackoff = 1 * time.Second
	config.MaxBackoff = 10 * time.Second

	client, err := NewHTTPClient(config, logger.NewFromEnv())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test exponential backoff progression
	for attempt := 0; attempt < 5; attempt++ {
		backoff := client.calculateBackoff(attempt)

		// Should be at least 100ms (minimum)
		if backoff < 100*time.Millisecond {
			t.Errorf("backoff too short for attempt %d: %v", attempt, backoff)
		}

		// Should not exceed MaxBackoff
		if backoff > config.MaxBackoff {
			t.Errorf("backoff exceeds max for attempt %d: %v > %v", attempt, backoff, config.MaxBackoff)
		}

		t.Logf("Attempt %d: backoff = %v", attempt, backoff)
	}
}

func TestConfig_WithMethods(t *testing.T) {
	config := NewConfig("http://localhost:8899", "ws://localhost:8900")

	// Test WithTimeout
	newConfig := config.WithTimeout(60 * time.Second)
	if newConfig.RequestTimeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", newConfig.RequestTimeout)
	}

	// Test WithRetries
	newConfig2 := config.WithRetries(5)
	if newConfig2.MaxRetries != 5 {
		t.Errorf("expected max retries 5, got %d", newConfig2.MaxRetries)
	}

	// Original config should be unchanged
	if config.RequestTimeout != 30*time.Second {
		t.Error("original config was modified")
	}
}

func TestErrors_AdditionalCoverage(t *testing.T) {
	// Test ValidationError.Error method
	valErr := ValidationError{
		Field:  "test",
		Value:  "invalid",
		Reason: "too short",
	}
	errorMsg := valErr.Error()
	if !strings.Contains(errorMsg, "test") || !strings.Contains(errorMsg, "invalid") {
		t.Errorf("ValidationError message missing expected content: %s", errorMsg)
	}

	// Test IsTemporary function
	tempErr := &NetworkError{Err: errors.New("temp"), Attempt: 1, Final: false}
	if !IsTemporary(tempErr) {
		t.Error("expected NetworkError to be temporary")
	}

	regularErr := errors.New("regular error")
	if IsTemporary(regularErr) {
		t.Error("expected regular error to not be temporary")
	}
}

func TestHTTPClient_Close(t *testing.T) {
	config := NewConfig("http://localhost:8899", "ws://localhost:8900")
	client, err := NewHTTPClient(config, logger.NewFromEnv())
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test Close method
	err = client.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestSignatureInfo_GetTime(t *testing.T) {
	// Test with valid block time
	blockTime := int64(1694019123)
	sigInfo := &SignatureInfo{BlockTime: &blockTime}

	expectedTime := time.Unix(1694019123, 0)
	actualTime := sigInfo.GetTime()

	if !actualTime.Equal(expectedTime) {
		t.Errorf("expected time %v, got %v", expectedTime, actualTime)
	}

	// Test with nil block time
	sigInfoNil := &SignatureInfo{BlockTime: nil}
	zeroTime := sigInfoNil.GetTime()

	if !zeroTime.IsZero() {
		t.Errorf("expected zero time for nil block time, got %v", zeroTime)
	}
}

func TestAccountInfo_UnmarshalJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
	}{
		{
			name:        "data as string",
			jsonData:    `{"data": "dGVzdA==", "lamports": 1000}`,
			expectError: false,
		},
		{
			name:        "data as array",
			jsonData:    `{"data": ["dGVzdA==", "base64"], "lamports": 1000}`,
			expectError: false,
		},
		{
			name:        "invalid base64",
			jsonData:    `{"data": "invalid!@#", "lamports": 1000}`,
			expectError: true,
		},
		{
			name:        "empty data array",
			jsonData:    `{"data": [], "lamports": 1000}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var account AccountInfo
			err := json.Unmarshal([]byte(tt.jsonData), &account)

			if tt.expectError && err == nil {
				t.Error("expected unmarshal error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected unmarshal error: %v", err)
			}
		})
	}
}

// Helper function to load test data from files
func loadTestData(t *testing.T, filename string) string {
	t.Helper()

	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load test data %s: %v", filename, err)
	}

	return string(data)
}
