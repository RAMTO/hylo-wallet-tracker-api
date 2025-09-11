package tokens

import (
	"os"
	"testing"

	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/utils"
)

func TestTokenConstants(t *testing.T) {
	t.Run("token mint addresses are valid", func(t *testing.T) {
		// Test that all token mint addresses are valid Solana addresses (44 characters)
		if err := HyUSDMint.Validate(); err != nil {
			t.Errorf("HyUSDMint validation failed: %v", err)
		}

		if err := SHyUSDMint.Validate(); err != nil {
			t.Errorf("SHyUSDMint validation failed: %v", err)
		}

		if err := XSOLMint.Validate(); err != nil {
			t.Errorf("XSOLMint validation failed: %v", err)
		}

		if err := USDCMint.Validate(); err != nil {
			t.Errorf("USDCMint validation failed: %v", err)
		}

		if err := JitoSOLMint.Validate(); err != nil {
			t.Errorf("JitoSOLMint validation failed: %v", err)
		}
	})

	t.Run("token mint addresses match documentation", func(t *testing.T) {
		// Verify against known addresses from Hylo documentation
		expectedHyUSD := string(HyUSDMint)
		expectedSHyUSD := string(SHyUSDMint)
		expectedXSOL := string(XSOLMint)

		if string(HyUSDMint) != expectedHyUSD {
			t.Errorf("HyUSDMint mismatch: expected %s, got %s", expectedHyUSD, HyUSDMint)
		}

		if string(SHyUSDMint) != expectedSHyUSD {
			t.Errorf("SHyUSDMint mismatch: expected %s, got %s", expectedSHyUSD, SHyUSDMint)
		}

		if string(XSOLMint) != expectedXSOL {
			t.Errorf("XSOLMint mismatch: expected %s, got %s", expectedXSOL, XSOLMint)
		}
	})

	t.Run("supported tokens list", func(t *testing.T) {
		supportedMints := GetSupportedTokenMints()

		if len(supportedMints) != 5 {
			t.Errorf("Expected 5 supported mints, got %d", len(supportedMints))
		}

		// Verify all expected mints are included
		mintMap := make(map[solana.Address]bool)
		for _, mint := range supportedMints {
			mintMap[mint] = true
		}

		if !mintMap[HyUSDMint] {
			t.Errorf("HyUSDMint not found in supported mints")
		}

		if !mintMap[SHyUSDMint] {
			t.Errorf("SHyUSDMint not found in supported mints")
		}

		if !mintMap[XSOLMint] {
			t.Errorf("XSOLMint not found in supported mints")
		}

		if !mintMap[USDCMint] {
			t.Errorf("USDCMint not found in supported mints")
		}

		if !mintMap[JitoSOLMint] {
			t.Errorf("JitoSOLMint not found in supported mints")
		}
	})
}

func TestTokenHelperFunctions(t *testing.T) {
	t.Run("IsValidTokenMint", func(t *testing.T) {
		// Valid mints
		if !IsValidTokenMint(HyUSDMint) {
			t.Errorf("HyUSDMint should be valid")
		}

		if !IsValidTokenMint(SHyUSDMint) {
			t.Errorf("SHyUSDMint should be valid")
		}

		if !IsValidTokenMint(XSOLMint) {
			t.Errorf("XSOLMint should be valid")
		}

		if !IsValidTokenMint(USDCMint) {
			t.Errorf("USDCMint should be valid")
		}

		if !IsValidTokenMint(JitoSOLMint) {
			t.Errorf("JitoSOLMint should be valid")
		}

		// Invalid mint
		invalidMint := solana.Address(TestInvalidMint)
		if IsValidTokenMint(invalidMint) {
			t.Errorf("Invalid mint should not be valid")
		}
	})

	t.Run("GetTokenSymbol", func(t *testing.T) {
		testCases := []struct {
			mint           solana.Address
			expectedSymbol string
		}{
			{HyUSDMint, HyUSDSymbol},
			{SHyUSDMint, SHyUSDSymbol},
			{XSOLMint, XSOLSymbol},
			{USDCMint, USDCSymbol},
			{JitoSOLMint, JitoSOLSymbol},
			{solana.Address(TestInvalidMint), ""},
		}

		for _, tc := range testCases {
			symbol := GetTokenSymbol(tc.mint)
			if symbol != tc.expectedSymbol {
				t.Errorf("GetTokenSymbol(%s): expected %s, got %s",
					tc.mint, tc.expectedSymbol, symbol)
			}
		}
	})

	t.Run("GetTokenDecimals", func(t *testing.T) {
		testCases := []struct {
			mint             solana.Address
			expectedDecimals uint8
		}{
			{HyUSDMint, HyUSDDecimals},
			{SHyUSDMint, SHyUSDDecimals},
			{XSOLMint, XSOLDecimals},
			{USDCMint, USDCDecimals},
			{JitoSOLMint, JitoSOLDecimals},
			{solana.Address(TestInvalidMint), 0},
		}

		for _, tc := range testCases {
			decimals := GetTokenDecimals(tc.mint)
			if decimals != tc.expectedDecimals {
				t.Errorf("GetTokenDecimals(%s): expected %d, got %d",
					tc.mint, tc.expectedDecimals, decimals)
			}
		}
	})

	t.Run("GetTokenName", func(t *testing.T) {
		testCases := []struct {
			mint         solana.Address
			expectedName string
		}{
			{HyUSDMint, HyUSDName},
			{SHyUSDMint, SHyUSDName},
			{XSOLMint, XSOLName},
			{USDCMint, USDCName},
			{JitoSOLMint, JitoSOLName},
			{solana.Address(TestInvalidMint), ""},
		}

		for _, tc := range testCases {
			name := GetTokenName(tc.mint)
			if name != tc.expectedName {
				t.Errorf("GetTokenName(%s): expected %s, got %s",
					tc.mint, tc.expectedName, name)
			}
		}
	})
}

