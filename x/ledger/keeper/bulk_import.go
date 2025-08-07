package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// TODO move this to init genesis.
// BulkImportLedgerData imports ledger data from genesis state. This function assumes that ledger classes, status
// types, entry types, and bucket types are already created before calling this function.
func (k Keeper) BulkImportLedgerData(ctx sdk.Context, genesisState types.GenesisState) error {
	ctx.Logger().Info("Starting bulk import of ledger data",
		"ledger_to_entries", len(genesisState.LedgerToEntries))

	// Import ledgers and their entries
	for _, ledgerToEntries := range genesisState.LedgerToEntries {
		var ledgerClassId string

		// Determine the ledger class ID and get maintainer address
		if ledgerToEntries.Ledger != nil {
			// If we have a ledger object, use its ledger class ID
			ledgerClassId = ledgerToEntries.Ledger.LedgerClassId
		} else {
			// If we don't have a ledger object, get it from the existing ledger
			existingLedger, err := k.GetLedger(ctx, ledgerToEntries.LedgerKey)
			if err != nil {
				return fmt.Errorf("failed to get existing ledger: %w", err)
			}
			if existingLedger == nil {
				return fmt.Errorf("ledger %s does not exist and no ledger object provided", ledgerToEntries.LedgerKey.NftId)
			}
			ledgerClassId = existingLedger.LedgerClassId
		}

		// Get the maintainer address from the ledger class
		_, err := k.RequireGetLedgerClass(ctx, ledgerClassId)
		if err != nil {
			return types.NewErrCodeNotFound("ledger_class")
		}

		// Create the ledger only if it doesn't already exist
		if ledgerToEntries.Ledger != nil && !k.HasLedger(ctx, ledgerToEntries.Ledger.Key) {
			if err := k.AddLedger(ctx, *ledgerToEntries.Ledger); err != nil {
				return fmt.Errorf("failed to create ledger: %w", err)
			}
			ctx.Logger().Info("Created ledger", "nft_id", ledgerToEntries.Ledger.Key.NftId, "asset_class", ledgerToEntries.Ledger.Key.AssetClassId)
		} else if ledgerToEntries.Ledger != nil {
			ctx.Logger().Info("Ledger already exists, skipping creation", "nft_id", ledgerToEntries.Ledger.Key.NftId, "asset_class", ledgerToEntries.Ledger.Key.AssetClassId)
		}

		// If the ledger doesn't exist, we can't add entries to it
		if !k.HasLedger(ctx, ledgerToEntries.LedgerKey) {
			return fmt.Errorf("ledger %s does not exist", ledgerToEntries.LedgerKey.NftId)
		}

		// Add ledger entries
		if len(ledgerToEntries.Entries) > 0 {
			entries := make([]*types.LedgerEntry, len(ledgerToEntries.Entries))
			for i, entry := range ledgerToEntries.Entries {
				entries[i] = entry
			}

			if err := k.AppendEntries(ctx, ledgerToEntries.LedgerKey, entries); err != nil {
				return fmt.Errorf("failed to append entries for ledger key %s: %w", ledgerToEntries.LedgerKey.NftId, err)
			}
			ctx.Logger().Info("Added ledger entries", "ledger_key", ledgerToEntries.LedgerKey.NftId, "count", len(entries))
		}
	}

	ctx.Logger().Info("Successfully completed bulk import of ledger data",
		"ledger_to_entries", len(genesisState.LedgerToEntries))

	return nil
}
