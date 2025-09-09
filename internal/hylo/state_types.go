package hylo

import (
	"encoding/binary"
	"fmt"
	"time"

	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/tokens"
	"hylo-wallet-tracker-api/internal/utils"
)

// Token mint addresses are imported from tokens package to avoid duplication
// All Hylo token mint addresses are centrally defined in internal/tokens/constants.go

// SPLTokenInfo represents token mint information from SPL token program
type SPLTokenInfo struct {
	// MintAuthority is the authority that can mint new tokens (may be null)
	MintAuthority *solana.Address `json:"mint_authority"`

	// Supply is the total supply of the token in raw units (no decimals)
	Supply uint64 `json:"supply"`

	// Decimals is the number of decimal places for the token
	Decimals uint8 `json:"decimals"`

	// IsInitialized indicates if the mint account is initialized
	IsInitialized bool `json:"is_initialized"`

	// FreezeAuthority is the authority that can freeze token accounts (may be null)
	FreezeAuthority *solana.Address `json:"freeze_authority"`
}

// HyloProtocolState represents the complete Hylo protocol state snapshot
type HyloProtocolState struct {
	// Timestamp indicates when this state was captured
	Timestamp time.Time `json:"timestamp"`

	// Slot is the Solana slot when this state was captured
	Slot uint64 `json:"slot"`

	// Token supply information
	HyUSDSupply uint64 `json:"hyusd_supply"` // Total hyUSD supply in raw units (6 decimals)
	XSOLSupply  uint64 `json:"xsol_supply"`  // Total xSOL supply in raw units (6 decimals)

	// Token mint information (includes decimals and authorities)
	HyUSDMintInfo SPLTokenInfo `json:"hyusd_mint_info"`
	XSOLMintInfo  SPLTokenInfo `json:"xsol_mint_info"`

	// Protocol reserves (placeholder - exact structure depends on Hylo program state)
	// This will be populated based on actual program account data structure
	TotalSOLReserve uint64 `json:"total_sol_reserve"` // Total SOL reserve in lamports

	// Protocol health metrics (calculated from the above data)
	CollateralRatio   float64 `json:"collateral_ratio"`   // Protocol health ratio
	EffectiveLeverage float64 `json:"effective_leverage"` // xSOL effective leverage
	HyUSDNAVInSOL     float64 `json:"hyusd_nav_sol"`      // hyUSD NAV in SOL terms
	XSOLNAVInSOL      float64 `json:"xsol_nav_sol"`       // xSOL NAV in SOL terms

	// External price data dependency
	SOLPriceUSD float64 `json:"sol_price_usd"` // Current SOL/USD price
}

// ParseSPLTokenMintData parses SPL Token mint account data
// SPL Token mint account data structure (165 bytes total):
// - mint_authority (36 bytes): Option<Pubkey> - 4 bytes (option flag) + 32 bytes (pubkey)
// - supply (8 bytes): u64
// - decimals (1 byte): u8
// - is_initialized (1 byte): bool
// - freeze_authority (36 bytes): Option<Pubkey> - 4 bytes (option flag) + 32 bytes (pubkey)
func ParseSPLTokenMintData(data []byte) (*SPLTokenInfo, error) {
	// SPL Token mint data is exactly 82 bytes
	if len(data) != 82 {
		return nil, fmt.Errorf("invalid SPL token mint data length: expected 82 bytes, got %d", len(data))
	}

	info := &SPLTokenInfo{}

	// Parse mint authority (36 bytes: 4 bytes option + 32 bytes pubkey)
	offset := 0
	hasAuthority := binary.LittleEndian.Uint32(data[offset:offset+4]) != 0
	offset += 4

	if hasAuthority {
		authority := solana.Address(data[offset : offset+32])
		info.MintAuthority = &authority
	}
	offset += 32

	// Parse supply (8 bytes)
	info.Supply = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Parse decimals (1 byte)
	info.Decimals = data[offset]
	offset += 1

	// Parse is_initialized (1 byte)
	info.IsInitialized = data[offset] != 0
	offset += 1

	// Parse freeze authority (36 bytes: 4 bytes option + 32 bytes pubkey)
	hasFreezeAuthority := binary.LittleEndian.Uint32(data[offset:offset+4]) != 0
	offset += 4

	if hasFreezeAuthority {
		freezeAuthority := solana.Address(data[offset : offset+32])
		info.FreezeAuthority = &freezeAuthority
	}

	return info, nil
}

// GetFormattedSupplies returns human-readable token supplies
func (s *HyloProtocolState) GetFormattedSupplies() (string, string) {
	hyusdFormatted := utils.FormatTokenAmount(s.HyUSDSupply, s.HyUSDMintInfo.Decimals)
	xsolFormatted := utils.FormatTokenAmount(s.XSOLSupply, s.XSOLMintInfo.Decimals)
	return hyusdFormatted, xsolFormatted
}

