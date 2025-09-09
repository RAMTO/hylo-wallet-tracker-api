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
