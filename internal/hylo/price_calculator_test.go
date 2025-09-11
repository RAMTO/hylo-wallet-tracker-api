package hylo

import (
	"testing"
	"time"
)

func TestCalculateHistoricalXSOLPrice(t *testing.T) {
	tests := []struct {
		name     string
		trade    *XSOLTrade
		expected *string
	}{
		{
			name: "valid hyUSD trade - basic calculation",
			trade: &XSOLTrade{
				XSOLAmount:    "5.0",
				CounterAmount: "1250.0",
				CounterAsset:  "hyUSD",
			},
			expected: stringPtr("250.000"), // 1250 / 5 = 250
		},
		{
			name: "valid hyUSD trade - decimal amounts",
			trade: &XSOLTrade{
				XSOLAmount:    "2.5",
				CounterAmount: "500.75",
				CounterAsset:  "hyUSD",
			},
			expected: stringPtr("200.300"), // 500.75 / 2.5 = 200.3
		},
		{
			name: "valid hyUSD trade - small amounts",
			trade: &XSOLTrade{
				XSOLAmount:    "0.1",
				CounterAmount: "24.5",
				CounterAsset:  "hyUSD",
			},
			expected: stringPtr("245.000"), // 24.5 / 0.1 = 245
		},
		{
			name: "valid hyUSD trade - high price",
			trade: &XSOLTrade{
				XSOLAmount:    "1.0",
				CounterAmount: "5000.0",
				CounterAsset:  "hyUSD",
			},
			expected: stringPtr("5000.000"), // 5000 / 1 = 5000
		},
		{
			name: "SOL trade - should return nil",
			trade: &XSOLTrade{
				XSOLAmount:    "2.0",
				CounterAmount: "100.0",
				CounterAsset:  "SOL",
			},
			expected: nil,
		},
		{
			name: "USDC trade - should return nil",
			trade: &XSOLTrade{
				XSOLAmount:    "1.5",
				CounterAmount: "300.0",
				CounterAsset:  "USDC",
			},
			expected: nil,
		},
		{
			name: "jitoSOL trade - should return nil",
			trade: &XSOLTrade{
				XSOLAmount:    "3.0",
				CounterAmount: "150.0",
				CounterAsset:  "jitoSOL",
			},
			expected: nil,
		},
		{
			name: "zero xSOL amount",
			trade: &XSOLTrade{
				XSOLAmount:    "0.0",
				CounterAmount: "100.0",
				CounterAsset:  "hyUSD",
			},
			expected: nil,
		},
		{
			name: "zero hyUSD amount",
			trade: &XSOLTrade{
				XSOLAmount:    "1.0",
				CounterAmount: "0.0",
				CounterAsset:  "hyUSD",
			},
			expected: nil,
		},
		{
			name: "negative xSOL amount",
			trade: &XSOLTrade{
				XSOLAmount:    "-1.0",
				CounterAmount: "100.0",
				CounterAsset:  "hyUSD",
			},
			expected: nil,
		},
		{
			name: "empty xSOL amount",
			trade: &XSOLTrade{
				XSOLAmount:    "",
				CounterAmount: "100.0",
				CounterAsset:  "hyUSD",
			},
			expected: nil,
		},
		{
			name: "empty hyUSD amount",
			trade: &XSOLTrade{
				XSOLAmount:    "1.0",
				CounterAmount: "",
				CounterAsset:  "hyUSD",
			},
			expected: nil,
		},
		{
			name: "invalid xSOL amount format",
			trade: &XSOLTrade{
				XSOLAmount:    "not-a-number",
				CounterAmount: "100.0",
				CounterAsset:  "hyUSD",
			},
			expected: nil,
		},
		{
			name: "invalid hyUSD amount format",
			trade: &XSOLTrade{
				XSOLAmount:    "1.0",
				CounterAmount: "invalid",
				CounterAsset:  "hyUSD",
			},
			expected: nil,
		},
		{
			name: "price too low - below $1 threshold",
			trade: &XSOLTrade{
				XSOLAmount:    "1000.0", // Very high xSOL amount
				CounterAmount: "500.0",  // Moderate hyUSD amount
				CounterAsset:  "hyUSD",
			},
			expected: nil, // Price = 500/1000 = 0.5, below $1 threshold
		},
		{
			name: "price too high - above $10,000 threshold",
			trade: &XSOLTrade{
				XSOLAmount:    "0.01",   // Very small xSOL amount
				CounterAmount: "500000", // Very high hyUSD amount
				CounterAsset:  "hyUSD",
			},
			expected: nil, // Price = 500000/0.01 = 50,000,000, above $10,000 threshold
		},
		{
			name: "edge case - exactly $1 price (should be valid)",
			trade: &XSOLTrade{
				XSOLAmount:    "100.0",
				CounterAmount: "100.0",
				CounterAsset:  "hyUSD",
			},
			expected: stringPtr("1.000"), // 100 / 100 = 1.0
		},
		{
			name: "edge case - exactly $10,000 price (should be valid)",
			trade: &XSOLTrade{
				XSOLAmount:    "1.0",
				CounterAmount: "10000.0",
				CounterAsset:  "hyUSD",
			},
			expected: stringPtr("10000.000"), // 10000 / 1 = 10000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateHistoricalXSOLPrice(tt.trade)

			// Check if both are nil
			if tt.expected == nil && result == nil {
				return // Test passed
			}

			// Check if one is nil but the other isn't
			if tt.expected == nil || result == nil {
				t.Errorf("Expected %v, got %v", tt.expected, result)
				return
			}

			// Compare string values
			if *tt.expected != *result {
				t.Errorf("Expected price %s, got %s", *tt.expected, *result)
			}
		})
	}
}

