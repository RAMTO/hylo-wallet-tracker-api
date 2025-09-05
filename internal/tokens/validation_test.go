package tokens

import (
	"strings"
	"testing"

	"hylo-wallet-tracker-api/internal/solana"
)

func TestSanitizeAddress(t *testing.T) {
	t.Run("valid addresses", func(t *testing.T) {
		validAddresses := []string{
			"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
			"11111111111111111111111111111111111111111111",
			"5YMkXAYccHSGnHn9nob9xEvv6Pvka9DZWH7nTbotTu9E",
			"HnnGv3HrSqjRpgdFmx7vQGjntNEoex1SU4e9Lxcxuihz",
			"4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
		}

		for _, addr := range validAddresses {
			sanitized, err := SanitizeAddress(addr)
			if err != nil {
				t.Errorf("SanitizeAddress(%s) failed: %v", addr, err)
			}

			if string(sanitized) != addr {
				t.Errorf("SanitizeAddress(%s) changed address: got %s", addr, sanitized)
			}
		}
	})

	t.Run("addresses with whitespace", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{" A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g"},
			{"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g ", "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g"},
			{" A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g ", "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g"},
			{"\tA3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g\n", "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g"},
		}

		for _, tc := range testCases {
			sanitized, err := SanitizeAddress(tc.input)
			if err != nil {
				t.Errorf("SanitizeAddress(%q) failed: %v", tc.input, err)
			}

			if string(sanitized) != tc.expected {
				t.Errorf("SanitizeAddress(%q): expected %s, got %s", tc.input, tc.expected, sanitized)
			}
		}
	})

	t.Run("invalid addresses", func(t *testing.T) {
		invalidAddresses := []struct {
			address string
			reason  string
		}{
			{"", "empty address"},
			{"short", "too short"},
			{"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6gTOOLONG", "too long"},
			{"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc0g", "contains invalid character '0'"},
			{"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDycIg", "contains invalid character 'I'"},
			{"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDycOg", "contains invalid character 'O'"},
			{"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyclg", "contains invalid character 'l'"},
		}

		for _, tc := range invalidAddresses {
			_, err := SanitizeAddress(tc.address)
			if err == nil {
				t.Errorf("SanitizeAddress(%s) should have failed (%s)", tc.address, tc.reason)
			}

			// Check that it returns AddressValidationError
			if validationErr, ok := err.(*AddressValidationError); ok {
				if validationErr.Address != tc.address {
					t.Errorf("Error should contain original address: expected %s, got %s",
						tc.address, validationErr.Address)
				}
			} else {
				t.Errorf("Should return AddressValidationError, got %T", err)
			}
		}
	})
}

func TestValidateWalletAddress(t *testing.T) {
	t.Run("valid wallet addresses", func(t *testing.T) {
		validAddresses := []string{
			"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
			"11111111111111111111111111111111111111111111",
		}

		for _, addr := range validAddresses {
			if err := ValidateWalletAddress(addr); err != nil {
				t.Errorf("ValidateWalletAddress(%s) failed: %v", addr, err)
			}
		}
	})

	t.Run("invalid wallet addresses", func(t *testing.T) {
		invalidAddresses := []string{
			"",
			"short",
			"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6gTOOLONG",
			"invalid-characters-here-123456789012345678",
		}

		for _, addr := range invalidAddresses {
			if err := ValidateWalletAddress(addr); err == nil {
				t.Errorf("ValidateWalletAddress(%s) should have failed", addr)
			}
		}
	})
}

func TestValidateTokenMintAddress(t *testing.T) {
	config := NewConfig()

	t.Run("supported token mints", func(t *testing.T) {
		supportedMints := []string{
			string(HyUSDMint),
			string(SHyUSDMint),
			string(XSOLMint),
		}

		for _, mint := range supportedMints {
			if err := ValidateTokenMintAddress(mint, config); err != nil {
				t.Errorf("ValidateTokenMintAddress(%s) should be valid: %v", mint, err)
			}
		}
	})

	t.Run("unsupported but valid mint", func(t *testing.T) {
		// Valid address format but not a Hylo token
		unsupportedMint := "11111111111111111111111111111111111111111111"

		err := ValidateTokenMintAddress(unsupportedMint, config)
		if err == nil {
			t.Errorf("ValidateTokenMintAddress(%s) should fail for unsupported token", unsupportedMint)
		}

		// Should be AddressValidationError
		if validationErr, ok := err.(*AddressValidationError); ok {
			if !strings.Contains(validationErr.Reason, "unsupported") {
				t.Errorf("Error should mention unsupported token: %s", validationErr.Reason)
			}
		}
	})

	t.Run("invalid mint format", func(t *testing.T) {
		invalidMints := []string{
			"",
			"short",
			"toolong1111111111111111111111111111111111111111111",
		}

		for _, mint := range invalidMints {
			if err := ValidateTokenMintAddress(mint, config); err == nil {
				t.Errorf("ValidateTokenMintAddress(%s) should fail", mint)
			}
		}
	})

	t.Run("nil config allows any valid address", func(t *testing.T) {
		validMint := "11111111111111111111111111111111111111111111"

		if err := ValidateTokenMintAddress(validMint, nil); err != nil {
			t.Errorf("ValidateTokenMintAddress with nil config should allow valid addresses: %v", err)
		}
	})
}

