// validation_basic.go contains validation functions for basic data uniformity.
// No data access, or operations should be performed on data from this file.
package keeper

import (
	"time"

	"cosmossdk.io/math"
	"github.com/provenance-io/provenance/x/ledger"
)

func ValidateLedgerBasic(l *ledger.Ledger) error {
	if emptyString(&l.LedgerClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "ledger_class_id")
	}
	if emptyString(&l.NftId) {
		return NewLedgerCodedError(ErrCodeMissingField, "nft_id")
	}

	epochTime, _ := time.Parse("2006-01-02", "1970-01-01")
	epoch := DaysSinceEpoch(epochTime.UTC())

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

func ValidateLedgerEntryBasic(e *ledger.LedgerEntry) error {
	if emptyString(&e.CorrelationId) {
		return NewLedgerCodedError(ErrCodeMissingField, "correlation_id")
	} else {
		if !isCorrelationIDValid(e.CorrelationId) {
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
