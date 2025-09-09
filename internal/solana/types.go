package solana

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Commitment represents the level of commitment desired when querying state
type Commitment string

const (
	CommitmentProcessed Commitment = "processed"
	CommitmentConfirmed Commitment = "confirmed"
	CommitmentFinalized Commitment = "finalized"
)

// Validate checks if the commitment level is valid
func (c Commitment) Validate() error {
	switch c {
	case CommitmentProcessed, CommitmentConfirmed, CommitmentFinalized:
		return nil
	default:
		return fmt.Errorf("invalid commitment level: %s", c)
	}
}

// Address represents a Solana public key/address
type Address string

// Validate checks if the address is a valid base58 encoded public key (44 characters)
func (a Address) Validate() error {
	if len(string(a)) != 44 {
		return fmt.Errorf("invalid address length: %d, expected 44", len(string(a)))
	}
	// Basic validation - could be enhanced with full base58 decoding
	return nil
}

// String returns the string representation of the address
func (a Address) String() string {
	return string(a)
}

// Signature represents a transaction signature
type Signature string

// Validate checks if the signature is valid
func (s Signature) Validate() error {
	if len(string(s)) < 64 || len(string(s)) > 128 {
		return fmt.Errorf("invalid signature length: %d", len(string(s)))
	}
	return nil
}

// String returns the string representation of the signature
func (s Signature) String() string {
	return string(s)
}

// Slot represents a Solana slot number
type Slot uint64

// AccountInfo represents account data from Solana
type AccountInfo struct {
	Lamports   uint64 `json:"lamports"`
	Data       []byte `json:"data"`
	Owner      string `json:"owner"`
	Executable bool   `json:"executable"`
	RentEpoch  uint64 `json:"rentEpoch"`
}

// UnmarshalJSON handles the custom data field decoding and large numbers
func (ai *AccountInfo) UnmarshalJSON(data []byte) error {
	// Use a temporary struct to handle all fields properly
	aux := struct {
		Lamports   uint64      `json:"lamports"`
		Data       interface{} `json:"data"`
		Owner      string      `json:"owner"`
		Executable bool        `json:"executable"`
		RentEpoch  interface{} `json:"rentEpoch"` // Handle as interface for large numbers
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Set the simple fields
	ai.Lamports = aux.Lamports
	ai.Owner = aux.Owner
	ai.Executable = aux.Executable

	// Handle RentEpoch - safely convert large numbers
	switch v := aux.RentEpoch.(type) {
	case float64:
		// For very large numbers that exceed uint64, cap at MaxUint64
		if v > 9223372036854775807 { // MaxInt64, close to MaxUint64
			ai.RentEpoch = 18446744073709551615 // MaxUint64
		} else {
			ai.RentEpoch = uint64(v)
		}
	case int64:
		ai.RentEpoch = uint64(v)
	case string:
		// Try to parse string as uint64, fallback to 0
		if parsed, err := json.Number(v).Int64(); err == nil {
			ai.RentEpoch = uint64(parsed)
		} else {
			ai.RentEpoch = 0
		}
	default:
		ai.RentEpoch = 0 // Safe fallback
	}

	// Handle data field which can be array or base64 string
	switch v := aux.Data.(type) {
	case []interface{}:
		// Data as array format [base64_string, encoding]
		if len(v) >= 1 {
			if dataStr, ok := v[0].(string); ok {
				decoded, err := base64.StdEncoding.DecodeString(dataStr)
				if err != nil {
					return fmt.Errorf("failed to decode base64 data: %w", err)
				}
				ai.Data = decoded
			}
		}
	case string:
		// Data as direct base64 string
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return fmt.Errorf("failed to decode base64 data: %w", err)
		}
		ai.Data = decoded
	}

	return nil
}

// TransactionDetails represents transaction information
type TransactionDetails struct {
	BlockTime   *int64      `json:"blockTime"`
	Meta        *TxMeta     `json:"meta"`
	Slot        Slot        `json:"slot"`
	Transaction Transaction `json:"transaction"`
}

// TxMeta contains transaction metadata
type TxMeta struct {
	Err               interface{}    `json:"err"`
	Fee               uint64         `json:"fee"`
	LogMessages       []string       `json:"logMessages"`
	PreBalances       []uint64       `json:"preBalances"`
	PostBalances      []uint64       `json:"postBalances"`
	PreTokenBalances  []TokenBalance `json:"preTokenBalances"`
	PostTokenBalances []TokenBalance `json:"postTokenBalances"`
}

// TokenBalance represents a token balance in transaction metadata
type TokenBalance struct {
	AccountIndex  uint32         `json:"accountIndex"`
	Mint          string         `json:"mint"`
	Owner         *string        `json:"owner,omitempty"`
	ProgramId     *string        `json:"programId,omitempty"`
	UITokenAmount *UITokenAmount `json:"uiTokenAmount,omitempty"`
}

// UITokenAmount provides the token amount in different formats
type UITokenAmount struct {
	Amount         string   `json:"amount"`         // Raw amount as string
	Decimals       uint8    `json:"decimals"`       // Number of decimals
	UIAmount       *float64 `json:"uiAmount"`       // Human readable amount
	UIAmountString string   `json:"uiAmountString"` // Human readable as string
}

// Transaction contains the actual transaction data
type Transaction struct {
	Message    TxMessage `json:"message"`
	Signatures []string  `json:"signatures"`
}

// TxMessage contains the transaction message
type TxMessage struct {
	AccountKeys     []string        `json:"accountKeys"`
	Instructions    []TxInstruction `json:"instructions"`
	RecentBlockhash string          `json:"recentBlockhash"`
}

// TxInstruction represents an instruction in a transaction
type TxInstruction struct {
	ProgramIdIndex uint8   `json:"programIdIndex"`
	Accounts       []uint8 `json:"accounts"`
	Data           string  `json:"data"`
}

// SignatureInfo represents signature information from getSignaturesForAddress
type SignatureInfo struct {
	Signature          string      `json:"signature"`
	Slot               Slot        `json:"slot"`
	BlockTime          *int64      `json:"blockTime"`
	ConfirmationStatus *string     `json:"confirmationStatus"`
	Err                interface{} `json:"err"`
	Memo               *string     `json:"memo"`
}

// GetTime returns the block time as a time.Time, or zero time if nil
func (si *SignatureInfo) GetTime() time.Time {
	if si.BlockTime == nil {
		return time.Time{}
	}
	return time.Unix(*si.BlockTime, 0)
}
