package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Enhanced Helper Functions for consistent response handling

// writeJSONError writes a structured error response with categorization
func (s *Server) writeJSONError(w http.ResponseWriter, statusCode int, message string, details string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		BaseResponse: BaseResponse{
			Timestamp: getCurrentTimestamp(),
			RequestID: extractRequestID(w),
		},
		Error:   message,
		Details: details,
		Code:    code,
	}

	json.NewEncoder(w).Encode(response)
}

// writeJSONSuccess writes a structured success response
func (s *Server) writeJSONSuccess(w http.ResponseWriter, data interface{}) {
	s.writeJSONSuccessWithCode(w, http.StatusOK, data)
}

// writeJSONSuccessWithCode writes a success response with custom status code
func (s *Server) writeJSONSuccessWithCode(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// For backward compatibility, return data directly (not wrapped in SuccessResponse)
	// This maintains current API response formats for /wallet/{address}/balances
	json.NewEncoder(w).Encode(data)
}

// Specialized error helper functions with predefined categories

// writeValidationError writes a validation error (400) with VALIDATION_ERROR code
func (s *Server) writeValidationError(w http.ResponseWriter, message string, details string) {
	s.writeJSONError(w, http.StatusBadRequest, message, details, ErrorCodeValidation)
}

// writeNotFoundError writes a not found error (404) with NOT_FOUND code
func (s *Server) writeNotFoundError(w http.ResponseWriter, resource string) {
	message := resource + " not found"
	s.writeJSONError(w, http.StatusNotFound, message, "", ErrorCodeNotFound)
}

// writeInternalError writes an internal server error (500) with INTERNAL_ERROR code
func (s *Server) writeInternalError(w http.ResponseWriter, details string) {
	s.writeJSONError(w, http.StatusInternalServerError, "Internal server error", details, ErrorCodeInternal)
}

// writeNetworkError writes a network error (502) with NETWORK_ERROR code
func (s *Server) writeNetworkError(w http.ResponseWriter, details string) {
	s.writeJSONError(w, http.StatusBadGateway, "Network connectivity issue", details, ErrorCodeNetwork)
}

// Error classification helpers

// isNetworkError checks if error is network-related
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	networkIndicators := []string{
		"connection refused",
		"timeout",
		"network",
		"dns",
		"unreachable",
		"rpc error",
		"dial tcp",
		"no such host",
		"connection reset",
	}

	for _, indicator := range networkIndicators {
		if strings.Contains(errStr, indicator) {
			return true
		}
	}
	return false
}

// isValidationError checks for validation-related errors
func isValidationError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "validation") ||
		strings.Contains(errStr, "malformed")
}
