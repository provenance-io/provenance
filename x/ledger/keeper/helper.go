// validation.go contains data logic validations across the LedgerKeeper and other module Keepers
package keeper

import (
	"sort"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/provenance-io/provenance/x/ledger"
)

// StrPtr returns a pointer to the string s.
func StrPtr(s string) *string {
	return &s
}

func getAddress(s *string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(*s)
	if err != nil || addr == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "nft_address")
	}

	return addr, nil
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

// validateCorrelationID validates that the provided string is a valid correlation ID.
// Returns true if the string is a valid correlation ID (non-empty and max 50 characters), false otherwise.
func isCorrelationIDValid(correlationID string) bool {
	if len(correlationID) == 0 || len(correlationID) > 50 {
		return false
	}
	return true
}

// parseIS08601Date parses an ISO 8601 date string 2006-01-02 into a time.Time object.
func parseIS08601Date(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// sortLedgerEntries sorts the ledger entries by effective date and then by sequence.
func sortLedgerEntries(entries []*ledger.LedgerEntry) {
	sort.Slice(entries, func(i, j int) bool {
		effectiveDateI, err := parseIS08601Date((entries)[i].EffectiveDate)
		if err != nil {
			return false
		}
		effectiveDateJ, err := parseIS08601Date((entries)[j].EffectiveDate)
		if err != nil {
			return false
		}

		if effectiveDateI.Equal(effectiveDateJ) {
			return (entries)[i].Sequence < (entries)[j].Sequence
		}
		return effectiveDateI.Before(effectiveDateJ)
	})
}
