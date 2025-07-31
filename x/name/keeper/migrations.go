package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/provenance-io/provenance/x/name/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// MigrateKVToCollections2to3 migrates the name module data from the legacy KV store layout
// to the new collections-based layout (version 2 to version 3).
func (m Migrator) MigrateKVToCollections2to3(ctx sdk.Context) error {
	fmt.Println("RUNNING NAME MODULE MIGRATION")
	ctx.Logger().Info("Migrating name module from KV store to collections (v2 to v3)...")

	store := ctx.KVStore(m.keeper.storeKey)

	// Migrate parameters
	if bz := store.Get(types.NameParamStoreKey); bz != nil {
		var params types.Params
		m.keeper.cdc.MustUnmarshal(bz, &params)
		if err := m.keeper.paramsStore.Set(ctx, params); err != nil {
			return err
		}
		ctx.Logger().Info("Migrated name module parameters to collections store.")
	}

	// Migrate name records
	ctx.Logger().Info("Migrating name records...")
	namePrefix := types.NameKeyPrefix
	nameIter := store.Iterator(namePrefix, storetypes.PrefixEndBytes(namePrefix))
	defer nameIter.Close()
	count := 0
	for ; nameIter.Valid(); nameIter.Next() {
		var record types.NameRecord
		m.keeper.cdc.MustUnmarshal(nameIter.Value(), &record)

		// Use the exact key from the store
		key := nameIter.Key()

		// Set in collections
		if err := m.keeper.nameRecords.Set(ctx, key, record); err != nil {
			return err
		}
		count++
	}
	ctx.Logger().Info(fmt.Sprintf("Migrated %d name records.", count))
	// Migrate address index
	addrPrefix := types.AddressKeyPrefix
	addrIter := store.Iterator(addrPrefix, storetypes.PrefixEndBytes(addrPrefix))
	defer addrIter.Close()
	addrCount := 0
	for ; addrIter.Valid(); addrIter.Next() {
		var record types.NameRecord
		m.keeper.cdc.MustUnmarshal(addrIter.Value(), &record)

		// Use the exact key from the store
		key := addrIter.Key()

		// Set in collections address index
		if err := m.keeper.addrIndex.Set(ctx, key, record); err != nil {
			return err
		}
		addrCount++
	}
	ctx.Logger().Info(fmt.Sprintf("Migrated %d address index records.", addrCount))
	ctx.Logger().Info("Name module migration to collections (v2 to v3) completed successfully.")
	return nil
}
