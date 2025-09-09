package utils

import "testing"

func TestFormatTokenAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   uint64
		decimals uint8
		expected string
	}{
		{"Zero amount", 0, 6, "0"},
		{"Small amount", 1, 6, "0.000001"},
		{"One unit", 1000000, 6, "1"},
		{"Decimal amount", 1500000, 6, "1.5"},
		{"Large amount", 123456789000, 6, "123456.789"},
		{"No decimal places", 1000, 0, "1000"},
		{"SOL amount", 1000000000, 9, "1"},
		{"Fractional SOL", 1500000000, 9, "1.5"},
		{"Trailing zeros removed", 1100000, 6, "1.1"},
		{"Many decimal places", 1234567, 6, "1.234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTokenAmount(tt.amount, tt.decimals)
			if result != tt.expected {
				t.Errorf("FormatTokenAmount(%d, %d) = %s, want %s",
					tt.amount, tt.decimals, result, tt.expected)
			}
		})
	}
}

func TestParseDecimalAmount(t *testing.T) {
	tests := []struct {
		name        string
		decimalStr  string
		decimals    uint8
		expectedRaw uint64
		wantErr     bool
	}{
		{"Zero string", "0", 6, 0, false},
		{"Empty string", "", 6, 0, false},
		{"Integer", "123", 6, 123000000, false},
		{"Decimal", "1.5", 6, 1500000, false},
		{"Small decimal", "0.000001", 6, 1, false},
		{"Large decimal", "123456.789", 6, 123456789000, false},
		{"No decimals token", "1000", 0, 1000, false},
		{"SOL amount", "1.5", 9, 1500000000, false},
		{"Invalid format", "1.2.3", 6, 0, true},
		{"Too many decimals", "1.1234567", 6, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDecimalAmount(tt.decimalStr, tt.decimals)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDecimalAmount(%s, %d) should have failed", tt.decimalStr, tt.decimals)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDecimalAmount(%s, %d) failed: %v", tt.decimalStr, tt.decimals, err)
				return
			}

			if result != tt.expectedRaw {
				t.Errorf("ParseDecimalAmount(%s, %d): expected %d, got %d",
					tt.decimalStr, tt.decimals, tt.expectedRaw, result)
			}
		})
	}
}

// Test round-trip conversion
func TestFormatParseCycle(t *testing.T) {
	testCases := []struct {
		amount   uint64
		decimals uint8
	}{
		{0, 6},
		{1, 6},
		{1000000, 6},
		{1500000, 6},
		{123456789000, 6},
		{1000000000, 9},
	}

	for _, tc := range testCases {
		// Format -> Parse should return original
		formatted := FormatTokenAmount(tc.amount, tc.decimals)
		parsed, err := ParseDecimalAmount(formatted, tc.decimals)

		if err != nil {
			t.Errorf("Round-trip failed for %d (decimals=%d): parse error %v", tc.amount, tc.decimals, err)
			continue
		}

		if parsed != tc.amount {
			t.Errorf("Round-trip failed for %d (decimals=%d): got %d", tc.amount, tc.decimals, parsed)
		}
	}
}
