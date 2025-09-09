package trades

import (
	"context"
	"errors"
	"testing"

	"hylo-wallet-tracker-api/internal/hylo"
	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/tokens"
)

// mockHTTPClient implements HTTPClientInterface for testing
type mockHTTPClient struct {
	getAccountFunc              func(ctx context.Context, address solana.Address, commitment solana.Commitment) (*solana.AccountInfo, error)
	getSignaturesForAddressFunc func(ctx context.Context, address solana.Address, before string, limit int) ([]solana.SignatureInfo, error)
	getTransactionFunc          func(ctx context.Context, signature solana.Signature) (*solana.TransactionDetails, error)
}

func (m *mockHTTPClient) GetAccount(ctx context.Context, address solana.Address, commitment solana.Commitment) (*solana.AccountInfo, error) {
	if m.getAccountFunc != nil {
		return m.getAccountFunc(ctx, address, commitment)
	}
	return &solana.AccountInfo{}, nil
}

func (m *mockHTTPClient) GetSignaturesForAddress(ctx context.Context, address solana.Address, before string, limit int) ([]solana.SignatureInfo, error) {
	if m.getSignaturesForAddressFunc != nil {
		return m.getSignaturesForAddressFunc(ctx, address, before, limit)
	}
	return []solana.SignatureInfo{}, nil
}

func (m *mockHTTPClient) GetTransaction(ctx context.Context, signature solana.Signature) (*solana.TransactionDetails, error) {
	if m.getTransactionFunc != nil {
		return m.getTransactionFunc(ctx, signature)
	}
	return &solana.TransactionDetails{}, nil
}

func TestNewTradeService(t *testing.T) {
	validClient := &mockHTTPClient{}
	validTokenConfig := tokens.NewConfig()
	validHyloConfig := hylo.NewConfig()

	tests := []struct {
		name        string
		httpClient  HTTPClientInterface
		tokenConfig *tokens.Config
		hyloConfig  *hylo.Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid configuration",
			httpClient:  validClient,
			tokenConfig: validTokenConfig,
			hyloConfig:  validHyloConfig,
			expectError: false,
		},
		{
			name:        "nil http client",
			httpClient:  nil,
			tokenConfig: validTokenConfig,
			hyloConfig:  validHyloConfig,
			expectError: true,
			errorMsg:    "httpClient cannot be nil",
		},
		{
			name:        "nil token config",
			httpClient:  validClient,
			tokenConfig: nil,
			hyloConfig:  validHyloConfig,
			expectError: true,
			errorMsg:    "tokenConfig cannot be nil",
		},
		{
			name:        "nil hylo config",
			httpClient:  validClient,
			tokenConfig: validTokenConfig,
			hyloConfig:  nil,
			expectError: true,
			errorMsg:    "hyloConfig cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewTradeService(tt.httpClient, tt.tokenConfig, tt.hyloConfig)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error %q, got %q", tt.errorMsg, err.Error())
				}
				if service != nil {
					t.Error("expected nil service on error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if service == nil {
				t.Error("expected service, got nil")
				return
			}

			// Verify service is properly initialized
			if service.httpClient == nil {
				t.Error("httpClient not set")
			}
			if service.tokenConfig == nil {
				t.Error("tokenConfig not set")
			}
			if service.hyloConfig == nil {
				t.Error("hyloConfig not set")
			}
			if service.options == nil {
				t.Error("options not set")
			}
		})
	}
}

