// validation.go contains data logic validations across the LedgerKeeper and other module Keepers
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

// Returns an error if the Ledger exists
func (k LedgerKeeper) validateLedgerNotExists(ctx sdk.Context, l *ledger.Ledger) error {
	if k.HasLedger(ctx, l.NftAddress) {
		return NewLedgerCodedError(ErrCodeAlreadyExists, "object[ledger]")
	}

	return nil
}
