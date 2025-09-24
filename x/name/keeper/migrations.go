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

// Legacy prefixes - OLD values moved here for migration
var (
	LegacyNameKeyPrefix     = []byte{0x03}
	LegacyAddressKeyPrefix  = []byte{0x05}
	LegacyNameParamStoreKey = []byte{0x06}
)

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// MigrateKVToCollections2to3 migrates the name module data from the legacy KV store layout
// to the new collections-based layout (version 2 to version 3).
func (m Migrator) MigrateKVToCollections2to3(ctx sdk.Context) error {
	ctx.Logger().Info("Migrating name module from KV store to collections (v2 to v3)...")

	store := m.keeper.storeService.OpenKVStore(ctx)

	// Step 1: Load legacy records
	legacyIter, err := store.Iterator(LegacyNameKeyPrefix, storetypes.PrefixEndBytes(LegacyNameKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create legacy iterator: %w", err)
	}
	defer legacyIter.Close()

	var recordsToMigrate []types.NameRecord
	for ; legacyIter.Valid(); legacyIter.Next() {
		var record types.NameRecord
		if err := m.keeper.cdc.Unmarshal(legacyIter.Value(), &record); err != nil {
			continue // skip bad records silently
		}
		recordsToMigrate = append(recordsToMigrate, record)
	}

	// Step 2: Migrate records to collections
	for _, record := range recordsToMigrate {
		if err := m.keeper.nameRecords.Set(ctx, record.Name, record); err != nil {
			return fmt.Errorf("failed to migrate record %s: %w", record.Name, err)
		}
	}

	// Step 3: Migrate params if they exist
	if exists, _ := store.Has(LegacyNameParamStoreKey); exists {
		bz, err := store.Get(LegacyNameParamStoreKey)
		if err == nil {
			var params types.Params
			if err := m.keeper.cdc.Unmarshal(bz, &params); err == nil {
				if err := m.keeper.paramsStore.Set(ctx, params); err != nil {
					return fmt.Errorf("failed to migrate params: %w", err)
				}
			}
		}
	}

	ctx.Logger().Info(fmt.Sprintf("Successfully migrated %d name records to collections", len(recordsToMigrate)))
	ctx.Logger().Info("Name module migration to collections (v2 to v3) completed successfully.")
	return nil
}