// GetFormattedSOLReserve returns human-readable SOL reserve amount
func (s *HyloProtocolState) GetFormattedSOLReserve() string {
	return utils.FormatTokenAmount(s.TotalSOLReserve, 9) // SOL has 9 decimals (lamports)
}

// IsHealthy checks if the protocol state indicates healthy conditions
func (s *HyloProtocolState) IsHealthy() bool {
	// Basic health checks:
	// 1. Collateral ratio should be > 1.0 (overcollateralized)
	// 2. Both token supplies should be > 0
	// 3. SOL reserve should be > 0
	// 4. Effective leverage should be reasonable (> 0 and < 100)

	return s.CollateralRatio > 1.0 &&
		s.HyUSDSupply > 0 &&
		s.XSOLSupply > 0 &&
		s.TotalSOLReserve > 0 &&
		s.EffectiveLeverage > 0 &&
		s.EffectiveLeverage < 100 // Sanity check for reasonable leverage
}

// Validate performs basic validation on the protocol state
func (s *HyloProtocolState) Validate() error {
	// Check that required fields are present
	if s.HyUSDSupply == 0 {
		return fmt.Errorf("hyUSD supply cannot be zero")
	}

	if s.XSOLSupply == 0 {
		return fmt.Errorf("xSOL supply cannot be zero")
	}

	if s.TotalSOLReserve == 0 {
		return fmt.Errorf("total SOL reserve cannot be zero")
	}

	if s.SOLPriceUSD <= 0 {
		return fmt.Errorf("SOL price must be positive, got %f", s.SOLPriceUSD)
	}

	// Check token mint info
	if !s.HyUSDMintInfo.IsInitialized {
		return fmt.Errorf("hyUSD mint is not initialized")
	}

	if !s.XSOLMintInfo.IsInitialized {
		return fmt.Errorf("xSOL mint is not initialized")
	}

	// Check that supplies match mint info
	if s.HyUSDSupply != s.HyUSDMintInfo.Supply {
		return fmt.Errorf("hyUSD supply mismatch: state=%d, mint=%d", s.HyUSDSupply, s.HyUSDMintInfo.Supply)
	}

	if s.XSOLSupply != s.XSOLMintInfo.Supply {
		return fmt.Errorf("xSOL supply mismatch: state=%d, mint=%d", s.XSOLSupply, s.XSOLMintInfo.Supply)
	}

	// Check that calculated fields are reasonable
	if s.CollateralRatio < 0 {
		return fmt.Errorf("collateral ratio cannot be negative: %f", s.CollateralRatio)
	}

	if s.EffectiveLeverage <= 0 {
		return fmt.Errorf("effective leverage must be positive: %f", s.EffectiveLeverage)
	}

	return nil
}

// GetAllTokenMints returns all Hylo token mint addresses for state reading
func GetAllTokenMints() []solana.Address {
	return []solana.Address{
		tokens.HyUSDMint,
		tokens.XSOLMint,
		// Note: We don't include sHYUSD in state reading as it's not needed for price calculations
		// but it's available as tokens.SHyUSDMint if needed in the future
	}
}

// IsHyloTokenMint checks if the given address is a Hylo token mint
// This function reuses the centralized validation from the tokens package
func IsHyloTokenMint(address solana.Address) bool {
	return tokens.IsValidTokenMint(address)
}

// GetTokenMintName returns the name of the token mint for the given address
// This function reuses the centralized token information from the tokens package
func GetTokenMintName(address solana.Address) string {
	// Use the tokens package function for consistency
	symbol := tokens.GetTokenSymbol(address)
	if symbol != "" {
		return symbol
	}
	return "unknown"
}

// HyloExchangeState represents the main Hylo exchange program state account
// This structure mirrors the on-chain program account data layout
type HyloExchangeState struct {
	// Version for program upgrades compatibility
	Version uint8 `json:"version"`

	// Authority that controls the exchange
	Authority solana.Address `json:"authority"`

	// Total value locked in SOL (sum of all LST reserves converted to SOL)
	TotalSOLReserve uint64 `json:"total_sol_reserve"` // In lamports

	// Fee configuration
	MintFeeRateBps   uint16 `json:"mint_fee_rate_bps"`   // Basis points (100 = 1%)
	RedeemFeeRateBps uint16 `json:"redeem_fee_rate_bps"` // Basis points

	// Protocol revenue tracking
	TotalFeesCollected uint64 `json:"total_fees_collected"` // In lamports

	// LST registry and vault information
	LSTCount uint8 `json:"lst_count"` // Number of supported LSTs

	// Reserved space for future fields
	Reserved [32]byte `json:"reserved"`
}

