package hylo

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/tokens"
)

// StateReader provides functionality to read Hylo protocol state from the Solana blockchain
type StateReader struct {
	solanaClient *solana.HTTPClient
	config       *Config
}

// NewStateReader creates a new StateReader with the provided Solana HTTP client
func NewStateReader(solanaClient *solana.HTTPClient, config *Config) *StateReader {
	if config == nil {
		config = NewConfig() // Use default config if none provided
	}

	return &StateReader{
		solanaClient: solanaClient,
		config:       config,
	}
}

// ReadProtocolState reads the complete Hylo protocol state from on-chain data
// This includes token supplies, reserves, and calculates derived metrics
func (r *StateReader) ReadProtocolState(ctx context.Context, solPriceUSD float64) (*HyloProtocolState, error) {
	if solPriceUSD <= 0 {
		return nil, fmt.Errorf("SOL price must be positive, got %f", solPriceUSD)
	}

	// Read token mint information for both hyUSD and xSOL using tokens package constants
	hyusdMintInfo, err := r.readTokenMintInfo(ctx, tokens.HyUSDMint)
	if err != nil {
		return nil, fmt.Errorf("failed to read hyUSD mint info: %w", err)
	}

	xsolMintInfo, err := r.readTokenMintInfo(ctx, tokens.XSOLMint)
	if err != nil {
		return nil, fmt.Errorf("failed to read xSOL mint info: %w", err)
	}

	// TODO: Read actual protocol reserve data from Hylo program accounts
	// For now, we'll use a placeholder calculation based on token supplies
	// This needs to be implemented based on actual Hylo program state structure
	totalSOLReserve, err := r.estimateSOLReserve(ctx, hyusdMintInfo.Supply, xsolMintInfo.Supply, solPriceUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate SOL reserve: %w", err)
	}

	// Create the protocol state
	state := &HyloProtocolState{
		Timestamp:       time.Now(),
		Slot:            0, // TODO: Get actual slot from latest response
		HyUSDSupply:     hyusdMintInfo.Supply,
		XSOLSupply:      xsolMintInfo.Supply,
		HyUSDMintInfo:   *hyusdMintInfo,
		XSOLMintInfo:    *xsolMintInfo,
		TotalSOLReserve: totalSOLReserve,
		SOLPriceUSD:     solPriceUSD,
	}

	// Calculate derived metrics using the price calculator
	if err := r.calculateDerivedMetrics(state); err != nil {
		return nil, fmt.Errorf("failed to calculate derived metrics: %w", err)
	}

	// Validate the final state
	if err := state.Validate(); err != nil {
		return nil, fmt.Errorf("invalid protocol state: %w", err)
	}

	return state, nil
}

// readTokenMintInfo reads SPL token mint information for a given token mint address
func (r *StateReader) readTokenMintInfo(ctx context.Context, mintAddress solana.Address) (*SPLTokenInfo, error) {
	// Get account info for the token mint
	accountInfo, err := r.solanaClient.GetAccount(ctx, mintAddress, solana.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info for mint %s: %w", mintAddress, err)
	}

	// Decode the base64 data
	data, err := base64.StdEncoding.DecodeString(string(accountInfo.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode account data: %w", err)
	}

	// Parse SPL Token mint data
	mintInfo, err := ParseSPLTokenMintData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SPL token mint data: %w", err)
	}

	return mintInfo, nil
}

// estimateSOLReserve provides an estimated SOL reserve calculation
// TODO: Replace this with actual Hylo program state reading
// This is a placeholder implementation that estimates based on token supplies and price
func (r *StateReader) estimateSOLReserve(ctx context.Context, hyusdSupply, xsolSupply uint64, solPriceUSD float64) (uint64, error) {
	// This is a simplified estimation and should be replaced with actual program state reading
	// The actual implementation would read the Hylo program's state accounts to get the real reserve

	// Estimate based on the assumption that:
	// 1. hyUSD supply represents roughly $1 worth of backing per token
	// 2. xSOL represents additional leveraged exposure
	// 3. The protocol is overcollateralized

	// Convert hyUSD supply to SOL equivalent (hyUSD has 6 decimals)
	hyusdSupplyFloat := float64(hyusdSupply) / 1e6 // Convert to actual hyUSD amount
	solNeededForHyUSD := hyusdSupplyFloat / solPriceUSD

	// Convert xSOL supply to estimated SOL backing (xSOL has 6 decimals)
	// Assume each xSOL is backed by approximately 2-3 SOL on average (leveraged exposure)
	xsolSupplyFloat := float64(xsolSupply) / 1e6 // Convert to actual xSOL amount
	estimatedSOLForXSOL := xsolSupplyFloat * 2.5 // Conservative estimate

	// Total estimated SOL reserve with some buffer for overcollateralization
	totalEstimatedSOL := (solNeededForHyUSD + estimatedSOLForXSOL) * 1.2 // 20% buffer

	// Convert back to lamports (SOL has 9 decimals)
	totalSOLReserveLamports := uint64(totalEstimatedSOL * 1e9)

	// Sanity check: ensure we have a reasonable minimum reserve
	minimumReserve := uint64(100 * 1e9) // Minimum 100 SOL reserve
	if totalSOLReserveLamports < minimumReserve {
		totalSOLReserveLamports = minimumReserve
	}

	return totalSOLReserveLamports, nil
}

