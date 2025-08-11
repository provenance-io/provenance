package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// BulkCreate creates ledgers and their entries. This function assumes that ledger classes, status types, entry types,
// and bucket types are already created before calling this function.
func (k Keeper) BulkCreate(ctx sdk.Context, ledgers []*types.LedgerToEntries) error {
	// Import ledgers and their entries
	for _, ledgerToEntries := range ledgers {
		// If we have a ledger object, add it.
		// All errors are handled in the AddLedger function (dupe, ledger class, etc).
		if ledgerToEntries.Ledger != nil {
			if err := k.AddLedger(ctx, *ledgerToEntries.Ledger); err != nil {
				return err
			}
		}

		// Add ledger entries
		// All errors are handled in the AppendEntries function (dupes, ledger, etc).
		if err := k.AppendEntries(ctx, ledgerToEntries.LedgerKey, ledgerToEntries.Entries); err != nil {
			return err
		}
	}

	return nil
}