func TestValidateATAAddress(t *testing.T) {
	t.Run("valid ATA addresses", func(t *testing.T) {
		// Use known addresses as examples of valid format
		validATAs := []string{
			"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
			"11111111111111111111111111111111111111111111",
			"5YMkXAYccHSGnHn9nob9xEvv6Pvka9DZWH7nTbotTu9E",
		}

		for _, ata := range validATAs {
			if err := ValidateATAAddress(ata); err != nil {
				t.Errorf("ValidateATAAddress(%s) failed: %v", ata, err)
			}
		}
	})

	t.Run("invalid ATA addresses", func(t *testing.T) {
		invalidATAs := []string{
			"",
			"short",
			"toolong1111111111111111111111111111111111111111111",
		}

		for _, ata := range invalidATAs {
			if err := ValidateATAAddress(ata); err == nil {
				t.Errorf("ValidateATAAddress(%s) should have failed", ata)
			}
		}
	})
}

func TestValidateATADerivation(t *testing.T) {
	wallet := "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g"
	mint := string(HyUSDMint)

	t.Run("correct ATA derivation", func(t *testing.T) {
		// Derive correct ATA
		walletAddr := solana.Address(wallet)
		mintAddr := solana.Address(mint)
		correctATA, err := DeriveAssociatedTokenAddress(walletAddr, mintAddr)
		if err != nil {
			t.Errorf("Failed to derive ATA: %v", err)
			return
		}

		// Validate the derivation
		err = ValidateATADerivation(string(correctATA), wallet, mint)
		if err != nil {
			t.Errorf("ValidateATADerivation should pass for correct ATA: %v", err)
		}
	})

	t.Run("incorrect ATA derivation", func(t *testing.T) {
		wrongATA := "11111111111111111111111111111111111111111111"

		err := ValidateATADerivation(wrongATA, wallet, mint)
		if err == nil {
			t.Errorf("ValidateATADerivation should fail for incorrect ATA")
		}

		// Should be AddressValidationError
		if validationErr, ok := err.(*AddressValidationError); ok {
			if !strings.Contains(validationErr.Reason, "does not match") {
				t.Errorf("Error should mention derivation mismatch: %s", validationErr.Reason)
			}
		}
	})

	t.Run("invalid inputs", func(t *testing.T) {
		testCases := []struct {
			ata    string
			wallet string
			mint   string
			name   string
		}{
			{"invalid", wallet, mint, "invalid ATA"},
			{"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", "invalid", mint, "invalid wallet"},
			{"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", wallet, "invalid", "invalid mint"},
		}

		for _, tc := range testCases {
			err := ValidateATADerivation(tc.ata, tc.wallet, tc.mint)
			if err == nil {
				t.Errorf("ValidateATADerivation should fail for %s", tc.name)
			}
		}
	})
}

func TestIsValidBase58Character(t *testing.T) {
	t.Run("valid base58 characters", func(t *testing.T) {
		validChars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

		for i, char := range validChars {
			if !IsValidBase58Character(byte(char)) {
				t.Errorf("Character %c at position %d should be valid base58", char, i)
			}
		}
	})

	t.Run("invalid base58 characters", func(t *testing.T) {
		invalidChars := []byte{'0', 'O', 'I', 'l', ' ', '\n', '\t', '!', '@', '#'}

		for _, char := range invalidChars {
			if IsValidBase58Character(char) {
				t.Errorf("Character %c should not be valid base58", char)
			}
		}
	})
}