func TestParseDecimalAmount(t *testing.T) {
	tests := []struct {
		name        string
		amountStr   string
		expected    float64
		expectError bool
	}{
		{
			name:        "valid integer",
			amountStr:   "100",
			expected:    100.0,
			expectError: false,
		},
		{
			name:        "valid decimal",
			amountStr:   "123.456",
			expected:    123.456,
			expectError: false,
		},
		{
			name:        "zero",
			amountStr:   "0.0",
			expected:    0.0,
			expectError: false,
		},
		{
			name:        "very small decimal",
			amountStr:   "0.000001",
			expected:    0.000001,
			expectError: false,
		},
		{
			name:        "large number",
			amountStr:   "999999.999999",
			expected:    999999.999999,
			expectError: false,
		},
		{
			name:        "empty string",
			amountStr:   "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid format",
			amountStr:   "not-a-number",
			expected:    0,
			expectError: true,
		},
		{
			name:        "multiple decimals",
			amountStr:   "12.34.56",
			expected:    0,
			expectError: true,
		},
		{
			name:        "scientific notation",
			amountStr:   "1e6",
			expected:    1000000.0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDecimalAmount(tt.amountStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.amountStr)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.amountStr, err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

// Test helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Comprehensive integration test for historical price in parsed trades
func TestParseTransactionWithHistoricalPrice(t *testing.T) {
	tests := []struct {
		name                    string
		trade                   *XSOLTrade
		expectedHistoricalPrice *string
	}{
		{
			name: "hyUSD BUY trade with historical price",
			trade: &XSOLTrade{
				Signature:     "testSigHistoricalBuy",
				Slot:          12345,
				BlockTime:     time.Now().Unix(),
				Side:          TradeSideBuy,
				XSOLAmount:    "2.0",
				CounterAmount: "500.0",
				CounterAsset:  "hyUSD",
				Timestamp:     time.Now(),
				ExplorerURL:   "https://solscan.io/tx/testSigHistoricalBuy",
			},
			expectedHistoricalPrice: stringPtr("250.000"), // 500 / 2 = 250
		},
		{
			name: "hyUSD SELL trade with historical price",
			trade: &XSOLTrade{
				Signature:     "testSigHistoricalSell",
				Slot:          12346,
				BlockTime:     time.Now().Unix(),
				Side:          TradeSideSell,
				XSOLAmount:    "1.5",
				CounterAmount: "375.75",
				CounterAsset:  "hyUSD",
				Timestamp:     time.Now(),
				ExplorerURL:   "https://solscan.io/tx/testSigHistoricalSell",
			},
			expectedHistoricalPrice: stringPtr("250.500"), // 375.75 / 1.5 = 250.5
		},
		{
			name: "SOL BUY trade - no historical price",
			trade: &XSOLTrade{
				Signature:     "testSigSOLBuy",
				Slot:          12347,
				BlockTime:     time.Now().Unix(),
				Side:          TradeSideBuy,
				XSOLAmount:    "3.0",
				CounterAmount: "150.0",
				CounterAsset:  "SOL",
				Timestamp:     time.Now(),
				ExplorerURL:   "https://solscan.io/tx/testSigSOLBuy",
			},
			expectedHistoricalPrice: nil, // SOL trades should not have historical price
		},
		{
			name: "USDC trade - no historical price",
			trade: &XSOLTrade{
				Signature:     "testSigUSDC",
				Slot:          12348,
				BlockTime:     time.Now().Unix(),
				Side:          TradeSideBuy,
				XSOLAmount:    "1.0",
				CounterAmount: "200.0",
				CounterAsset:  "USDC",
				Timestamp:     time.Now(),
				ExplorerURL:   "https://solscan.io/tx/testSigUSDC",
			},
			expectedHistoricalPrice: nil, // USDC trades should not have historical price
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate historical price (this simulates what happens in the parser)
			tt.trade.HistoricalPriceUSD = CalculateHistoricalXSOLPrice(tt.trade)

			// Check if both are nil
			if tt.expectedHistoricalPrice == nil && tt.trade.HistoricalPriceUSD == nil {
				return // Test passed
			}

			// Check if one is nil but the other isn't
			if tt.expectedHistoricalPrice == nil || tt.trade.HistoricalPriceUSD == nil {
				t.Errorf("Expected historical price %v, got %v", tt.expectedHistoricalPrice, tt.trade.HistoricalPriceUSD)
				return
			}

			// Compare string values
			if *tt.expectedHistoricalPrice != *tt.trade.HistoricalPriceUSD {
				t.Errorf("Expected historical price %s, got %s", *tt.expectedHistoricalPrice, *tt.trade.HistoricalPriceUSD)
			}
		})
	}
}
