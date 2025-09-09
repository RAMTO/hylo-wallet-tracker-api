package hylo

import "hylo-wallet-tracker-api/internal/solana"

// Hylo Protocol Program Constants
// These addresses are for Solana mainnet-beta as specified in Hylo documentation

const (
	// ExchangeProgramID is the main Hylo exchange program for xSOL mint/redeem operations
	// Source: docs/01-hylo-documentation.md - Exchange v0.1 program address
	ExchangeProgramID = "HYEXCHtHkBagdStcJCp3xbbb9B7sdMdWXFNj6mdsG4hn"

	// StabilityPoolProgramID is the Hylo stability pool program for sHYUSD operations
	// Source: docs/01-hylo-documentation.md - Stability Pool v0.1 program address
	StabilityPoolProgramID = "HysTabVUfmQBFcmzu1ctRd1Y1fxd66RBpboy1bmtDSQQ"
)

// Trade Instruction Constants
// These instruction names are used to identify xSOL trades in transaction parsing
const (
	// MintLeverCoinInstruction identifies BUY xSOL operations
	// When users mint xSOL (levercoin), they are buying leveraged SOL exposure
	// Source: mint_levercoin function in exchange_client.rs
	MintLeverCoinInstruction = "mint_levercoin"

	// RedeemLeverCoinInstruction identifies SELL xSOL operations
	// When users redeem xSOL (levercoin), they are selling their leveraged SOL position
	// Source: redeem_levercoin function in exchange_client.rs
	RedeemLeverCoinInstruction = "redeem_levercoin"
)

// Trade Side Constants for consistent classification
const (
	TradeSideBuy  = "BUY"  // User mints xSOL (acquires leveraged SOL exposure)
	TradeSideSell = "SELL" // User redeems xSOL (closes leveraged SOL position)
)

// Hylo Program IDs as solana.Address types for consistency with existing codebase
var (
	// ExchangeProgram represents the main Hylo exchange program address
	// Used for identifying transactions that involve xSOL mint/redeem operations
	ExchangeProgram = solana.Address(ExchangeProgramID)

	// StabilityPoolProgram represents the Hylo stability pool program address
	// Used for identifying transactions that involve sHYUSD staking operations
	StabilityPoolProgram = solana.Address(StabilityPoolProgramID)
)

// GetSupportedHyloPrograms returns all Hylo program addresses for transaction filtering
// Used to identify whether a transaction involves Hylo protocol operations
func GetSupportedHyloPrograms() []solana.Address {
	return []solana.Address{
		ExchangeProgram,
		StabilityPoolProgram,
	}
}

// IsHyloProgram checks if the given address is a supported Hylo program
// Used in transaction parsing to filter for Hylo-related transactions only
func IsHyloProgram(address solana.Address) bool {
	supportedPrograms := GetSupportedHyloPrograms()
	for _, program := range supportedPrograms {
		if address == program {
			return true
		}
	}
	return false
}

// IsXSOLTradeInstruction checks if the instruction name represents an xSOL trade
// Used in transaction parsing to identify mint/redeem operations specifically
func IsXSOLTradeInstruction(instructionName string) bool {
	return instructionName == MintLeverCoinInstruction || instructionName == RedeemLeverCoinInstruction
}

// GetTradeSideFromInstruction returns the trade side based on instruction name
// Returns TradeSideBuy for mint operations, TradeSideSell for redeem operations
func GetTradeSideFromInstruction(instructionName string) string {
	switch instructionName {
	case MintLeverCoinInstruction:
		return TradeSideBuy
	case RedeemLeverCoinInstruction:
		return TradeSideSell
	default:
		return ""
	}
}

// Hylo Protocol State Account Seeds
// These are used to derive Program Derived Addresses (PDAs) for protocol state accounts
const (
	// HyloStateSeed is used to derive the main Hylo protocol state account
	// This account contains the core protocol parameters and Total SOL Reserve
	HyloStateSeed = "hylo"

	// StabilityPoolConfigSeed is used to derive the stability pool configuration account
	StabilityPoolConfigSeed = "pool_config"

	// LSTRegistrySeed is used to derive the LST registry account
	LSTRegistrySeed = "lst_registry"
)

// GetHyloStateAddress returns the main Hylo protocol state account address
// This account contains the actual Total SOL Reserve that we need for price calculations
// For now, we use a placeholder address - this should be replaced with actual PDA derivation
func GetHyloStateAddress(programID solana.Address) solana.Address {
	// TODO: Implement proper PDA derivation using crypto/hash functions
	// For now, return the program ID as a placeholder
	// In production, this should derive the actual PDA with seeds
	return programID
}

// GetStabilityPoolConfigAddress returns the stability pool configuration account address
func GetStabilityPoolConfigAddress(programID solana.Address) solana.Address {
	// TODO: Implement proper PDA derivation
	return programID
}

// GetLSTRegistryAddress returns the LST registry account address
func GetLSTRegistryAddress(programID solana.Address) solana.Address {
	// TODO: Implement proper PDA derivation
	return programID
}
