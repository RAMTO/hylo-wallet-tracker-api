package hylo

import (
	"context"
	"fmt"

	"hylo-wallet-tracker-api/internal/price"
	"hylo-wallet-tracker-api/internal/solana"
)

// PriceService provides a high-level interface for xSOL price calculation
// It integrates the StateReader and PriceCalculator to provide complete pricing functionality
type PriceService struct {
	stateReader       *StateReader
	priceCalculator   *PriceCalculator
	dexScreenerClient *price.DexScreenerClient
}

// NewPriceService creates a new PriceService with all required dependencies
func NewPriceService(solanaClient *solana.HTTPClient, config *Config, priceConfig *price.PriceConfig) *PriceService {
	// Create state reader
	stateReader := NewStateReader(solanaClient, config)

	// Create price calculator
	priceCalculator := NewPriceCalculator(stateReader)

	// Create DexScreener client for SOL/USD prices
	dexScreenerClient := price.NewDexScreenerClient(priceConfig)

	return &PriceService{
		stateReader:       stateReader,
		priceCalculator:   priceCalculator,
		dexScreenerClient: dexScreenerClient,
	}
}

// GetCurrentXSOLPrice fetches the current xSOL price in both SOL and USD terms
// This method handles the complete workflow:
// 1. Fetch SOL/USD price from DexScreener
// 2. Read Hylo protocol state from on-chain data
// 3. Calculate xSOL price using Hylo equations
func (ps *PriceService) GetCurrentXSOLPrice(ctx context.Context) (*price.XSOLPrice, error) {
	// Step 1: Fetch current SOL/USD price
	solPrice, err := ps.dexScreenerClient.FetchSOLPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SOL/USD price: %w", err)
	}

	// Step 2: Calculate xSOL price using the SOL price
	xsolPrice, err := ps.priceCalculator.CalculateXSOLPriceWithSOLPrice(ctx, solPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate xSOL price: %w", err)
	}

	return xsolPrice, nil
}

// GetCombinedPriceResponse returns the complete price response for the /price API endpoint
// Includes SOL/USD, xSOL/SOL, and xSOL/USD prices in the expected API format
func (ps *PriceService) GetCombinedPriceResponse(ctx context.Context) (*price.CombinedPriceResponse, error) {
	// Step 1: Fetch current SOL/USD price
	solPrice, err := ps.dexScreenerClient.FetchSOLPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SOL/USD price: %w", err)
	}

	// Step 2: Calculate combined price response
	response, err := ps.priceCalculator.CalculateCombinedPriceResponse(ctx, solPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate combined price response: %w", err)
	}

	return response, nil
}

// GetProtocolHealthStatus returns comprehensive protocol health information
// Useful for monitoring and debugging price calculation issues
func (ps *PriceService) GetProtocolHealthStatus(ctx context.Context) (map[string]interface{}, error) {
	// Step 1: Fetch current SOL/USD price
	solPrice, err := ps.dexScreenerClient.FetchSOLPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SOL/USD price: %w", err)
	}

	// Step 2: Get health status from state reader
	healthStatus, err := ps.stateReader.GetHealthStatus(ctx, solPrice.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to get protocol health status: %w", err)
	}

	// Step 3: Add SOL/USD price information
	healthStatus["sol_price_usd"] = solPrice.Price
	healthStatus["sol_price_source"] = solPrice.Source
	healthStatus["sol_price_pair"] = solPrice.Pair
	healthStatus["sol_price_timestamp"] = solPrice.Timestamp

	return healthStatus, nil
}

// ValidateSystemHealth performs comprehensive validation of all system components
// Ensures that all dependencies are working correctly for price calculation
func (ps *PriceService) ValidateSystemHealth(ctx context.Context) error {
	// Step 1: Validate Solana connectivity by checking token mints
	if err := ps.stateReader.ValidateAllHyloTokenMints(ctx); err != nil {
		return fmt.Errorf("Solana connectivity validation failed: %w", err)
	}

	// Step 2: Validate DexScreener connectivity by fetching SOL price
	solPrice, err := ps.dexScreenerClient.FetchSOLPrice(ctx)
	if err != nil {
		return fmt.Errorf("DexScreener connectivity validation failed: %w", err)
	}

	// Step 3: Validate price calculation by reading protocol state
	protocolState, err := ps.stateReader.ReadProtocolState(ctx, solPrice.Price)
	if err != nil {
		return fmt.Errorf("protocol state reading validation failed: %w", err)
	}

	// Step 4: Validate price calculation
	if err := ps.priceCalculator.ValidateProtocolState(protocolState); err != nil {
		return fmt.Errorf("price calculation validation failed: %w", err)
	}

	// Step 5: Perform a full price calculation to ensure everything works
	_, err = ps.priceCalculator.CalculateXSOLPrice(protocolState, solPrice.Price)
	if err != nil {
		return fmt.Errorf("xSOL price calculation validation failed: %w", err)
	}

	return nil
}

// GetPriceCalculationDetails returns detailed information about the price calculation
// This is useful for debugging and understanding how prices are derived
func (ps *PriceService) GetPriceCalculationDetails(ctx context.Context) (map[string]interface{}, error) {
	// Step 1: Fetch current SOL/USD price
	solPrice, err := ps.dexScreenerClient.FetchSOLPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SOL/USD price: %w", err)
	}

	// Step 2: Read protocol state
	protocolState, err := ps.stateReader.ReadProtocolState(ctx, solPrice.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol state: %w", err)
	}

	// Step 3: Get detailed calculation information
	details, err := ps.priceCalculator.GetPriceCalculationDetails(protocolState, solPrice.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to get price calculation details: %w", err)
	}

	// Step 4: Add SOL price source information
	details["sol_price_source"] = solPrice.Source
	details["sol_price_pair"] = solPrice.Pair
	details["sol_price_liquidity"] = solPrice.Liquidity
	details["sol_price_volume_24h"] = solPrice.Volume24h

	return details, nil
}

// Close performs cleanup of all resources
func (ps *PriceService) Close() error {
	// Close DexScreener client
	if err := ps.dexScreenerClient.Close(); err != nil {
		return fmt.Errorf("failed to close DexScreener client: %w", err)
	}

	return nil
}

// GetStateReader returns the underlying StateReader for advanced usage
func (ps *PriceService) GetStateReader() *StateReader {
	return ps.stateReader
}

// GetPriceCalculator returns the underlying PriceCalculator for advanced usage
func (ps *PriceService) GetPriceCalculator() *PriceCalculator {
	return ps.priceCalculator
}

// GetDexScreenerClient returns the underlying DexScreenerClient for advanced usage
func (ps *PriceService) GetDexScreenerClient() *price.DexScreenerClient {
	return ps.dexScreenerClient
}
