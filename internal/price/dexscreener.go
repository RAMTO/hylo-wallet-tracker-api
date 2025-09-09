package price

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DexScreenerClient handles interactions with the DexScreener API for SOL price data
type DexScreenerClient struct {
	// httpClient is the underlying HTTP client with configured timeouts
	httpClient *http.Client

	// config holds the client configuration
	config *PriceConfig

	// baseURL is the DexScreener API base URL
	baseURL string

	// rateLimiter handles request rate limiting
	rateLimiter *rateLimiter

	// lastRequest tracks the last request time for rate limiting
	lastRequest time.Time
	requestMu   sync.Mutex
}

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewDexScreenerClient creates a new DexScreener API client with the given configuration
func NewDexScreenerClient(config *PriceConfig) *DexScreenerClient {
	if config == nil {
		config = DefaultConfig()
	}

	client := &DexScreenerClient{
		httpClient: &http.Client{
			Timeout: config.DexScreenerTimeout,
		},
		config:  config,
		baseURL: strings.TrimSuffix(config.DexScreenerURL, "/"),
		rateLimiter: &rateLimiter{
			tokens:     config.RequestsPerMinute,
			maxTokens:  config.RequestsPerMinute,
			refillRate: config.RateLimitWindow / time.Duration(config.RequestsPerMinute),
		},
	}

	return client
}

// FetchSOLPrice fetches the current SOL/USD price from DexScreener
// Returns the best price based on liquidity and trading volume
func (c *DexScreenerClient) FetchSOLPrice(ctx context.Context) (*SOLUSDPrice, error) {
	const op = "FetchSOLPrice"

	// Apply rate limiting
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, NewPriceError(op, err).WithSource("rate_limit").WithRetryable(false)
	}

	// Build the request URL for SOL pairs
	requestURL := fmt.Sprintf("%s/latest/dex/tokens/So11111111111111111111111111111111111111112", c.baseURL)

	// Fetch data with retries
	response, err := c.fetchWithRetry(ctx, requestURL)
	if err != nil {
		// If this is already a PriceError, preserve it; otherwise wrap it
		var priceErr *PriceError
		if errors.As(err, &priceErr) {
			return nil, err // Already wrapped with proper retry/status info
		}
		return nil, NewPriceError(op, err).WithSource("dexscreener")
	}

	// Parse the response
	solPrice, err := c.parseSOLPriceResponse(response)
	if err != nil {
		return nil, NewPriceError(op, err).WithSource("parsing")
	}

	// Validate the price
	if !c.config.IsValidSOLPrice(solPrice.Price) {
		return nil, NewValidationError(op,
			fmt.Sprintf("price %f outside valid range [%f, %f]",
				solPrice.Price, c.config.SOLUSDMinPrice, c.config.SOLUSDMaxPrice),
			solPrice.Price)
	}

	return solPrice, nil
}

// waitForRateLimit blocks until a request can be made according to rate limiting rules
func (c *DexScreenerClient) waitForRateLimit(ctx context.Context) error {
	c.requestMu.Lock()
	defer c.requestMu.Unlock()

	// Check if we need to wait based on rate limiting
	if !c.rateLimiter.allowRequest() {
		// Calculate wait time until next token is available
		waitTime := c.rateLimiter.timeUntilNextToken()

		// Use a timer that respects context cancellation
		timer := time.NewTimer(waitTime)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return fmt.Errorf("rate limit wait cancelled: %w", ctx.Err())
		case <-timer.C:
			// Wait completed, proceed with request
		}
	}

	c.lastRequest = time.Now()
	return nil
}

// fetchWithRetry performs HTTP request with exponential backoff retry logic
func (c *DexScreenerClient) fetchWithRetry(ctx context.Context, requestURL string) (*DexScreenerResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		// Add backoff delay for retry attempts
		if attempt > 0 {
			backoffDelay := c.config.CalculateBackoff(attempt - 1)

			timer := time.NewTimer(backoffDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, fmt.Errorf("request cancelled during backoff: %w", ctx.Err())
			case <-timer.C:
				// Continue with retry attempt
			}
		}

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add proper headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "HyloWalletTracker/1.0")

		// Execute the request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = WrapNetworkError("http_request", err)

			// Check if this is a retryable error
			if !IsRetryable(lastErr) || attempt == c.config.MaxRetries {
				return nil, lastErr
			}
			continue // Retry the request
		}

		// Handle HTTP response
		result, err := c.handleHTTPResponse(resp)
		if err != nil {
			lastErr = err

			// Check if we should retry based on error type
			if !IsRetryable(err) || attempt == c.config.MaxRetries {
				return nil, err
			}
			continue // Retry the request
		}

		return result, nil
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
	}

	return nil, fmt.Errorf("request failed after %d attempts", c.config.MaxRetries+1)
}