func TestTokenInfo(t *testing.T) {
	t.Run("valid token info", func(t *testing.T) {
		tokenInfo := TokenInfo{
			Mint:     HyUSDMint,
			Symbol:   HyUSDSymbol,
			Name:     HyUSDName,
			Decimals: HyUSDDecimals,
		}

		if err := tokenInfo.Validate(); err != nil {
			t.Errorf("Valid token info should not have validation error: %v", err)
		}
	})

	t.Run("invalid token info validation", func(t *testing.T) {
		testCases := []struct {
			name      string
			tokenInfo TokenInfo
			shouldErr bool
		}{
			{
				name: "invalid mint address",
				tokenInfo: TokenInfo{
					Mint:     solana.Address(TestTooShortMint),
					Symbol:   "TEST",
					Name:     "Test Token",
					Decimals: 6,
				},
				shouldErr: true,
			},
			{
				name: "empty symbol",
				tokenInfo: TokenInfo{
					Mint:     HyUSDMint,
					Symbol:   "",
					Name:     "Test Token",
					Decimals: 6,
				},
				shouldErr: true,
			},
			{
				name: "empty name",
				tokenInfo: TokenInfo{
					Mint:     HyUSDMint,
					Symbol:   "TEST",
					Name:     "",
					Decimals: 6,
				},
				shouldErr: true,
			},
			{
				name: "excessive decimals",
				tokenInfo: TokenInfo{
					Mint:     HyUSDMint,
					Symbol:   "TEST",
					Name:     "Test Token",
					Decimals: 25,
				},
				shouldErr: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.tokenInfo.Validate()
				if tc.shouldErr && err == nil {
					t.Errorf("Expected validation error for %s", tc.name)
				}
				if !tc.shouldErr && err != nil {
					t.Errorf("Unexpected validation error for %s: %v", tc.name, err)
				}
			})
		}
	})
}

