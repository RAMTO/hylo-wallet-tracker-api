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
	HyUSDDecimals   = 6 // Standard stablecoin precision (6 decimals)
	SHyUSDDecimals  = 6 // Staked hyUSD shares (same as hyUSD)
	XSOLDecimals    = 6 // xSOL token precision (6 decimals, same as other Hylo tokens)
	USDCDecimals    = 6 // USDC standard precision (6 decimals)
	JitoSOLDecimals = 9 // jitoSOL liquid staking token precision (9 decimals, same as SOL)
	SOLDecimals     = 9 // SOL native token precision (9 decimals - lamports)

	// Token Symbols for display and identification
	HyUSDSymbol   = "hyUSD"
	SHyUSDSymbol  = "sHYUSD"
	XSOLSymbol    = "xSOL"
	USDCSymbol    = "USDC"
	JitoSOLSymbol = "jitoSOL"

	// Token Display Names for user interfaces
	HyUSDName   = "Hylo USD Stablecoin"
	SHyUSDName  = "Staked Hylo USD"
	XSOLName    = "Leveraged SOL Token"
	USDCName    = "USD Coin"
	JitoSOLName = "Jito Staked SOL"
)

// Token Mint Addresses (Solana mainnet-beta)
// Hylo Protocol tokens and supported external tokens
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

	// USDCMint is the mint address for USDC (USD Coin) stablecoin
	// Official USD-pegged stablecoin by Centre/Coinbase
	// Source: https://solscan.io/token/EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
	USDCMint = solana.Address("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")

	// JitoSOLMint is the mint address for jitoSOL (Jito Staked SOL)
	// Liquid staking derivative that includes MEV rewards
	// Source: https://solscan.io/token/J1toso1uCk3RLmjorhTtrVwY9HJ7X8V9yYac6Y7kGCPn
	JitoSOLMint = solana.Address("J1toso1uCk3RLmjorhTtrVwY9HJ7X8V9yYac6Y7kGCPn")
)

// GetSupportedTokenMints returns all supported token mint addresses
// Used for validation and iteration over all supported tokens
func GetSupportedTokenMints() []solana.Address {
	return []solana.Address{
		HyUSDMint,
		SHyUSDMint,
		XSOLMint,
		USDCMint,
		JitoSOLMint,
	}
}

// IsValidTokenMint checks if the given address is a supported token mint
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
	case USDCMint:
		return USDCSymbol
	case JitoSOLMint:
		return JitoSOLSymbol
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
	case USDCMint:
		return USDCDecimals
	case JitoSOLMint:
		return JitoSOLDecimals
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
	case USDCMint:
		return USDCName
	case JitoSOLMint:
		return JitoSOLName
	default:
		return ""
	}
}
