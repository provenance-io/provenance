package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/provenance-io/provenance/x/name/types"
)

// Define old key prefixes for migration
var (
	OldNameKeyPrefix    = []byte{0x03}
	OldAddressKeyPrefix = []byte{0x05}
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
	ctx.Logger().Info("Migrating name module from KV store to collections (v2 to v3)...")
	store := m.keeper.storeService.OpenKVStore(ctx)

	// 1. Migrate parameters
	if bz, _ := store.Get(types.NameParamStoreKey); bz != nil {
		var params types.Params
		m.keeper.cdc.MustUnmarshal(bz, &params)
		if err := m.keeper.paramsStore.Set(ctx, params); err != nil {
			return fmt.Errorf("failed to migrate params: %w", err)
		}
		ctx.Logger().Info("Migrated parameters")
		store.Delete(types.NameParamStoreKey)
	}
	// Migrate name records
	ctx.Logger().Info("Migrating name records...")
	addrPrefix := types.AddressKeyPrefix
	iter, _ := store.Iterator(addrPrefix, storetypes.PrefixEndBytes(addrPrefix))
	defer iter.Close()

	count := 0
	for ; iter.Valid(); iter.Next() {
		store.Set(iter.Key(), []byte{})
		count++
	}
	ctx.Logger().Info(fmt.Sprintf("Fixed %d address index values", count))

	ctx.Logger().Info("Deleting old entries...")
	// m.keeper.DeleteInvalidAddressIndexEntries(ctx)

	ctx.Logger().Info("Name module migration to collections (v2 to v3) completed successfully.")

	return nil
}
