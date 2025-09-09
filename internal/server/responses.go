package server

import (
	"net/http"
	"time"
)

// Base response structures for consistent API responses
type BaseResponse struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	BaseResponse
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    string `json:"code,omitempty"` // For categorizing errors
}

// SuccessResponse represents a structured success response wrapper (optional - for future use)
type SuccessResponse struct {
	BaseResponse
	Data interface{} `json:"data,omitempty"`
}

// HealthResponse represents the health check response (matches current format)
type HealthResponse struct {
	Status    string      `json:"status"`
	Solana    interface{} `json:"solana"`
	Timestamp string      `json:"timestamp"`
}

// Error codes for categorization - helps with monitoring and debugging
const (
	ErrorCodeValidation   = "VALIDATION_ERROR"
	ErrorCodeNotFound     = "NOT_FOUND"
	ErrorCodeNetwork      = "NETWORK_ERROR"
	ErrorCodeInternal     = "INTERNAL_ERROR"
	ErrorCodeRateLimit    = "RATE_LIMIT"
	ErrorCodeUnauthorized = "UNAUTHORIZED"
)

// Helper function to create timestamp in consistent format
func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// extractRequestID extracts request ID from response writer (for future middleware integration)
func extractRequestID(w http.ResponseWriter) string {
	// This can be enhanced later with request ID middleware
	// For now, return empty string
	return ""
}