func TestGetWalletTrades(t *testing.T) {
	// Sample test data
	testWallet := solana.Address("A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g")
	testXSOLATA := solana.Address("Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ")

	tests := []struct {
		name                 string
		walletAddr           solana.Address
		limit                int
		before               string
		mockSignatures       []solana.SignatureInfo
		mockTransactions     map[string]*solana.TransactionDetails
		mockSignaturesError  error
		mockTransactionError error
		expectedTradeCount   int
		expectError          bool
		errorMsg             string
	}{
		{
			name:       "successful trade fetch",
			walletAddr: testWallet,
			limit:      10,
			before:     "",
			mockSignatures: []solana.SignatureInfo{
				{
					Signature: "sig1",
					Slot:      365528388,
					BlockTime: int64Ptr(1757360079),
					Err:       nil,
				},
				{
					Signature: "sig2",
					Slot:      365528387,
					BlockTime: int64Ptr(1757360078),
					Err:       nil,
				},
			},
			mockTransactions: map[string]*solana.TransactionDetails{
				"sig1": createMockTradeTransaction("sig1", 365528388, 1757360079, testXSOLATA, "1000000", "2000000", hylo.TradeSideBuy),
				"sig2": createMockTradeTransaction("sig2", 365528387, 1757360078, testXSOLATA, "3000000", "1500000", hylo.TradeSideSell),
			},
			expectedTradeCount: 2,
		},
		{
			name:               "no signatures found",
			walletAddr:         testWallet,
			limit:              10,
			before:             "",
			mockSignatures:     []solana.SignatureInfo{},
			mockTransactions:   map[string]*solana.TransactionDetails{},
			expectedTradeCount: 0,
		},
		{
			name:       "signatures with failed transactions",
			walletAddr: testWallet,
			limit:      10,
			before:     "",
			mockSignatures: []solana.SignatureInfo{
				{
					Signature: "sigFailed",
					Slot:      365528388,
					BlockTime: int64Ptr(1757360079),
					Err:       "InstructionError",
				},
				{
					Signature: "sigSuccess",
					Slot:      365528387,
					BlockTime: int64Ptr(1757360078),
					Err:       nil,
				},
			},
			mockTransactions: map[string]*solana.TransactionDetails{
				"sigSuccess": createMockTradeTransaction("sigSuccess", 365528387, 1757360078, testXSOLATA, "1000000", "2000000", hylo.TradeSideBuy),
			},
			expectedTradeCount: 1, // Only the successful one
		},
		{
			name:        "invalid wallet address",
			walletAddr:  "invalid-address",
			limit:       10,
			before:      "",
			expectError: true,
		},
		{
			name:                "signature fetch error",
			walletAddr:          testWallet,
			limit:               10,
			before:              "",
			mockSignaturesError: errors.New("RPC error"),
			expectError:         true,
			errorMsg:            "failed to fetch transaction signatures",
		},
		{
			name:       "limit enforcement",
			walletAddr: testWallet,
			limit:      1, // Only want 1 trade
			before:     "",
			mockSignatures: []solana.SignatureInfo{
				{
					Signature: "sig1",
					Slot:      365528388,
					BlockTime: int64Ptr(1757360079),
					Err:       nil,
				},
				{
					Signature: "sig2",
					Slot:      365528387,
					BlockTime: int64Ptr(1757360078),
					Err:       nil,
				},
			},
			mockTransactions: map[string]*solana.TransactionDetails{
				"sig1": createMockTradeTransaction("sig1", 365528388, 1757360079, testXSOLATA, "1000000", "2000000", hylo.TradeSideBuy),
				"sig2": createMockTradeTransaction("sig2", 365528387, 1757360078, testXSOLATA, "3000000", "1500000", hylo.TradeSideSell),
			},
			expectedTradeCount: 1, // Should only return 1 despite 2 available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockHTTPClient{
				getSignaturesForAddressFunc: func(ctx context.Context, address solana.Address, before string, limit int) ([]solana.SignatureInfo, error) {
					if tt.mockSignaturesError != nil {
						return nil, tt.mockSignaturesError
					}
					return tt.mockSignatures, nil
				},
				getTransactionFunc: func(ctx context.Context, signature solana.Signature) (*solana.TransactionDetails, error) {
					if tt.mockTransactionError != nil {
						return nil, tt.mockTransactionError
					}
					if tx, ok := tt.mockTransactions[string(signature)]; ok {
						return tx, nil
					}
					return nil, errors.New("transaction not found")
				},
			}

			// Create service
			tokenConfig := tokens.NewConfig()
			hyloConfig := hylo.NewConfig()
			service, err := NewTradeService(mockClient, tokenConfig, hyloConfig)
			if err != nil {
				t.Fatalf("failed to create service: %v", err)
			}

			// Test GetWalletTrades
			ctx := context.Background()
			result, err := service.GetWalletTrades(ctx, tt.walletAddr, tt.limit, tt.before)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result, got nil")
				return
			}

			// Verify trade count
			if len(result.Trades) != tt.expectedTradeCount {
				t.Errorf("expected %d trades, got %d", tt.expectedTradeCount, len(result.Trades))
			}

			// Verify response structure
			if result.WalletAddress != string(tt.walletAddr) {
				t.Errorf("expected wallet address %q, got %q", tt.walletAddr, result.WalletAddress)
			}

			if result.Count != len(result.Trades) {
				t.Errorf("expected count %d, got %d", len(result.Trades), result.Count)
			}

			if result.Pagination.Count != len(result.Trades) {
				t.Errorf("expected pagination count %d, got %d", len(result.Trades), result.Pagination.Count)
			}

			if result.Pagination.Limit != tt.limit {
				t.Errorf("expected pagination limit %d, got %d", tt.limit, result.Pagination.Limit)
			}

			// Verify trades are sorted by slot (newest first)
			for i := 1; i < len(result.Trades); i++ {
				if result.Trades[i-1].Slot < result.Trades[i].Slot {
					t.Error("trades are not sorted by slot (newest first)")
					break
				}
			}
		})
	}
}

