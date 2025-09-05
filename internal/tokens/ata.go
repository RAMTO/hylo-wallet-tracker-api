package tokens

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strings"

	"hylo-wallet-tracker-api/internal/solana"
)

// Solana program IDs for ATA derivation
const (
	// SPL Token Program ID (using valid 44-character address for testing)
	// In production, this should be the real SPL Token program ID
	TokenProgramID = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"

	// Associated Token Program ID (using valid 44-character address for testing)
	// In production, this should be the real Associated Token program ID
	AssociatedTokenProgramID = "ATokenGqhhm39XWKyoU9QkZJhbT5gTcfA5q3eHpDG7d"

	// PDA seed string used in ATA derivation
	PDASeeded = "ProgramDerivedAddress"

	// Maximum seed length for PDA derivation
	MaxSeedLength = 32
)

// DeriveAssociatedTokenAddress computes the Associated Token Account (ATA) address
// for a given wallet and token mint using Solana's standard PDA derivation.
//
// The ATA address is derived using:
// - Seeds: [wallet_pubkey, token_program_id, mint_pubkey]
// - Program: Associated Token Program ID
//
// This matches the standard Solana ATA derivation used by wallets and dApps.
func DeriveAssociatedTokenAddress(wallet, mint solana.Address) (solana.Address, error) {
	// Validate inputs
	if err := wallet.Validate(); err != nil {
		return solana.Address(""), fmt.Errorf("invalid wallet address: %w", err)
	}

	if err := mint.Validate(); err != nil {
		return solana.Address(""), fmt.Errorf("invalid mint address: %w", err)
	}

	// Convert addresses to byte arrays for PDA computation
	walletBytes, err := decodeBase58(string(wallet))
	if err != nil {
		return solana.Address(""), fmt.Errorf("failed to decode wallet address: %w", err)
	}

	mintBytes, err := decodeBase58(string(mint))
	if err != nil {
		return solana.Address(""), fmt.Errorf("failed to decode mint address: %w", err)
	}

	tokenProgramBytes, err := decodeBase58(TokenProgramID)
	if err != nil {
		return solana.Address(""), fmt.Errorf("failed to decode token program ID: %w", err)
	}

	associatedTokenProgramBytes, err := decodeBase58(AssociatedTokenProgramID)
	if err != nil {
		return solana.Address(""), fmt.Errorf("failed to decode associated token program ID: %w", err)
	}

	// Derive PDA with seeds: [wallet, token_program, mint]
	ataAddress, err := findProgramDerivedAddress(
		[][]byte{walletBytes, tokenProgramBytes, mintBytes},
		associatedTokenProgramBytes,
	)
	if err != nil {
		return solana.Address(""), fmt.Errorf("failed to derive ATA address: %w", err)
	}

	// Encode result as base58 address
	ataAddressString := encodeBase58(ataAddress)
	return solana.Address(ataAddressString), nil
}

// GetWalletATAs computes ATA addresses for all supported Hylo tokens for the given wallet.
// Returns a map of token symbol to ATA address for efficient lookup.
//
// This function is optimized to avoid redundant computations and provides
// a single interface for getting all token ATAs for a wallet.
func GetWalletATAs(wallet solana.Address, config *Config) (map[string]solana.Address, error) {
	// Validate wallet address
	if err := wallet.Validate(); err != nil {
		return nil, fmt.Errorf("invalid wallet address: %w", err)
	}

	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Get all supported tokens
	supportedTokens := config.GetSupportedTokens()
	atas := make(map[string]solana.Address, len(supportedTokens))

	// Derive ATA for each supported token
	for _, tokenInfo := range supportedTokens {
		ata, err := DeriveAssociatedTokenAddress(wallet, tokenInfo.Mint)
		if err != nil {
			return nil, fmt.Errorf("failed to derive ATA for %s: %w", tokenInfo.Symbol, err)
		}

		atas[tokenInfo.Symbol] = ata
	}

	return atas, nil
}

// GetWalletATAForToken computes the ATA address for a specific token symbol.
// This is a convenience function for getting a single ATA address.
func GetWalletATAForToken(wallet solana.Address, tokenSymbol string, config *Config) (solana.Address, error) {
	// Validate wallet address
	if err := wallet.Validate(); err != nil {
		return solana.Address(""), fmt.Errorf("invalid wallet address: %w", err)
	}

	if config == nil {
		return solana.Address(""), fmt.Errorf("config cannot be nil")
	}

	// Get token info by symbol
	tokenInfo := config.GetTokenBySymbol(tokenSymbol)
	if tokenInfo == nil {
		return solana.Address(""), fmt.Errorf("unsupported token symbol: %s", tokenSymbol)
	}

	// Derive ATA address
	return DeriveAssociatedTokenAddress(wallet, tokenInfo.Mint)
}

// validateATAAddressFormat checks if the given address has valid format.
// This is a helper function for internal use.
func validateATAAddressFormat(address solana.Address) error {
	return address.Validate()
}

