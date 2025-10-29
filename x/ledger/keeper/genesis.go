package keeper

import (
	"context"
	"fmt"
	"strconv"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// ExportGenesis exports the current keeper state of the ledger module.
func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
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

// ExportLedgerClasses mutates the GenesisState value for LedgerClasses to match the exported data from the keeper's LedgerClasses collection.
func (k Keeper) ExportLedgerClasses(ctx context.Context, genesis *types.GenesisState) {
	genesis.LedgerClasses = nil // Make sure we're starting fresh.
	err := k.LedgerClasses.Walk(ctx, nil, func(_ string, ledgerClass types.LedgerClass) (stop bool, err error) {
		genesis.LedgerClasses = append(genesis.LedgerClasses, ledgerClass)
		return false, nil
	})
	if err != nil {
		panic(fmt.Errorf("error walking ledger classes: %w", err))
	}
}

// ExportLedgerClassEntryTypes mutates the GenesisState value for LedgerClassEntryTypes to match the exported data from the keeper's LedgerClassEntryTypes collection.
func (k Keeper) ExportLedgerClassEntryTypes(ctx context.Context, genesis *types.GenesisState) {
	genesis.LedgerClassEntryTypes = nil // Make sure we're starting fresh.
	err := k.LedgerClassEntryTypes.Walk(ctx, nil, func(key collections.Pair[string, int32], entryType types.LedgerClassEntryType) (stop bool, err error) {
		genesis.LedgerClassEntryTypes = append(genesis.LedgerClassEntryTypes, types.GenesisLedgerClassEntryType{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: strconv.Itoa(int(key.K2())),
			},
			EntryType: entryType,
		})
		return false, nil
	})
	if err != nil {
		panic(fmt.Errorf("error walking ledger class entry types: %w", err))
	}
}

// ExportLedgerClassStatusTypes mutates the GenesisState value for LedgerClassStatusTypes to match the exported data from the keeper's LedgerClassStatusTypes collection.
func (k Keeper) ExportLedgerClassStatusTypes(ctx context.Context, genesis *types.GenesisState) {
	genesis.LedgerClassStatusTypes = nil // Make sure we're starting fresh.
	err := k.LedgerClassStatusTypes.Walk(ctx, nil, func(key collections.Pair[string, int32], statusType types.LedgerClassStatusType) (stop bool, err error) {
		genesis.LedgerClassStatusTypes = append(genesis.LedgerClassStatusTypes, types.GenesisLedgerClassStatusType{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: strconv.Itoa(int(key.K2())),
			},
			StatusType: statusType,
		})
		return false, nil
	})
	if err != nil {
		panic(fmt.Errorf("error walking ledger class status types: %w", err))
	}
}

// ExportLedgerClassBucketTypes mutates the GenesisState value for LedgerClassBucketTypes to match the exported data from the keeper's LedgerClassBucketTypes collection.
func (k Keeper) ExportLedgerClassBucketTypes(ctx context.Context, genesis *types.GenesisState) {
	genesis.LedgerClassBucketTypes = nil // Make sure we're starting fresh.
	err := k.LedgerClassBucketTypes.Walk(ctx, nil, func(key collections.Pair[string, int32], bucketType types.LedgerClassBucketType) (stop bool, err error) {
		genesis.LedgerClassBucketTypes = append(genesis.LedgerClassBucketTypes, types.GenesisLedgerClassBucketType{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: strconv.Itoa(int(key.K2())),
			},
			BucketType: bucketType,
		})
		return false, nil
	})
	if err != nil {
		panic(fmt.Errorf("error walking ledger class bucket types: %w", err))
	}
}

// ExportLedgers mutates the GenesisState value for Ledgers to match the exported data from the keeper's Ledgers collection.
func (k Keeper) ExportLedgers(ctx context.Context, genesis *types.GenesisState) {
	genesis.Ledgers = nil // Make sure we're starting fresh.
	err := k.Ledgers.Walk(ctx, nil, func(key string, ledger types.Ledger) (stop bool, err error) {
		// Because the ledger key was removed for storage efficiency reasons, we need to reconstruct it from the key.
		ledger.Key, err = types.StringToLedgerKey(key)
		if err != nil {
			return true, fmt.Errorf("invalid ledger entry key %q: %w", key, err)
		}
		genesis.Ledgers = append(genesis.Ledgers, types.GenesisLedger{Ledger: ledger})
		return false, nil
	})
	if err != nil {
		panic(fmt.Errorf("error walking ledgers: %w", err))
	}
}

// ExportLedgerEntries mutates the GenesisState value for LedgerEntries to match the exported data from the keeper's LedgerEntries collection.
func (k Keeper) ExportLedgerEntries(ctx context.Context, genesis *types.GenesisState) {
	genesis.LedgerEntries = nil // Make sure we're starting fresh.
	err := k.LedgerEntries.Walk(ctx, nil, func(key collections.Pair[string, string], entry types.LedgerEntry) (stop bool, err error) {
		ledgerKey, err := types.StringToLedgerKey(key.K1())
		if err != nil {
			return true, fmt.Errorf("invalid ledger entry key %q: %w", key.K1(), err)
		}
		genesis.LedgerEntries = append(genesis.LedgerEntries, types.GenesisLedgerEntry{
			Key:   ledgerKey,
			Entry: entry,
		})
		return false, nil
	})
	if err != nil {
		panic(fmt.Errorf("error walking ledger entries: %w", err))
	}
}

