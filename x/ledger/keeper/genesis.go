package keeper

import (
	"strconv"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// ExportGenesis exports the current keeper state of the ledger module.
// This exports data in the format that matches test.json for bulk import.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Generate the initial genesis state.
	state := &types.GenesisState{}

	// Ledger Classes
	k.ExportLedgerClasses(ctx, state)
	k.ExportLedgerClassEntryTypes(ctx, state)
	k.ExportLedgerClassStatusTypes(ctx, state)
	k.ExportLedgerClassBucketTypes(ctx, state)

	// Ledgers
	k.ExportLedgers(ctx, state)
	k.ExportLedgerEntries(ctx, state)
	k.ExportStoredSettlementInstructions(ctx, state)

	return state
}

// Mutates the GenesisState value for LedgerClasses to match the exported data from the keeper's LedgerClasses collection.
func (k Keeper) ExportLedgerClasses(ctx sdk.Context, genesis *types.GenesisState) {
	// Get all ledger classes.
	ledgerClassIter, err := k.LedgerClasses.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer ledgerClassIter.Close()

	ledgerClasses := make([]types.LedgerClass, 0)
	for ; ledgerClassIter.Valid(); ledgerClassIter.Next() {
		_, err := ledgerClassIter.Key()
		if err != nil {
			panic(err)
		}
		ledgerClass, err := ledgerClassIter.Value()
		if err != nil {
			panic(err)
		}

		ledgerClasses = append(ledgerClasses, ledgerClass)
	}

	genesis.LedgerClasses = ledgerClasses
}

// Mutates the GenesisState value for LedgerClassEntryTypes to match the exported data from the keeper's LedgerClassEntryTypes collection.
func (k Keeper) ExportLedgerClassEntryTypes(ctx sdk.Context, genesis *types.GenesisState) {
	// Get all entry types.
	entryTypeIter, err := k.LedgerClassEntryTypes.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer entryTypeIter.Close()

	entryTypes := make([]types.GenesisLedgerClassEntryType, 0)
	for ; entryTypeIter.Valid(); entryTypeIter.Next() {
		key, err := entryTypeIter.Key()
		if err != nil {
			panic(err)
		}
		entryType, err := entryTypeIter.Value()
		if err != nil {
			panic(err)
		}

		entryTypes = append(entryTypes, types.GenesisLedgerClassEntryType{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: strconv.Itoa(int(key.K2())),
			},
			EntryType: entryType,
		})
	}

	genesis.LedgerClassEntryTypes = entryTypes
}

// Mutates the GenesisState value for LedgerClassStatusTypes to match the exported data from the keeper's LedgerClassStatusTypes collection.
func (k Keeper) ExportLedgerClassStatusTypes(ctx sdk.Context, genesis *types.GenesisState) {
	// Get all status types.
	statusTypeIter, err := k.LedgerClassStatusTypes.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer statusTypeIter.Close()

	statusTypes := make([]types.GenesisLedgerClassStatusType, 0)
	for ; statusTypeIter.Valid(); statusTypeIter.Next() {
		key, err := statusTypeIter.Key()
		if err != nil {
			panic(err)
		}
		statusType, err := statusTypeIter.Value()
		if err != nil {
			panic(err)
		}

		statusTypes = append(statusTypes, types.GenesisLedgerClassStatusType{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: strconv.Itoa(int(key.K2())),
			},
			StatusType: statusType,
		})
	}

	genesis.LedgerClassStatusTypes = statusTypes
}

// Mutates the GenesisState value for LedgerClassBucketTypes to match the exported data from the keeper's LedgerClassBucketTypes collection.
func (k Keeper) ExportLedgerClassBucketTypes(ctx sdk.Context, genesis *types.GenesisState) {
	// Get all bucket types.
	bucketTypeIter, err := k.LedgerClassBucketTypes.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer bucketTypeIter.Close()

	bucketTypes := make([]types.GenesisLedgerClassBucketType, 0)
	for ; bucketTypeIter.Valid(); bucketTypeIter.Next() {
		key, err := bucketTypeIter.Key()
		if err != nil {
			panic(err)
		}
		bucketType, err := bucketTypeIter.Value()
		if err != nil {
			panic(err)
		}

		bucketTypes = append(bucketTypes, types.GenesisLedgerClassBucketType{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: strconv.Itoa(int(key.K2())),
			},
			BucketType: bucketType,
		})
	}

	genesis.LedgerClassBucketTypes = bucketTypes
}

// Mutates the GenesisState value for Ledgers to match the exported data from the keeper's Ledgers collection.
func (k Keeper) ExportLedgers(ctx sdk.Context, genesis *types.GenesisState) {
	// Get all ledgers.
	ledgerIter, err := k.Ledgers.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer ledgerIter.Close()

	ledgers := make([]types.GenesisLedger, 0)
	for ; ledgerIter.Valid(); ledgerIter.Next() {
		key, err := ledgerIter.Key()
		if err != nil {
			panic(err)
		}
		ledger, err := ledgerIter.Value()
		if err != nil {
			panic(err)
		}

		// Because the ledger key was removed for storage efficiency reasons, we need to reconstruct it from the key.
		ledger.Key, err = types.StringToLedgerKey(key)
		if err != nil {
			panic(err)
		}

		ledgers = append(ledgers, types.GenesisLedger{
			Ledger: ledger,
		})
	}

	genesis.Ledgers = ledgers
}

