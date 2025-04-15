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

func DaysSinceEpoch(date time.Time) int32 {
	return int32(date.Sub(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24)
}

func StrToDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
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

// sortLedgerEntries sorts the ledger entries by effective date and then by sequence.
func sortLedgerEntries(entries []*ledger.LedgerEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].EffectiveDate == entries[j].EffectiveDate {
			return (entries)[i].Sequence < (entries)[j].Sequence
		}
		return entries[i].EffectiveDate < entries[j].EffectiveDate
	})
}
