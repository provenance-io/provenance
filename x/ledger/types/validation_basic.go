// validation_basic.go contains validation functions for basic data uniformity.
// No data access, or operations should be performed on data from this file.
package types

import (
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/provenance-io/provenance/x/ledger/helper"
)

func ValidateLedgerClassBasic(l *LedgerClass) error {
	if emptyString(&l.LedgerClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "ledger_class_id")
	}

	if emptyString(&l.AssetClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "asset_class_id")
	}

	if emptyString(&l.Denom) {
		return NewLedgerCodedError(ErrCodeMissingField, "denom")
	}

	maintainerAddress := l.MaintainerAddress
	if emptyString(&maintainerAddress) {
		return NewLedgerCodedError(ErrCodeMissingField, "maintainer_address")
	}

	return nil
}

func ValidateLedgerKeyBasic(key *LedgerKey) error {
	if emptyString(&key.NftId) {
		return NewLedgerCodedError(ErrCodeMissingField, "nft_id")
	}
	if emptyString(&key.AssetClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "asset_class_id")
	}
	return nil
}

func ValidateLedgerBasic(l *Ledger) error {
	if l == nil {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger", "ledger cannot be nil")
	}

	if err := ValidateLedgerKeyBasic(l.Key); err != nil {
		return err
	}

	// Validate the LedgerClassId field
	if emptyString(&l.LedgerClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "ledger_class_id")
	}

	keyError := ValidateLedgerKeyBasic(l.Key)
	if keyError != nil {
		return keyError
	}

	epochTime, _ := time.Parse("2006-01-02", "1970-01-01")
	epoch := helper.DaysSinceEpoch(epochTime.UTC())

	// Validate next payment date format if provided
	if l.NextPmtDate < epoch {
		return NewLedgerCodedError(ErrCodeInvalidField, "next_pmt_date", "must be after 1970-01-01")
	}

	// Validate next payment amount if provided
	if l.NextPmtAmt < 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "next_pmt_amt", "must be a non-negative integer")
	}

	// Validate interest rate if provided
	if l.InterestRate < 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "interest_rate", "must be a non-negative integer")
	}

	// Validate maturity date format if provided
	if l.MaturityDate < epoch {
		return NewLedgerCodedError(ErrCodeInvalidField, "maturity_date", "must be after 1970-01-01")
	}

	return nil
}

func ValidateLedgerEntryBasic(e *LedgerEntry) error {
	if emptyString(&e.CorrelationId) {
		return NewLedgerCodedError(ErrCodeMissingField, "correlation_id")
	} else {
		if !IsCorrelationIDValid(e.CorrelationId) {
			return NewLedgerCodedError(ErrCodeInvalidField, "correlation_id", "must be a valid string that is less than 50 characters")
		}
	}

	if e.PostedDate <= 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "posted_date", "must be a valid integer")
	}

	if e.EffectiveDate <= 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "effective_date", "must be a valid integer")
	}

	// Validate amounts are non-negative
	if e.TotalAmt.LT(math.NewInt(0)) {
		return NewLedgerCodedError(ErrCodeInvalidField, "total_amt", "must be a non-negative integer")
	}

	return nil
}

func ValidateBucketBalance(bb *BucketBalance) error {
	if bb.BucketTypeId <= 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "bucket_type_id", "must be a positive integer")
	}

	return nil
}

func ValidateFundTransferBasic(ft *FundTransfer) error {
	if err := ValidateLedgerKeyBasic(ft.Key); err != nil {
		return err
	}

	if ft.LedgerEntryCorrelationId != "" {
		if !IsCorrelationIDValid(ft.LedgerEntryCorrelationId) {
			return NewLedgerCodedError(ErrCodeInvalidField, "ledger_entry_correlation_id", "must be a valid string that is less than 50 characters")
		}
	}

	if ft.Amount.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "amount", "must be a non-negative integer")
	}

	if ft.Memo != "" {
		if len(ft.Memo) > 50 {
			return NewLedgerCodedError(ErrCodeInvalidField, "memo", "must be less than 50 characters")
		}
	}

	if ft.Status != 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "status", "must remain unspecified since it is used internally to track the state of settlement of future transfers")
	}

	return nil
}

func ValidateFundTransferWithSettlementBasic(ft *FundTransferWithSettlement) error {
	if err := ValidateLedgerKeyBasic(ft.Key); err != nil {
		return err
	}

	// Validate that the correlation ID is valid
	if !IsCorrelationIDValid(ft.LedgerEntryCorrelationId) {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger_entry_correlation_id", "must be a valid string that is less than 50 characters")
	}

	// Validate that the settlement instructions are valid
	for _, settlementInstruction := range ft.SettlementInstructions {
		if err := ValidateSettlementInstructionBasic(settlementInstruction); err != nil {
			return err
		}
	}

	return nil
}

func ValidateSettlementInstructionBasic(si *SettlementInstruction) error {
	if si.Amount.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "amount", "must be a non-negative integer")
	}

	if si.Memo != "" {
		if len(si.Memo) > 50 {
			return NewLedgerCodedError(ErrCodeInvalidField, "memo", "must be less than 50 characters")
		}
	}

	if si.RecipientAddress == "" {
		return NewLedgerCodedError(ErrCodeMissingField, "recipient_address")
	}

	return nil
}

// Returns true if the string is nil or empty(TrimSpace(*s))
func emptyString(s *string) bool {
	if s == nil || strings.TrimSpace(*s) == "" {
		return true
	}
	return false
}

// validateCorrelationID validates that the provided string is a valid correlation ID.
// Returns true if the string is a valid correlation ID (non-empty and max 50 characters), false otherwise.
func IsCorrelationIDValid(correlationID string) bool {
	if len(correlationID) == 0 || len(correlationID) > 50 {
		return false
	}
	return true
}
