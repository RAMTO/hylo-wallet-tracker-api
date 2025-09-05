package tokens

import "hylo-wallet-tracker-api/internal/solana"

// Hylo Protocol Token Constants
// These addresses are for Solana mainnet-beta as specified in Hylo documentation

const (
	// Solana Program IDs
	// SPLTokenProgramID is the official SPL Token Program on Solana mainnet-beta
	// This program manages all SPL token accounts and operations
	SPLTokenProgramID = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"

	// AssociatedTokenProgramID is the official Associated Token Program on Solana
	// This program manages Associated Token Account (ATA) creation and management
	AssociatedTokenProgramID = "ATokenGqhhm39XWKyoU9QkZJhbT5gTcfA5q3eHpDG7d"

	// SystemProgramID is the system program that manages account creation and transfers
	SystemProgramID = "11111111111111111111111111111111"

	// TokenDecimals defines the decimal precision for each token
	// Based on Solana SPL token standards and Hylo protocol specifications
	HyUSDDecimals  = 6 // Standard stablecoin precision (6 decimals)
	SHyUSDDecimals = 6 // Staked hyUSD shares (same as hyUSD)
	XSOLDecimals   = 9 // SOL-based token precision (9 decimals like SOL)

	// Token Symbols for display and identification
	HyUSDSymbol  = "hyUSD"
	SHyUSDSymbol = "sHYUSD"
	XSOLSymbol   = "xSOL"

	// Token Display Names for user interfaces
	HyUSDName  = "Hylo USD Stablecoin"
	SHyUSDName = "Staked Hylo USD"
	XSOLName   = "Leveraged SOL Token"
)

// Hylo Token Mint Addresses (Solana mainnet-beta)
// Source: https://hylo.so documentation - Token Mints section
var (
	// HyUSDMint is the mint address for hyUSD stablecoin
	// Always pegged to 1 USD, backed by SOL LST collateral pool
	HyUSDMint = solana.Address("5YMkXAYccHSGnHn9nob9xEvv6Pvka9DZWH7nTbotTu9E")

	// SHyUSDMint is the mint address for sHYUSD (staked hyUSD)
	// LP token representing staked hyUSD in the stability pool
	// Earns yield from LST staking rewards distributed to stability pool
	SHyUSDMint = solana.Address("HnnGv3HrSqjRpgdFmx7vQGjntNEoex1SU4e9Lxcxuihz")

	// XSOLMint is the mint address for xSOL (leveraged SOL exposure)
	// Price = (Collateral TVL - hyUSD Supply) / xSOL Supply
	// Absorbs SOL price volatility to maintain hyUSD 1:1 USD peg
	XSOLMint = solana.Address("4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs")
)

// GetSupportedTokenMints returns all supported Hylo token mint addresses
// Used for validation and iteration over all supported tokens
func GetSupportedTokenMints() []solana.Address {
	return []solana.Address{
		HyUSDMint,
		SHyUSDMint,
		XSOLMint,
	}
}

// IsValidTokenMint checks if the given address is a supported Hylo token mint
func IsValidTokenMint(mint solana.Address) bool {
	supportedMints := GetSupportedTokenMints()
	for _, supportedMint := range supportedMints {
		if mint == supportedMint {
			return true
		}
	}
	return false
}

// GetTokenSymbol returns the symbol for a given token mint address
// Returns empty string if the mint is not supported
func GetTokenSymbol(mint solana.Address) string {
	switch mint {
	case HyUSDMint:
		return HyUSDSymbol
	case SHyUSDMint:
		return SHyUSDSymbol
	case XSOLMint:
		return XSOLSymbol
	default:
		return ""
	}
}

// GetTokenDecimals returns the decimal precision for a given token mint
// Returns 0 if the mint is not supported
func GetTokenDecimals(mint solana.Address) uint8 {
	switch mint {
	case HyUSDMint:
		return HyUSDDecimals
	case SHyUSDMint:
		return SHyUSDDecimals
	case XSOLMint:
		return XSOLDecimals
	default:
		return 0
	}
}

// GetTokenName returns the display name for a given token mint address
// Returns empty string if the mint is not supported
func GetTokenName(mint solana.Address) string {
	switch mint {
	case HyUSDMint:
		return HyUSDName
	case SHyUSDMint:
		return SHyUSDName
	case XSOLMint:
		return XSOLName
	default:
		return ""
	}
}