func TestProcessSignatures(t *testing.T) {
	testXSOLATA := solana.Address("Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ")

	signatures := []solana.SignatureInfo{
		{
			Signature: "sig1",
			Slot:      365528388,
			BlockTime: int64Ptr(1757360079),
			Err:       nil,
		},
		{
			Signature: "sigFailed",
			Slot:      365528389,
			BlockTime: int64Ptr(1757360080),
			Err:       "InstructionError", // This should be skipped
		},
		{
			Signature: "sig2",
			Slot:      365528387,
			BlockTime: int64Ptr(1757360078),
			Err:       nil,
		},
	}

	mockTransactions := map[string]*solana.TransactionDetails{
		"sig1": createMockTradeTransaction("sig1", 365528388, 1757360079, testXSOLATA, "1000000", "2000000", hylo.TradeSideBuy),
		"sig2": createMockTradeTransaction("sig2", 365528387, 1757360078, testXSOLATA, "3000000", "1500000", hylo.TradeSideSell),
	}

	mockClient := &mockHTTPClient{
		getTransactionFunc: func(ctx context.Context, signature solana.Signature) (*solana.TransactionDetails, error) {
			if tx, ok := mockTransactions[string(signature)]; ok {
				return tx, nil
			}
			return nil, errors.New("transaction not found")
		},
	}

	// Create service
	tokenConfig := tokens.NewConfig()
	hyloConfig := hylo.NewConfig()
	service, err := NewTradeService(mockClient, tokenConfig, hyloConfig)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	ctx := context.Background()
	trades, err := service.processSignatures(ctx, signatures, testXSOLATA, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Should get 2 trades (sig1 and sig2), sigFailed should be skipped
	expectedCount := 2
	if len(trades) != expectedCount {
		t.Errorf("expected %d trades, got %d", expectedCount, len(trades))
	}

	// Verify trades are sorted by slot (newest first)
	if len(trades) >= 2 && trades[0].Slot < trades[1].Slot {
		t.Error("trades are not sorted by slot (newest first)")
	}

	// Verify trade content
	if len(trades) > 0 {
		trade := trades[0]
		if trade.Signature == "" {
			t.Error("trade signature is empty")
		}
		if trade.Side != hylo.TradeSideBuy && trade.Side != hylo.TradeSideSell {
			t.Errorf("invalid trade side: %s", trade.Side)
		}
	}
}

func TestGetServiceHealth(t *testing.T) {
	mockClient := &mockHTTPClient{}
	tokenConfig := tokens.NewConfig()
	hyloConfig := hylo.NewConfig()

	service, err := NewTradeService(mockClient, tokenConfig, hyloConfig)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	ctx := context.Background()
	health := service.GetServiceHealth(ctx)

	// Verify health response structure
	if health["service"] != "TradeService" {
		t.Errorf("expected service name 'TradeService', got %v", health["service"])
	}

	if health["status"] == nil {
		t.Error("status field is missing")
	}

	if health["config"] == nil {
		t.Error("config field is missing")
	}

	// Verify config contains expected fields
	config, ok := health["config"].(map[string]interface{})
	if !ok {
		t.Error("config is not a map")
		return
	}

	if config["default_limit"] == nil {
		t.Error("default_limit is missing from config")
	}
	if config["max_limit"] == nil {
		t.Error("max_limit is missing from config")
	}
	if config["hylo_programs"] == nil {
		t.Error("hylo_programs is missing from config")
	}
}

func TestValidateTradeRequest(t *testing.T) {
	options := DefaultTradeServiceOptions()

	tests := []struct {
		name        string
		req         *TradeRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			req: &TradeRequest{
				WalletAddress: "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
				Limit:         10,
				Before:        "",
			},
			expectError: false,
		},
		{
			name: "empty wallet address",
			req: &TradeRequest{
				WalletAddress: "",
				Limit:         10,
			},
			expectError: true,
		},
		{
			name: "zero limit gets default",
			req: &TradeRequest{
				WalletAddress: "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
				Limit:         0,
			},
			expectError: false,
		},
		{
			name: "limit exceeds max",
			req: &TradeRequest{
				WalletAddress: "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g",
				Limit:         100, // Exceeds default max of 50
			},
			expectError: false, // Should be capped to max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLimit := tt.req.Limit
			err := ValidateTradeRequest(tt.req, options)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify limit enforcement
			if originalLimit == 0 && tt.req.Limit != options.DefaultLimit {
				t.Errorf("expected default limit %d, got %d", options.DefaultLimit, tt.req.Limit)
			}
			if originalLimit > options.MaxLimit && tt.req.Limit != options.MaxLimit {
				t.Errorf("expected max limit %d, got %d", options.MaxLimit, tt.req.Limit)
			}
		})
	}
}

// Helper functions for tests

func createMockTradeTransaction(signature string, slot uint64, blockTime int64, xsolATA solana.Address, preAmount, postAmount, side string) *solana.TransactionDetails {
	return &solana.TransactionDetails{
		BlockTime: &blockTime,
		Slot:      solana.Slot(slot),
		Meta: &solana.TxMeta{
			Err: nil,
			PreTokenBalances: []solana.TokenBalance{
				{
					AccountIndex: 3, // xSOL ATA is at index 3
					Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
					UITokenAmount: &solana.UITokenAmount{
						Amount:   preAmount,
						Decimals: 6,
					},
				},
			},
			PostTokenBalances: []solana.TokenBalance{
				{
					AccountIndex: 3,
					Mint:         "4sWNB8zGWHkh6UnmwiEtzNxL4XrN7uK9tosbESbJFfVs",
					UITokenAmount: &solana.UITokenAmount{
						Amount:   postAmount,
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
					string(xsolATA), // xSOL ATA (index 3)
				},
			},
			Signatures: []string{signature},
		},
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
