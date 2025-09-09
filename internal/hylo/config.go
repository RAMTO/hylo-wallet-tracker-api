package hylo

import (
	"fmt"
	"os"
	"strings"

	"hylo-wallet-tracker-api/internal/solana"
)

// Config holds Hylo protocol configuration and provides program registry functionality
type Config struct {
	// ExchangeProgramID can be overridden via HYLO_EXCHANGE_PROGRAM_ID environment variable
	ExchangeProgramID solana.Address

	// StabilityPoolProgramID can be overridden via HYLO_STABILITY_POOL_PROGRAM_ID environment variable
	StabilityPoolProgramID solana.Address

	// programRegistry is an internal map for fast program lookups
	programRegistry map[solana.Address]string
}

// NewConfig creates a new Hylo configuration with default mainnet addresses
// Environment variables can override default program addresses for different environments
func NewConfig() *Config {
	config := &Config{
		// Default to mainnet addresses from constants
		ExchangeProgramID:      ExchangeProgram,
		StabilityPoolProgramID: StabilityPoolProgram,
	}

	// Load configuration from environment variables
	config.loadFromEnvironment()

	// Build the program registry for fast lookups
	config.buildProgramRegistry()

	return config
}

// loadFromEnvironment loads program addresses from environment variables
// This allows different environments (testnet, devnet) to use different addresses
func (c *Config) loadFromEnvironment() {
	// Load exchange program ID if provided
	if exchangeID := os.Getenv("HYLO_EXCHANGE_PROGRAM_ID"); exchangeID != "" {
		c.ExchangeProgramID = solana.Address(strings.TrimSpace(exchangeID))
	}

	// Load stability pool program ID if provided
	if stabilityPoolID := os.Getenv("HYLO_STABILITY_POOL_PROGRAM_ID"); stabilityPoolID != "" {
		c.StabilityPoolProgramID = solana.Address(strings.TrimSpace(stabilityPoolID))
	}
}

// buildProgramRegistry builds an internal registry for fast program lookups
// Maps program address to program name for identification purposes
func (c *Config) buildProgramRegistry() {
	c.programRegistry = make(map[solana.Address]string)
	c.programRegistry[c.ExchangeProgramID] = "Exchange"
	c.programRegistry[c.StabilityPoolProgramID] = "StabilityPool"
}

// Validate ensures all program addresses are valid and properly configured
func (c *Config) Validate() error {
	// Validate exchange program address
	if err := c.ExchangeProgramID.Validate(); err != nil {
		return fmt.Errorf("invalid exchange program ID: %w", err)
	}

	// Validate stability pool program address
	if err := c.StabilityPoolProgramID.Validate(); err != nil {
		return fmt.Errorf("invalid stability pool program ID: %w", err)
	}

	// Check for duplicate program addresses (shouldn't be the same)
	if c.ExchangeProgramID == c.StabilityPoolProgramID {
		return fmt.Errorf("exchange and stability pool programs cannot have the same address")
	}

	return nil
}

// IsHyloProgramID checks if the given address is a supported Hylo program
// Returns true if the address matches any configured Hylo program
func (c *Config) IsHyloProgramID(address solana.Address) bool {
	_, exists := c.programRegistry[address]
	return exists
}

// GetProgramName returns the name of the program for the given address
// Returns empty string if the address is not a known Hylo program
func (c *Config) GetProgramName(address solana.Address) string {
	name, exists := c.programRegistry[address]
	if !exists {
		return ""
	}
	return name
}

// GetSupportedPrograms returns all configured Hylo program addresses
// Used for transaction filtering and program identification
func (c *Config) GetSupportedPrograms() []solana.Address {
	return []solana.Address{
		c.ExchangeProgramID,
		c.StabilityPoolProgramID,
	}
}

// GetExchangeProgramID returns the configured exchange program address
func (c *Config) GetExchangeProgramID() solana.Address {
	return c.ExchangeProgramID
}

// GetStabilityPoolProgramID returns the configured stability pool program address
func (c *Config) GetStabilityPoolProgramID() solana.Address {
	return c.StabilityPoolProgramID
}

// IsExchangeProgram checks if the given address is the exchange program
func (c *Config) IsExchangeProgram(address solana.Address) bool {
	return address == c.ExchangeProgramID
}

// IsStabilityPoolProgram checks if the given address is the stability pool program
func (c *Config) IsStabilityPoolProgram(address solana.Address) bool {
	return address == c.StabilityPoolProgramID
}
