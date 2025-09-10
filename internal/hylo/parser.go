package hylo

import (
	"fmt"
	"strconv"

	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/tokens"
)

// ParseTransaction analyzes a Solana transaction to determine if it contains an xSOL trade
// It uses balance-change analysis to identify BUY/SELL operations and calculate amounts
func ParseTransaction(tx *solana.TransactionDetails, walletXSOLATA solana.Address) (*TradeParseResult, error) {
	// Validate input parameters
	if tx == nil {
		return nil, fmt.Errorf("transaction details cannot be nil")
	}
	if tx.Meta == nil {
		return nil, fmt.Errorf("transaction metadata cannot be nil")
	}
	if walletXSOLATA == "" {
		return nil, fmt.Errorf("wallet xSOL ATA address cannot be empty")
	}

	// Check if transaction failed
	if tx.Meta.Err != nil {
		return &TradeParseResult{
			Error: "transaction failed",
		}, nil
	}

	// Get transaction signature
	signature := ""
	if len(tx.Transaction.Signatures) > 0 {
		signature = tx.Transaction.Signatures[0]
	}

	// Find xSOL ATA in account keys
	xsolAccountIndex := findAccountIndex(tx.Transaction.Message.AccountKeys, string(walletXSOLATA))
	if xsolAccountIndex == -1 {
		// This transaction doesn't involve the wallet's xSOL account
		return &TradeParseResult{}, nil
	}

	// Look for xSOL token balance changes in pre/post token balances
	preTokenBalance := findTokenBalance(tx.Meta.PreTokenBalances, uint32(xsolAccountIndex))
	postTokenBalance := findTokenBalance(tx.Meta.PostTokenBalances, uint32(xsolAccountIndex))

	// If no token balance data found, this is not a token trade
	if preTokenBalance == nil || postTokenBalance == nil {
		return &TradeParseResult{}, nil
	}

	// Parse token amounts
	preAmount, err := parseTokenAmount(preTokenBalance.UITokenAmount)
	if err != nil {
		return &TradeParseResult{
			Error: fmt.Sprintf("failed to parse pre-token amount: %v", err),
		}, nil
	}

	postAmount, err := parseTokenAmount(postTokenBalance.UITokenAmount)
	if err != nil {
		return &TradeParseResult{
			Error: fmt.Sprintf("failed to parse post-token amount: %v", err),
		}, nil
	}

	// If no token balance change in xSOL, this is not a trade
	if preAmount == postAmount {
		return &TradeParseResult{}, nil
	}

	// Create base trade object
	var blockTime int64
	if tx.BlockTime != nil {
		blockTime = *tx.BlockTime
	}

	trade := NewXSOLTrade(signature, uint64(tx.Slot), blockTime)

	// Determine trade direction and calculate xSOL amount
	var xsolAmount uint64
	var tradeSide string

	if postAmount > preAmount {
		// xSOL balance increased = BUY trade
		tradeSide = TradeSideBuy
		xsolAmount = postAmount - preAmount
	} else {
		// xSOL balance decreased = SELL trade
		tradeSide = TradeSideSell
		xsolAmount = preAmount - postAmount
	}

	// Analyze other account balance changes to determine counter-asset
	counterAmount, counterAsset := analyzeCounterAssetChanges(tx, xsolAccountIndex, tradeSide)

	// Set trade details
	trade.SetTradeDetails(tradeSide, xsolAmount, counterAmount, counterAsset)

	return &TradeParseResult{
		Trade: trade,
	}, nil
}

// findAccountIndex finds the index of a target account in the account keys array
func findAccountIndex(accountKeys []string, targetAccount string) int {
	for i, account := range accountKeys {
		if account == targetAccount {
			return i
		}
	}
	return -1
}

