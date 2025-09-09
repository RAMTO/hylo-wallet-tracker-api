package hylo

import (
	"testing"

	"hylo-wallet-tracker-api/internal/solana"
)

func TestParseTransaction(t *testing.T) {
	tests := []struct {
		name          string
		tx            *solana.TransactionDetails
		walletXSOLATA solana.Address
		expectedTrade *XSOLTrade
		expectedError string
		expectNoTrade bool
	}{
		{
			name: "successful BUY trade",
			tx: &solana.TransactionDetails{
				BlockTime: int64Ptr(1757360079),
				Slot:      365528388,
				Meta: &solana.TxMeta{
					Err: nil,
					PreTokenBalances: []solana.TokenBalance{
						{
							AccountIndex: 3,
							Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs", // xSOL mint
							UITokenAmount: &solana.UITokenAmount{
								Amount:   "1000000", // 1 xSOL (6 decimals)
								Decimals: 6,
							},
						},
					},
					PostTokenBalances: []solana.TokenBalance{
						{
							AccountIndex: 3,
							Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
							UITokenAmount: &solana.UITokenAmount{
								Amount:   "2500000", // 2.5 xSOL
								Decimals: 6,
							},
						},
					},
				},
				Transaction: solana.Transaction{
					Message: solana.TxMessage{
						AccountKeys: []string{
							"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g", // wallet
							"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",  // SPL token program
							"4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs", // xSOL mint
							"Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ", // xSOL ATA (index 3)
						},
					},
					Signatures: []string{"testSig123"},
				},
			},
			walletXSOLATA: "Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
			expectedTrade: &XSOLTrade{
				Signature:    "testSig123",
				Slot:         365528388,
				BlockTime:    1757360079,
				Side:         TradeSideBuy,
				XSOLAmount:   "1.5", // 1,500,000 / 1e6
				CounterAsset: "SOL",
			},
		},
		{
			name: "successful SELL trade",
			tx: &solana.TransactionDetails{
				BlockTime: int64Ptr(1757360079),
				Slot:      365528388,
				Meta: &solana.TxMeta{
					Err: nil,
					PreTokenBalances: []solana.TokenBalance{
						{
							AccountIndex: 3,
							Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
							UITokenAmount: &solana.UITokenAmount{
								Amount:   "5000000", // 5 xSOL
								Decimals: 6,
							},
						},
					},
					PostTokenBalances: []solana.TokenBalance{
						{
							AccountIndex: 3,
							Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
							UITokenAmount: &solana.UITokenAmount{
								Amount:   "2500000", // 2.5 xSOL
								Decimals: 6,
							},
						},
					},
				},
				Transaction: solana.Transaction{
					Message: solana.TxMessage{
						AccountKeys: []string{
							"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
							"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
							"4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
							"Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
						},
					},
					Signatures: []string{"testSig456"},
				},
			},
			walletXSOLATA: "Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
			expectedTrade: &XSOLTrade{
				Signature:    "testSig456",
				Slot:         365528388,
				BlockTime:    1757360079,
				Side:         TradeSideSell,
				XSOLAmount:   "2.5", // 2,500,000 / 1e6
				CounterAsset: "SOL",
			},
		},
		{
			name: "no balance change - should not be a trade",
			tx: &solana.TransactionDetails{
				BlockTime: int64Ptr(1757360079),
				Slot:      365528388,
				Meta: &solana.TxMeta{
					Err: nil,
					PreTokenBalances: []solana.TokenBalance{
						{
							AccountIndex: 3,
							Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
							UITokenAmount: &solana.UITokenAmount{
								Amount:   "1000000",
								Decimals: 6,
							},
						},
					},
					PostTokenBalances: []solana.TokenBalance{
						{
							AccountIndex: 3,
							Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
							UITokenAmount: &solana.UITokenAmount{
								Amount:   "1000000", // Same amount
								Decimals: 6,
							},
						},
					},
				},
				Transaction: solana.Transaction{
					Message: solana.TxMessage{
						AccountKeys: []string{
							"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
							"Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
						},
					},
					Signatures: []string{"testSig789"},
				},
			},
			walletXSOLATA: "Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
			expectNoTrade: true,
		},
		{
			name: "failed transaction - should not be a trade",
			tx: &solana.TransactionDetails{
				BlockTime: int64Ptr(1757360079),
				Slot:      365528388,
				Meta: &solana.TxMeta{
					Err: "InstructionError", // Failed transaction
					PreTokenBalances: []solana.TokenBalance{
						{
							AccountIndex: 1,
							Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
							UITokenAmount: &solana.UITokenAmount{
								Amount:   "1000000",
								Decimals: 6,
							},
						},
					},
					PostTokenBalances: []solana.TokenBalance{
						{
							AccountIndex: 1,
							Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
							UITokenAmount: &solana.UITokenAmount{
								Amount:   "2000000",
								Decimals: 6,
							},
						},
					},
				},
				Transaction: solana.Transaction{
					Message: solana.TxMessage{
						AccountKeys: []string{
							"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
							"Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
						},
					},
					Signatures: []string{"testSigFailed"},
				},
			},
			walletXSOLATA: "Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
			expectNoTrade: true,
		},
		{
			name: "xSOL ATA not in transaction - should not be a trade",
			tx: &solana.TransactionDetails{
				BlockTime: int64Ptr(1757360079),
				Slot:      365528388,
				Meta: &solana.TxMeta{
					Err:               nil,
					PreTokenBalances:  []solana.TokenBalance{},
					PostTokenBalances: []solana.TokenBalance{},
				},
				Transaction: solana.Transaction{
					Message: solana.TxMessage{
						AccountKeys: []string{
							"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
							"OtherAccountNotXSOL11111111111111111111111",
						},
					},
					Signatures: []string{"testSigNoXSOL"},
				},
			},
			walletXSOLATA: "Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
			expectNoTrade: true,
		},
		{
			name:          "nil transaction",
			tx:            nil,
			walletXSOLATA: "Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
			expectedError: "transaction details cannot be nil",
		},
		{
			name: "nil transaction meta",
			tx: &solana.TransactionDetails{
				Meta: nil,
			},
			walletXSOLATA: "Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ",
			expectedError: "transaction metadata cannot be nil",
		},
		{
			name: "empty wallet ATA",
			tx: &solana.TransactionDetails{
				Meta: &solana.TxMeta{},
			},
			walletXSOLATA: "",
			expectedError: "wallet xSOL ATA address cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTransaction(tt.tx, tt.walletXSOLATA)

			// Check for expected errors
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			// Check for no error when not expected
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check for no trade cases
			if tt.expectNoTrade {
				if result == nil {
					t.Error("expected TradeParseResult, got nil")
					return
				}
				if result.Trade != nil {
					t.Errorf("expected no trade, got: %+v", result.Trade)
				}
				return
			}

			// Check for successful trade parsing
			if result == nil || result.Trade == nil {
				t.Error("expected trade result, got nil")
				return
			}

			trade := result.Trade
			expected := tt.expectedTrade

			if trade.Signature != expected.Signature {
				t.Errorf("expected signature %q, got %q", expected.Signature, trade.Signature)
			}
			if trade.Slot != expected.Slot {
				t.Errorf("expected slot %d, got %d", expected.Slot, trade.Slot)
			}
			if trade.BlockTime != expected.BlockTime {
				t.Errorf("expected blockTime %d, got %d", expected.BlockTime, trade.BlockTime)
			}
			if trade.Side != expected.Side {
				t.Errorf("expected side %q, got %q", expected.Side, trade.Side)
			}
			if trade.XSOLAmount != expected.XSOLAmount {
				t.Errorf("expected xSOL amount %q, got %q", expected.XSOLAmount, trade.XSOLAmount)
			}
			if trade.CounterAsset != expected.CounterAsset {
				t.Errorf("expected counter asset %q, got %q", expected.CounterAsset, trade.CounterAsset)
			}

			// Verify timestamp is set correctly
			if trade.Timestamp.IsZero() && trade.BlockTime > 0 {
				t.Error("expected timestamp to be set for valid blockTime")
			}

			// Verify explorer URL is generated
			expectedURL := "https://solscan.io/tx/" + trade.Signature
			if trade.ExplorerURL != expectedURL {
				t.Errorf("expected explorer URL %q, got %q", expectedURL, trade.ExplorerURL)
			}
		})
	}
}

