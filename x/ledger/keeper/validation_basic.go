// validation_basic.go contains validation functions for basic data uniformity.
// No data access, or operations should be performed on data from this file.
package keeper

import (
	"strings"

	"github.com/google/uuid"
	"github.com/provenance-io/provenance/x/ledger"
)

func validateLedgerBasic(l *ledger.Ledger) error {
	if emptyString(&l.Denom) {
		return NewLedgerCodedError(ErrCodeMissingField, "denom")
	}
	if emptyString(&l.NftAddress) {
		return NewLedgerCodedError(ErrCodeMissingField, "nft_address")
	}

	return nil
}

func validateLedgerEntryBasic(e *ledger.LedgerEntry) error {
	if emptyString(&e.Uuid) {
		return NewLedgerCodedError(ErrCodeMissingField, "uuid")
	} else {
		if !isUUIDValid(e.Uuid) {
			return NewLedgerCodedError(ErrCodeInvalidField, "uuid")
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

// Returns true if the string is nil or empty(TrimSpace(*s))
func emptyString(s *string) bool {
	if s == nil || strings.TrimSpace(*s) == "" {
		return true
	}
	return false
}

// validateUUID validates that the provided string is a valid UUID.
// Returns true if the string is a valid UUID, false otherwise.
func isUUIDValid(uuidStr string) bool {
	if _, err := uuid.Parse(uuidStr); err != nil {
		return false
	}
	return true
}
