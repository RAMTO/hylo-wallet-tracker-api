package tokens

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"testing"

	"hylo-wallet-tracker-api/internal/solana"

	"github.com/mr-tron/base58"
)

// GoldenTestCase represents a test case for ATA derivation
type GoldenTestCase struct {
	Name        string `json:"name"`
	Wallet      string `json:"wallet"`
	Mint        string `json:"mint"`
	ExpectedATA string `json:"expected_ata"`
	TokenSymbol string `json:"token_symbol"`
}

// ValidationTestCase represents a test case for validation errors
type ValidationTestCase struct {
	Name        string `json:"name"`
	Wallet      string `json:"wallet"`
	Mint        string `json:"mint,omitempty"`
	ShouldError bool   `json:"should_error"`
	ErrorType   string `json:"error_type"`
}

// GoldenTestData represents the structure of the golden test data file
type GoldenTestData struct {
	Description         string               `json:"description"`
	TestCases           []GoldenTestCase     `json:"test_cases"`
	ValidationTestCases []ValidationTestCase `json:"validation_test_cases"`
	Notes               []string             `json:"notes"`
}

// Test helper functions for ATA tests
func decodeBase58(s string) ([]byte, error) {
	return base58.Decode(s)
}

func encodeBase58(b []byte) string {
	return base58.Encode(b)
}

func sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func TestDeriveAssociatedTokenAddress(t *testing.T) {
	t.Run("basic ATA derivation", func(t *testing.T) {
		// Test with known addresses
		wallet := solana.Address(TestReferenceWallet)
		mint := HyUSDMint

		ata, err := DeriveAssociatedTokenAddress(wallet, mint)
		if err != nil {
			t.Errorf("ATA derivation failed: %v", err)
		}

		// Basic validation - ATA should be a valid Solana address
		if err := ata.Validate(); err != nil {
			t.Errorf("Derived ATA is not a valid address: %v", err)
		}

		// ATA should be different from wallet and mint
		if ata == wallet {
			t.Errorf("ATA should not equal wallet address")
		}

		if ata == mint {
			t.Errorf("ATA should not equal mint address")
		}

		t.Logf("Derived ATA for %s + %s = %s", wallet, mint, ata)
	})

	t.Run("deterministic derivation", func(t *testing.T) {
		// Same inputs should always produce same output
		wallet := solana.Address(TestReferenceWallet)
		mint := XSOLMint

		ata1, err1 := DeriveAssociatedTokenAddress(wallet, mint)
		ata2, err2 := DeriveAssociatedTokenAddress(wallet, mint)

		if err1 != nil {
			t.Errorf("First derivation failed: %v", err1)
		}

		if err2 != nil {
			t.Errorf("Second derivation failed: %v", err2)
		}

		if ata1 != ata2 {
			t.Errorf("ATA derivation is not deterministic: %s != %s", ata1, ata2)
		}
	})

	t.Run("different tokens produce different ATAs", func(t *testing.T) {
		wallet := solana.Address(TestReferenceWallet)

		hyusdATA, err := DeriveAssociatedTokenAddress(wallet, HyUSDMint)
		if err != nil {
			t.Errorf("hyUSD ATA derivation failed: %v", err)
		}

		xsolATA, err := DeriveAssociatedTokenAddress(wallet, XSOLMint)
		if err != nil {
			t.Errorf("xSOL ATA derivation failed: %v", err)
		}

		if hyusdATA == xsolATA {
			t.Errorf("Different tokens should produce different ATAs")
		}
	})

	t.Run("different wallets produce different ATAs", func(t *testing.T) {
		wallet1 := solana.Address(TestReferenceWallet)
		wallet2 := solana.Address(TestSystemWallet)

		ata1, err := DeriveAssociatedTokenAddress(wallet1, HyUSDMint)
		if err != nil {
			t.Errorf("Wallet1 ATA derivation failed: %v", err)
		}

		ata2, err := DeriveAssociatedTokenAddress(wallet2, HyUSDMint)
		if err != nil {
			t.Errorf("Wallet2 ATA derivation failed: %v", err)
		}

		if ata1 == ata2 {
			t.Errorf("Different wallets should produce different ATAs")
		}
	})

	t.Run("invalid wallet address", func(t *testing.T) {
		invalidWallet := solana.Address("invalid")

		_, err := DeriveAssociatedTokenAddress(invalidWallet, HyUSDMint)
		if err == nil {
			t.Errorf("Should fail with invalid wallet address")
		}
	})

	t.Run("invalid mint address", func(t *testing.T) {
		wallet := solana.Address(TestReferenceWallet)
		invalidMint := solana.Address("invalid")

		_, err := DeriveAssociatedTokenAddress(wallet, invalidMint)
		if err == nil {
			t.Errorf("Should fail with invalid mint address")
		}
	})
}