func TestFindAccountIndex(t *testing.T) {
	accountKeys := []string{
		"11111111111111111111111111111111",
		"A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
		"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
	}

	tests := []struct {
		name          string
		targetAccount string
		expectedIndex int
	}{
		{
			name:          "find existing account",
			targetAccount: "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
			expectedIndex: 1,
		},
		{
			name:          "find first account",
			targetAccount: "11111111111111111111111111111111",
			expectedIndex: 0,
		},
		{
			name:          "find last account",
			targetAccount: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
			expectedIndex: 2,
		},
		{
			name:          "account not found",
			targetAccount: "NotFoundAccount1111111111111111111111",
			expectedIndex: -1,
		},
		{
			name:          "empty target",
			targetAccount: "",
			expectedIndex: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := findAccountIndex(accountKeys, tt.targetAccount)
			if index != tt.expectedIndex {
				t.Errorf("expected index %d, got %d", tt.expectedIndex, index)
			}
		})
	}
}

func TestFindTokenBalance(t *testing.T) {
	tokenBalances := []solana.TokenBalance{
		{
			AccountIndex: 0,
			Mint:         "mint1",
			UITokenAmount: &solana.UITokenAmount{
				Amount:   "100",
				Decimals: 6,
			},
		},
		{
			AccountIndex: 2,
			Mint:         "mint2",
			UITokenAmount: &solana.UITokenAmount{
				Amount:   "200",
				Decimals: 9,
			},
		},
	}

	tests := []struct {
		name         string
		accountIndex uint32
		expectFound  bool
		expectedMint string
	}{
		{
			name:         "find existing balance",
			accountIndex: 0,
			expectFound:  true,
			expectedMint: "mint1",
		},
		{
			name:         "find another existing balance",
			accountIndex: 2,
			expectFound:  true,
			expectedMint: "mint2",
		},
		{
			name:         "balance not found",
			accountIndex: 5,
			expectFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findTokenBalance(tokenBalances, tt.accountIndex)

			if tt.expectFound {
				if result == nil {
					t.Error("expected to find token balance, got nil")
					return
				}
				if result.AccountIndex != tt.accountIndex {
					t.Errorf("expected account index %d, got %d", tt.accountIndex, result.AccountIndex)
				}
				if result.Mint != tt.expectedMint {
					t.Errorf("expected mint %q, got %q", tt.expectedMint, result.Mint)
				}
			} else {
				if result != nil {
					t.Errorf("expected nil, got: %+v", result)
				}
			}
		})
	}
}

