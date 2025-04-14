// validation_basic.go contains validation functions for basic data uniformity.
// No data access, or operations should be performed on data from this file.
package keeper

import (
	"cosmossdk.io/math"
	"github.com/provenance-io/provenance/x/ledger"
)

func ValidateLedgerBasic(l *ledger.Ledger) error {
	if emptyString(&l.Denom) {
		return NewLedgerCodedError(ErrCodeMissingField, "denom")
	}
	if emptyString(&l.NftAddress) {
		return NewLedgerCodedError(ErrCodeMissingField, "nft_address")
	}

	// Validate next payment date format if provided
	if !emptyString(&l.NextPmtDate) {
		if _, err := parseIS08601Date(l.NextPmtDate); err != nil {
			return NewLedgerCodedError(ErrCodeInvalidField, "next_pmt_date")
		}
	}

	// Validate next payment amount if provided
	if !emptyString(&l.NextPmtAmt) {
		if _, ok := math.NewIntFromString(l.NextPmtAmt); !ok {
			return NewLedgerCodedError(ErrCodeInvalidField, "next_pmt_amt")
		}
	}

	// Validate status if provided
	if !emptyString(&l.Status) {
		// Add any specific status validation here if needed
	}

	// Validate interest rate if provided
	if !emptyString(&l.InterestRate) {
		if _, ok := math.NewIntFromString(l.InterestRate); !ok {
			return NewLedgerCodedError(ErrCodeInvalidField, "interest_rate")
		}
	}

	// Validate maturity date format if provided
	if !emptyString(&l.MaturityDate) {
		if _, err := parseIS08601Date(l.MaturityDate); err != nil {
			return NewLedgerCodedError(ErrCodeInvalidField, "maturity_date")
		}
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

	// Validate dates are valid
	if _, err := parseIS08601Date(e.PostedDate); err != nil {
		return NewLedgerCodedError(ErrCodeInvalidField, "posted_date")
	}
	if _, err := parseIS08601Date(e.EffectiveDate); err != nil {
		return NewLedgerCodedError(ErrCodeInvalidField, "effective_date")
	}

	// Validate amounts are non-negative
	if e.Amt.IsNil() || e.Amt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "amount")
	}
	if e.PrinAppliedAmt.IsNil() || e.PrinAppliedAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "principal_applied_amount")
	}
	if e.PrinBalAmt.IsNil() {
		return NewLedgerCodedError(ErrCodeInvalidField, "principal_balance_amount")
	}
	if e.IntAppliedAmt.IsNil() || e.IntAppliedAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "interest_applied_amount")
	}
	if e.IntBalAmt.IsNil() {
		return NewLedgerCodedError(ErrCodeInvalidField, "interest_balance_amount")
	}
	if e.OtherAppliedAmt.IsNil() || e.OtherAppliedAmt.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "other_applied_amount")
	}
	if e.OtherBalAmt.IsNil() {
		return NewLedgerCodedError(ErrCodeInvalidField, "other_balance_amount")
	}

	return nil
}
