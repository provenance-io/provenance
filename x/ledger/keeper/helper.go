// validation.go contains data logic validations across the LedgerKeeper and other module Keepers
package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
)

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