func TestTokenBalance(t *testing.T) {
	t.Run("token balance creation and formatting", func(t *testing.T) {
		tokenInfo := TokenInfo{
			Mint:     HyUSDMint,
			Symbol:   HyUSDSymbol,
			Name:     HyUSDName,
			Decimals: HyUSDDecimals, // 6 decimals
		}

		testCases := []struct {
			rawAmount       uint64
			expectedDecimal string
		}{
			{0, "0"},
			{1, "0.000001"},
			{1000, "0.001"},
			{1000000, "1"},
			{1500000, "1.5"},
			{1234567, "1.234567"},
			{1000000000, "1000"},
			{1234567890, "1234.56789"},
		}

		for _, tc := range testCases {
			balance := NewTokenBalance(tokenInfo, tc.rawAmount)
			if balance.FormattedAmount != tc.expectedDecimal {
				t.Errorf("Raw amount %d: expected %s, got %s",
					tc.rawAmount, tc.expectedDecimal, balance.FormattedAmount)
			}

			if balance.RawAmount != tc.rawAmount {
				t.Errorf("Raw amount mismatch: expected %d, got %d",
					tc.rawAmount, balance.RawAmount)
			}
		}
	})

	t.Run("zero balance detection", func(t *testing.T) {
		tokenInfo := TokenInfo{
			Mint:     HyUSDMint,
			Symbol:   HyUSDSymbol,
			Name:     HyUSDName,
			Decimals: HyUSDDecimals,
		}

		zeroBalance := NewTokenBalance(tokenInfo, 0)
		if !zeroBalance.IsZero() {
			t.Errorf("Zero balance should be detected as zero")
		}

		nonZeroBalance := NewTokenBalance(tokenInfo, 1000000)
		if nonZeroBalance.IsZero() {
			t.Errorf("Non-zero balance should not be detected as zero")
		}
	})

	t.Run("USD value setting", func(t *testing.T) {
		tokenInfo := TokenInfo{
			Mint:     HyUSDMint,
			Symbol:   HyUSDSymbol,
			Name:     HyUSDName,
			Decimals: HyUSDDecimals,
		}

		balance := NewTokenBalance(tokenInfo, 1000000) // 1.0 hyUSD

		// Initially no USD value
		if balance.USDValue != nil {
			t.Errorf("Initial USD value should be nil")
		}

		// Set USD value
		balance.SetUSDValue(1.0)
		if balance.USDValue == nil || *balance.USDValue != 1.0 {
			t.Errorf("USD value should be set to 1.0")
		}
	})
}

func TestParseDecimalAmount(t *testing.T) {
	t.Run("valid decimal parsing", func(t *testing.T) {
		testCases := []struct {
			decimalStr  string
			decimals    uint8
			expectedRaw uint64
		}{
			{"0", 6, 0},
			{"", 6, 0},
			{"1", 6, 1000000},
			{"1.5", 6, 1500000},
			{"1.234567", 6, 1234567},
			{"1000", 6, 1000000000},
			{"0.000001", 6, 1},
			{"123.456", 3, 123456},
		}

		for _, tc := range testCases {
			rawAmount, err := utils.ParseDecimalAmount(tc.decimalStr, tc.decimals)
			if err != nil {
				t.Errorf("utils.ParseDecimalAmount(%s, %d) failed: %v",
					tc.decimalStr, tc.decimals, err)
				continue
			}

			if rawAmount != tc.expectedRaw {
				t.Errorf("utils.ParseDecimalAmount(%s, %d): expected %d, got %d",
					tc.decimalStr, tc.decimals, tc.expectedRaw, rawAmount)
			}
		}
	})

	t.Run("invalid decimal parsing", func(t *testing.T) {
		testCases := []struct {
			decimalStr string
			decimals   uint8
		}{
			{"1.23.45", 6},   // Multiple decimal points
			{"1.1234567", 6}, // Too many fractional digits
			{"invalid", 6},   // Non-numeric
		}

		for _, tc := range testCases {
			_, err := utils.ParseDecimalAmount(tc.decimalStr, tc.decimals)
			if err == nil {
				t.Errorf("utils.ParseDecimalAmount(%s, %d) should have failed",
					tc.decimalStr, tc.decimals)
			}
		}
	})
}

