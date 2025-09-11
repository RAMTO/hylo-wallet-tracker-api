package tokens

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"time"

	"hylo-wallet-tracker-api/internal/logger"
	solanainternal "hylo-wallet-tracker-api/internal/solana"

	"github.com/gagliardetto/solana-go"
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
	Mint solanainternal.Address `json:"mint"`
	// Owner is the account owner address (32 bytes)
	Owner solanainternal.Address `json:"owner"`
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
func ParseSPLTokenAccount(accountInfo *solanainternal.AccountInfo) (*SPLTokenAccount, error) {
	return ParseSPLTokenAccountWithContext(context.Background(), accountInfo, nil)
}

// ParseSPLTokenAccountWithContext parses SPL token account data with logging context
func ParseSPLTokenAccountWithContext(ctx context.Context, accountInfo *solanainternal.AccountInfo, log *logger.Logger) (*SPLTokenAccount, error) {
	startTime := time.Now()

	// Use default logger if none provided
	if log == nil {
		log = logger.NewFromEnv().WithComponent("token-parser")
	}

	log.DebugContext(ctx, "Starting SPL token account parsing",
		slog.String("owner", string(accountInfo.Owner)),
		slog.Int("data_size", len(accountInfo.Data)))

	// Validate account owner is SPL Token Program
	if accountInfo.Owner != SPLTokenProgramID {
		err := fmt.Errorf("invalid token account owner: expected %s, got %s",
			SPLTokenProgramID, accountInfo.Owner)
		log.LogParsingError(ctx, "parse_spl_token_account", "account_owner", err,
			slog.String("expected_owner", string(SPLTokenProgramID)),
			slog.String("actual_owner", string(accountInfo.Owner)))
		return nil, err
	}

	// Validate account data size
	if len(accountInfo.Data) != SPLTokenAccountSize {
		err := fmt.Errorf("invalid token account data size: expected %d bytes, got %d",
			SPLTokenAccountSize, len(accountInfo.Data))
		log.LogParsingError(ctx, "parse_spl_token_account", "data_size", err,
			slog.Int("expected_size", SPLTokenAccountSize),
			slog.Int("actual_size", len(accountInfo.Data)))
		return nil, err
	}

	// Extract mint address (bytes 0-31)
	mintBytes := accountInfo.Data[MintOffset : MintOffset+32]
	mint, err := bytesToAddressWithLogging(ctx, mintBytes, log, "mint")
	if err != nil {
		log.LogParsingError(ctx, "parse_spl_token_account", "mint_address", err)
		return nil, fmt.Errorf("failed to parse mint address: %w", err)
	}

	// Extract owner address (bytes 32-63)
	ownerBytes := accountInfo.Data[OwnerOffset : OwnerOffset+32]
	owner, err := bytesToAddressWithLogging(ctx, ownerBytes, log, "owner")
	if err != nil {
		log.LogParsingError(ctx, "parse_spl_token_account", "owner_address", err)
		return nil, fmt.Errorf("failed to parse owner address: %w", err)
	}

	// Extract amount (bytes 64-71, little-endian u64)
	amount := binary.LittleEndian.Uint64(accountInfo.Data[AmountOffset : AmountOffset+8])

	// Extract state (byte 108)
	state := accountInfo.Data[StateOffset]

	// Determine account status flags
	isInitialized := state == TokenStateInitialized || state == TokenStateFrozen
	isFrozen := state == TokenStateFrozen

	account := &SPLTokenAccount{
		Mint:          mint,
		Owner:         owner,
		Amount:        amount,
		State:         state,
		IsInitialized: isInitialized,
		IsFrozen:      isFrozen,
	}

	// Log successful parsing
	log.InfoContext(ctx, "Successfully parsed SPL token account",
		slog.String("mint", mint.String()),
		slog.String("owner", owner.String()),
		slog.Uint64("amount", amount),
		slog.Int("state", int(state)),
		slog.Bool("is_initialized", isInitialized),
		slog.Bool("is_frozen", isFrozen),
		slog.Duration("parse_time", time.Since(startTime)))

	return account, nil
}

// bytesToAddress converts 32-byte slice to base58-encoded Solana Address
func bytesToAddress(bytes []byte) (solanainternal.Address, error) {
	if len(bytes) != 32 {
		return "", fmt.Errorf("invalid address length: expected 32 bytes, got %d", len(bytes))
	}

	// Convert bytes to Solana PublicKey and then to string
	pubkey := solana.PublicKeyFromBytes(bytes)
	encoded := pubkey.String()

	return solanainternal.Address(encoded), nil
}

// bytesToAddressWithLogging wraps bytesToAddress with logging
func bytesToAddressWithLogging(ctx context.Context, bytes []byte, log *logger.Logger, addressType string) (solanainternal.Address, error) {
	log.DebugContext(ctx, "Converting bytes to address",
		slog.String("address_type", addressType),
		slog.Int("byte_length", len(bytes)))

	if len(bytes) != 32 {
		err := fmt.Errorf("invalid address length: expected 32 bytes, got %d", len(bytes))
		log.LogParsingError(ctx, "bytes_to_address", addressType, err,
			slog.Int("expected_length", 32),
			slog.Int("actual_length", len(bytes)))
		return "", err
	}

	address, err := bytesToAddress(bytes)
	if err != nil {
		log.LogParsingError(ctx, "bytes_to_address", addressType, err)
		return "", err
	}

	log.DebugContext(ctx, "Successfully converted bytes to address",
		slog.String("address_type", addressType),
		slog.String("address", address.String()))

	return address, nil
}

// ValidateTokenAccount performs additional validation on a parsed token account
func ValidateTokenAccount(account *SPLTokenAccount, expectedMint solanainternal.Address, expectedOwner solanainternal.Address) error {
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

// String returns a string representation of the token account
func (account *SPLTokenAccount) String() string {
	return fmt.Sprintf("SPLTokenAccount{Mint: %s, Owner: %s, Amount: %d, State: %d}",
		account.Mint, account.Owner, account.Amount, account.State)
}
