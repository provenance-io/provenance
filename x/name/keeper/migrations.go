package keeper

import (
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

func (m Migrator) MigrateKVToCollections2to3(ctx sdk.Context) error {

	store := ctx.KVStore(m.keeper.storeKey)

	// 1. Migrate parameters
	if bz := store.Get(types.NameParamStoreKey); bz != nil {
		var params types.Params
		m.keeper.cdc.MustUnmarshal(bz, &params)
		if err := m.keeper.ParamsStore.Set(ctx, params); err != nil {
			return err
		}
	}

	// 2. Migrate name records
	namePrefix := types.NameKeyPrefix
	nameIter := store.Iterator(namePrefix, storetypes.PrefixEndBytes(namePrefix))
	defer nameIter.Close()

	for ; nameIter.Valid(); nameIter.Next() {
		var record types.NameRecord
		m.keeper.cdc.MustUnmarshal(nameIter.Value(), &record)

		// Use the exact key from the store
		key := nameIter.Key()

		// Set in collections
		if err := m.keeper.NameRecords.Set(ctx, key, record); err != nil {
			return err
		}
	}

	// 3. Migrate address index
	addrPrefix := types.AddressKeyPrefix
	addrIter := store.Iterator(addrPrefix, storetypes.PrefixEndBytes(addrPrefix))
	defer addrIter.Close()

	for ; addrIter.Valid(); addrIter.Next() {
		var record types.NameRecord
		m.keeper.cdc.MustUnmarshal(addrIter.Value(), &record)

		// Use the exact key from the store
		key := addrIter.Key()

		// Set in collections address index
		if err := m.keeper.AddrIndex.Set(ctx, key, record); err != nil {
			return err
		}
	}

	return nil

}
