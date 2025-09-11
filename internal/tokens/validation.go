package tokens

import (
	"fmt"
	"regexp"
	"strings"

	"hylo-wallet-tracker-api/internal/solana"
)

// AddressValidationError represents validation errors for addresses
type AddressValidationError struct {
	Address string
	Reason  string
}

func (e *AddressValidationError) Error() string {
	return fmt.Sprintf("invalid address %s: %s", e.Address, e.Reason)
}

// SanitizeAddress cleans and normalizes a Solana address string
// Removes whitespace, converts to proper case, and validates format
func SanitizeAddress(address string) (solana.Address, error) {
	if address == "" {
		return solana.Address(""), &AddressValidationError{
			Address: address,
			Reason:  "address cannot be empty",
		}
	}

	// Trim whitespace
	cleaned := strings.TrimSpace(address)

	// Check length (Solana addresses are 32-44 characters in base58)
	if len(cleaned) < 32 || len(cleaned) > 44 {
		return solana.Address(""), &AddressValidationError{
			Address: address,
			Reason:  fmt.Sprintf("invalid length: expected 32-44 characters, got %d", len(cleaned)),
		}
	}

	// Validate character set (base58 alphabet)
	validChars := regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]+$`)
	if !validChars.MatchString(cleaned) {
		return solana.Address(""), &AddressValidationError{
			Address: address,
			Reason:  "contains invalid characters (must be valid base58)",
		}
	}

	// Create address and validate using existing solana.Address validation
	solanaAddr := solana.Address(cleaned)
	if err := solanaAddr.Validate(); err != nil {
		return solana.Address(""), &AddressValidationError{
			Address: address,
			Reason:  fmt.Sprintf("failed solana validation: %s", err.Error()),
		}
	}

	return solanaAddr, nil
}

// ValidateWalletAddress validates a wallet address with additional context
func ValidateWalletAddress(address string) error {
	_, err := SanitizeAddress(address)
	if err != nil {
		return fmt.Errorf("invalid wallet address: %w", err)
	}
	return nil
}

// ValidateTokenMintAddress validates a token mint address and checks if it's supported
func ValidateTokenMintAddress(address string, config *Config) error {
	sanitized, err := SanitizeAddress(address)
	if err != nil {
		return fmt.Errorf("invalid mint address: %w", err)
	}

	if config != nil && !config.IsTokenSupported(sanitized) {
		return &AddressValidationError{
			Address: address,
			Reason:  "unsupported token mint (not a Hylo token)",
		}
	}

	return nil
}

// ValidateATAAddress validates an Associated Token Account address
// This checks basic format but does not verify the address was derived correctly
func ValidateATAAddress(address string) error {
	_, err := SanitizeAddress(address)
	if err != nil {
		return fmt.Errorf("invalid ATA address: %w", err)
	}
	return nil
}

// ValidateATADerivation verifies that an ATA address was correctly derived
// from the given wallet and mint addresses
func ValidateATADerivation(ataAddress, walletAddress, mintAddress string) error {
	// Sanitize all addresses
	sanitizedATA, err := SanitizeAddress(ataAddress)
	if err != nil {
		return fmt.Errorf("invalid ATA address: %w", err)
	}

	sanitizedWallet, err := SanitizeAddress(walletAddress)
	if err != nil {
		return fmt.Errorf("invalid wallet address: %w", err)
	}

	sanitizedMint, err := SanitizeAddress(mintAddress)
	if err != nil {
		return fmt.Errorf("invalid mint address: %w", err)
	}

	// Derive the expected ATA address
	expectedATA, err := DeriveAssociatedTokenAddress(sanitizedWallet, sanitizedMint)
	if err != nil {
		return fmt.Errorf("failed to derive expected ATA: %w", err)
	}

	// Compare with provided ATA address
	if sanitizedATA != expectedATA {
		return &AddressValidationError{
			Address: ataAddress,
			Reason:  fmt.Sprintf("does not match derived ATA %s", expectedATA),
		}
	}

	return nil
}

// IsValidBase58Character checks if a character is valid in base58 encoding
func IsValidBase58Character(c byte) bool {
	// Base58 alphabet: 123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz
	// Excludes: 0 (zero), O (capital o), I (capital i), l (lowercase L)
	return (c >= '1' && c <= '9') ||
		(c >= 'A' && c <= 'H') ||
		(c >= 'J' && c <= 'N') ||
		(c >= 'P' && c <= 'Z') ||
		(c >= 'a' && c <= 'k') ||
		(c >= 'm' && c <= 'z')
}

// ValidateAddressBatch validates multiple addresses efficiently
func ValidateAddressBatch(addresses []string) []error {
	errors := make([]error, len(addresses))

	for i, address := range addresses {
		_, err := SanitizeAddress(address)
		if err != nil {
			errors[i] = fmt.Errorf("address %d: %w", i, err)
		}
	}

	return errors
}

// SuggestCorrection attempts to suggest corrections for common address input errors
func SuggestCorrection(address string) []string {
	var suggestions []string

	// If too short, might be missing characters
	if len(address) < 32 {
		suggestions = append(suggestions, "Address too short - Solana addresses are 32-44 characters")
	}

	// If too long, might have extra characters
	if len(address) > 44 {
		trimmed := strings.TrimSpace(address)
		if len(trimmed) >= 32 && len(trimmed) <= 44 {
			suggestions = append(suggestions, fmt.Sprintf("Try removing whitespace: '%s'", trimmed))
		} else {
			suggestions = append(suggestions, "Address too long - Solana addresses are 32-44 characters")
		}
	}

	// Check for common character substitutions
	if strings.Contains(address, "0") {
		corrected := strings.ReplaceAll(address, "0", "O")
		if len(corrected) == 44 {
			suggestions = append(suggestions, fmt.Sprintf("Try replacing '0' with 'O': %s", corrected))
		}
	}

	if strings.Contains(address, "I") {
		suggestions = append(suggestions, "Base58 does not contain 'I' - did you mean '1' or 'l'?")
	}

	if strings.Contains(address, "l") && strings.Contains(address, "I") {
		suggestions = append(suggestions, "Both 'I' and 'l' are not valid - check for correct characters")
	}

	return suggestions
}

// ValidateTokenConfiguration validates that all token mints in the configuration
// are valid Solana addresses and don't conflict with each other
func ValidateTokenConfiguration(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Get all supported tokens
	supportedTokens := config.GetSupportedTokens()
	if len(supportedTokens) == 0 {
		return fmt.Errorf("no supported tokens configured")
	}

	// Track mint addresses to check for duplicates
	mintAddresses := make(map[string]string) // mint -> symbol

	for _, tokenInfo := range supportedTokens {
		// Validate the mint address
		if err := ValidateTokenMintAddress(string(tokenInfo.Mint), nil); err != nil {
			return fmt.Errorf("invalid mint for %s: %w", tokenInfo.Symbol, err)
		}

		// Check for duplicate mint addresses
		mintStr := string(tokenInfo.Mint)
		if existingSymbol, exists := mintAddresses[mintStr]; exists {
			return fmt.Errorf("duplicate mint address %s used by both %s and %s",
				mintStr, existingSymbol, tokenInfo.Symbol)
		}

		mintAddresses[mintStr] = tokenInfo.Symbol
	}

	return nil
}