// ExportStoredSettlementInstructions mutates the GenesisState value for SettlementInstructions to match the exported data from the keeper's FundTransfersWithSettlement collection.
func (k Keeper) ExportStoredSettlementInstructions(ctx context.Context, genesis *types.GenesisState) {
	genesis.SettlementInstructions = nil // Make sure we're starting fresh.
	err := k.FundTransfersWithSettlement.Walk(ctx, nil, func(key collections.Pair[string, string], settlement types.StoredSettlementInstructions) (stop bool, err error) {
		genesis.SettlementInstructions = append(genesis.SettlementInstructions, types.GenesisStoredSettlementInstructions{
			Key: &types.GenesisPair{
				P1: key.K1(),
				P2: key.K2(),
			},
			SettlementInstructions: settlement,
		})
		return false, nil
	})
	if err != nil {
		panic(fmt.Errorf("error walking stored settlement instructions: %w", err))
	}
}

// InitGenesis writes the provided GenesisState to the ledger module collections/state.
func (k Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) {
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

// ImportLedgerClasses writes all of the LedgerClasses to the LedgerClasses state collection.
func (k Keeper) ImportLedgerClasses(ctx context.Context, state *types.GenesisState) {
	for i, ledgerClass := range state.LedgerClasses {
		if err := k.LedgerClasses.Set(ctx, ledgerClass.LedgerClassId, ledgerClass); err != nil {
			panic(fmt.Errorf("error storing LedgerClasses[%d]: %w", i, err))
		}
	}
}

// ImportLedgerClassEntryTypes writes all of the LedgerClassEntryTypes to the LedgerClassEntryTypes state collection.
func (k Keeper) ImportLedgerClassEntryTypes(ctx context.Context, state *types.GenesisState) {
	for i, l := range state.LedgerClassEntryTypes {
		// Parse the second key as an integer.
		id, err := strconv.ParseInt(l.Key.P2, 10, 32)
		if err != nil {
			panic(fmt.Errorf("invalid LedgerClassEntryTypes[%d].Key.P2: %w", i, err))
		}

		key := collections.Join(l.Key.P1, int32(id)) //nolint:gosec // Parsed above as 32-bits, so we know it fits.
		if err = k.LedgerClassEntryTypes.Set(ctx, key, l.EntryType); err != nil {
			panic(fmt.Errorf("error storing LedgerClassEntryTypes[%d]: %w", i, err))
		}
	}
}

// ImportLedgerClassStatusTypes writes all of the LedgerClassStatusTypes to the LedgerClassStatusTypes state collection.
func (k Keeper) ImportLedgerClassStatusTypes(ctx context.Context, state *types.GenesisState) {
	for i, l := range state.LedgerClassStatusTypes {
		// Parse the second key as an integer.
		id, err := strconv.ParseInt(l.Key.P2, 10, 32)
		if err != nil {
			panic(fmt.Errorf("invalid LedgerClassStatusTypes[%d].Key.P2: %w", i, err))
		}

		key := collections.Join(l.Key.P1, int32(id)) //nolint:gosec // Parsed above as 32-bits, so we know it fits.
		if err := k.LedgerClassStatusTypes.Set(ctx, key, l.StatusType); err != nil {
			panic(fmt.Errorf("error storing LedgerClassStatusTypes[%d]: %w", i, err))
		}
	}
}

// ImportLedgerClassBucketTypes writes all of the LedgerClassBucketTypes to the LedgerClassBucketTypes state collection.
func (k Keeper) ImportLedgerClassBucketTypes(ctx context.Context, state *types.GenesisState) {
	for i, l := range state.LedgerClassBucketTypes {
		// Parse the second key as an integer.
		id, err := strconv.ParseInt(l.Key.P2, 10, 32)
		if err != nil {
			panic(fmt.Errorf("invalid LedgerClassBucketTypes[%d].Key.P2: %w", i, err))
		}

		key := collections.Join(l.Key.P1, int32(id)) //nolint:gosec // Parsed above as 32-bits, so we know it fits.
		if err = k.LedgerClassBucketTypes.Set(ctx, key, l.BucketType); err != nil {
			panic(fmt.Errorf("error storing LedgerClassBucketTypes[%d]: %w", i, err))
		}
	}
}

// ImportLedgers writes all of the Ledgers to the Ledgers state collection.
func (k Keeper) ImportLedgers(ctx context.Context, state *types.GenesisState) {
	for i, l := range state.Ledgers {
		// Set the NextPmtAmt to zero if it's nil.
		if l.Ledger.NextPmtAmt.IsNil() {
			l.Ledger.NextPmtAmt = sdkmath.ZeroInt()
		}

		// Remove the key from the ledger to avoid storing it twice. Easy optimization.
		key := l.Ledger.Key.String()
		l.Ledger.Key = nil

		if err := k.Ledgers.Set(ctx, key, l.Ledger); err != nil {
			panic(fmt.Errorf("error storing Ledgers[%d]: %w", i, err))
		}
	}
}

// ImportLedgerEntries writes all of the LedgerEntries to the LedgerEntries state collection.
func (k Keeper) ImportLedgerEntries(ctx context.Context, state *types.GenesisState) {
	for i, le := range state.LedgerEntries {
		key := collections.Join(le.Key.String(), le.Entry.CorrelationId)

		if err := k.LedgerEntries.Set(ctx, key, le.Entry); err != nil {
			panic(fmt.Errorf("error storing LedgerEntries[%d]: %w", i, err))
		}
	}
}

// ImportStoredSettlementInstructions writes all of the SettlementInstructions to the FundTransfersWithSettlement state collection.
func (k Keeper) ImportStoredSettlementInstructions(ctx context.Context, state *types.GenesisState) {
	for i, si := range state.SettlementInstructions {
		key := collections.Join(si.Key.P1, si.Key.P2)
		if err := k.FundTransfersWithSettlement.Set(ctx, key, si.SettlementInstructions); err != nil {
			panic(fmt.Errorf("error storing SettlementInstructions[%d]: %w", i, err))
		}
	}
}
