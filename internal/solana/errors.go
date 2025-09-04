package solana

import (
	"errors"
	"fmt"
)

// Error types for Solana RPC operations
var (
	// ErrInvalidAddress indicates an invalid Solana address
	ErrInvalidAddress = errors.New("invalid solana address")

	// ErrInvalidSignature indicates an invalid transaction signature
	ErrInvalidSignature = errors.New("invalid transaction signature")

	// ErrInvalidCommitment indicates an invalid commitment level
	ErrInvalidCommitment = errors.New("invalid commitment level")

	// ErrConnectionClosed indicates the connection was closed
	ErrConnectionClosed = errors.New("connection closed")

	// ErrTimeout indicates a request timed out
	ErrTimeout = errors.New("request timeout")

	// ErrMaxRetriesExceeded indicates max retry attempts were exceeded
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")

	// ErrAccountNotFound indicates the account doesn't exist
	ErrAccountNotFound = errors.New("account not found")
)

// RPCError represents an error returned by the Solana RPC
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *RPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("RPC error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// IsRetryable returns true if this RPC error should be retried
func (e *RPCError) IsRetryable() bool {
	// Retry on server errors and rate limiting
	return e.Code >= 500 || e.Code == 429
}

// NetworkError wraps network-related errors with retry information
type NetworkError struct {
	Err     error
	Attempt int
	Final   bool
}

// Error implements the error interface
func (e *NetworkError) Error() string {
	if e.Final {
		return fmt.Sprintf("network error after %d attempts: %v", e.Attempt, e.Err)
	}
	return fmt.Sprintf("network error (attempt %d): %v", e.Attempt, e.Err)
}

// Unwrap returns the wrapped error for errors.Is/As compatibility
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if this network error should be retried
func (e *NetworkError) IsRetryable() bool {
	return !e.Final
}

// ValidationError wraps parameter validation errors
type ValidationError struct {
	Field  string
	Value  interface{}
	Reason string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Reason, e.Value)
}

// WrapNetworkError creates a NetworkError from a regular error
func WrapNetworkError(err error, attempt int, final bool) error {
	return &NetworkError{
		Err:     err,
		Attempt: attempt,
		Final:   final,
	}
}

// WrapRPCError creates an RPCError from HTTP response details
func WrapRPCError(statusCode int, message string, data interface{}) error {
	return &RPCError{
		Code:    statusCode,
		Message: message,
		Data:    data,
	}
}

// WrapValidationError creates a ValidationError
func WrapValidationError(field string, value interface{}, reason string) error {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
	}
}

// IsRetryable checks if an error should be retried
func IsRetryable(err error) bool {
	var rpcErr *RPCError
	var netErr *NetworkError

	if errors.As(err, &rpcErr) {
		return rpcErr.IsRetryable()
	}

	if errors.As(err, &netErr) {
		return netErr.IsRetryable()
	}

	// For other errors, assume not retryable by default

	return false
}

// IsTemporary checks if an error is temporary and might succeed on retry
func IsTemporary(err error) bool {
	// Check for temporary network errors
	if temp, ok := err.(interface{ Temporary() bool }); ok {
		return temp.Temporary()
	}

	return IsRetryable(err)
}
