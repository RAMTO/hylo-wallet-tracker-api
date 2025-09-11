package logger

import (
	"net/http"

	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request context
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Add request ID to context
		ctx := WithRequestID(r.Context(), requestID)

		// Add request ID to response headers for debugging
		w.Header().Set("X-Request-ID", requestID)

		// Continue with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