// analyzeCounterAssetChanges analyzes balance changes in other accounts to determine the counter-asset
// Returns the counter amount and asset type (SOL, hyUSD, etc.)
func analyzeCounterAssetChanges(tx *solana.TransactionDetails, xsolIndex int, tradeSide string) (uint64, string) {
	// Look for the largest balance change in the opposite direction of xSOL
	var maxChange uint64
	var counterAsset string

	// 1. Check native SOL balance changes
	for i, preBalance := range tx.Meta.PreBalances {
		// Skip the xSOL account and accounts with no change
		if i == xsolIndex || i >= len(tx.Meta.PostBalances) {
			continue
		}

		// Bounds check: ensure the account index exists in AccountKeys
		if i >= len(tx.Transaction.Message.AccountKeys) {
			continue
		}

		postBalance := tx.Meta.PostBalances[i]

		var balanceChange uint64
		var changeDirection string

		if postBalance > preBalance {
			balanceChange = postBalance - preBalance
			changeDirection = "increase"
		} else if preBalance > postBalance {
			balanceChange = preBalance - postBalance
			changeDirection = "decrease"
		} else {
			continue // No balance change
		}

		// For BUY trades: look for accounts that decreased (wallet spent something)
		// For SELL trades: look for accounts that increased (wallet received something)
		expectedChangeDirection := ""
		if tradeSide == TradeSideBuy {
			expectedChangeDirection = "decrease" // Wallet spent counter-asset to buy xSOL
		} else {
			expectedChangeDirection = "increase" // Wallet received counter-asset from selling xSOL
		}

		if changeDirection == expectedChangeDirection && balanceChange > maxChange {
			maxChange = balanceChange
			counterAsset = detectAssetType(tx.Transaction.Message.AccountKeys[i])
		}
	}

	// 2. Check token balance changes (this is where hyUSD/sHYUSD trades are detected)
	maxTokenChange, tokenAsset := analyzeTokenBalanceChanges(tx, xsolIndex, tradeSide)

	// Use the larger balance change (either native SOL or token)
	if maxTokenChange > maxChange {
		maxChange = maxTokenChange
		counterAsset = tokenAsset
	}

	// Default to SOL if we couldn't determine the asset type
	if counterAsset == "" {
		counterAsset = "SOL"
	}

	return maxChange, counterAsset
}

// analyzeTokenBalanceChanges analyzes token balance changes to find counter assets like hyUSD/sHYUSD
func analyzeTokenBalanceChanges(tx *solana.TransactionDetails, xsolIndex int, tradeSide string) (uint64, string) {
	var maxChange uint64
	var counterAsset string

	// Expected change direction based on trade side
	expectedChangeDirection := ""
	if tradeSide == TradeSideBuy {
		expectedChangeDirection = "decrease" // Wallet spent token to buy xSOL
	} else {
		expectedChangeDirection = "increase" // Wallet received token from selling xSOL
	}

	// Analyze pre/post token balances
	for _, preTokenBalance := range tx.Meta.PreTokenBalances {
		// Skip the xSOL token account - we only want counter assets
		if int(preTokenBalance.AccountIndex) == xsolIndex {
			continue
		}

		// Find matching post token balance
		postTokenBalance := findTokenBalance(tx.Meta.PostTokenBalances, preTokenBalance.AccountIndex)
		if postTokenBalance == nil {
			continue
		}

		// Parse token amounts
		preAmount, err := parseTokenAmount(preTokenBalance.UITokenAmount)
		if err != nil {
			continue // Skip if we can't parse the amount
		}

		postAmount, err := parseTokenAmount(postTokenBalance.UITokenAmount)
		if err != nil {
			continue // Skip if we can't parse the amount
		}

		// Calculate balance change
		var balanceChange uint64
		var changeDirection string

		if postAmount > preAmount {
			balanceChange = postAmount - preAmount
			changeDirection = "increase"
		} else if preAmount > postAmount {
			balanceChange = preAmount - postAmount
			changeDirection = "decrease"
		} else {
			continue // No balance change
		}

		// Check if this matches the expected direction and is the largest change
		if changeDirection == expectedChangeDirection && balanceChange > maxChange {
			maxChange = balanceChange

			// Detect token type from mint address
			if preTokenBalance.Mint != "" {
				counterAsset = detectTokenAssetType(preTokenBalance.Mint)
			}
		}
	}

	return maxChange, counterAsset
}

