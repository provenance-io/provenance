package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

// ExportGenesis exports the current keeper state of the ledger module.
// This exports data in the format that matches test.json for bulk import.
func (k Keeper) ExportGenesis(ctx sdk.Context) *ledger.GenesisState {
	state := &ledger.GenesisState{}

	// Group ledgers and entries by ledger key to create LedgerToEntries
	ledgerToEntriesMap := make(map[string]*ledger.LedgerToEntries)

	// Export ledgers
	ledgerIter, err := k.Ledgers.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer ledgerIter.Close()

	for ; ledgerIter.Valid(); ledgerIter.Next() {
		key, err := ledgerIter.Key()
		if err != nil {
			panic(err)
		}
		ledgerObj, err := ledgerIter.Value()
		if err != nil {
			panic(err)
		}

		// Reconstruct the ledger key from the storage key
		ledgerKey, err := ledger.StringToLedgerKey(key)
		if err != nil {
			panic(err)
		}

		// Create ledger key string for grouping
		ledgerKeyStr := ledgerKey.String()

		// Restore the key to the ledger object
		ledgerObj.Key = ledgerKey

		ledgerToEntriesMap[ledgerKeyStr] = &ledger.LedgerToEntries{
			LedgerKey: ledgerKey,
			Ledger:    &ledgerObj,
			Entries:   []*ledger.LedgerEntry{},
		}
	}

	// Export ledger entries and group by ledger key
	entryIter, err := k.LedgerEntries.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer entryIter.Close()

	for ; entryIter.Valid(); entryIter.Next() {
		key, err := entryIter.Key()
		if err != nil {
			panic(err)
		}
		entry, err := entryIter.Value()
		if err != nil {
			panic(err)
		}

		// Extract ledger key from the storage key
		ledgerKeyStr := key.K1()

		// Add entry to the appropriate LedgerToEntries
		if ledgerToEntries, exists := ledgerToEntriesMap[ledgerKeyStr]; exists {
			ledgerToEntries.Entries = append(ledgerToEntries.Entries, &entry)
		}
	}

	// Convert map to slice
	for _, ledgerToEntries := range ledgerToEntriesMap {
		state.LedgerToEntries = append(state.LedgerToEntries, *ledgerToEntries)
	}

	return state
}
