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
)
