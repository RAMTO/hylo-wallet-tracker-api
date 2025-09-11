package tokens

// Test Constants - For testing purposes only
// These constants are used exclusively in test files and should never be used in production code

const (
	// Test wallet addresses for consistent testing
	// TestReferenceWallet is the reference wallet address from PRD for consistent testing
	TestReferenceWallet = "A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g"

	// TestSystemWallet is a second reference wallet address for testing different wallets
	TestSystemWallet = "B4wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6h"

	// Test mock values for error scenarios and invalid inputs
	// TestInvalidProgramID is used to test invalid program ID error handling
	TestInvalidProgramID = "WrongProgramID111111111111111111111111"

	// TestUnsupportedMint is used to test unsupported token mint error handling
	TestUnsupportedMint = "8NmnwkuHr6mwegkxeU26LnHym7M2g1KWk8CzdDmNkLT6"

	// TestOwnerAddress is used for mock SPL token account owner in parser tests
	TestOwnerAddress = "WalletOwner111111111111111111111111111111"

	// TestMintAddress is used for mock SPL token account mint in parser tests
	TestMintAddress = "TestMint1111111111111111111111111111111"

	// Test mint addresses for environment variable override testing
	TestHyUSDMintOverride   = "TestHyUSDMintAddress1234567890123456789012"
	TestSHyUSDMintOverride  = "TestSHyUSDMintAddress123456789012345678901"
	TestXSOLMintOverride    = "TestXSOLMintAddress1234567890123456789012"
	TestUSDCMintOverride    = "TestUSDCMintAddress1234567890123456789012"
	TestJitoSOLMintOverride = "TestJitoSOLMintAddress12345678901234567890"

	// Test mint address for whitespace trimming test
	TestMintWithSpaces = "  TestHyUSDMintAddress1234567890123456789012  "
	TestMintTrimmed    = "TestHyUSDMintAddress1234567890123456789012"

	// Invalid mint addresses for error testing
	TestInvalidMint       = "InvalidMint123456789012345678901234567890"
	TestTooShortMint      = "toolong"
	TestDuplicateTestMint = "DuplicateTestMint1234567890123456789012"

	// Test transaction signatures for parser testing
	TestSignatureBuy       = "testSig123"
	TestSignatureSell      = "testSig456"
	TestSignatureNoTrade   = "testSig789"
	TestSignatureFailed    = "testSigFailed"
	TestSignatureNoXSOL    = "testSigNoXSOL"
	TestSignatureHyUSDBuy  = "testSigHyUSDBuy"
	TestSignatureHyUSDSell = "testHyUSDSellSig"
	TestSignatureSHyUSDBuy = "testSHYUSDBuySig"
	TestSignatureMixed     = "testMixedBalanceSig"
	TestSignatureSOLTrade  = "testSigSOLTrade"

	// Test Associated Token Account (ATA) addresses for parser testing
	TestXSOLATA1  = "Dqk1wW44Mw9LkKBcVjSfWDXeNYuNZ1KaXKFBAuVRzzhJ"
	TestXSOLATA2  = "xSOLATAaddress123456789abcdef"
	TestSOLATA    = "SOLATAaddress123456789abcdef"
	TestHyUSDATA  = "hyUSDATAaddress123456789abcdef"
	TestSHyUSDATA = "SHYUSDATAAddressHere123456789012345678901234"
	TestHyUSDATA2 = "HyUSDATAAddressHere1234567890123456789012345"

	// Test blockchain data for consistent testing
	TestBlockTime    = int64(1757360079)
	TestSlot         = uint64(365528388)
	TestAccountIndex = uint32(3)

	// Test token amounts (raw amounts without decimal adjustment)
	TestXSOLAmount1M   = "1000000"   // 1.0 xSOL
	TestXSOLAmount2M   = "2000000"   // 2.0 xSOL
	TestXSOLAmount2_5M = "2500000"   // 2.5 xSOL
	TestXSOLAmount3M   = "3000000"   // 3.0 xSOL
	TestXSOLAmount5M   = "5000000"   // 5.0 xSOL
	TestXSOLAmount25M  = "25000000"  // 25 xSOL
	TestXSOLAmount50M  = "50000000"  // 50 xSOL
	TestXSOLAmount100M = "100000000" // 100 xSOL
	TestXSOLAmount125M = "125000000" // 125 xSOL

	TestHyUSDAmount500M  = "500000000"  // 500 hyUSD
	TestHyUSDAmount700M  = "700000000"  // 700 hyUSD
	TestHyUSDAmount800M  = "800000000"  // 800 hyUSD
	TestHyUSDAmount1000M = "1000000000" // 1000 hyUSD
	TestHyUSDAmount1200M = "1200000000" // 1200 hyUSD
	TestHyUSDAmount1233M = "1233591985" // 1233.591985 hyUSD

	TestSHyUSDAmount300M = "300000000" // 300 sHYUSD
	TestSHyUSDAmount800M = "800000000" // 800 sHYUSD
)
