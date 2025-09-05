package tokens

import (
	"fmt"

	solanainternal "hylo-wallet-tracker-api/internal/solana"

	"github.com/gagliardetto/solana-go"
)

// ATA derivation constants
const (
	// PDA seed string used in ATA derivation
	PDASeeded = "ProgramDerivedAddress"

	// Maximum seed length for PDA derivation
	MaxSeedLength = 32
)

// DeriveAssociatedTokenAddress computes the Associated Token Account (ATA) address
// for a given wallet and token mint using Solana's standard derivation.
//
// This uses the official Solana Go library to ensure correct ATA derivation
// that matches wallets, explorers, and other Solana tools.
func DeriveAssociatedTokenAddress(wallet, mint solanainternal.Address) (solanainternal.Address, error) {
	// Validate inputs
	if err := wallet.Validate(); err != nil {
		return solanainternal.Address(""), fmt.Errorf("invalid wallet address: %w", err)
	}

	if err := mint.Validate(); err != nil {
		return solanainternal.Address(""), fmt.Errorf("invalid mint address: %w", err)
	}

	// Convert to Solana library types
	walletPubkey, err := solana.PublicKeyFromBase58(string(wallet))
	if err != nil {
		return solanainternal.Address(""), fmt.Errorf("failed to parse wallet address: %w", err)
	}

	mintPubkey, err := solana.PublicKeyFromBase58(string(mint))
	if err != nil {
		return solanainternal.Address(""), fmt.Errorf("failed to parse mint address: %w", err)
	}

	// Use the standard Solana library ATA derivation
	ataAddress, _, err := solana.FindAssociatedTokenAddress(walletPubkey, mintPubkey)
	if err != nil {
		return solanainternal.Address(""), fmt.Errorf("failed to derive ATA address: %w", err)
	}

	// Convert back to our internal Address type
	return solanainternal.Address(ataAddress.String()), nil
}

// GetWalletATAs computes ATA addresses for all supported Hylo tokens for the given wallet.
// Returns a map of token symbol to ATA address for efficient lookup.
//
// This function is optimized to avoid redundant computations and provides
// a single interface for getting all token ATAs for a wallet.
func GetWalletATAs(wallet solanainternal.Address, config *Config) (map[string]solanainternal.Address, error) {
	// Validate wallet address
	if err := wallet.Validate(); err != nil {
		return nil, fmt.Errorf("invalid wallet address: %w", err)
	}

	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Get all supported tokens
	supportedTokens := config.GetSupportedTokens()
	atas := make(map[string]solanainternal.Address, len(supportedTokens))

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
func GetWalletATAForToken(wallet solanainternal.Address, tokenSymbol string, config *Config) (solanainternal.Address, error) {
	// Validate wallet address
	if err := wallet.Validate(); err != nil {
		return solanainternal.Address(""), fmt.Errorf("invalid wallet address: %w", err)
	}

	if config == nil {
		return solanainternal.Address(""), fmt.Errorf("config cannot be nil")
	}

	// Get token info by symbol
	tokenInfo := config.GetTokenBySymbol(tokenSymbol)
	if tokenInfo == nil {
		return solanainternal.Address(""), fmt.Errorf("unsupported token symbol: %s", tokenSymbol)
	}

	// Derive ATA address
	return DeriveAssociatedTokenAddress(wallet, tokenInfo.Mint)
}

// validateATAAddressFormat checks if the given address has valid format.
// This is a helper function for internal use.
func validateATAAddressFormat(address solanainternal.Address) error {
	return address.Validate()
}
