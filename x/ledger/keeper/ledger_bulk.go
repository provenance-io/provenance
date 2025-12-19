package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// BulkCreate creates ledgers and their entries. This function assumes that ledger classes, status types, entry types,
// and bucket types are already created before calling this function.
func (k Keeper) BulkCreate(goCtx context.Context, ledgers []*types.LedgerAndEntries) error {
	// Import ledgers and their entries
	for i, ledgerAndEntries := range ledgers {
		// Use a cache context so that an entry is all or nothing.
		ctx, writeCache := sdk.UnwrapSDKContext(goCtx).CacheContext()

		// If we have a ledger object, add it.
		// All errors are handled in the AddLedger function (dupe, ledger class, etc).
		if ledgerAndEntries.Ledger != nil {
			if err := k.AddLedger(ctx, *ledgerAndEntries.Ledger); err != nil {
				return fmt.Errorf("[%d]: error adding ledger: %w", i, err)
			}
			// Charge for creating one ledger.
			antewrapper.ConsumeMsg(ctx, &types.MsgCreateLedgerRequest{})
		}

		// Add ledger entries
		// All errors are handled in the AppendEntries function (dupes, ledger, etc).
		if len(ledgerAndEntries.Entries) > 0 {
			key := ledgerAndEntries.GetKey()
			if key == nil {
				return fmt.Errorf("[%d]: no ledger key provided", i)
			}
			if err := k.AppendEntries(ctx, key, ledgerAndEntries.Entries); err != nil {
				return fmt.Errorf("[%d]: error appending entries: %w", i, err)
			}
			// Charge for appending entries.
			antewrapper.ConsumeMsg(ctx, &types.MsgAppendRequest{})
		}

		// Done with this entry, write it!
		writeCache()
	}

	return nil
}
