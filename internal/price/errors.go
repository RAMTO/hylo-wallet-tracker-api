package price

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Common price-related errors
var (
	// ErrPriceNotFound indicates no price data could be retrieved
	ErrPriceNotFound = errors.New("price data not found")

	// ErrPriceStale indicates cached price data is too old
	ErrPriceStale = errors.New("price data is stale")

	// ErrPriceInvalid indicates price data failed validation
	ErrPriceInvalid = errors.New("price data is invalid")

	// ErrRateLimited indicates API rate limit was exceeded
	ErrRateLimited = errors.New("API rate limit exceeded")

	// ErrAPIUnavailable indicates external price API is unavailable
	ErrAPIUnavailable = errors.New("price API unavailable")

	// ErrConfigurationInvalid indicates price service configuration is invalid
	ErrConfigurationInvalid = errors.New("price configuration invalid")

	// ErrCacheUnavailable indicates price cache is not available
	ErrCacheUnavailable = errors.New("price cache unavailable")
)

// PriceError wraps price-related errors with additional context
type PriceError struct {
	// Op is the operation that failed (e.g., "FetchSOLPrice", "ValidatePrice")
	Op string

	// Err is the underlying error
	Err error

	// Source identifies where the error occurred (e.g., "dexscreener", "cache", "validation")
	Source string

	// Timestamp indicates when the error occurred
	Timestamp time.Time

	// Retryable indicates if this error should be retried
	Retryable bool

	// HTTPStatus suggests appropriate HTTP status code for API responses
	HTTPStatus int
}

// Error implements the error interface
func (e *PriceError) Error() string {
	if e.Source != "" {
		return fmt.Sprintf("price %s [%s]: %v", e.Op, e.Source, e.Err)
	}
	return fmt.Sprintf("price %s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error for error unwrapping
func (e *PriceError) Unwrap() error {
	return e.Err
}

// Is implements error comparison for errors.Is()
func (e *PriceError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewPriceError creates a new PriceError with context
func NewPriceError(op string, err error) *PriceError {
	return &PriceError{
		Op:         op,
		Err:        err,
		Timestamp:  time.Now(),
		HTTPStatus: http.StatusInternalServerError, // Default to 500
		Retryable:  false,                          // Default to non-retryable
	}
}

// WithSource adds source context to the error
func (e *PriceError) WithSource(source string) *PriceError {
	e.Source = source
	return e
}

// WithRetryable marks the error as retryable or non-retryable
func (e *PriceError) WithRetryable(retryable bool) *PriceError {
	e.Retryable = retryable
	return e
}

// WithHTTPStatus sets the suggested HTTP status code
func (e *PriceError) WithHTTPStatus(status int) *PriceError {
	e.HTTPStatus = status
	return e
}

// API-specific error constructors following common patterns

// NewDexScreenerError creates an error for DexScreener API issues
func NewDexScreenerError(op string, err error, statusCode int) *PriceError {
	priceErr := NewPriceError(op, err).WithSource("dexscreener")

	// Determine if error is retryable based on HTTP status
	switch {
	case statusCode >= 500:
		// Server errors are usually retryable
		priceErr = priceErr.WithRetryable(true).WithHTTPStatus(http.StatusBadGateway)
	case statusCode == http.StatusTooManyRequests:
		// Rate limiting is retryable after delay
		priceErr = priceErr.WithRetryable(true).WithHTTPStatus(http.StatusServiceUnavailable)
	case statusCode >= 400:
		// Client errors are usually not retryable
		priceErr = priceErr.WithRetryable(false).WithHTTPStatus(http.StatusBadRequest)
	default:
		// Network errors, timeouts, etc. are retryable
		priceErr = priceErr.WithRetryable(true).WithHTTPStatus(http.StatusBadGateway)
	}

	return priceErr
}

// NewValidationError creates an error for price validation failures
func NewValidationError(op string, reason string, value interface{}) *PriceError {
	err := fmt.Errorf("validation failed: %s (value: %v)", reason, value)
	return NewPriceError(op, err).
		WithSource("validation").
		WithRetryable(false).
		WithHTTPStatus(http.StatusBadRequest)
}

// NewCacheError creates an error for cache-related issues
func NewCacheError(op string, err error) *PriceError {
	return NewPriceError(op, err).
		WithSource("cache").
		WithRetryable(true). // Cache errors are usually retryable
		WithHTTPStatus(http.StatusServiceUnavailable)
}

// NewConfigError creates an error for configuration issues
func NewConfigError(op string, err error) *PriceError {
	return NewPriceError(op, err).
		WithSource("config").
		WithRetryable(false). // Config errors are not retryable without fixing config
		WithHTTPStatus(http.StatusInternalServerError)
}

// Convenience functions for common error scenarios

// WrapNetworkError wraps network-related errors with appropriate context
func WrapNetworkError(op string, err error) *PriceError {
	return NewPriceError(op, fmt.Errorf("network error: %w", err)).
		WithSource("network").
		WithRetryable(true).
		WithHTTPStatus(http.StatusBadGateway)
}

// WrapTimeoutError wraps timeout errors with appropriate context
func WrapTimeoutError(op string, err error) *PriceError {
	return NewPriceError(op, fmt.Errorf("timeout: %w", err)).
		WithSource("timeout").
		WithRetryable(true).
		WithHTTPStatus(http.StatusGatewayTimeout)
}

// WrapParsingError wraps JSON parsing or data format errors
func WrapParsingError(op string, err error) *PriceError {
	return NewPriceError(op, fmt.Errorf("parsing error: %w", err)).
		WithSource("parsing").
		WithRetryable(false). // Parsing errors usually indicate API format changes
		WithHTTPStatus(http.StatusBadGateway)
}

// Error classification helpers for monitoring and alerting

// IsRetryable checks if an error should be retried
func IsRetryable(err error) bool {
	var priceErr *PriceError
	if errors.As(err, &priceErr) {
		return priceErr.Retryable
	}
	return false
}

// IsNetworkError checks if an error is network-related
func IsNetworkError(err error) bool {
	var priceErr *PriceError
	if errors.As(err, &priceErr) {
		return priceErr.Source == "network" || priceErr.Source == "timeout"
	}
	return false
}

// IsValidationError checks if an error is due to validation failure
func IsValidationError(err error) bool {
	var priceErr *PriceError
	if errors.As(err, &priceErr) {
		return priceErr.Source == "validation"
	}
	return false
}

// IsAPIError checks if an error is from external API
func IsAPIError(err error) bool {
	var priceErr *PriceError
	if errors.As(err, &priceErr) {
		return priceErr.Source == "dexscreener"
	}
	return false
}

// GetHTTPStatus extracts HTTP status code from error, defaults to 500
func GetHTTPStatus(err error) int {
	var priceErr *PriceError
	if errors.As(err, &priceErr) {
		return priceErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorSource extracts error source from wrapped error
func GetErrorSource(err error) string {
	var priceErr *PriceError
	if errors.As(err, &priceErr) {
		return priceErr.Source
	}
	return "unknown"
}