// Mutates the GenesisState value for LedgerEntries to match the exported data from the keeper's LedgerEntries collection.
func (k Keeper) ExportLedgerEntries(ctx sdk.Context, genesis *types.GenesisState) {
	// Get all entries.
	entryIter, err := k.LedgerEntries.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer entryIter.Close()

	entries := make([]types.GenesisLedgerEntry, 0)
	for ; entryIter.Valid(); entryIter.Next() {
		key, err := entryIter.Key()
		if err != nil {
			panic(err)
		}
		entry, err := entryIter.Value()
		if err != nil {
			panic(err)
		}

		entries = append(entries, types.GenesisLedgerEntry{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: key.K2(),
			},
			Entry: entry,
		})
	}

	genesis.LedgerEntries = entries
}

// Mutates the GenesisState value for SettlementInstructions to match the exported data from the keeper's FundTransfersWithSettlement collection.
func (k Keeper) ExportStoredSettlementInstructions(ctx sdk.Context, genesis *types.GenesisState) {
	// Get all settlement instructions.
	settlementIter, err := k.FundTransfersWithSettlement.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer settlementIter.Close()

	settlements := make([]types.GenesisStoredSettlementInstructions, 0)

	for ; settlementIter.Valid(); settlementIter.Next() {
		key, err := settlementIter.Key()
		if err != nil {
			panic(err)
		}
		settlement, err := settlementIter.Value()
		if err != nil {
			panic(err)
		}

		settlements = append(settlements, types.GenesisStoredSettlementInstructions{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: key.K2(),
			},
			SettlementInstructions: settlement,
		})
	}

	genesis.SettlementInstructions = settlements
}

func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	if state == nil {
		return
	}

	k.ImportLedgerClasses(ctx, state)
	k.ImportLedgerClassEntryTypes(ctx, state)
	k.ImportLedgerClassStatusTypes(ctx, state)
	k.ImportLedgerClassBucketTypes(ctx, state)
	k.ImportLedgers(ctx, state)
	k.ImportLedgerEntries(ctx, state)
	k.ImportStoredSettlementInstructions(ctx, state)
}

func (k Keeper) ImportLedgerClasses(ctx sdk.Context, state *types.GenesisState) {
	for _, ledgerClass := range state.LedgerClasses {
		if err := k.LedgerClasses.Set(ctx, ledgerClass.LedgerClassId, ledgerClass); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ImportLedgerClassEntryTypes(ctx sdk.Context, state *types.GenesisState) {
	for _, l := range state.LedgerClassEntryTypes {
		// Parse the second key as an integer.
		id, err := strconv.Atoi(l.Key.P2)
		if err != nil {
			panic(err)
		}

		key := collections.Join(l.Key.P1, int32(id)) //nolint:gosec // Controlled conversion
		if err := k.LedgerClassEntryTypes.Set(ctx, key, l.EntryType); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ImportLedgerClassStatusTypes(ctx sdk.Context, state *types.GenesisState) {
	for _, l := range state.LedgerClassStatusTypes {
		// Parse the second key as an integer.
		id, err := strconv.Atoi(l.Key.P2)
		if err != nil {
			panic(err)
		}

		key := collections.Join(l.Key.P1, int32(id)) //nolint:gosec // Controlled conversion
		if err := k.LedgerClassStatusTypes.Set(ctx, key, l.StatusType); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ImportLedgerClassBucketTypes(ctx sdk.Context, state *types.GenesisState) {
	for _, l := range state.LedgerClassBucketTypes {
		// Parse the second key as an integer.
		id, err := strconv.Atoi(l.Key.P2)
		if err != nil {
			panic(err)
		}

		key := collections.Join(l.Key.P1, int32(id)) //nolint:gosec // Controlled conversion
		if err := k.LedgerClassBucketTypes.Set(ctx, key, l.BucketType); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ImportLedgers(ctx sdk.Context, state *types.GenesisState) {
	for _, l := range state.Ledgers {
		// Remove the key from the ledger to avoid storing it twice. Easy optimization.
		key := l.Ledger.Key.String()
		l.Ledger.Key = nil

		if err := k.Ledgers.Set(ctx, key, l.Ledger); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ImportLedgerEntries(ctx sdk.Context, state *types.GenesisState) {
	for _, le := range state.LedgerEntries {
		key := collections.Join(le.Key.P1, le.Key.P2)

		if err := k.LedgerEntries.Set(ctx, key, le.Entry); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ImportStoredSettlementInstructions(ctx sdk.Context, state *types.GenesisState) {
	for _, si := range state.SettlementInstructions {
		key := collections.Join(si.Key.P1, si.Key.P2)
		if err := k.FundTransfersWithSettlement.Set(ctx, key, si.SettlementInstructions); err != nil {
			panic(err)
		}
	}
}