func TestValidateAddressBatch(t *testing.T) {
	t.Run("mixed valid and invalid addresses", func(t *testing.T) {
		addresses := []string{
			"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // valid
			"invalid", // invalid
			"11111111111111111111111111111111111111111111", // valid
			"", // invalid
			"5YMkXAYccHSGnHn9nob9xEvv6Pvka9DZWH7nTbotTu9E", // valid
		}

		errors := ValidateAddressBatch(addresses)

		if len(errors) != len(addresses) {
			t.Errorf("Expected %d errors, got %d", len(addresses), len(errors))
		}

		// Check specific results
		expectedResults := []bool{true, false, true, false, true} // true = valid (no error)

		for i, expected := range expectedResults {
			hasError := errors[i] != nil
			if expected && hasError {
				t.Errorf("Address %d should be valid but got error: %v", i, errors[i])
			}
			if !expected && !hasError {
				t.Errorf("Address %d should be invalid but got no error", i)
			}
		}
	})
}

func TestSuggestCorrection(t *testing.T) {
	t.Run("too short address", func(t *testing.T) {
		short := "short"
		suggestions := SuggestCorrection(short)

		if len(suggestions) == 0 {
			t.Errorf("Should provide suggestions for short address")
		}

		found := false
		for _, suggestion := range suggestions {
			if strings.Contains(suggestion, "too short") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should suggest address is too short")
		}
	})

	t.Run("too long address", func(t *testing.T) {
		long := "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6gTOOLONG"
		suggestions := SuggestCorrection(long)

		if len(suggestions) == 0 {
			t.Errorf("Should provide suggestions for long address")
		}

		found := false
		for _, suggestion := range suggestions {
			if strings.Contains(suggestion, "too long") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should suggest address is too long")
		}
	})

	t.Run("address with whitespace", func(t *testing.T) {
		withSpace := " A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g "
		suggestions := SuggestCorrection(withSpace)

		found := false
		for _, suggestion := range suggestions {
			if strings.Contains(suggestion, "whitespace") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should suggest removing whitespace")
		}
	})

	t.Run("address with zero character", func(t *testing.T) {
		withZero := "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc0g"
		suggestions := SuggestCorrection(withZero)

		found := false
		for _, suggestion := range suggestions {
			if strings.Contains(suggestion, "'0'") && strings.Contains(suggestion, "'O'") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should suggest replacing '0' with 'O'")
		}
	})

	t.Run("address with capital I", func(t *testing.T) {
		withI := "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDycIg"
		suggestions := SuggestCorrection(withI)

		found := false
		for _, suggestion := range suggestions {
			if strings.Contains(suggestion, "'I'") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Should suggest issue with 'I' character")
		}
	})
}

func TestValidateTokenConfiguration(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := NewConfig()

		if err := ValidateTokenConfiguration(config); err != nil {
			t.Errorf("ValidateTokenConfiguration should pass for valid config: %v", err)
		}
	})

	t.Run("nil configuration", func(t *testing.T) {
		if err := ValidateTokenConfiguration(nil); err == nil {
			t.Errorf("ValidateTokenConfiguration should fail for nil config")
		}
	})

	t.Run("configuration with duplicate mints", func(t *testing.T) {
		// Create config with duplicate mint addresses
		duplicateConfig := &Config{
			HyUSDMint:  HyUSDMint,
			SHyUSDMint: HyUSDMint, // Duplicate!
			XSOLMint:   XSOLMint,
		}
		duplicateConfig.buildTokenRegistry()

		if err := ValidateTokenConfiguration(duplicateConfig); err == nil {
			t.Errorf("ValidateTokenConfiguration should fail for duplicate mints")
		}
	})

	t.Run("configuration with invalid mint", func(t *testing.T) {
		// Create config with invalid mint address
		invalidConfig := &Config{
			HyUSDMint:  solana.Address("invalid"),
			SHyUSDMint: SHyUSDMint,
			XSOLMint:   XSOLMint,
		}
		invalidConfig.buildTokenRegistry()

		if err := ValidateTokenConfiguration(invalidConfig); err == nil {
			t.Errorf("ValidateTokenConfiguration should fail for invalid mint")
		}
	})
}

func TestAddressValidationError(t *testing.T) {
	t.Run("error formatting", func(t *testing.T) {
		err := &AddressValidationError{
			Address: "invalid",
			Reason:  "too short",
		}

		expected := "invalid address invalid: too short"
		if err.Error() != expected {
			t.Errorf("Error message format: expected %q, got %q", expected, err.Error())
		}
	})
}
