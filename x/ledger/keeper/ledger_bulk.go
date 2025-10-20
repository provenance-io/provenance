package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// BulkCreate creates ledgers and their entries. This function assumes that ledger classes, status types, entry types,
// and bucket types are already created before calling this function.
func (k Keeper) BulkCreate(goCtx context.Context, ledgers []*types.LedgerAndEntries) error {
	// Import ledgers and their entries
	for _, ledgerAndEntries := range ledgers {
		// Use a cache context so that an entry is all or nothing.
		ctx, writeCache := sdk.UnwrapSDKContext(goCtx).CacheContext()

		// If we have a ledger object, add it.
		// All errors are handled in the AddLedger function (dupe, ledger class, etc).
		if ledgerAndEntries.Ledger != nil {
			if err := k.AddLedger(ctx, *ledgerAndEntries.Ledger); err != nil {
				return err
			}
		}

		// Add ledger entries
		// All errors are handled in the AppendEntries function (dupes, ledger, etc).
		if len(ledgerAndEntries.Entries) > 0 {
			key := ledgerAndEntries.LedgerKey
			if key == nil && ledgerAndEntries.Ledger != nil {
				key = ledgerAndEntries.Ledger.Key
			}
			if err := k.AppendEntries(ctx, key, ledgerAndEntries.Entries); err != nil {
				return err
			}
		}

		// Charge for creating one ledger.
		antewrapper.ConsumeMsg(ctx, &types.MsgCreateLedgerRequest{})
		// Done with this entry, write it!
		writeCache()
	}

	return nil
}