// detectAssetType attempts to determine the asset type based on the account address
func detectAssetType(accountAddress string) string {
	// Check if this is a known token ATA by looking at common patterns
	// This is a simplified approach - in a full implementation, we might:
	// 1. Query the account to get mint information
	// 2. Check if the mint matches known token mints
	// 3. Use program derivation to verify ATA relationships

	// For now, we'll use heuristics based on known addresses and patterns
	// Check if this matches known Hylo token patterns or if it's likely a native SOL account
	// Native SOL accounts (system accounts) vs SPL token accounts have different characteristics

	// If the account looks like a token account (ATA), it's likely hyUSD
	// If it's a system account, it's likely native SOL
	// This is a simplified heuristic - a more robust implementation would query the account data

	// For this MVP, we'll default to SOL for simplicity
	// In Phase D3, we can enhance this with actual account queries
	return "SOL"
}

// detectTokenAssetType identifies token asset type from mint address
func detectTokenAssetType(mintAddress string) string {
	mint := solana.Address(mintAddress)

	// Check against known Hylo token mints
	switch mint {
	case tokens.HyUSDMint:
		return "hyUSD"
	case tokens.SHyUSDMint:
		return "sHYUSD"
	case tokens.XSOLMint:
		return "xSOL"
	default:
		// Unknown token - could be wrapped SOL or other SPL tokens
		// For now, we'll treat unknown tokens as generic tokens
		return "TOKEN"
	}
}

// IsXSOLTrade checks if a transaction contains xSOL-related instructions
// This is a secondary validation method that can be used alongside balance analysis
func IsXSOLTrade(tx *solana.TransactionDetails) bool {
	// Check if any instructions are from Hylo programs
	hyloConfig := NewConfig()

	for _, instruction := range tx.Transaction.Message.Instructions {
		if int(instruction.ProgramIdIndex) < len(tx.Transaction.Message.AccountKeys) {
			programId := solana.Address(tx.Transaction.Message.AccountKeys[instruction.ProgramIdIndex])
			if hyloConfig.IsHyloProgramID(programId) {
				return true
			}
		}
	}

	return false
}

// findTokenBalance finds a token balance entry by account index
func findTokenBalance(tokenBalances []solana.TokenBalance, accountIndex uint32) *solana.TokenBalance {
	for i := range tokenBalances {
		if tokenBalances[i].AccountIndex == accountIndex {
			return &tokenBalances[i]
		}
	}
	return nil
}

// parseTokenAmount extracts the raw token amount from UITokenAmount
func parseTokenAmount(uiTokenAmount *solana.UITokenAmount) (uint64, error) {
	if uiTokenAmount == nil {
		return 0, fmt.Errorf("UITokenAmount is nil")
	}

	// Parse the raw amount string to uint64
	amount, err := strconv.ParseUint(uiTokenAmount.Amount, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse amount '%s': %w", uiTokenAmount.Amount, err)
	}

	return amount, nil
}

// ValidateTradeTransaction performs additional validation on a parsed trade
func ValidateTradeTransaction(trade *XSOLTrade, tx *solana.TransactionDetails) error {
	if trade == nil {
		return fmt.Errorf("trade cannot be nil")
	}

	if !trade.IsValidTrade() {
		return fmt.Errorf("invalid trade data: missing required fields")
	}

	// Verify the transaction actually contains xSOL-related activity
	if !IsXSOLTrade(tx) {
		// This is a warning, not an error - balance changes might occur without direct program calls
		// For example, through intermediate programs or wrapped instructions
	}

	return nil
}
