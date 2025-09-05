package tokens

import (
	"encoding/binary"
	"testing"

	"hylo-wallet-tracker-api/internal/solana"
)

func TestParseSPLTokenAccount(t *testing.T) {
	tests := []struct {
		name        string
		accountInfo *solana.AccountInfo
		wantErr     bool
		errContains string
		validate    func(*testing.T, *SPLTokenAccount)
	}{
		{
			name: "Valid token account",
			accountInfo: &solana.AccountInfo{
				Owner:      SPLTokenProgramID,
				Data:       createValidTokenAccountData(),
				Lamports:   2039280,
				Executable: false,
				RentEpoch:  514,
			},
			wantErr: false,
			validate: func(t *testing.T, account *SPLTokenAccount) {
				// Check that mint and owner are valid addresses (all-zero bytes encode to shorter addresses)
				if len(account.Mint) == 0 {
					t.Errorf("Expected mint to be non-empty address")
				}
				if len(account.Owner) == 0 {
					t.Errorf("Expected owner to be non-empty address")
				}
				if account.Amount != 1000000 {
					t.Errorf("Expected amount 1000000, got %d", account.Amount)
				}
				if account.State != TokenStateInitialized {
					t.Errorf("Expected state Initialized, got %d", account.State)
				}
				if !account.IsInitialized {
					t.Errorf("Expected IsInitialized to be true")
				}
				if account.IsFrozen {
					t.Errorf("Expected IsFrozen to be false")
				}
			},
		},
		{
			name: "Frozen token account",
			accountInfo: &solana.AccountInfo{
				Owner: SPLTokenProgramID,
				Data:  createFrozenTokenAccountData(),
			},
			wantErr: false,
			validate: func(t *testing.T, account *SPLTokenAccount) {
				if account.State != TokenStateFrozen {
					t.Errorf("Expected state Frozen, got %d", account.State)
				}
				if !account.IsInitialized {
					t.Errorf("Expected IsInitialized to be true for frozen account")
				}
				if !account.IsFrozen {
					t.Errorf("Expected IsFrozen to be true")
				}
			},
		},
		{
			name: "Wrong owner",
			accountInfo: &solana.AccountInfo{
				Owner: "WrongProgramID111111111111111111111111",
				Data:  createValidTokenAccountData(),
			},
			wantErr:     true,
			errContains: "invalid token account owner",
		},
		{
			name: "Wrong data size",
			accountInfo: &solana.AccountInfo{
				Owner: SPLTokenProgramID,
				Data:  make([]byte, 100), // Wrong size
			},
			wantErr:     true,
			errContains: "invalid token account data size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := ParseSPLTokenAccount(tt.accountInfo)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if account == nil {
				t.Errorf("Expected account to be non-nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, account)
			}
		})
	}
}

