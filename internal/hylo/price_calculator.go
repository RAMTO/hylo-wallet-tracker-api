package hylo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"hylo-wallet-tracker-api/internal/price"
)

// PriceCalculator computes xSOL prices using Hylo protocol equations
// and integrates with SOL/USD price data from the price package
type PriceCalculator struct {
	stateReader *StateReader
}

// NewPriceCalculator creates a new PriceCalculator with the provided StateReader
func NewPriceCalculator(stateReader *StateReader) *PriceCalculator {
	return &PriceCalculator{
		stateReader: stateReader,
	}
}

// CalculateXSOLPrice computes the xSOL price in both SOL and USD terms
// using the current protocol state and provided SOL/USD price
func (calc *PriceCalculator) CalculateXSOLPrice(protocolState *HyloProtocolState, solPriceUSD float64) (*price.XSOLPrice, error) {
	if protocolState == nil {
		return nil, fmt.Errorf("protocol state cannot be nil")
	}

	if solPriceUSD <= 0 {
		return nil, fmt.Errorf("SOL price must be positive, got %f", solPriceUSD)
	}

	// Ensure the protocol state includes the same SOL price for consistency
	if protocolState.SOLPriceUSD != solPriceUSD {
		// Update protocol state with new SOL price and recalculate derived metrics
		protocolState.SOLPriceUSD = solPriceUSD
		if err := calc.stateReader.calculateDerivedMetrics(protocolState); err != nil {
			return nil, fmt.Errorf("failed to recalculate derived metrics: %w", err)
		}
	}

	// The xSOL NAV in SOL is already calculated in the protocol state
	// using the formula: xSOL_NAV_in_SOL = (Total SOL in Reserve - (hyUSD_NAV_in_SOL × hyUSD Supply)) / xSOL Supply
	xsolPriceInSOL := protocolState.XSOLNAVInSOL

	// Calculate xSOL price in USD
	// Formula: xSOL_Price_USD = xSOL_NAV_in_SOL × SOL_Price_USD
	xsolPriceInUSD := xsolPriceInSOL * solPriceUSD

	// Validate calculated prices
	if xsolPriceInSOL <= 0 {
		return nil, fmt.Errorf("calculated xSOL price in SOL is not positive: %f", xsolPriceInSOL)
	}

	if xsolPriceInUSD <= 0 {
		return nil, fmt.Errorf("calculated xSOL price in USD is not positive: %f", xsolPriceInUSD)
	}

	return &price.XSOLPrice{
		PriceInSOL:        xsolPriceInSOL,
		PriceInUSD:        xsolPriceInUSD,
		Timestamp:         time.Now(),
		CollateralRatio:   protocolState.CollateralRatio,
		EffectiveLeverage: protocolState.EffectiveLeverage,
	}, nil
}

// CalculateXSOLPriceWithSOLPrice computes xSOL price by first reading protocol state
// and then calculating the price using the provided SOL/USD price
func (calc *PriceCalculator) CalculateXSOLPriceWithSOLPrice(ctx context.Context, solUSDPrice *price.SOLUSDPrice) (*price.XSOLPrice, error) {
	if solUSDPrice == nil {
		return nil, fmt.Errorf("SOL/USD price cannot be nil")
	}

	if !solUSDPrice.IsValid(0, 1000000) { // Very wide bounds for safety
		return nil, fmt.Errorf("SOL/USD price is invalid")
	}

	// Read current protocol state
	protocolState, err := calc.stateReader.ReadProtocolState(ctx, solUSDPrice.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol state: %w", err)
	}

	// Calculate xSOL price
	xsolPrice, err := calc.CalculateXSOLPrice(protocolState, solUSDPrice.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate xSOL price: %w", err)
	}

	return xsolPrice, nil
}

// CalculateCombinedPriceResponse creates a complete price response combining SOL/USD and xSOL prices
// This matches the expected response format for the /price API endpoint
func (calc *PriceCalculator) CalculateCombinedPriceResponse(ctx context.Context, solUSDPrice *price.SOLUSDPrice) (*price.CombinedPriceResponse, error) {
	if solUSDPrice == nil {
		return nil, fmt.Errorf("SOL/USD price cannot be nil")
	}

	// Calculate xSOL price using the SOL/USD price
	xsolPrice, err := calc.CalculateXSOLPriceWithSOLPrice(ctx, solUSDPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate xSOL price: %w", err)
	}

	// Create combined response
	return &price.CombinedPriceResponse{
		SOLUSD:    solUSDPrice.Price,
		XSOLInSOL: xsolPrice.PriceInSOL,
		XSOLInUSD: xsolPrice.PriceInUSD,
		UpdatedAt: time.Now(),
	}, nil
}

// ValidateProtocolState performs comprehensive validation of protocol state for price calculation
func (calc *PriceCalculator) ValidateProtocolState(protocolState *HyloProtocolState) error {
	if protocolState == nil {
		return fmt.Errorf("protocol state cannot be nil")
	}

	// Use the built-in validation from the state type
	if err := protocolState.Validate(); err != nil {
		return fmt.Errorf("protocol state validation failed: %w", err)
	}

	// Additional validation specific to price calculation
	if protocolState.XSOLNAVInSOL <= 0 {
		return fmt.Errorf("xSOL NAV in SOL must be positive for price calculation: %f", protocolState.XSOLNAVInSOL)
	}

	if protocolState.HyUSDNAVInSOL <= 0 {
		return fmt.Errorf("hyUSD NAV in SOL must be positive for price calculation: %f", protocolState.HyUSDNAVInSOL)
	}

	// Check for reasonable collateral ratio (should be > 1 for healthy protocol)
	if protocolState.CollateralRatio < 1.0 {
		return fmt.Errorf("collateral ratio indicates undercollateralization: %f", protocolState.CollateralRatio)
	}

	// Check for reasonable effective leverage (should be > 1 but not extreme)
	if protocolState.EffectiveLeverage < 1.0 {
		return fmt.Errorf("effective leverage is below 1.0, indicating no leverage: %f", protocolState.EffectiveLeverage)
	}

	if protocolState.EffectiveLeverage > 50.0 {
		return fmt.Errorf("effective leverage is extremely high, possible calculation error: %f", protocolState.EffectiveLeverage)
	}

	return nil
}

