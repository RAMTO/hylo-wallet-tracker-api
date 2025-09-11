package logger

import (
	"context"
	"log/slog"
	"net/http"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeExternalAPI ErrorType = "external_api"
	ErrorTypeInternal    ErrorType = "internal"
	ErrorTypeParsing     ErrorType = "parsing"
)

// LogHandlerError logs errors that occur in HTTP handlers
func (l *Logger) LogHandlerError(ctx context.Context, operation string, err error, additionalFields ...slog.Attr) {
	requestID := GetRequestID(ctx)

	attrs := []slog.Attr{
		slog.String("component", "handler"),
		slog.String("operation", operation),
		slog.Group("error",
			slog.String("message", err.Error()),
		),
	}

	// Add request ID if available
	if requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	// Add any additional fields
	attrs = append(attrs, additionalFields...)

	l.LogAttrs(ctx, slog.LevelError, "Handler error occurred", attrs...)
}

// LogExternalAPIError logs errors from external API calls
func (l *Logger) LogExternalAPIError(ctx context.Context, service, endpoint string, err error, statusCode int, additionalFields ...slog.Attr) {
	requestID := GetRequestID(ctx)

	attrs := []slog.Attr{
		slog.String("component", "handler"),
		slog.String("external_service", service),
		slog.String("endpoint", endpoint),
		slog.Group("error",
			slog.String("type", string(ErrorTypeExternalAPI)),
			slog.String("message", err.Error()),
		),
		slog.Int("status_code", statusCode),
	}

	// Add request ID if available
	if requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	// Add any additional fields
	attrs = append(attrs, additionalFields...)

	l.LogAttrs(ctx, slog.LevelError, "External API call failed", attrs...)
}

// LogValidationError logs input validation errors
func (l *Logger) LogValidationError(ctx context.Context, operation, field string, value interface{}, err error) {
	requestID := GetRequestID(ctx)

	attrs := []slog.Attr{
		slog.String("component", "handler"),
		slog.String("operation", operation),
		slog.Group("error",
			slog.String("type", string(ErrorTypeValidation)),
			slog.String("message", err.Error()),
			slog.String("field", field),
			slog.Any("invalid_value", value),
		),
	}

	// Add request ID if available
	if requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	l.LogAttrs(ctx, slog.LevelError, "Validation error", attrs...)
}

// LogParsingError logs data parsing errors
func (l *Logger) LogParsingError(ctx context.Context, operation, dataType string, err error, additionalFields ...slog.Attr) {
	requestID := GetRequestID(ctx)

	attrs := []slog.Attr{
		slog.String("component", "handler"),
		slog.String("operation", operation),
		slog.Group("error",
			slog.String("type", string(ErrorTypeParsing)),
			slog.String("message", err.Error()),
			slog.String("data_type", dataType),
		),
	}

	// Add request ID if available
	if requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	// Add any additional fields
	attrs = append(attrs, additionalFields...)

	l.LogAttrs(ctx, slog.LevelError, "Parsing error occurred", attrs...)
}

// LogHTTPError logs HTTP-related errors with request context
func (l *Logger) LogHTTPError(ctx context.Context, operation string, r *http.Request, statusCode int, err error, additionalFields ...slog.Attr) {
	requestID := GetRequestID(ctx)

	attrs := []slog.Attr{
		slog.String("component", "handler"),
		slog.String("operation", operation),
		slog.Group("http",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status_code", statusCode),
		),
		slog.Group("error",
			slog.String("type", string(ErrorTypeInternal)),
			slog.String("message", err.Error()),
		),
	}

	// Add request ID if available
	if requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	// Add any additional fields
	attrs = append(attrs, additionalFields...)

	l.LogAttrs(ctx, slog.LevelError, "HTTP error", attrs...)
}

// WithWalletAddress creates a logger with wallet address context
func (l *Logger) WithWalletAddress(address string) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("wallet_address", address)),
	}
}

// WithTokenCount creates a logger with token count context
func (l *Logger) WithTokenCount(count int) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.Int("token_count", count)),
	}
}
