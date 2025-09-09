package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewDexScreenerClient(t *testing.T) {
	config := DefaultConfig()
	client := NewDexScreenerClient(config)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.config != config {
		t.Error("Expected client config to match provided config")
	}

	if client.httpClient == nil {
		t.Error("Expected non-nil HTTP client")
	}

	if client.rateLimiter == nil {
		t.Error("Expected non-nil rate limiter")
	}

	expectedBaseURL := strings.TrimSuffix(config.DexScreenerURL, "/")
	if client.baseURL != expectedBaseURL {
		t.Errorf("Expected baseURL %s, got %s", expectedBaseURL, client.baseURL)
	}
}

func TestNewDexScreenerClientWithNilConfig(t *testing.T) {
	client := NewDexScreenerClient(nil)

	if client == nil {
		t.Fatal("Expected non-nil client even with nil config")
	}

	if client.config == nil {
		t.Error("Expected client to use default config when nil provided")
	}
}

func TestDexScreenerClient_FetchSOLPrice_Success(t *testing.T) {
	// Create mock response
	mockResponse := DexScreenerResponse{
		SchemaVersion: "1.0.0",
		Pairs: []DexScreenerPair{
			{
				ChainID:  "solana",
				DexID:    "raydium",
				PriceUSD: "150.50",
				BaseToken: Token{
					Symbol: "SOL",
					Name:   "Solana",
				},
				QuoteToken: Token{
					Symbol: "USDC",
					Name:   "USD Coin",
				},
				Liquidity: Liquidity{
					USD: 1000000,
				},
				Volume: Volume{
					H24: 500000,
				},
			},
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "So11111111111111111111111111111111111111112") {
			t.Errorf("Expected SOL token address in URL path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	config := DefaultConfig()
	config.DexScreenerURL = server.URL
	config.SOLUSDMinPrice = 100.0
	config.SOLUSDMaxPrice = 200.0
	client := NewDexScreenerClient(config)

	ctx := context.Background()
	price, err := client.FetchSOLPrice(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if price == nil {
		t.Fatal("Expected non-nil price")
	}

	expectedPrice := 150.50
	if price.Price != expectedPrice {
		t.Errorf("Expected price %f, got %f", expectedPrice, price.Price)
	}

	if price.Source != "dexscreener" {
		t.Errorf("Expected source 'dexscreener', got %s", price.Source)
	}

	if price.Pair != "SOL/USDC" {
		t.Errorf("Expected pair 'SOL/USDC', got %s", price.Pair)
	}

	if price.Liquidity != 1000000 {
		t.Errorf("Expected liquidity 1000000, got %f", price.Liquidity)
	}

	if price.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestDexScreenerClient_FetchSOLPrice_PriceValidation(t *testing.T) {
	testCases := []struct {
		name        string
		priceUSD    string
		minPrice    float64
		maxPrice    float64
		expectError bool
	}{
		{
			name:        "Valid price",
			priceUSD:    "150.0",
			minPrice:    100.0,
			maxPrice:    200.0,
			expectError: false,
		},
		{
			name:        "Price too low",
			priceUSD:    "50.0",
			minPrice:    100.0,
			maxPrice:    200.0,
			expectError: true,
		},
		{
			name:        "Price too high",
			priceUSD:    "300.0",
			minPrice:    100.0,
			maxPrice:    200.0,
			expectError: true,
		},
		{
			name:        "Zero price",
			priceUSD:    "0",
			minPrice:    100.0,
			maxPrice:    200.0,
			expectError: true,
		},
		{
			name:        "Invalid price format",
			priceUSD:    "invalid",
			minPrice:    100.0,
			maxPrice:    200.0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockResponse := DexScreenerResponse{
				Pairs: []DexScreenerPair{
					{
						PriceUSD:   tc.priceUSD,
						BaseToken:  Token{Symbol: "SOL"},
						QuoteToken: Token{Symbol: "USDC"},
						Liquidity:  Liquidity{USD: 1000000},
						Volume:     Volume{H24: 500000},
					},
				},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(mockResponse)
			}))
			defer server.Close()

			config := DefaultConfig()
			config.DexScreenerURL = server.URL
			config.SOLUSDMinPrice = tc.minPrice
			config.SOLUSDMaxPrice = tc.maxPrice
			client := NewDexScreenerClient(config)

			_, err := client.FetchSOLPrice(context.Background())

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDexScreenerClient_FetchSOLPrice_BestPairSelection(t *testing.T) {
	mockResponse := DexScreenerResponse{
		Pairs: []DexScreenerPair{
			{
				PriceUSD:   "150.0",
				BaseToken:  Token{Symbol: "SOL"},
				QuoteToken: Token{Symbol: "USDT"},
				Liquidity:  Liquidity{USD: 500000}, // Lower liquidity
				Volume:     Volume{H24: 300000},
			},
			{
				PriceUSD:   "150.5",
				BaseToken:  Token{Symbol: "SOL"},
				QuoteToken: Token{Symbol: "USDC"},
				Liquidity:  Liquidity{USD: 2000000}, // Higher liquidity - should be selected
				Volume:     Volume{H24: 800000},
			},
			{
				PriceUSD:   "149.8",
				BaseToken:  Token{Symbol: "SOL"},
				QuoteToken: Token{Symbol: "USDT"},
				Liquidity:  Liquidity{USD: 1000000},
				Volume:     Volume{H24: 400000},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.DexScreenerURL = server.URL
	client := NewDexScreenerClient(config)

	price, err := client.FetchSOLPrice(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should select the pair with highest liquidity (second pair)
	expectedPrice := 150.5
	if price.Price != expectedPrice {
		t.Errorf("Expected price %f (highest liquidity pair), got %f", expectedPrice, price.Price)
	}

	if price.Pair != "SOL/USDC" {
		t.Errorf("Expected pair 'SOL/USDC', got %s", price.Pair)
	}
}

func TestDexScreenerClient_FetchSOLPrice_HTTPErrors(t *testing.T) {
	testCases := []struct {
		name            string
		statusCode      int
		expectRetryable bool
	}{
		{"Server Error", http.StatusInternalServerError, true},
		{"Bad Gateway", http.StatusBadGateway, true},
		{"Service Unavailable", http.StatusServiceUnavailable, true},
		{"Too Many Requests", http.StatusTooManyRequests, true},
		{"Not Found", http.StatusNotFound, false},
		{"Bad Request", http.StatusBadRequest, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(fmt.Sprintf("HTTP %d Error", tc.statusCode)))
			}))
			defer server.Close()

			config := DefaultConfig()
			config.DexScreenerURL = server.URL
			config.MaxRetries = 0 // Disable retries for testing
			client := NewDexScreenerClient(config)

			_, err := client.FetchSOLPrice(context.Background())

			if err == nil {
				t.Error("Expected error for HTTP error response")
			}

			// Check if error has expected retryable property
			if IsRetryable(err) != tc.expectRetryable {
				t.Errorf("Expected retryable=%v, got retryable=%v", tc.expectRetryable, IsRetryable(err))
			}
		})
	}
}

func TestDexScreenerClient_FetchSOLPrice_EmptyResponse(t *testing.T) {
	testCases := []struct {
		name     string
		response DexScreenerResponse
	}{
		{
			name:     "Empty pairs",
			response: DexScreenerResponse{Pairs: []DexScreenerPair{}},
		},
		{
			name: "No valid pairs",
			response: DexScreenerResponse{
				Pairs: []DexScreenerPair{
					{PriceUSD: "0", BaseToken: Token{Symbol: "SOL"}},
					{PriceUSD: "", BaseToken: Token{Symbol: "SOL"}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tc.response)
			}))
			defer server.Close()

			config := DefaultConfig()
			config.DexScreenerURL = server.URL
			client := NewDexScreenerClient(config)

			_, err := client.FetchSOLPrice(context.Background())

			if err == nil {
				t.Error("Expected error for empty/invalid response")
			}
		})
	}
}

func TestDexScreenerClient_FetchSOLPrice_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DexScreenerResponse{})
	}))
	defer server.Close()

	config := DefaultConfig()
	config.DexScreenerURL = server.URL
	client := NewDexScreenerClient(config)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.FetchSOLPrice(ctx)

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	if !strings.Contains(err.Error(), "context") {
		t.Errorf("Expected context-related error, got: %v", err)
	}
}