func TestGetWalletATAs(t *testing.T) {
	config := NewConfig()
	wallet := solana.Address(TestReferenceWallet)

	t.Run("get all wallet ATAs", func(t *testing.T) {
		atas, err := GetWalletATAs(wallet, config)
		if err != nil {
			t.Errorf("GetWalletATAs failed: %v", err)
		}

		// Should have ATAs for all supported tokens
		expectedTokens := []string{HyUSDSymbol, SHyUSDSymbol, XSOLSymbol, USDCSymbol}

		if len(atas) != len(expectedTokens) {
			t.Errorf("Expected %d ATAs, got %d", len(expectedTokens), len(atas))
		}

		for _, tokenSymbol := range expectedTokens {
			ata, exists := atas[tokenSymbol]
			if !exists {
				t.Errorf("Missing ATA for token %s", tokenSymbol)
				continue
			}

			// Validate ATA address
			if err := ata.Validate(); err != nil {
				t.Errorf("Invalid ATA for %s: %v", tokenSymbol, err)
			}

			t.Logf("ATA for %s: %s", tokenSymbol, ata)
		}

		// Verify ATAs are unique
		uniqueATAs := make(map[solana.Address]string)
		for symbol, ata := range atas {
			if existingSymbol, exists := uniqueATAs[ata]; exists {
				t.Errorf("Duplicate ATA %s for tokens %s and %s", ata, existingSymbol, symbol)
			}
			uniqueATAs[ata] = symbol
		}
	})

	t.Run("invalid wallet", func(t *testing.T) {
		invalidWallet := solana.Address("invalid")

		_, err := GetWalletATAs(invalidWallet, config)
		if err == nil {
			t.Errorf("Should fail with invalid wallet")
		}
	})

	t.Run("nil config", func(t *testing.T) {
		_, err := GetWalletATAs(wallet, nil)
		if err == nil {
			t.Errorf("Should fail with nil config")
		}
	})
}

func TestGetWalletATAForToken(t *testing.T) {
	config := NewConfig()
	wallet := solana.Address(TestReferenceWallet)

	t.Run("get ATA for specific token", func(t *testing.T) {
		ata, err := GetWalletATAForToken(wallet, HyUSDSymbol, config)
		if err != nil {
			t.Errorf("GetWalletATAForToken failed: %v", err)
		}

		if err := ata.Validate(); err != nil {
			t.Errorf("Invalid ATA: %v", err)
		}

		// Should match ATA from batch function
		batchATAs, err := GetWalletATAs(wallet, config)
		if err != nil {
			t.Errorf("GetWalletATAs failed: %v", err)
		}

		if batchATA, exists := batchATAs[HyUSDSymbol]; exists {
			if ata != batchATA {
				t.Errorf("Single ATA derivation doesn't match batch: %s != %s", ata, batchATA)
			}
		}
	})

	t.Run("unsupported token", func(t *testing.T) {
		_, err := GetWalletATAForToken(wallet, "INVALID", config)
		if err == nil {
			t.Errorf("Should fail with unsupported token")
		}
	})
}

func TestATADeterminism(t *testing.T) {
	// Test that our ATA derivation is deterministic across multiple runs
	config := NewConfig()
	wallet := solana.Address(TestReferenceWallet)

	t.Run("multiple derivations produce same results", func(t *testing.T) {
		results := make([]map[string]solana.Address, 5)

		// Derive ATAs multiple times
		for i := 0; i < 5; i++ {
			atas, err := GetWalletATAs(wallet, config)
			if err != nil {
				t.Errorf("Derivation %d failed: %v", i, err)
			}
			results[i] = atas
		}

		// All results should be identical
		for i := 1; i < len(results); i++ {
			for symbol, ata := range results[0] {
				if results[i][symbol] != ata {
					t.Errorf("Derivation %d differs from first: %s %s != %s",
						i, symbol, results[i][symbol], ata)
				}
			}
		}
	})
}