// handleHTTPResponse processes HTTP response and handles various status codes
func (c *DexScreenerClient) handleHTTPResponse(resp *http.Response) (*DexScreenerResponse, error) {
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, WrapNetworkError("read_response", err)
	}

	// Handle different HTTP status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Success - parse JSON response
		var dexResponse DexScreenerResponse
		if err := json.Unmarshal(body, &dexResponse); err != nil {
			return nil, WrapParsingError("json_unmarshal", err)
		}
		return &dexResponse, nil

	case http.StatusTooManyRequests:
		return nil, NewDexScreenerError("rate_limited", ErrRateLimited, resp.StatusCode)

	case http.StatusNotFound:
		return nil, NewDexScreenerError("not_found", ErrPriceNotFound, resp.StatusCode)

	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return nil, NewDexScreenerError("server_error",
			fmt.Errorf("server error: %s", resp.Status), resp.StatusCode)

	default:
		// Other client/server errors
		return nil, NewDexScreenerError("http_error",
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)), resp.StatusCode)
	}
}

// parseSOLPriceResponse extracts the best SOL price from DexScreener response
func (c *DexScreenerClient) parseSOLPriceResponse(response *DexScreenerResponse) (*SOLUSDPrice, error) {
	if response == nil || len(response.Pairs) == 0 {
		return nil, fmt.Errorf("no trading pairs found in response")
	}

	// Find the best pair based on liquidity and volume
	var bestPair *DexScreenerPair
	var bestScore float64

	for i := range response.Pairs {
		pair := &response.Pairs[i]

		// Skip pairs that don't have USD pricing
		if pair.PriceUSD == "" || pair.PriceUSD == "0" {
			continue
		}

		// Parse the price
		price, err := strconv.ParseFloat(pair.PriceUSD, 64)
		if err != nil || price <= 0 {
			continue
		}

		// Validate price is within reasonable bounds
		if !c.config.IsValidSOLPrice(price) {
			continue
		}

		// Calculate a score based on liquidity and volume
		// Prefer pairs with higher liquidity and volume
		liquidityScore := pair.Liquidity.USD
		volumeScore := pair.Volume.H24

		// Combined score favoring liquidity more heavily
		score := liquidityScore + (volumeScore * 0.1)

		if bestPair == nil || score > bestScore {
			bestPair = pair
			bestScore = score
		}
	}

	if bestPair == nil {
		return nil, fmt.Errorf("no valid trading pairs found")
	}

	// Parse the final price
	price, err := strconv.ParseFloat(bestPair.PriceUSD, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse best pair price: %w", err)
	}

	// Determine the pair symbol (prefer major stablecoins)
	pairSymbol := fmt.Sprintf("SOL/%s", bestPair.QuoteToken.Symbol)

	return &SOLUSDPrice{
		Price:     price,
		Timestamp: time.Now(),
		Source:    "dexscreener",
		Pair:      pairSymbol,
		Liquidity: bestPair.Liquidity.USD,
		Volume24h: bestPair.Volume.H24,
	}, nil
}

// Rate limiter implementation

// allowRequest checks if a request can be made and consumes a token if available
func (r *rateLimiter) allowRequest() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on elapsed time
	r.refillTokens()

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

// refillTokens adds tokens to the bucket based on elapsed time
func (r *rateLimiter) refillTokens() {
	now := time.Now()
	if r.lastRefill.IsZero() {
		r.lastRefill = now
		return
	}

	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed / r.refillRate)

	if tokensToAdd > 0 {
		r.tokens += tokensToAdd
		if r.tokens > r.maxTokens {
			r.tokens = r.maxTokens
		}
		r.lastRefill = now
	}
}

// timeUntilNextToken calculates how long to wait until the next token is available
func (r *rateLimiter) timeUntilNextToken() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tokens > 0 {
		return 0
	}

	// Calculate time until next refill
	return r.refillRate - time.Since(r.lastRefill)
}

// Close performs cleanup (currently no-op but provided for interface consistency)
func (c *DexScreenerClient) Close() error {
	// Close HTTP client connections if needed
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// GetConfig returns the client configuration (for testing/debugging)
func (c *DexScreenerClient) GetConfig() *PriceConfig {
	return c.config
}