// GetPriceCalculationDetails returns detailed information about the price calculation
// This is useful for debugging and monitoring price calculation health
func (calc *PriceCalculator) GetPriceCalculationDetails(protocolState *HyloProtocolState, solPriceUSD float64) (map[string]interface{}, error) {
	if err := calc.ValidateProtocolState(protocolState); err != nil {
		return nil, fmt.Errorf("invalid protocol state: %w", err)
	}

	// Calculate xSOL price
	xsolPrice, err := calc.CalculateXSOLPrice(protocolState, solPriceUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate xSOL price: %w", err)
	}

	// Get formatted token supplies
	hyusdFormatted, xsolFormatted := protocolState.GetFormattedSupplies()

	// Convert raw values to human-readable formats
	solReserveFormatted := protocolState.GetFormattedSOLReserve()

	return map[string]interface{}{
		// Input data
		"sol_price_usd": solPriceUSD,
		"hyusd_supply":  hyusdFormatted,
		"xsol_supply":   xsolFormatted,
		"sol_reserve":   solReserveFormatted,

		// Calculated NAVs
		"hyusd_nav_sol": protocolState.HyUSDNAVInSOL,
		"xsol_nav_sol":  protocolState.XSOLNAVInSOL,

		// Final prices
		"xsol_price_sol": xsolPrice.PriceInSOL,
		"xsol_price_usd": xsolPrice.PriceInUSD,

		// Protocol health metrics
		"collateral_ratio":   protocolState.CollateralRatio,
		"effective_leverage": protocolState.EffectiveLeverage,
		"protocol_healthy":   protocolState.IsHealthy(),

		// Calculation metadata
		"timestamp":             protocolState.Timestamp,
		"calculation_timestamp": time.Now(),
	}, nil
}

// EstimateXSOLPriceImpact estimates how xSOL price would change with different SOL/USD prices
// This is useful for stress testing and understanding price sensitivity
func (calc *PriceCalculator) EstimateXSOLPriceImpact(protocolState *HyloProtocolState, basePriceUSD float64, priceMultipliers []float64) ([]map[string]interface{}, error) {
	if err := calc.ValidateProtocolState(protocolState); err != nil {
		return nil, fmt.Errorf("invalid protocol state: %w", err)
	}

	results := make([]map[string]interface{}, len(priceMultipliers))

	for i, multiplier := range priceMultipliers {
		newSOLPrice := basePriceUSD * multiplier

		// Calculate xSOL price at this SOL price
		xsolPrice, err := calc.CalculateXSOLPrice(protocolState, newSOLPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate xSOL price for multiplier %f: %w", multiplier, err)
		}

		results[i] = map[string]interface{}{
			"sol_price_multiplier": multiplier,
			"sol_price_usd":        newSOLPrice,
			"xsol_price_sol":       xsolPrice.PriceInSOL,
			"xsol_price_usd":       xsolPrice.PriceInUSD,
			"collateral_ratio":     xsolPrice.CollateralRatio,
			"effective_leverage":   xsolPrice.EffectiveLeverage,
		}
	}

	return results, nil
}

// IsCalculationStale checks if the protocol state timestamp is too old for reliable calculation
func (calc *PriceCalculator) IsCalculationStale(protocolState *HyloProtocolState, maxAge time.Duration) bool {
	if protocolState == nil {
		return true
	}

	return time.Since(protocolState.Timestamp) > maxAge
}

// CalculateHistoricalXSOLPrice calculates historical xSOL price from trade data
// Only calculates for hyUSD trades, returns nil for SOL trades
// Formula: price = hyUSD_amount / xSOL_amount (hyUSD ≈ $1 by design)
func CalculateHistoricalXSOLPrice(trade *XSOLTrade) *string {
	// Skip if not hyUSD trade
	if trade.CounterAsset != "hyUSD" {
		return nil
	}

	// Parse amounts (handle decimal strings)
	xsolAmount, err := parseDecimalAmount(trade.XSOLAmount)
	if err != nil || xsolAmount <= 0 {
		return nil
	}

	hyusdAmount, err := parseDecimalAmount(trade.CounterAmount)
	if err != nil || hyusdAmount <= 0 {
		return nil
	}

	// Calculate: price = hyUSD_amount / xSOL_amount
	price := hyusdAmount / xsolAmount

	// Sanity check: xSOL price should be reasonable ($1-$10,000 range)
	if price < 1.0 || price > 10000.0 {
		return nil // Skip unrealistic prices
	}

	// Format to 3 decimal places for USD price
	formatted := fmt.Sprintf("%.3f", price)
	return &formatted
}

// parseDecimalAmount parses decimal amount string to float64
// Helper function for historical price calculation
func parseDecimalAmount(amountStr string) (float64, error) {
	if amountStr == "" {
		return 0, fmt.Errorf("empty amount string")
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse amount '%s': %w", amountStr, err)
	}

	return amount, nil
}
