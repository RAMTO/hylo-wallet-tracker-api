package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"time"

	"hylo-wallet-tracker-api/internal/logger"
)

// HTTPClient provides HTTP-based Solana RPC functionality
type HTTPClient struct {
	config     *Config
	logger     *logger.Logger
	httpClient *http.Client
	rpcID      int
}

// NewHTTPClient creates a new HTTP client for Solana RPC
func NewHTTPClient(config *Config, serviceLogger *logger.Logger) (*HTTPClient, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	clientLogger := serviceLogger.WithComponent("solana-http-client")

	client := &HTTPClient{
		config: config,
		logger: clientLogger,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
		rpcID: 1,
	}

	clientLogger.InfoContext(context.Background(), "Solana HTTP client created",
		slog.String("rpc_url", config.HttpURL),
		slog.Duration("timeout", config.RequestTimeout))

	return client, nil
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error"`
}

// GetAccount fetches account information for the given address
func (c *HTTPClient) GetAccount(ctx context.Context, address Address, commitment Commitment) (*AccountInfo, error) {
	// Validate inputs
	if err := address.Validate(); err != nil {
		return nil, WrapValidationError("address", address, err.Error())
	}

	if err := commitment.Validate(); err != nil {
		return nil, WrapValidationError("commitment", commitment, err.Error())
	}

	params := []interface{}{
		address.String(),
		map[string]interface{}{
			"encoding":   "base64",
			"commitment": string(commitment),
		},
	}

	var response struct {
		Context struct {
			Slot Slot `json:"slot"`
		} `json:"context"`
		Value *AccountInfo `json:"value"`
	}

	if err := c.request(ctx, "getAccountInfo", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	// Account not found
	if response.Value == nil {
		return nil, ErrAccountNotFound
	}

	return response.Value, nil
}

// GetTransaction fetches transaction details for the given signature
func (c *HTTPClient) GetTransaction(ctx context.Context, signature Signature) (*TransactionDetails, error) {
	// Validate signature
	if err := signature.Validate(); err != nil {
		return nil, WrapValidationError("signature", signature, err.Error())
	}

	params := []interface{}{
		signature.String(),
		map[string]interface{}{
			"encoding":                       "json",
			"commitment":                     "confirmed",
			"maxSupportedTransactionVersion": 0,
			"rewards":                        false,
		},
	}

	var response TransactionDetails

	if err := c.request(ctx, "getTransaction", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &response, nil
}

// GetSignaturesForAddress fetches signatures for the given address
func (c *HTTPClient) GetSignaturesForAddress(ctx context.Context, address Address, before string, limit int) ([]SignatureInfo, error) {
	// Validate address
	if err := address.Validate(); err != nil {
		return nil, WrapValidationError("address", address, err.Error())
	}

	// Validate limit
	if limit <= 0 || limit > 1000 {
		return nil, WrapValidationError("limit", limit, "must be between 1 and 1000")
	}

	params := []interface{}{
		address.String(),
		map[string]interface{}{
			"limit":      limit,
			"commitment": "finalized", // Use finalized for historical data
		},
	}

	// Add before parameter if provided
	if before != "" {
		params[1].(map[string]interface{})["before"] = before
	}

	var response []SignatureInfo

	if err := c.request(ctx, "getSignaturesForAddress", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get signatures: %w", err)
	}

	return response, nil
}

// request performs a JSON-RPC request with retry logic
func (c *HTTPClient) request(ctx context.Context, method string, params interface{}, result interface{}) error {
	startTime := time.Now()
	var lastErr error

	// Log request start
	c.logger.DebugContext(ctx, "Starting Solana RPC request",
		slog.String("method", method),
		slog.Int("max_retries", c.config.MaxRetries))

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		// Calculate backoff delay (exponential with jitter)
		if attempt > 0 {
			delay := c.calculateBackoff(attempt - 1)

			// Log retry attempt
			c.logger.WarnContext(ctx, "Retrying Solana RPC request",
				slog.String("method", method),
				slog.Int("attempt", attempt+1),
				slog.Int("max_retries", c.config.MaxRetries+1),
				slog.Duration("backoff_delay", delay),
				slog.String("previous_error", lastErr.Error()))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := c.doRequest(ctx, method, params, result)
		if err == nil {
			totalTime := time.Since(startTime)

			// Log successful request
			if totalTime > 3*time.Second {
				c.logger.WarnContext(ctx, "Slow Solana RPC request completed",
					slog.String("method", method),
					slog.Duration("total_time", totalTime),
					slog.Int("attempts", attempt+1))
			} else {
				c.logger.InfoContext(ctx, "Solana RPC request completed",
					slog.String("method", method),
					slog.Duration("total_time", totalTime),
					slog.Int("attempts", attempt+1))
			}

			return nil // Success
		}

		lastErr = err

		// Don't retry on validation errors or non-retryable errors
		if !IsRetryable(err) {
			c.logger.LogExternalAPIError(ctx, "solana-rpc", method, err, 0,
				slog.Duration("total_time", time.Since(startTime)),
				slog.Int("attempts", attempt+1),
				slog.String("error_type", "non_retryable"))
			return err
		}

		// Don't retry on the last attempt
		if attempt == c.config.MaxRetries {
			break
		}
	}

	finalError := WrapNetworkError(lastErr, c.config.MaxRetries+1, true)
	c.logger.LogExternalAPIError(ctx, "solana-rpc", method, finalError, 0,
		slog.Duration("total_time", time.Since(startTime)),
		slog.Int("total_attempts", c.config.MaxRetries+1),
		slog.String("error_type", "max_retries_exceeded"))

	return finalError
}

// doRequest performs a single JSON-RPC request without retry
func (c *HTTPClient) doRequest(ctx context.Context, method string, params interface{}, result interface{}) error {
	// Create JSON-RPC request
	c.rpcID++
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.rpcID,
		Method:  method,
		Params:  params,
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.HttpURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "hylo-wallet-tracker/1.0")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return WrapNetworkError(err, 1, false)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WrapNetworkError(err, 1, false)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return WrapRPCError(resp.StatusCode, string(body), nil)
	}

	// Parse JSON-RPC response
	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return fmt.Errorf("failed to parse RPC response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return rpcResp.Error
	}

	// Marshal result back to JSON and unmarshal to target type
	// This handles type conversion properly
	resultBytes, err := json.Marshal(rpcResp.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultBytes, result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}

// calculateBackoff returns the backoff delay for the given attempt with jitter
func (c *HTTPClient) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: BaseBackoff * 2^attempt
	backoff := time.Duration(float64(c.config.BaseBackoff) * math.Pow(2, float64(attempt)))

	// Cap at MaxBackoff
	if backoff > c.config.MaxBackoff {
		backoff = c.config.MaxBackoff
	}

	// Add jitter (±25% random variance)
	jitter := time.Duration(rand.Float64() * 0.5 * float64(backoff)) // 0-50% of backoff
	if rand.Intn(2) == 0 {
		backoff -= jitter // 50% chance to subtract
	} else {
		backoff += jitter // 50% chance to add
	}

	// Ensure backoff doesn't exceed MaxBackoff after jitter
	if backoff > c.config.MaxBackoff {
		backoff = c.config.MaxBackoff
	}

	// Ensure minimum 100ms backoff
	if backoff < 100*time.Millisecond {
		backoff = 100 * time.Millisecond
	}

	return backoff
}

// Close closes the HTTP client
func (c *HTTPClient) Close() error {
	c.logger.InfoContext(context.Background(), "Closing Solana HTTP client")
	// HTTP client doesn't need explicit closing, but this provides
	// a consistent interface with WebSocket client
	return nil
}

// Health returns the health status of the HTTP client
func (c *HTTPClient) Health(ctx context.Context) error {
	// Simple health check by making a basic RPC call
	c.logger.DebugContext(ctx, "Performing Solana HTTP client health check")

	_, err := c.GetAccount(ctx, Address("A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g"), CommitmentFinalized)

	// Account not found is expected for the null address, so it's healthy
	if errors.Is(err, ErrAccountNotFound) {
		c.logger.DebugContext(ctx, "Solana HTTP client health check passed (account not found as expected)")
		return nil
	}

	if err != nil {
		c.logger.LogExternalAPIError(ctx, "solana-rpc", "health-check", err, 0)
	}

	return err
}