func TestParseTokenAmount(t *testing.T) {
	tests := []struct {
		name           string
		uiTokenAmount  *solana.UITokenAmount
		expectedAmount uint64
		expectedError  string
	}{
		{
			name: "valid amount",
			uiTokenAmount: &solana.UITokenAmount{
				Amount:   "1000000",
				Decimals: 6,
			},
			expectedAmount: 1000000,
		},
		{
			name: "zero amount",
			uiTokenAmount: &solana.UITokenAmount{
				Amount:   "0",
				Decimals: 6,
			},
			expectedAmount: 0,
		},
		{
			name: "large amount",
			uiTokenAmount: &solana.UITokenAmount{
				Amount:   "18446744073709551615", // Max uint64
				Decimals: 18,
			},
			expectedAmount: 18446744073709551615,
		},
		{
			name:          "nil UITokenAmount",
			uiTokenAmount: nil,
			expectedError: "UITokenAmount is nil",
		},
		{
			name: "invalid amount string",
			uiTokenAmount: &solana.UITokenAmount{
				Amount:   "not-a-number",
				Decimals: 6,
			},
			expectedError: "failed to parse amount 'not-a-number':",
		},
		{
			name: "negative amount",
			uiTokenAmount: &solana.UITokenAmount{
				Amount:   "-1000",
				Decimals: 6,
			},
			expectedError: "failed to parse amount '-1000':",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := parseTokenAmount(tt.uiTokenAmount)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if !contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if amount != tt.expectedAmount {
				t.Errorf("expected amount %d, got %d", tt.expectedAmount, amount)
			}
		})
	}
}

func TestValidateTradeTransaction(t *testing.T) {
	validTrade := &XSOLTrade{
		Signature:    "validSig",
		Side:         TradeSideBuy,
		XSOLAmount:   "1.0",
		CounterAsset: "SOL",
	}
	validTrade.XSOLAmountRaw = 1000000

	tests := []struct {
		name          string
		trade         *XSOLTrade
		tx            *solana.TransactionDetails
		expectedError string
	}{
		{
			name:  "valid trade",
			trade: validTrade,
			tx: &solana.TransactionDetails{
				Meta: &solana.TxMeta{},
				Transaction: solana.Transaction{
					Message: solana.TxMessage{},
				},
			},
		},
		{
			name:          "nil trade",
			trade:         nil,
			tx:            &solana.TransactionDetails{},
			expectedError: "trade cannot be nil",
		},
		{
			name: "invalid trade data",
			trade: &XSOLTrade{
				Signature: "", // Missing required field
			},
			tx:            &solana.TransactionDetails{},
			expectedError: "invalid trade data: missing required fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTradeTransaction(tt.trade, tt.tx)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helper functions for tests
func int64Ptr(v int64) *int64 {
	return &v
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexContains(s, substr) >= 0)))
}

func indexContains(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
