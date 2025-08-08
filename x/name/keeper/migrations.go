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
	ctx.Logger().Info("Migrating name module from KV store to collections (v2 to v3)...")

	store := m.keeper.storeService.OpenKVStore(ctx)

	ctx.Logger().Info("Phase 1: Collecting existing name records...")
	var recordsToMigrate []struct {
		name   string
		record types.NameRecord
	}

	// Iterate over name records.
	namePrefix := types.NameKeyPrefix
	iter, err := store.Iterator(namePrefix, storetypes.PrefixEndBytes(namePrefix))
	if err != nil {
		return fmt.Errorf("failed to create name records iterator: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		if len(key) <= len(namePrefix) {
			continue
		}
		var record types.NameRecord
		if err := m.keeper.cdc.Unmarshal(value, &record); err != nil {
			ctx.Logger().Error("failed to unmarshal name record during migration",
				"key", fmt.Sprintf("%x", key), "error", err)
			continue
		}

		if record.Name == "" {
			ctx.Logger().Error("name record has empty name field",
				"key", fmt.Sprintf("%x", key))
			continue
		}
		recordsToMigrate = append(recordsToMigrate, struct {
			name   string
			record types.NameRecord
		}{
			name:   record.Name,
			record: record,
		})
	}

	ctx.Logger().Info(fmt.Sprintf("Found %d name records to migrate", len(recordsToMigrate)))

	ctx.Logger().Info("Phase 2: Migrating records to collections...")
	for _, item := range recordsToMigrate {
		if err := m.keeper.nameRecords.Set(ctx, item.name, item.record); err != nil {
			ctx.Logger().Error("failed to set record in nameRecords", "name", item.name, "address", item.record.Address, "error", err)
			return fmt.Errorf("failed to migrate name record %s: %w", item.name, err)
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Successfully migrated %d name records to collections", len(recordsToMigrate)))

	ctx.Logger().Info("Name module migration to collections (v2 to v3) completed successfully.")

	return nil
}
