// validation_basic.go contains validation functions for basic data uniformity.
// No data access, or operations should be performed on data from this file.
package keeper

import (
	"github.com/provenance-io/provenance/x/ledger"
)

func ValidateLedgerBasic(l *ledger.Ledger) error {
	if emptyString(&l.Denom) {
		return NewLedgerCodedError(ErrCodeMissingField, "denom")
	}
	if emptyString(&l.NftAddress) {
		return NewLedgerCodedError(ErrCodeMissingField, "nft_address")
	}

	return nil
}

func ValidateLedgerEntryBasic(e *ledger.LedgerEntry) error {
	if emptyString(&e.CorrelationId) {
		return NewLedgerCodedError(ErrCodeMissingField, "correlation_id")
	} else {
		if !isCorrelationIDValid(e.CorrelationId) {
			return NewLedgerCodedError(ErrCodeInvalidField, "correlation_id")
		}
	}

	// Validate entry type is set
	if e.Type == ledger.LedgerEntryType_Unspecified {
		return NewLedgerCodedError(ErrCodeMissingField, "type")
	}

	// Validate dates are set
	if e.PostedDate.IsZero() {
		return NewLedgerCodedError(ErrCodeMissingField, "posted_date")
	}
	if e.EffectiveDate.IsZero() {
		return NewLedgerCodedError(ErrCodeMissingField, "effective_date")
	}

	// Validate amounts are non-negative
	if e.Amt.IsNil() || e.Amt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "amount")
	}
	if e.PrinAppliedAmt.IsNil() || e.PrinAppliedAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "principal_applied_amount")
	}
	if e.PrinBalAmt.IsNil() || e.PrinBalAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "principal_balance_amount")
	}
	if e.IntAppliedAmt.IsNil() || e.IntAppliedAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "interest_applied_amount")
	}
	if e.IntBalAmt.IsNil() || e.IntBalAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "interest_balance_amount")
	}
	if e.OtherAppliedAmt.IsNil() || e.OtherAppliedAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "other_applied_amount")
	}
	if e.OtherBalAmt.IsNil() || e.OtherBalAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "other_balance_amount")
	}

	return nil
}
