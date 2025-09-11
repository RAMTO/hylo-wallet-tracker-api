package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger with service-specific metadata
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level       string // debug, info, warn, error
	Format      string // json, text
	ServiceName string
	Version     string
}

// New creates a new configured logger instance
func New(config Config) *Logger {
	// Set default values
	if config.ServiceName == "" {
		config.ServiceName = "wallet-tracker-api"
	}
	if config.Version == "" {
		config.Version = "dev"
	}
	if config.Level == "" {
		config.Level = "info"
	}
	if config.Format == "" {
		config.Format = "json"
	}

	// Parse log level
	level := parseLogLevel(config.Level)

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if strings.ToLower(config.Format) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	// Create logger with service metadata
	logger := slog.New(handler).With(
		slog.String("service", config.ServiceName),
		slog.String("version", config.Version),
	)

	return &Logger{Logger: logger}
}

// NewFromEnv creates a logger from environment variables
func NewFromEnv() *Logger {
	return New(Config{
		Level:       getEnv("LOG_LEVEL", "info"),
		Format:      getEnv("LOG_FORMAT", "json"),
		ServiceName: getEnv("SERVICE_NAME", "wallet-tracker-api"),
		Version:     getEnv("SERVICE_VERSION", "dev"),
	})
}

// WithComponent adds component information to logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("component", component)),
	}
}

// WithOperation adds operation context to logger
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("operation", operation)),
	}
}

// parseLogLevel converts string level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getEnv gets environment variable with default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RequestIDKey is the context key for request ID
type contextKey string

const RequestIDKey contextKey = "request_id"

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