func TestWalletBalances(t *testing.T) {
	testWallet := solana.Address(TestReferenceWallet)
	testSlot := solana.Slot(12345)

	t.Run("wallet balances creation", func(t *testing.T) {
		wb := NewWalletBalances(testWallet, testSlot)

		if wb.Wallet != testWallet {
			t.Errorf("Wallet address mismatch")
		}

		if wb.Slot != testSlot {
			t.Errorf("Slot mismatch")
		}

		if wb.Balances == nil {
			t.Errorf("Balances map should be initialized")
		}

		if len(wb.Balances) != 0 {
			t.Errorf("Initial balances should be empty")
		}

		if wb.HasAnyBalance() {
			t.Errorf("Empty wallet should not have any balance")
		}
	})

	t.Run("adding and retrieving balances", func(t *testing.T) {
		wb := NewWalletBalances(testWallet, testSlot)

		// Create token balance
		tokenInfo := TokenInfo{
			Mint:     HyUSDMint,
			Symbol:   HyUSDSymbol,
			Name:     HyUSDName,
			Decimals: HyUSDDecimals,
		}
		balance := NewTokenBalance(tokenInfo, 1000000) // 1.0 hyUSD

		// Add balance
		wb.AddBalance(balance)

		// Retrieve balance
		retrievedBalance, exists := wb.GetBalance(HyUSDSymbol)
		if !exists {
			t.Errorf("Added balance should exist")
		}

		if retrievedBalance.RawAmount != 1000000 {
			t.Errorf("Retrieved balance amount mismatch")
		}

		// Test convenience methods
		hyusdBalance, exists := wb.GetHyUSDBalance()
		if !exists || hyusdBalance.RawAmount != 1000000 {
			t.Errorf("GetHyUSDBalance failed")
		}

		if !wb.HasAnyBalance() {
			t.Errorf("Wallet should have balance after adding")
		}
	})

	t.Run("USD value calculation", func(t *testing.T) {
		wb := NewWalletBalances(testWallet, testSlot)

		// Add multiple balances with USD values
		hyusdBalance := NewTokenBalance(TokenInfo{
			Mint:     HyUSDMint,
			Symbol:   HyUSDSymbol,
			Name:     HyUSDName,
			Decimals: HyUSDDecimals,
		}, 1000000) // 1.0 hyUSD
		hyusdBalance.SetUSDValue(1.0)

		xsolBalance := NewTokenBalance(TokenInfo{
			Mint:     XSOLMint,
			Symbol:   XSOLSymbol,
			Name:     XSOLName,
			Decimals: XSOLDecimals,
		}, 500000000) // 0.5 xSOL
		xsolBalance.SetUSDValue(100.0) // Assume 1 xSOL = $200

		wb.AddBalance(hyusdBalance)
		wb.AddBalance(xsolBalance)

		// Calculate total USD value
		wb.CalculateTotalUSDValue()

		if wb.TotalUSDValue == nil {
			t.Errorf("Total USD value should be calculated")
		}

		expectedTotal := 101.0 // $1 + $100
		if *wb.TotalUSDValue != expectedTotal {
			t.Errorf("Total USD value: expected %.2f, got %.2f",
				expectedTotal, *wb.TotalUSDValue)
		}
	})

	t.Run("wallet balances convenience methods", func(t *testing.T) {
		wb := NewWalletBalances(testWallet, testSlot)

		// Add all token types
		hyusdBalance := NewTokenBalance(TokenInfo{
			Mint:     HyUSDMint,
			Symbol:   HyUSDSymbol,
			Name:     HyUSDName,
			Decimals: HyUSDDecimals,
		}, 1000000)

		shyusdBalance := NewTokenBalance(TokenInfo{
			Mint:     SHyUSDMint,
			Symbol:   SHyUSDSymbol,
			Name:     SHyUSDName,
			Decimals: SHyUSDDecimals,
		}, 2000000)

		xsolBalance := NewTokenBalance(TokenInfo{
			Mint:     XSOLMint,
			Symbol:   XSOLSymbol,
			Name:     XSOLName,
			Decimals: XSOLDecimals,
		}, 500000000)

		wb.AddBalance(hyusdBalance)
		wb.AddBalance(shyusdBalance)
		wb.AddBalance(xsolBalance)

		// Test convenience methods
		hyusdRetrieved, exists := wb.GetHyUSDBalance()
		if !exists || hyusdRetrieved.RawAmount != 1000000 {
			t.Errorf("GetHyUSDBalance failed")
		}

		shyusdRetrieved, exists := wb.GetSHyUSDBalance()
		if !exists || shyusdRetrieved.RawAmount != 2000000 {
			t.Errorf("GetSHyUSDBalance failed")
		}

		xsolRetrieved, exists := wb.GetXSOLBalance()
		if !exists || xsolRetrieved.RawAmount != 500000000 {
			t.Errorf("GetXSOLBalance failed")
		}
	})

	t.Run("wallet balances validation", func(t *testing.T) {
		// Test valid wallet balances
		wb := NewWalletBalances(testWallet, testSlot)
		hyusdBalance := NewTokenBalance(TokenInfo{
			Mint:     HyUSDMint,
			Symbol:   HyUSDSymbol,
			Name:     HyUSDName,
			Decimals: HyUSDDecimals,
		}, 1000000)
		wb.AddBalance(hyusdBalance)

		if err := wb.Validate(); err != nil {
			t.Errorf("Valid wallet balances should pass validation: %v", err)
		}

		// Test invalid wallet address
		invalidWB := &WalletBalances{
			Wallet:   solana.Address(TestTooShortMint), // Use invalid short address
			Slot:     testSlot,
			Balances: make(map[string]*TokenBalance),
		}
		if err := invalidWB.Validate(); err == nil {
			t.Errorf("Should fail validation with invalid wallet address")
		}

		// Test nil balances map
		nilMapWB := &WalletBalances{
			Wallet:   testWallet,
			Slot:     testSlot,
			Balances: nil,
		}
		if err := nilMapWB.Validate(); err == nil {
			t.Errorf("Should fail validation with nil balances map")
		}

		// Test nil balance in map
		nilBalanceWB := &WalletBalances{
			Wallet: testWallet,
			Slot:   testSlot,
			Balances: map[string]*TokenBalance{
				HyUSDSymbol: nil,
			},
		}
		if err := nilBalanceWB.Validate(); err == nil {
			t.Errorf("Should fail validation with nil balance")
		}

		// Test symbol mismatch
		mismatchWB := &WalletBalances{
			Wallet: testWallet,
			Slot:   testSlot,
			Balances: map[string]*TokenBalance{
				"WRONG": hyusdBalance, // Wrong key for hyUSD balance
			},
		}
		if err := mismatchWB.Validate(); err == nil {
			t.Errorf("Should fail validation with symbol mismatch")
		}
	})
}

