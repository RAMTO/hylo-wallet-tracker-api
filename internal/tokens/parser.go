package tokens

import (
	"encoding/binary"
	"fmt"
	"strings"

	"hylo-wallet-tracker-api/internal/solana"
)

// SPL Token Account Layout Constants
// Based on: https://github.com/solana-labs/solana-program-library
const (
	// Total size of SPL token account data
	SPLTokenAccountSize = 165

	// Field offsets in the 165-byte SPL token account structure
	MintOffset   = 0   // mint: Pubkey (32 bytes)
	OwnerOffset  = 32  // owner: Pubkey (32 bytes)
	AmountOffset = 64  // amount: u64 (8 bytes)
	StateOffset  = 108 // state: u8 (1 byte)
)

// Token Account State values
const (
	TokenStateUninitialized = 0
	TokenStateInitialized   = 1
	TokenStateFrozen        = 2
)

// SPLTokenAccount represents a parsed SPL token account
type SPLTokenAccount struct {
	// Mint is the token mint address (32 bytes)
	Mint solana.Address `json:"mint"`
	// Owner is the account owner address (32 bytes)
	Owner solana.Address `json:"owner"`
	// Amount is the raw token amount as stored on-chain (8 bytes, u64)
	Amount uint64 `json:"amount"`
	// State indicates the account state (1 byte)
	State uint8 `json:"state"`
	// IsInitialized indicates if the account is properly initialized
	IsInitialized bool `json:"isInitialized"`
	// IsFrozen indicates if the account is frozen
	IsFrozen bool `json:"isFrozen"`
}

// ParseSPLTokenAccount parses SPL token account data from Solana AccountInfo
// Returns the parsed token account structure or error if data is invalid
func ParseSPLTokenAccount(accountInfo *solana.AccountInfo) (*SPLTokenAccount, error) {
	// Validate account owner is SPL Token Program
	if accountInfo.Owner != SPLTokenProgramID {
		return nil, fmt.Errorf("invalid token account owner: expected %s, got %s",
			SPLTokenProgramID, accountInfo.Owner)
	}

	// Validate account data size
	if len(accountInfo.Data) != SPLTokenAccountSize {
		return nil, fmt.Errorf("invalid token account data size: expected %d bytes, got %d",
			SPLTokenAccountSize, len(accountInfo.Data))
	}

	// Extract mint address (bytes 0-31)
	mintBytes := accountInfo.Data[MintOffset : MintOffset+32]
	mint, err := bytesToAddress(mintBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mint address: %w", err)
	}
	// fmt.Printf("DEBUG: Parsed mint from account data: %s\n", mint)

	// Extract owner address (bytes 32-63)
	ownerBytes := accountInfo.Data[OwnerOffset : OwnerOffset+32]
	owner, err := bytesToAddress(ownerBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse owner address: %w", err)
	}

	// Extract amount (bytes 64-71, little-endian u64)
	amount := binary.LittleEndian.Uint64(accountInfo.Data[AmountOffset : AmountOffset+8])

	// Extract state (byte 108)
	state := accountInfo.Data[StateOffset]

	// Determine account status flags
	isInitialized := state == TokenStateInitialized || state == TokenStateFrozen
	isFrozen := state == TokenStateFrozen

	return &SPLTokenAccount{
		Mint:          mint,
		Owner:         owner,
		Amount:        amount,
		State:         state,
		IsInitialized: isInitialized,
		IsFrozen:      isFrozen,
	}, nil
}

// bytesToAddress converts 32-byte slice to base58-encoded Solana Address
func bytesToAddress(bytes []byte) (solana.Address, error) {
	if len(bytes) != 32 {
		return "", fmt.Errorf("invalid address length: expected 32 bytes, got %d", len(bytes))
	}

	// Debug: Show the raw bytes
	// fmt.Printf("DEBUG: bytesToAddress input bytes (first 8): %v\n", bytes[:8])

	// Encode bytes to base58 for Solana address
	// Using the same base58 encoding utility from ata.go
	encoded := encodeBase58(bytes)
	if encoded == "" {
		return "", fmt.Errorf("failed to encode address: invalid length")
	}

	// fmt.Printf("DEBUG: bytesToAddress result: %s\n", encoded)
	return solana.Address(encoded), nil
}

// ValidateTokenAccount performs additional validation on a parsed token account
func ValidateTokenAccount(account *SPLTokenAccount, expectedMint solana.Address, expectedOwner solana.Address) error {
	// Validate account is initialized
	if !account.IsInitialized {
		return fmt.Errorf("token account is not initialized (state: %d)", account.State)
	}

	// Validate mint matches expected
	if expectedMint != "" && account.Mint != expectedMint {
		return fmt.Errorf("mint mismatch: expected %s, got %s", expectedMint, account.Mint)
	}

	// Validate owner matches expected
	if expectedOwner != "" && account.Owner != expectedOwner {
		return fmt.Errorf("owner mismatch: expected %s, got %s", expectedOwner, account.Owner)
	}

	return nil
}

// IsZeroBalance returns true if the token account has a zero balance
func (account *SPLTokenAccount) IsZeroBalance() bool {
	return account.Amount == 0
}

// GetFormattedAmount returns the token amount formatted with decimals
func (account *SPLTokenAccount) GetFormattedAmount(decimals uint8) string {
	return formatTokenAmount(account.Amount, decimals)
}

// formatTokenAmount formats raw token amount with proper decimal precision
func formatTokenAmount(rawAmount uint64, decimals uint8) string {
	if rawAmount == 0 {
		return "0"
	}

	// Simple decimal formatting for display
	divisor := uint64(1)
	for i := uint8(0); i < decimals; i++ {
		divisor *= 10
	}

	integerPart := rawAmount / divisor
	fractionalPart := rawAmount % divisor

	if fractionalPart == 0 {
		return fmt.Sprintf("%d", integerPart)
	}

	// Format with decimals, removing trailing zeros
	fracStr := fmt.Sprintf("%0*d", decimals, fractionalPart)
	fracStr = strings.TrimRight(fracStr, "0")

	if fracStr == "" {
		return fmt.Sprintf("%d", integerPart)
	}

	return fmt.Sprintf("%d.%s", integerPart, fracStr)
}

// String returns a string representation of the token account
func (account *SPLTokenAccount) String() string {
	return fmt.Sprintf("SPLTokenAccount{Mint: %s, Owner: %s, Amount: %d, State: %d}",
		account.Mint, account.Owner, account.Amount, account.State)
}