// calculateDerivedMetrics calculates NAVs, collateral ratio, and effective leverage
// Uses the Hylo equations documented in docs/01-hylo-documentation.md
func (r *StateReader) calculateDerivedMetrics(state *HyloProtocolState) error {
	// Calculate hyUSD NAV in SOL
	// Formula: hyUSD_NAV_in_SOL = 1 / SOL_Price_USD
	state.HyUSDNAVInSOL = 1.0 / state.SOLPriceUSD

	// Convert supplies to actual token amounts (accounting for decimals)
	hyusdActualSupply := float64(state.HyUSDSupply) / 1e6         // hyUSD has 6 decimals
	xsolActualSupply := float64(state.XSOLSupply) / 1e6           // xSOL has 6 decimals
	totalSOLReserveActual := float64(state.TotalSOLReserve) / 1e9 // SOL has 9 decimals (lamports)

	// Calculate xSOL NAV in SOL
	// Formula: xSOL_NAV_in_SOL = (Total SOL in Reserve - (hyUSD_NAV_in_SOL × hyUSD Supply)) / xSOL Supply
	hyusdSOLValue := state.HyUSDNAVInSOL * hyusdActualSupply
	if xsolActualSupply == 0 {
		return fmt.Errorf("xSOL supply cannot be zero for NAV calculation")
	}
	state.XSOLNAVInSOL = (totalSOLReserveActual - hyusdSOLValue) / xsolActualSupply

	// Validate that xSOL NAV is positive
	if state.XSOLNAVInSOL <= 0 {
		return fmt.Errorf("calculated xSOL NAV in SOL is not positive: %f", state.XSOLNAVInSOL)
	}

	// Calculate Collateral Ratio
	// Formula: Collateral Ratio = Total SOL in Reserve / (hyUSD NAV in SOL × hyUSD Supply)
	hyusdTotalSOLValue := state.HyUSDNAVInSOL * hyusdActualSupply
	if hyusdTotalSOLValue == 0 {
		return fmt.Errorf("hyUSD total SOL value cannot be zero for collateral ratio calculation")
	}
	state.CollateralRatio = totalSOLReserveActual / hyusdTotalSOLValue

	// Calculate Effective Leverage for xSOL
	// Formula: Effective Leverage_xSOL = Total SOL in Reserve / Market Cap_xSOL
	// Market Cap_xSOL = xSOL_NAV_in_SOL × xSOL Supply
	xsolMarketCapInSOL := state.XSOLNAVInSOL * xsolActualSupply
	if xsolMarketCapInSOL == 0 {
		return fmt.Errorf("xSOL market cap in SOL cannot be zero for effective leverage calculation")
	}
	state.EffectiveLeverage = totalSOLReserveActual / xsolMarketCapInSOL

	return nil
}

// ReadTokenSupplies reads just the token supplies for both hyUSD and xSOL
// This is a lighter weight operation compared to ReadProtocolState
func (r *StateReader) ReadTokenSupplies(ctx context.Context) (hyusdSupply, xsolSupply uint64, err error) {
	// Read hyUSD supply using tokens package constants
	hyusdMintInfo, err := r.readTokenMintInfo(ctx, tokens.HyUSDMint)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read hyUSD mint info: %w", err)
	}

	// Read xSOL supply using tokens package constants
	xsolMintInfo, err := r.readTokenMintInfo(ctx, tokens.XSOLMint)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read xSOL mint info: %w", err)
	}

	return hyusdMintInfo.Supply, xsolMintInfo.Supply, nil
}

// ValidateTokenMint checks if a token mint is properly initialized and accessible
func (r *StateReader) ValidateTokenMint(ctx context.Context, mintAddress solana.Address) error {
	mintInfo, err := r.readTokenMintInfo(ctx, mintAddress)
	if err != nil {
		return fmt.Errorf("failed to read token mint %s: %w", mintAddress, err)
	}

	if !mintInfo.IsInitialized {
		return fmt.Errorf("token mint %s is not initialized", mintAddress)
	}

	return nil
}

// ValidateAllHyloTokenMints validates that all required Hylo token mints are accessible
func (r *StateReader) ValidateAllHyloTokenMints(ctx context.Context) error {
	tokenMints := GetAllTokenMints()

	for _, mint := range tokenMints {
		if err := r.ValidateTokenMint(ctx, mint); err != nil {
			return fmt.Errorf("validation failed for %s mint: %w", tokens.GetTokenSymbol(mint), err)
		}
	}

	return nil
}

// GetHealthStatus returns a summary of the protocol health based on current state
func (r *StateReader) GetHealthStatus(ctx context.Context, solPriceUSD float64) (map[string]interface{}, error) {
	state, err := r.ReadProtocolState(ctx, solPriceUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol state: %w", err)
	}

	hyusdFormatted, xsolFormatted := state.GetFormattedSupplies()

	return map[string]interface{}{
		"healthy":            state.IsHealthy(),
		"collateral_ratio":   state.CollateralRatio,
		"effective_leverage": state.EffectiveLeverage,
		"hyusd_supply":       hyusdFormatted,
		"xsol_supply":        xsolFormatted,
		"sol_reserve":        state.GetFormattedSOLReserve(),
		"hyusd_nav_sol":      state.HyUSDNAVInSOL,
		"xsol_nav_sol":       state.XSOLNAVInSOL,
		"timestamp":          state.Timestamp,
	}, nil
}