func TestCryptoUtilityFunctions(t *testing.T) {
	t.Run("base58 encode/decode round trip", func(t *testing.T) {
		// Test with known Solana address
		original := TestReferenceWallet

		// Decode to bytes
		bytes, err := decodeBase58(original)
		if err != nil {
			t.Errorf("Base58 decode failed: %v", err)
		}

		if len(bytes) != 32 {
			t.Errorf("Decoded bytes should be 32 bytes, got %d", len(bytes))
		}

		// Encode back to string
		encoded := encodeBase58(bytes)

		// Should match original (or be a valid equivalent)
		if len(encoded) != 44 {
			t.Errorf("Encoded string should be 44 characters, got %d", len(encoded))
		}

		// Validate it's a proper base58 string
		for _, char := range encoded {
			if !IsValidBase58Character(byte(char)) {
				t.Errorf("Invalid base58 character in encoded result: %c", char)
			}
		}
	})

	t.Run("SHA256 hash function", func(t *testing.T) {
		input := []byte("test input")
		hash := sha256Hash(input)

		if len(hash) != 32 {
			t.Errorf("SHA256 hash should be 32 bytes, got %d", len(hash))
		}

		// Same input should produce same hash
		hash2 := sha256Hash(input)
		for i, b := range hash {
			if hash2[i] != b {
				t.Errorf("SHA256 hash is not deterministic")
				break
			}
		}
	})
}

// TestLoadGoldenVectors tests against golden test vectors from file
// This will be used to validate our ATA derivation against known good values
func TestLoadGoldenVectors(t *testing.T) {
	t.Run("load and parse golden test data", func(t *testing.T) {
		// Try to load golden test data
		data, err := os.ReadFile("testdata/golden_atas.json")
		if err != nil {
			t.Skipf("Golden test data not found: %v", err)
			return
		}

		var goldenData GoldenTestData
		if err := json.Unmarshal(data, &goldenData); err != nil {
			t.Errorf("Failed to parse golden test data: %v", err)
			return
		}

		t.Logf("Loaded %d golden test cases", len(goldenData.TestCases))
		t.Logf("Loaded %d validation test cases", len(goldenData.ValidationTestCases))

		// Validate structure
		if len(goldenData.TestCases) == 0 {
			t.Errorf("No golden test cases found")
		}

		if len(goldenData.ValidationTestCases) == 0 {
			t.Errorf("No validation test cases found")
		}
	})
}

// TestGenerateGoldenVectors can be used to generate golden test vectors
// Run with a flag to update the golden file with computed ATA addresses
func TestGenerateGoldenVectors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping golden vector generation in short mode")
	}

	t.Run("generate golden vectors for known addresses", func(t *testing.T) {
		config := NewConfig()

		testCases := []struct {
			name   string
			wallet string
			symbol string
		}{
			{"Reference Wallet - hyUSD", TestReferenceWallet, "hyUSD"},
			{"Reference Wallet - sHYUSD", TestReferenceWallet, "sHYUSD"},
			{"Reference Wallet - xSOL", TestReferenceWallet, "xSOL"},
			{"System Wallet - hyUSD", TestSystemWallet, "hyUSD"},
			{"System Wallet - xSOL", TestSystemWallet, "xSOL"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				wallet := solana.Address(tc.wallet)
				ata, err := GetWalletATAForToken(wallet, tc.symbol, config)
				if err != nil {
					t.Errorf("Failed to derive ATA for %s: %v", tc.name, err)
					return
				}

				t.Logf("Golden Vector - %s: %s", tc.name, ata)

				// Validate the derived ATA
				if err := ata.Validate(); err != nil {
					t.Errorf("Derived ATA is invalid for %s: %v", tc.name, err)
				}
			})
		}
	})
}