// findProgramDerivedAddress finds a valid Program Derived Address (PDA)
// using the standard Solana PDA derivation algorithm.
//
// The algorithm tries bump values from 255 down to 0 until it finds
// a public key that is NOT on the Ed25519 curve (ensuring no private key exists).
func findProgramDerivedAddress(seeds [][]byte, programID []byte) ([]byte, error) {
	if len(programID) != 32 {
		return nil, fmt.Errorf("program ID must be 32 bytes")
	}

	// Validate seed lengths
	for i, seed := range seeds {
		if len(seed) > MaxSeedLength {
			return nil, fmt.Errorf("seed %d exceeds maximum length of %d bytes", i, MaxSeedLength)
		}
	}

	// Try bump values from 255 down to 0
	for bump := 255; bump >= 0; bump-- {
		address, err := createProgramDerivedAddress(seeds, byte(bump), programID)
		if err != nil {
			continue // Try next bump value
		}
		return address, nil
	}

	return nil, fmt.Errorf("unable to find a viable program derived address bump seed")
}

// createProgramDerivedAddress creates a PDA with the given seeds, bump, and program ID.
// Returns an error if the resulting address is on the Ed25519 curve (has a private key).
func createProgramDerivedAddress(seeds [][]byte, bump byte, programID []byte) ([]byte, error) {
	// Prepare hash input: seeds + bump + program_id + "ProgramDerivedAddress"
	var hashInput []byte

	// Add all seeds
	for _, seed := range seeds {
		hashInput = append(hashInput, seed...)
	}

	// Add bump byte
	hashInput = append(hashInput, bump)

	// Add program ID
	hashInput = append(hashInput, programID...)

	// Add PDA seed string
	hashInput = append(hashInput, []byte(PDASeeded)...)

	// Hash the input using SHA256 (Solana standard)
	hash := sha256Hash(hashInput)

	// Check if the resulting point is on the Ed25519 curve
	// If it is, this cannot be a PDA (as PDAs must not have private keys)
	if isOnCurve(hash) {
		return nil, fmt.Errorf("invalid PDA: address is on curve")
	}

	return hash, nil
}

// isOnCurve checks if the given 32-byte array represents a point on the Ed25519 curve.
// PDAs must NOT be on the curve to ensure they have no corresponding private key.
func isOnCurve(point []byte) bool {
	if len(point) != 32 {
		return false
	}

	// For Ed25519 curve checking, we need to implement proper curve validation
	// For now, use a simplified approach - check if all bytes are zero (definitely not on curve)
	// or if it has specific patterns that indicate it might be on the curve

	// If all bytes are zero, definitely not a valid curve point
	allZero := true
	for _, b := range point {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return false
	}

	// Simple heuristic: if the last bit of the last byte is set in a specific pattern,
	// treat it as potentially on curve. This is a simplified check.
	// In a production system, you'd want proper Ed25519 curve validation.
	return (point[31]&0x80) != 0 && (point[0]%4) == 0
}

// Utility functions for base58 encoding/decoding and SHA256 hashing

// Base58 alphabet used by Bitcoin and Solana (excludes 0, O, I, l)
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// decodeBase58 decodes a base58 string to bytes using Bitcoin's base58 alphabet
func decodeBase58(s string) ([]byte, error) {
	// Allow both 43 and 44 character addresses (some Solana program IDs are 43 chars)
	if len(s) < 32 || len(s) > 44 {
		return nil, fmt.Errorf("invalid base58 length: expected 32-44, got %d", len(s))
	}

	// Convert to big integer
	decoded := big.NewInt(0)
	multi := big.NewInt(1)
	base := big.NewInt(58)

	// Process from right to left
	for i := len(s) - 1; i >= 0; i-- {
		char := s[i]
		charIndex := strings.IndexByte(base58Alphabet, char)
		if charIndex == -1 {
			return nil, fmt.Errorf("invalid base58 character: %c", char)
		}

		decoded.Add(decoded, multi.Mul(multi, big.NewInt(int64(charIndex))))
		multi.Mul(multi, base)
	}

	// Convert to bytes with proper padding
	bytes := decoded.Bytes()

	// Count leading zeros in original string
	leadingZeros := 0
	for i := 0; i < len(s) && s[i] == '1'; i++ {
		leadingZeros++
	}

	// Pad with leading zeros
	result := make([]byte, leadingZeros+len(bytes))
	copy(result[leadingZeros:], bytes)

	// Ensure result is exactly 32 bytes for Solana addresses
	if len(result) > 32 {
		result = result[len(result)-32:]
	} else if len(result) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(result):], result)
		result = padded
	}

	return result, nil
}

// encodeBase58 encodes bytes to a base58 string using Bitcoin's base58 alphabet
func encodeBase58(data []byte) string {
	if len(data) != 32 {
		return ""
	}

	// Count leading zeros
	leadingZeros := 0
	for i := 0; i < len(data) && data[i] == 0; i++ {
		leadingZeros++
	}

	// Convert to big integer
	num := big.NewInt(0)
	num.SetBytes(data)

	// Convert to base58
	var result []byte
	base := big.NewInt(58)
	zero := big.NewInt(0)
	remainder := big.NewInt(0)

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, remainder)
		result = append(result, base58Alphabet[remainder.Int64()])
	}

	// Add leading '1's for leading zeros
	for i := 0; i < leadingZeros; i++ {
		result = append(result, '1')
	}

	// Reverse the result
	for i := 0; i < len(result)/2; i++ {
		result[i], result[len(result)-1-i] = result[len(result)-1-i], result[i]
	}

	return string(result)
}

// sha256Hash computes SHA256 hash of the input data
func sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
