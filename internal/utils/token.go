package utils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// FormatTokenAmount formats a raw token amount with proper decimal precision using big.Int math
// This is the centralized, most robust implementation for all token amount formatting
// across the entire codebase to ensure consistent precision.
func FormatTokenAmount(rawAmount uint64, decimals uint8) string {
	if rawAmount == 0 {
		return "0"
	}

	// Convert raw amount to big.Int for precise decimal math
	rawBig := new(big.Int).SetUint64(rawAmount)

	// Calculate the divisor (10^decimals)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// Perform division to get integer part
	integerPart := new(big.Int).Div(rawBig, divisor)

	// Get remainder for fractional part
	remainder := new(big.Int).Mod(rawBig, divisor)

	// If no fractional part, return integer part
	if remainder.Cmp(big.NewInt(0)) == 0 {
		return integerPart.String()
	}

	// Format fractional part with leading zeros if needed
	fracStr := remainder.String()
	if len(fracStr) < int(decimals) {
		// Pad with leading zeros
		fracStr = strings.Repeat("0", int(decimals)-len(fracStr)) + fracStr
	}

	// Remove trailing zeros from fractional part
	fracStr = strings.TrimRight(fracStr, "0")

	// If no fractional digits remain, return just integer part
	if fracStr == "" {
		return integerPart.String()
	}

	return integerPart.String() + "." + fracStr
}

// ParseDecimalAmount parses a decimal string into raw token amount
// This is the reverse of FormatTokenAmount() - useful for testing and input parsing
func ParseDecimalAmount(decimalStr string, decimals uint8) (uint64, error) {
	if decimalStr == "" || decimalStr == "0" {
		return 0, nil
	}

	// Split on decimal point
	parts := strings.Split(decimalStr, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid decimal format: %s", decimalStr)
	}

	integerPart := parts[0]
	fractionalPart := ""

	if len(parts) == 2 {
		fractionalPart = parts[1]
	}

	// Validate fractional part doesn't exceed token decimals
	if len(fractionalPart) > int(decimals) {
		return 0, fmt.Errorf("fractional part exceeds token decimals (%d): %s", decimals, decimalStr)
	}

	// Pad fractional part with trailing zeros to match token decimals
	for len(fractionalPart) < int(decimals) {
		fractionalPart += "0"
	}

	// Combine integer and fractional parts
	combinedStr := integerPart + fractionalPart

	// Parse as uint64
	rawAmount, err := strconv.ParseUint(combinedStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse decimal amount: %w", err)
	}

	return rawAmount, nil
}