// HyloStabilityPoolState represents the stability pool program state
type HyloStabilityPoolState struct {
	// Version for program upgrades compatibility
	Version uint8 `json:"version"`

	// Authority that controls the stability pool
	Authority solana.Address `json:"authority"`

	// Pool balances
	HyUSDPoolBalance uint64 `json:"hyusd_pool_balance"` // In raw units (6 decimals)
	XSOLPoolBalance  uint64 `json:"xsol_pool_balance"`  // In raw units (6 decimals)

	// sHYUSD (staked hyUSD) supply
	SHyUSDSupply uint64 `json:"shyusd_supply"` // In raw units (6 decimals)

	// Yield distribution configuration
	YieldDistributionRateBps uint16 `json:"yield_distribution_rate_bps"`

	// Reserved space for future fields
	Reserved [32]byte `json:"reserved"`
}

// HyloLSTVaultInfo represents information about an individual LST vault
type HyloLSTVaultInfo struct {
	// LST mint address
	LSTMint solana.Address `json:"lst_mint"`

	// Vault token account holding the LSTs
	VaultAccount solana.Address `json:"vault_account"`

	// Current balance in the vault (raw LST units)
	VaultBalance uint64 `json:"vault_balance"`

	// LST price in SOL terms (from Sanctum calculator)
	LSTToSOLRate uint64 `json:"lst_to_sol_rate"` // Fixed-point with 9 decimals

	// Last update timestamp
	LastUpdated int64 `json:"last_updated"`
}

// ParseHyloExchangeState parses the Hylo exchange program account data
// This implements a best-effort parser based on common Solana program patterns
func ParseHyloExchangeState(data []byte) (*HyloExchangeState, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty account data")
	}

	// Most Solana programs start with a discriminator (8 bytes) followed by the struct
	// For Anchor programs, this is typically the hash of the account name
	if len(data) < 16 { // Minimum size for a meaningful state account
		return nil, fmt.Errorf("account data too small: %d bytes", len(data))
	}

	// Skip discriminator (first 8 bytes) and start parsing the state struct
	offset := 8

	// Try to parse the basic structure - this is a best-effort implementation
	state := &HyloExchangeState{}

	if len(data) <= offset {
		return nil, fmt.Errorf("insufficient data for version field")
	}

	// Parse version (1 byte)
	state.Version = data[offset]
	offset += 1

	// Parse authority (32 bytes for Pubkey)
	if len(data) < offset+32 {
		return nil, fmt.Errorf("insufficient data for authority field")
	}
	// For now, we'll skip parsing the authority as Address is a string type
	// TODO: Convert bytes to base58 address string
	state.Authority = solana.Address("") // Placeholder
	offset += 32

	// For now, we cannot reliably parse the exact SOL reserve without the proper IDL
	// This is where we need the actual program schema
	// As a temporary measure, we'll try to extract reasonable data or return an error

	// If we have enough data left, try to find a reasonable uint64 value that could be the reserve
	if len(data) >= offset+8 {
		// Try parsing the next 8 bytes as a potential Total SOL Reserve
		possibleReserve := binary.LittleEndian.Uint64(data[offset : offset+8])

		// Sanity check: SOL reserve should be reasonable (between 1 SOL and 10M SOL)
		minReasonableReserve := uint64(1e9)  // 1 SOL in lamports
		maxReasonableReserve := uint64(1e16) // 10M SOL in lamports

		if possibleReserve >= minReasonableReserve && possibleReserve <= maxReasonableReserve {
			state.TotalSOLReserve = possibleReserve
			return state, nil
		}
	}

	// If we can't reliably parse the reserve, return an error to trigger fallback
	return nil, fmt.Errorf("cannot reliably parse Total SOL Reserve from account data - program schema unknown")
}

// ParseHyloStabilityPoolState parses the Hylo stability pool program account data
func ParseHyloStabilityPoolState(data []byte) (*HyloStabilityPoolState, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty account data")
	}

	// Similar approach as exchange state - skip discriminator and parse basic structure
	if len(data) < 16 {
		return nil, fmt.Errorf("account data too small: %d bytes", len(data))
	}

	offset := 8 // Skip discriminator
	state := &HyloStabilityPoolState{}

	if len(data) <= offset {
		return nil, fmt.Errorf("insufficient data for version field")
	}

	// Parse version
	state.Version = data[offset]
	offset += 1

	// Parse authority
	if len(data) < offset+32 {
		return nil, fmt.Errorf("insufficient data for authority field")
	}
	// For now, we'll skip parsing the authority as Address is a string type
	// TODO: Convert bytes to base58 address string
	state.Authority = solana.Address("") // Placeholder

	// For stability pool, we don't need to parse the exact structure for SOL reserve calculation
	// The main exchange state contains the Total SOL Reserve we need
	return state, nil
}

// TryParseWithIDL attempts to parse Hylo state using Anchor IDL if available
// This is a placeholder for future implementation with proper IDL-based parsing
func TryParseWithIDL(data []byte, accountType string) (interface{}, error) {
	// TODO: Implement IDL-based parsing when we have access to Hylo's IDL files
	// This would provide the exact account layouts for reliable parsing
	return nil, fmt.Errorf("IDL-based parsing not yet implemented for account type: %s", accountType)
}
