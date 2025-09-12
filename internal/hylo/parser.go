package hylo

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"hylo-wallet-tracker-api/internal/logger"
	"hylo-wallet-tracker-api/internal/solana"
	"hylo-wallet-tracker-api/internal/tokens"
)

// ParseTransaction analyzes a Solana transaction to determine if it contains an xSOL trade
// It uses balance-change analysis to identify BUY/SELL operations and calculate amounts
func ParseTransaction(tx *solana.TransactionDetails, walletXSOLATA solana.Address) (*TradeParseResult, error) {
	return ParseTransactionWithContext(context.Background(), tx, walletXSOLATA, nil)
}

// ParseTransactionWithContext analyzes a Solana transaction with logging context
func ParseTransactionWithContext(ctx context.Context, tx *solana.TransactionDetails, walletXSOLATA solana.Address, log *logger.Logger) (*TradeParseResult, error) {
	startTime := time.Now()

	// Use default logger if none provided
	if log == nil {
		log = logger.NewFromEnv().WithComponent("hylo-parser")
	}

	// Get transaction signature for logging
	signature := ""
	if tx != nil && len(tx.Transaction.Signatures) > 0 {
		signature = tx.Transaction.Signatures[0]
	}

	log.DebugContext(ctx, "Starting transaction parsing",
		slog.String("signature", signature),
		slog.String("ata_address", walletXSOLATA.String()))

	// Validate input parameters
	if tx == nil {
		log.LogParsingError(ctx, "parse_transaction", "transaction_details", fmt.Errorf("transaction details cannot be nil"))
		return nil, fmt.Errorf("transaction details cannot be nil")
	}
	if tx.Meta == nil {
		log.LogParsingError(ctx, "parse_transaction", "transaction_meta", fmt.Errorf("transaction metadata cannot be nil"),
			slog.String("signature", signature))
		return nil, fmt.Errorf("transaction metadata cannot be nil")
	}
	if walletXSOLATA == "" {
		log.LogValidationError(ctx, "parse_transaction", "ata_address", walletXSOLATA, fmt.Errorf("wallet xSOL ATA address cannot be empty"))
		return nil, fmt.Errorf("wallet xSOL ATA address cannot be empty")
	}

	// Check if transaction failed
	if tx.Meta.Err != nil {
		log.WarnContext(ctx, "Transaction failed, skipping trade parsing",
			slog.String("signature", signature),
			slog.Any("error", tx.Meta.Err))
		return &TradeParseResult{
			Error: "transaction failed",
		}, nil
	}

	// Find xSOL ATA in account keys
	xsolAccountIndex := findAccountIndex(tx.Transaction.Message.AccountKeys, string(walletXSOLATA))
	if xsolAccountIndex == -1 {
		log.DebugContext(ctx, "Transaction doesn't involve wallet's xSOL account",
			slog.String("signature", signature),
			slog.String("ata_address", walletXSOLATA.String()))
		// This transaction doesn't involve the wallet's xSOL account
		return &TradeParseResult{}, nil
	}

	log.DebugContext(ctx, "Found xSOL account in transaction",
		slog.String("signature", signature),
		slog.Int("account_index", xsolAccountIndex))

	// PRIORITY: Check for Hylo program instructions first
	// This handles cases where users trade via Hylo Exchange, including first-time trades
	hyloInstructionType := detectHyloInstructions(tx)
	if hyloInstructionType != "" {
		log.DebugContext(ctx, "Detected Hylo instruction, parsing as trade",
			slog.String("signature", signature),
			slog.String("instruction_type", hyloInstructionType))
		return parseHyloTrade(ctx, tx, walletXSOLATA, xsolAccountIndex, hyloInstructionType, signature, log)
	}

	// FALLBACK: Look for xSOL token balance changes for non-Hylo transactions
	preTokenBalance := findTokenBalance(tx.Meta.PreTokenBalances, uint32(xsolAccountIndex))
	postTokenBalance := findTokenBalance(tx.Meta.PostTokenBalances, uint32(xsolAccountIndex))

	// Handle different balance scenarios for non-Hylo transactions:
	// 1. Both pre and post balance exist -> external trade or transfer
	// 2. Only post balance exists -> initial funding/transfer (RECEIVE)
	// 3. Neither exist -> not a token transaction
	if preTokenBalance == nil && postTokenBalance == nil {
		log.DebugContext(ctx, "No token balance data found, not a token transaction",
			slog.String("signature", signature))
		return &TradeParseResult{}, nil
	}

	// Handle initial funding case (only post balance, no pre balance) - for non-Hylo transactions
	if preTokenBalance == nil && postTokenBalance != nil {
		return parseInitialFundingTransaction(ctx, tx, postTokenBalance, walletXSOLATA, signature, log)
	}

	// Handle case where pre balance exists but post balance doesn't (shouldn't happen in normal cases)
	if preTokenBalance != nil && postTokenBalance == nil {
		log.DebugContext(ctx, "Pre-balance exists but no post-balance, unusual transaction",
			slog.String("signature", signature))
		return &TradeParseResult{}, nil
	}

	// Parse token amounts
	preAmount, err := parseTokenAmountWithLogging(ctx, preTokenBalance.UITokenAmount, log, "pre-amount")
	if err != nil {
		log.LogParsingError(ctx, "parse_transaction", "pre_token_amount", err,
			slog.String("signature", signature))
		return &TradeParseResult{
			Error: fmt.Sprintf("failed to parse pre-token amount: %v", err),
		}, nil
	}

	postAmount, err := parseTokenAmountWithLogging(ctx, postTokenBalance.UITokenAmount, log, "post-amount")
	if err != nil {
		log.LogParsingError(ctx, "parse_transaction", "post_token_amount", err,
			slog.String("signature", signature))
		return &TradeParseResult{
			Error: fmt.Sprintf("failed to parse post-token amount: %v", err),
		}, nil
	}

	// If no token balance change in xSOL, this is not a trade
	if preAmount == postAmount {
		log.DebugContext(ctx, "No xSOL balance change detected, not a trade",
			slog.String("signature", signature),
			slog.Uint64("amount", preAmount))
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
		log.DebugContext(ctx, "Detected BUY trade",
			slog.String("signature", signature),
			slog.Uint64("pre_amount", preAmount),
			slog.Uint64("post_amount", postAmount),
			slog.Uint64("xsol_amount", xsolAmount))
	} else {
		// xSOL balance decreased = SELL trade
		tradeSide = TradeSideSell
		xsolAmount = preAmount - postAmount
		log.DebugContext(ctx, "Detected SELL trade",
			slog.String("signature", signature),
			slog.Uint64("pre_amount", preAmount),
			slog.Uint64("post_amount", postAmount),
			slog.Uint64("xsol_amount", xsolAmount))
	}

	// Analyze other account balance changes to determine counter-asset
	counterAmount, counterAsset := analyzeCounterAssetChangesWithLogging(ctx, tx, xsolAccountIndex, tradeSide, log)

	// Set trade details
	trade.SetTradeDetails(tradeSide, xsolAmount, counterAmount, counterAsset)

	// Calculate historical price for hyUSD trades
	trade.HistoricalPriceUSD = CalculateHistoricalXSOLPrice(trade)

	// Log successful trade parsing with historical price info
	priceInfo := "N/A (SOL trade)"
	if trade.HistoricalPriceUSD != nil {
		priceInfo = fmt.Sprintf("$%s", *trade.HistoricalPriceUSD)
	}

	log.InfoContext(ctx, "Successfully parsed xSOL trade",
		slog.String("signature", signature),
		slog.String("side", tradeSide),
		slog.Uint64("xsol_amount", xsolAmount),
		slog.Uint64("counter_amount", counterAmount),
		slog.String("counter_asset", counterAsset),
		slog.String("historical_price", priceInfo),
		slog.Duration("parse_time", time.Since(startTime)))

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

// getAssetPriority returns priority score for counter asset selection
// Higher priority assets are preferred over lower priority ones in multi-hop transactions
func getAssetPriority(asset string) int {
	switch asset {
	case "hyUSD":
		return 100 // Hylo native stablecoin - highest priority
	case "USDC":
		return 90 // USD stablecoin - very high priority
	case "sHYUSD":
		return 80 // Staked hyUSD - high priority
	case "SOL":
		return 50 // Native SOL - medium priority
	case "jitoSOL":
		return 30 // Liquid staking token - lower priority (often intermediate)
	default:
		return 10 // Unknown tokens - lowest priority
	}
}

// shouldReplaceCounterAsset determines if a new candidate should replace the current counter asset
// Prioritizes higher-priority assets (stablecoins) over lower-priority ones (intermediate assets)
func shouldReplaceCounterAsset(currentAsset string, currentAmount uint64, newAsset string, newAmount uint64) bool {
	currentPriority := getAssetPriority(currentAsset)
	newPriority := getAssetPriority(newAsset)

	// If priorities are different, choose higher priority asset
	if newPriority != currentPriority {
		return newPriority > currentPriority
	}

	// If same priority, choose larger amount (original logic)
	return newAmount > currentAmount
}

// analyzeTokenBalanceChanges analyzes token balance changes to find counter assets like hyUSD/sHYUSD
// Uses asset priority to prefer stablecoins (USDC, hyUSD) over intermediate assets (jitoSOL) in multi-hop transactions
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

		// Check if this matches the expected direction
		if changeDirection == expectedChangeDirection {
			// Detect token type from mint address
			var candidateAsset string
			if preTokenBalance.Mint != "" {
				candidateAsset = detectTokenAssetType(preTokenBalance.Mint)
			}

			// Use priority-based selection instead of just largest amount
			// This prefers stablecoins (USDC, hyUSD) over intermediate assets (jitoSOL)
			if counterAsset == "" || shouldReplaceCounterAsset(counterAsset, maxChange, candidateAsset, balanceChange) {
				maxChange = balanceChange
				counterAsset = candidateAsset
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

	// Check against known token mints (Hylo protocol + supported external tokens)
	switch mint {
	case tokens.HyUSDMint:
		return "hyUSD"
	case tokens.SHyUSDMint:
		return "sHYUSD"
	case tokens.XSOLMint:
		return "xSOL"
	case tokens.USDCMint:
		return "USDC"
	case tokens.JitoSOLMint:
		return "jitoSOL"
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

// parseTokenAmountWithLogging wraps parseTokenAmount with logging
func parseTokenAmountWithLogging(ctx context.Context, uiTokenAmount *solana.UITokenAmount, log *logger.Logger, amountType string) (uint64, error) {
	if uiTokenAmount == nil {
		log.LogParsingError(ctx, "parse_token_amount", amountType, fmt.Errorf("UITokenAmount is nil"))
		return 0, fmt.Errorf("UITokenAmount is nil")
	}

	log.DebugContext(ctx, "Parsing token amount",
		slog.String("amount_type", amountType),
		slog.String("raw_amount", uiTokenAmount.Amount),
		slog.Int("decimals", int(uiTokenAmount.Decimals)))

	amount, err := parseTokenAmount(uiTokenAmount)
	if err != nil {
		log.LogParsingError(ctx, "parse_token_amount", amountType, err,
			slog.String("raw_amount", uiTokenAmount.Amount))
		return 0, err
	}

	log.DebugContext(ctx, "Successfully parsed token amount",
		slog.String("amount_type", amountType),
		slog.Uint64("parsed_amount", amount))

	return amount, nil
}

// analyzeCounterAssetChangesWithLogging wraps analyzeCounterAssetChanges with logging
func analyzeCounterAssetChangesWithLogging(ctx context.Context, tx *solana.TransactionDetails, xsolIndex int, tradeSide string, log *logger.Logger) (uint64, string) {
	log.DebugContext(ctx, "Analyzing counter asset changes",
		slog.Int("xsol_index", xsolIndex),
		slog.String("trade_side", tradeSide))

	counterAmount, counterAsset := analyzeCounterAssetChanges(tx, xsolIndex, tradeSide)

	log.DebugContext(ctx, "Counter asset analysis completed",
		slog.Uint64("counter_amount", counterAmount),
		slog.String("counter_asset", counterAsset))

	return counterAmount, counterAsset
}

// detectHyloInstructions checks if the transaction contains Hylo program instructions
func detectHyloInstructions(tx *solana.TransactionDetails) string {
	hyloConfig := NewConfig()

	// Check each instruction for Hylo program involvement
	for _, instruction := range tx.Transaction.Message.Instructions {
		if int(instruction.ProgramIdIndex) >= len(tx.Transaction.Message.AccountKeys) {
			continue
		}

		programId := solana.Address(tx.Transaction.Message.AccountKeys[instruction.ProgramIdIndex])

		// Check if this is a Hylo Exchange program instruction
		if programId == hyloConfig.GetExchangeProgramID() {
			// For now, we'll infer the instruction type from the transaction structure
			// In a more sophisticated implementation, we'd decode the instruction data

			// Check transaction structure to determine if it's mint or redeem
			// This is a simplified approach - in reality we'd need instruction parsing
			return inferHyloInstructionType(tx)
		}
	}

	return ""
}

// inferHyloInstructionType infers whether this is a mint_levercoin or redeem_levercoin operation
// This is a simplified approach based on balance changes until we implement full instruction parsing
func inferHyloInstructionType(tx *solana.TransactionDetails) string {
	// Look for xSOL balance changes to determine operation type
	for _, preBalance := range tx.Meta.PreTokenBalances {
		for _, postBalance := range tx.Meta.PostTokenBalances {
			if preBalance.AccountIndex == postBalance.AccountIndex {
				// Check if this is the xSOL token
				if preBalance.Mint == string(tokens.XSOLMint) {
					preAmount, _ := parseTokenAmount(preBalance.UITokenAmount)
					postAmount, _ := parseTokenAmount(postBalance.UITokenAmount)

					if postAmount > preAmount {
						return MintLeverCoinInstruction // BUY - xSOL increased
					} else if preAmount > postAmount {
						return RedeemLeverCoinInstruction // SELL - xSOL decreased
					}
				}
			}
		}
	}

	// If we can't find matching pre/post balances, check if there's only a post balance (first-time mint)
	for _, postBalance := range tx.Meta.PostTokenBalances {
		if postBalance.Mint == string(tokens.XSOLMint) {
			// Check if there's no corresponding pre balance (new account)
			hasPreBalance := false
			for _, preBalance := range tx.Meta.PreTokenBalances {
				if preBalance.AccountIndex == postBalance.AccountIndex {
					hasPreBalance = true
					break
				}
			}
			if !hasPreBalance {
				return MintLeverCoinInstruction // This is a first-time mint (BUY)
			}
		}
	}

	return ""
}

// parseHyloTrade parses a transaction that contains Hylo program instructions
func parseHyloTrade(ctx context.Context, tx *solana.TransactionDetails, walletXSOLATA solana.Address, xsolAccountIndex int, instructionType, signature string, log *logger.Logger) (*TradeParseResult, error) {
	startTime := time.Now()

	log.DebugContext(ctx, "Parsing Hylo trade transaction",
		slog.String("signature", signature),
		slog.String("instruction_type", instructionType),
		slog.String("ata_address", walletXSOLATA.String()))

	// Get pre and post balances for xSOL
	preTokenBalance := findTokenBalance(tx.Meta.PreTokenBalances, uint32(xsolAccountIndex))
	postTokenBalance := findTokenBalance(tx.Meta.PostTokenBalances, uint32(xsolAccountIndex))

	// Handle different scenarios
	var preAmount, postAmount uint64
	var err error

	// Parse pre-amount (may be 0 for first-time trades)
	if preTokenBalance != nil {
		preAmount, err = parseTokenAmountWithLogging(ctx, preTokenBalance.UITokenAmount, log, "pre-amount")
		if err != nil {
			log.LogParsingError(ctx, "parse_hylo_trade", "pre_token_amount", err,
				slog.String("signature", signature))
			return &TradeParseResult{
				Error: fmt.Sprintf("failed to parse pre-token amount: %v", err),
			}, nil
		}
	}

	// Parse post-amount (should always exist for successful trades)
	if postTokenBalance == nil {
		log.LogParsingError(ctx, "parse_hylo_trade", "post_token_balance", fmt.Errorf("post token balance not found"),
			slog.String("signature", signature))
		return &TradeParseResult{
			Error: "post token balance not found for Hylo trade",
		}, nil
	}

	postAmount, err = parseTokenAmountWithLogging(ctx, postTokenBalance.UITokenAmount, log, "post-amount")
	if err != nil {
		log.LogParsingError(ctx, "parse_hylo_trade", "post_token_amount", err,
			slog.String("signature", signature))
		return &TradeParseResult{
			Error: fmt.Sprintf("failed to parse post-token amount: %v", err),
		}, nil
	}

	// Calculate xSOL amount and determine trade direction
	var xsolAmount uint64
	var tradeSide string

	if instructionType == MintLeverCoinInstruction {
		tradeSide = TradeSideBuy
		xsolAmount = postAmount - preAmount
		log.DebugContext(ctx, "Detected Hylo BUY trade",
			slog.String("signature", signature),
			slog.Uint64("pre_amount", preAmount),
			slog.Uint64("post_amount", postAmount),
			slog.Uint64("xsol_amount", xsolAmount))
	} else if instructionType == RedeemLeverCoinInstruction {
		tradeSide = TradeSideSell
		xsolAmount = preAmount - postAmount
		log.DebugContext(ctx, "Detected Hylo SELL trade",
			slog.String("signature", signature),
			slog.Uint64("pre_amount", preAmount),
			slog.Uint64("post_amount", postAmount),
			slog.Uint64("xsol_amount", xsolAmount))
	} else {
		log.WarnContext(ctx, "Unknown Hylo instruction type",
			slog.String("signature", signature),
			slog.String("instruction_type", instructionType))
		return &TradeParseResult{
			Error: fmt.Sprintf("unknown Hylo instruction type: %s", instructionType),
		}, nil
	}

	// Skip if no actual xSOL change occurred
	if xsolAmount == 0 {
		log.DebugContext(ctx, "No xSOL balance change in Hylo trade",
			slog.String("signature", signature))
		return &TradeParseResult{}, nil
	}

	// Create base trade object
	var blockTime int64
	if tx.BlockTime != nil {
		blockTime = *tx.BlockTime
	}

	trade := NewXSOLTrade(signature, uint64(tx.Slot), blockTime)

	// Analyze other account balance changes to determine counter-asset
	counterAmount, counterAsset := analyzeCounterAssetChangesWithLogging(ctx, tx, xsolAccountIndex, tradeSide, log)

	// Set trade details
	trade.SetTradeDetails(tradeSide, xsolAmount, counterAmount, counterAsset)

	// Calculate historical price for hyUSD trades
	trade.HistoricalPriceUSD = CalculateHistoricalXSOLPrice(trade)

	// Log successful trade parsing with historical price info
	priceInfo := "N/A (SOL trade)"
	if trade.HistoricalPriceUSD != nil {
		priceInfo = fmt.Sprintf("$%s", *trade.HistoricalPriceUSD)
	}

	log.InfoContext(ctx, "Successfully parsed Hylo xSOL trade",
		slog.String("signature", signature),
		slog.String("side", tradeSide),
		slog.Uint64("xsol_amount", xsolAmount),
		slog.Uint64("counter_amount", counterAmount),
		slog.String("counter_asset", counterAsset),
		slog.String("historical_price", priceInfo),
		slog.Duration("parse_time", time.Since(startTime)))

	return &TradeParseResult{
		Trade: trade,
	}, nil
}

// parseInitialFundingTransaction handles transactions where xSOL tokens are received for the first time
// This covers initial mints, transfers, or other funding operations (NON-Hylo transactions only)
func parseInitialFundingTransaction(ctx context.Context, tx *solana.TransactionDetails, postTokenBalance *solana.TokenBalance, walletXSOLATA solana.Address, signature string, log *logger.Logger) (*TradeParseResult, error) {
	startTime := time.Now()

	log.DebugContext(ctx, "Parsing initial funding transaction",
		slog.String("signature", signature),
		slog.String("ata_address", walletXSOLATA.String()))

	// Parse the post-balance amount (this is the amount received)
	receivedAmount, err := parseTokenAmountWithLogging(ctx, postTokenBalance.UITokenAmount, log, "received-amount")
	if err != nil {
		log.LogParsingError(ctx, "parse_initial_funding", "post_token_amount", err,
			slog.String("signature", signature))
		return &TradeParseResult{
			Error: fmt.Sprintf("failed to parse received token amount: %v", err),
		}, nil
	}

	// Skip if no tokens were actually received
	if receivedAmount == 0 {
		log.DebugContext(ctx, "No tokens received in initial funding transaction",
			slog.String("signature", signature))
		return &TradeParseResult{}, nil
	}

	// Create base trade object
	var blockTime int64
	if tx.BlockTime != nil {
		blockTime = *tx.BlockTime
	}

	trade := NewXSOLTrade(signature, uint64(tx.Slot), blockTime)

	// Set trade details for RECEIVE operation
	// For initial funding, there's no counter asset exchange, so we leave it empty
	trade.SetTradeDetails(TradeSideReceive, receivedAmount, 0, "")

	log.InfoContext(ctx, "Successfully parsed initial xSOL funding",
		slog.String("signature", signature),
		slog.String("side", TradeSideReceive),
		slog.Uint64("xsol_amount", receivedAmount),
		slog.Duration("parse_time", time.Since(startTime)))

	return &TradeParseResult{
		Trade: trade,
	}, nil
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