func TestConfig(t *testing.T) {
	t.Run("default config creation", func(t *testing.T) {
		config := NewConfig()

		if err := config.Validate(); err != nil {
			t.Errorf("Default config validation failed: %v", err)
		}

		// Verify default addresses
		if config.HyUSDMint != HyUSDMint {
			t.Errorf("Default hyUSD mint mismatch")
		}

		if config.SHyUSDMint != SHyUSDMint {
			t.Errorf("Default sHYUSD mint mismatch")
		}

		if config.XSOLMint != XSOLMint {
			t.Errorf("Default xSOL mint mismatch")
		}
	})

	t.Run("environment variable override", func(t *testing.T) {
		// Set all environment variables
		os.Setenv("HYUSD_MINT", TestHyUSDMintOverride)
		os.Setenv("SHYUSD_MINT", TestSHyUSDMintOverride)
		os.Setenv("XSOL_MINT", TestXSOLMintOverride)
		os.Setenv("USDC_MINT", TestUSDCMintOverride)
		os.Setenv("JITOSOL_MINT", TestJitoSOLMintOverride)

		defer func() {
			os.Unsetenv("HYUSD_MINT")
			os.Unsetenv("SHYUSD_MINT")
			os.Unsetenv("XSOL_MINT")
			os.Unsetenv("USDC_MINT")
			os.Unsetenv("JITOSOL_MINT")
		}()

		config := NewConfig()

		if string(config.HyUSDMint) != TestHyUSDMintOverride {
			t.Errorf("HYUSD_MINT override failed: expected %s, got %s",
				TestHyUSDMintOverride, config.HyUSDMint)
		}

		if string(config.SHyUSDMint) != TestSHyUSDMintOverride {
			t.Errorf("SHYUSD_MINT override failed: expected %s, got %s",
				TestSHyUSDMintOverride, config.SHyUSDMint)
		}

		if string(config.XSOLMint) != TestXSOLMintOverride {
			t.Errorf("XSOL_MINT override failed: expected %s, got %s",
				TestXSOLMintOverride, config.XSOLMint)
		}

		if string(config.USDCMint) != TestUSDCMintOverride {
			t.Errorf("USDC_MINT override failed: expected %s, got %s",
				TestUSDCMintOverride, config.USDCMint)
		}

		if string(config.JitoSOLMint) != TestJitoSOLMintOverride {
			t.Errorf("JITOSOL_MINT override failed: expected %s, got %s",
				TestJitoSOLMintOverride, config.JitoSOLMint)
		}
	})

	t.Run("environment variable with spaces", func(t *testing.T) {
		// Test trimming of environment variables
		os.Setenv("HYUSD_MINT", TestMintWithSpaces)
		defer os.Unsetenv("HYUSD_MINT")

		config := NewConfig()

		if string(config.HyUSDMint) != TestMintTrimmed {
			t.Errorf("Environment variable should be trimmed: expected %s, got %s",
				TestMintTrimmed, config.HyUSDMint)
		}
	})

	t.Run("token registry functionality", func(t *testing.T) {
		config := NewConfig()

		// Test token info retrieval
		hyusdInfo := config.GetTokenInfo(HyUSDMint)
		if hyusdInfo == nil {
			t.Errorf("Should find hyUSD token info")
		}
		if hyusdInfo.Symbol != HyUSDSymbol {
			t.Errorf("hyUSD symbol mismatch")
		}

		// Test unsupported token
		invalidMint := solana.Address(TestInvalidMint)
		invalidInfo := config.GetTokenInfo(invalidMint)
		if invalidInfo != nil {
			t.Errorf("Should not find info for invalid mint")
		}

		// Test symbol lookup
		xsolInfo := config.GetTokenBySymbol(XSOLSymbol)
		if xsolInfo == nil || xsolInfo.Mint != XSOLMint {
			t.Errorf("Symbol lookup failed for xSOL")
		}

		// Test supported tokens list
		supportedTokens := config.GetSupportedTokens()
		if len(supportedTokens) != 5 {
			t.Errorf("Expected 5 supported tokens, got %d", len(supportedTokens))
		}
	})

	t.Run("config validation errors", func(t *testing.T) {
		// Test duplicate mint addresses
		duplicateConfig := &Config{
			HyUSDMint:   HyUSDMint,
			SHyUSDMint:  HyUSDMint, // Duplicate!
			XSOLMint:    XSOLMint,
			USDCMint:    USDCMint,
			JitoSOLMint: JitoSOLMint,
		}
		duplicateConfig.buildTokenRegistry()

		if err := duplicateConfig.Validate(); err == nil {
			t.Errorf("Should fail validation with duplicate mints")
		}

		// Test invalid mint address lengths
		invalidMintConfig := &Config{
			HyUSDMint:   solana.Address(TestTooShortMint), // Too short
			SHyUSDMint:  SHyUSDMint,
			XSOLMint:    XSOLMint,
			USDCMint:    USDCMint,
			JitoSOLMint: JitoSOLMint,
		}
		invalidMintConfig.buildTokenRegistry()

		if err := invalidMintConfig.Validate(); err == nil {
			t.Errorf("Should fail validation with invalid mint address")
		}
	})

	t.Run("amount formatting and parsing", func(t *testing.T) {
		config := NewConfig()

		// Test amount formatting
		formatted := config.FormatAmount(HyUSDMint, 1500000)
		if formatted != "1.5" {
			t.Errorf("Amount formatting failed: expected 1.5, got %s", formatted)
		}

		// Test amount formatting for unsupported mint
		invalidMint := solana.Address(TestInvalidMint)
		formattedInvalid := config.FormatAmount(invalidMint, 1000000)
		if formattedInvalid != "0" {
			t.Errorf("Unsupported mint formatting should return '0', got %s", formattedInvalid)
		}

		// Test amount parsing
		rawAmount, err := config.ParseAmount(HyUSDMint, "1.5")
		if err != nil {
			t.Errorf("Amount parsing failed: %v", err)
		}
		if rawAmount != 1500000 {
			t.Errorf("Parsed amount mismatch: expected 1500000, got %d", rawAmount)
		}

		// Test unsupported mint parsing
		_, err = config.ParseAmount(invalidMint, "1.0")
		if err == nil {
			t.Errorf("Should fail parsing for unsupported mint")
		}
	})

	t.Run("token support checking", func(t *testing.T) {
		config := NewConfig()

		// Test supported tokens
		if !config.IsTokenSupported(HyUSDMint) {
			t.Errorf("hyUSD should be supported")
		}

		if !config.IsTokenSupported(SHyUSDMint) {
			t.Errorf("sHYUSD should be supported")
		}

		if !config.IsTokenSupported(XSOLMint) {
			t.Errorf("xSOL should be supported")
		}

		if !config.IsTokenSupported(USDCMint) {
			t.Errorf("USDC should be supported")
		}

		if !config.IsTokenSupported(JitoSOLMint) {
			t.Errorf("jitoSOL should be supported")
		}

		// Test unsupported token
		invalidMint := solana.Address(TestInvalidMint)
		if config.IsTokenSupported(invalidMint) {
			t.Errorf("Invalid mint should not be supported")
		}
	})

	t.Run("supported mints list", func(t *testing.T) {
		config := NewConfig()

		supportedMints := config.GetSupportedMints()
		if len(supportedMints) != 5 {
			t.Errorf("Expected 5 supported mints, got %d", len(supportedMints))
		}

		// Verify order and content
		expectedMints := []solana.Address{HyUSDMint, SHyUSDMint, XSOLMint, USDCMint, JitoSOLMint}
		for i, expectedMint := range expectedMints {
			if i >= len(supportedMints) || supportedMints[i] != expectedMint {
				t.Errorf("Mint mismatch at index %d: expected %s, got %s",
					i, expectedMint, supportedMints[i])
			}
		}
	})

	t.Run("mint by symbol lookup", func(t *testing.T) {
		config := NewConfig()

		// Test valid symbols
		hyusdMint := config.GetMintBySymbol(HyUSDSymbol)
		if hyusdMint != HyUSDMint {
			t.Errorf("GetMintBySymbol(%s): expected %s, got %s",
				HyUSDSymbol, HyUSDMint, hyusdMint)
		}

		shyusdMint := config.GetMintBySymbol(SHyUSDSymbol)
		if shyusdMint != SHyUSDMint {
			t.Errorf("GetMintBySymbol(%s): expected %s, got %s",
				SHyUSDSymbol, SHyUSDMint, shyusdMint)
		}

		xsolMint := config.GetMintBySymbol(XSOLSymbol)
		if xsolMint != XSOLMint {
			t.Errorf("GetMintBySymbol(%s): expected %s, got %s",
				XSOLSymbol, XSOLMint, xsolMint)
		}

		// Test invalid symbol
		invalidMint := config.GetMintBySymbol("INVALID")
		if invalidMint != solana.Address("") {
			t.Errorf("Invalid symbol should return empty address")
		}
	})

	t.Run("token balance creation by config", func(t *testing.T) {
		config := NewConfig()

		// Test NewTokenBalance by mint
		balance := config.NewTokenBalance(HyUSDMint, 1500000)
		if balance == nil {
			t.Errorf("NewTokenBalance should create valid balance")
		}
		if balance.RawAmount != 1500000 {
			t.Errorf("Balance raw amount mismatch")
		}
		if balance.TokenInfo.Symbol != HyUSDSymbol {
			t.Errorf("Balance symbol mismatch")
		}

		// Test NewTokenBalance with unsupported mint
		invalidMint := solana.Address(TestInvalidMint)
		invalidBalance := config.NewTokenBalance(invalidMint, 1000000)
		if invalidBalance != nil {
			t.Errorf("NewTokenBalance should return nil for unsupported mint")
		}

		// Test NewTokenBalanceBySymbol
		balanceBySymbol := config.NewTokenBalanceBySymbol(XSOLSymbol, 500000000)
		if balanceBySymbol == nil {
			t.Errorf("NewTokenBalanceBySymbol should create valid balance")
		}
		if balanceBySymbol.RawAmount != 500000000 {
			t.Errorf("Balance by symbol raw amount mismatch")
		}
		if balanceBySymbol.TokenInfo.Symbol != XSOLSymbol {
			t.Errorf("Balance by symbol symbol mismatch")
		}

		// Test NewTokenBalanceBySymbol with invalid symbol
		invalidBalanceBySymbol := config.NewTokenBalanceBySymbol("INVALID", 1000000)
		if invalidBalanceBySymbol != nil {
			t.Errorf("NewTokenBalanceBySymbol should return nil for invalid symbol")
		}
	})

	t.Run("token balance validation", func(t *testing.T) {
		config := NewConfig()

		// Create valid token balances
		validBalances := map[string]*TokenBalance{
			HyUSDSymbol: config.NewTokenBalance(HyUSDMint, 1000000),
			XSOLSymbol:  config.NewTokenBalance(XSOLMint, 500000000),
		}

		// Test valid balances
		if err := config.ValidateTokenBalances(validBalances); err != nil {
			t.Errorf("Valid balances should pass validation: %v", err)
		}

		// Test with unsupported symbol
		invalidBalances := map[string]*TokenBalance{
			"INVALID": config.NewTokenBalance(HyUSDMint, 1000000),
		}
		if err := config.ValidateTokenBalances(invalidBalances); err == nil {
			t.Errorf("Should fail validation with unsupported symbol")
		}

		// Test with mint mismatch
		mismatchBalances := map[string]*TokenBalance{
			HyUSDSymbol: config.NewTokenBalance(XSOLMint, 1000000), // Wrong mint for symbol
		}
		if err := config.ValidateTokenBalances(mismatchBalances); err == nil {
			t.Errorf("Should fail validation with mint mismatch")
		}
	})
}