func TestValidateTokenAccount(t *testing.T) {
	validAccount := &SPLTokenAccount{
		Mint:          "HyUSDMint1111111111111111111111111111111",
		Owner:         "WalletOwner111111111111111111111111111111",
		Amount:        1000000,
		State:         TokenStateInitialized,
		IsInitialized: true,
		IsFrozen:      false,
	}

	tests := []struct {
		name          string
		account       *SPLTokenAccount
		expectedMint  solana.Address
		expectedOwner solana.Address
		wantErr       bool
		errContains   string
	}{
		{
			name:          "Valid account",
			account:       validAccount,
			expectedMint:  "HyUSDMint1111111111111111111111111111111",
			expectedOwner: "WalletOwner111111111111111111111111111111",
			wantErr:       false,
		},
		{
			name: "Uninitialized account",
			account: &SPLTokenAccount{
				State:         TokenStateUninitialized,
				IsInitialized: false,
			},
			wantErr:     true,
			errContains: "token account is not initialized",
		},
		{
			name:         "Mint mismatch",
			account:      validAccount,
			expectedMint: "DifferentMint1111111111111111111111111111",
			wantErr:      true,
			errContains:  "mint mismatch",
		},
		{
			name:          "Owner mismatch",
			account:       validAccount,
			expectedOwner: "DifferentOwner111111111111111111111111111",
			wantErr:       true,
			errContains:   "owner mismatch",
		},
		{
			name:    "No validation required",
			account: validAccount,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTokenAccount(tt.account, tt.expectedMint, tt.expectedOwner)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSPLTokenAccountMethods(t *testing.T) {
	t.Run("IsZeroBalance", func(t *testing.T) {
		account := &SPLTokenAccount{Amount: 0}
		if !account.IsZeroBalance() {
			t.Errorf("Expected IsZeroBalance to return true for zero amount")
		}

		account.Amount = 1000
		if account.IsZeroBalance() {
			t.Errorf("Expected IsZeroBalance to return false for non-zero amount")
		}
	})

	t.Run("GetFormattedAmount", func(t *testing.T) {
		account := &SPLTokenAccount{Amount: 1500000}

		// Test with 6 decimals (hyUSD)
		formatted := account.GetFormattedAmount(6)
		expected := "1.5"
		if formatted != expected {
			t.Errorf("Expected formatted amount '%s', got '%s'", expected, formatted)
		}

		// Test with 9 decimals (SOL)
		account.Amount = 1000000000
		formatted = account.GetFormattedAmount(9)
		expected = "1"
		if formatted != expected {
			t.Errorf("Expected formatted amount '%s', got '%s'", expected, formatted)
		}

		// Test zero amount
		account.Amount = 0
		formatted = account.GetFormattedAmount(6)
		expected = "0"
		if formatted != expected {
			t.Errorf("Expected formatted amount '%s', got '%s'", expected, formatted)
		}
	})

	t.Run("String", func(t *testing.T) {
		account := &SPLTokenAccount{
			Mint:   "TestMint1111111111111111111111111111111",
			Owner:  "TestOwner111111111111111111111111111111",
			Amount: 1000,
			State:  TokenStateInitialized,
		}

		str := account.String()
		if !contains(str, "TestMint") || !contains(str, "TestOwner") || !contains(str, "1000") {
			t.Errorf("String representation missing expected components: %s", str)
		}
	})
}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTokenAmount(tt.amount, tt.decimals)
			if result != tt.expected {
				t.Errorf("formatTokenAmount(%d, %d) = %s, want %s",
					tt.amount, tt.decimals, result, tt.expected)
			}
		})
	}
}

func TestBytesToAddress(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		wantErr bool
	}{
		{
			name:    "Valid 32-byte address",
			bytes:   make([]byte, 32),
			wantErr: false,
		},
		{
			name:    "Invalid length - too short",
			bytes:   make([]byte, 20),
			wantErr: true,
		},
		{
			name:    "Invalid length - too long",
			bytes:   make([]byte, 40),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := bytesToAddress(tt.bytes)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if address == "" {
					t.Errorf("Expected non-empty address")
				}
			}
		})
	}
}

// Helper functions for test data creation

func createValidTokenAccountData() []byte {
	data := make([]byte, SPLTokenAccountSize)

	// Set mint (bytes 0-31) - valid base58 data
	copy(data[MintOffset:], make([]byte, 32))

	// Set owner (bytes 32-63) - valid base58 data
	copy(data[OwnerOffset:], make([]byte, 32))

	// Set amount (bytes 64-71) - 1,000,000 tokens
	binary.LittleEndian.PutUint64(data[AmountOffset:], 1000000)

	// Set state (byte 108) - Initialized
	data[StateOffset] = TokenStateInitialized

	return data
}

func createFrozenTokenAccountData() []byte {
	data := createValidTokenAccountData()

	// Set state to frozen
	data[StateOffset] = TokenStateFrozen

	return data
}

func createInvalidTokenAccountData() []byte {
	data := make([]byte, SPLTokenAccountSize)

	// Create data that will make encodeBase58 return empty string (fail)
	// Use a pattern that causes encoding failure - all zero bytes will succeed
	// but we need a pattern that returns empty from encodeBase58

	// First, create valid structure but with specific bytes that cause encoding issues
	copy(data[MintOffset:], make([]byte, 32))
	copy(data[OwnerOffset:], make([]byte, 32))

	// Set state to initialized
	data[StateOffset] = TokenStateInitialized

	return data
}

// Helper function to check if string contains substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || (len(str) > len(substr) &&
		(str[:len(substr)] == substr || str[len(str)-len(substr):] == substr ||
			containsHelper(str, substr))))
}

func containsHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