func TestDexScreenerClient_RateLimiting(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		mockResponse := DexScreenerResponse{
			Pairs: []DexScreenerPair{
				{
					PriceUSD:   "150.0",
					BaseToken:  Token{Symbol: "SOL"},
					QuoteToken: Token{Symbol: "USDC"},
					Liquidity:  Liquidity{USD: 1000000},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.DexScreenerURL = server.URL
	config.RequestsPerMinute = 2 // Very low rate limit for testing
	config.RateLimitWindow = 1 * time.Second
	client := NewDexScreenerClient(config)

	ctx := context.Background()

	// First request should succeed immediately
	start := time.Now()
	_, err := client.FetchSOLPrice(ctx)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}

	// Second request should also succeed but may be delayed
	_, err = client.FetchSOLPrice(ctx)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}

	duration := time.Since(start)

	// Third request should be rate limited and take some time
	start = time.Now()
	_, err = client.FetchSOLPrice(ctx)
	if err != nil {
		t.Fatalf("Third request failed: %v", err)
	}

	thirdRequestDuration := time.Since(start)

	// Third request should have taken longer due to rate limiting
	if thirdRequestDuration < 100*time.Millisecond {
		t.Errorf("Expected rate limiting delay, but request completed too quickly: %v", thirdRequestDuration)
	}

	t.Logf("Total calls made: %d", callCount)
	t.Logf("First two requests took: %v", duration)
	t.Logf("Third request took: %v", thirdRequestDuration)
}

func TestDexScreenerClient_RetryLogic(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// First two attempts fail with server error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Third attempt succeeds
		mockResponse := DexScreenerResponse{
			Pairs: []DexScreenerPair{
				{
					PriceUSD:   "150.0",
					BaseToken:  Token{Symbol: "SOL"},
					QuoteToken: Token{Symbol: "USDC"},
					Liquidity:  Liquidity{USD: 1000000},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.DexScreenerURL = server.URL
	config.MaxRetries = 3
	config.BaseBackoff = 10 * time.Millisecond // Fast backoff for testing
	client := NewDexScreenerClient(config)

	start := time.Now()
	price, err := client.FetchSOLPrice(context.Background())
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if price == nil {
		t.Fatal("Expected non-nil price")
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	// Should have taken some time due to backoff delays
	if duration < 20*time.Millisecond {
		t.Errorf("Expected retry delays, but request completed too quickly: %v", duration)
	}

	t.Logf("Request completed in %v after %d attempts", duration, attemptCount)
}

func TestDexScreenerClient_Close(t *testing.T) {
	client := NewDexScreenerClient(DefaultConfig())

	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error from Close(), got: %v", err)
	}
}

func TestDexScreenerClient_GetConfig(t *testing.T) {
	config := DefaultConfig()
	client := NewDexScreenerClient(config)

	returnedConfig := client.GetConfig()
	if returnedConfig != config {
		t.Error("Expected GetConfig() to return the same config instance")
	}
}
