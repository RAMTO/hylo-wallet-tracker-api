package solana

import (
	"errors"
	"time"
)

// Config holds configuration for Solana RPC client
type Config struct {
	// HTTP RPC endpoint URL
	HttpURL string

	// WebSocket RPC endpoint URL
	WebSocketURL string

	// HTTP request timeout
	RequestTimeout time.Duration

	// Maximum number of HTTP retries on transient failures
	MaxRetries int

	// Base backoff duration for exponential backoff
	// Actual backoff will be: BaseBackoff * (2 ^ attempt) with jitter
	BaseBackoff time.Duration

	// Maximum backoff duration (cap for exponential backoff)
	MaxBackoff time.Duration

	// WebSocket heartbeat interval for keepalive pings
	HeartbeatInterval time.Duration

	// WebSocket reconnect timeout
	ReconnectTimeout time.Duration
}

// NewConfig creates a new Config with sensible defaults
func NewConfig(httpURL, wsURL string) *Config {
	return &Config{
		HttpURL:           httpURL,
		WebSocketURL:      wsURL,
		RequestTimeout:    30 * time.Second,
		MaxRetries:        3,
		BaseBackoff:       1 * time.Second,
		MaxBackoff:        30 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		ReconnectTimeout:  60 * time.Second,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.HttpURL == "" {
		return errors.New("HttpURL is required")
	}

	if c.WebSocketURL == "" {
		return errors.New("WebSocketURL is required")
	}

	if c.RequestTimeout <= 0 {
		return errors.New("RequestTimeout must be positive")
	}

	if c.MaxRetries < 0 {
		return errors.New("MaxRetries cannot be negative")
	}

	if c.BaseBackoff <= 0 {
		return errors.New("BaseBackoff must be positive")
	}

	if c.MaxBackoff < c.BaseBackoff {
		return errors.New("MaxBackoff must be >= BaseBackoff")
	}

	if c.HeartbeatInterval <= 0 {
		return errors.New("HeartbeatInterval must be positive")
	}

	if c.ReconnectTimeout <= 0 {
		return errors.New("ReconnectTimeout must be positive")
	}

	return nil
}

// WithTimeout returns a new config with the specified request timeout
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	newConfig := *c
	newConfig.RequestTimeout = timeout
	return &newConfig
}

// WithRetries returns a new config with the specified max retries
func (c *Config) WithRetries(retries int) *Config {
	newConfig := *c
	newConfig.MaxRetries = retries
	return &newConfig
}
