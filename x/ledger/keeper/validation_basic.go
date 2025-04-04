// validation_basic.go contains validation functions for basic data uniformity.
// No data access, or operations should be performed on data from this file.
package keeper

import (
	"strings"

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
