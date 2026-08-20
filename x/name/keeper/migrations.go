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
	logger := m.keeper.Logger(ctx)
	logger.Info("Migrating the name module from raw KV to collections (v2 to v3)...")

	if err := m.migrateV2Params(ctx); err != nil {
		return err
	}

	records, err := m.readV2NameRecords(ctx)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Found %d name record(s) to migrate.", len(records)))

	for _, record := range records {
		if err = m.keeper.nameRecords.Set(ctx, record.Name, record); err != nil {
			return fmt.Errorf("could not migrate the name record %q: %w", record.Name, err)
		}
	}

	logger.Info(fmt.Sprintf("Done migrating %d name record(s) to collections (v2 to v3).", len(records)))
	return nil
}

// readV2NameRecords reads every name record written under the pre-collections layout.
//
// It lives in its own function so the deferred Close runs before the caller writes anything,
// on every path out including the error ones.
func (m Migrator) readV2NameRecords(ctx sdk.Context) ([]types.NameRecord, error) {
	store := m.keeper.storeService.OpenKVStore(ctx)

	iter, err := store.Iterator(LegacyNameKeyPrefix, storetypes.PrefixEndBytes(LegacyNameKeyPrefix))
	if err != nil {
		return nil, fmt.Errorf("could not iterate the legacy name records: %w", err)
	}
	defer iter.Close() //nolint:errcheck // close error safe to ignore in this context.

	var records []types.NameRecord
	for ; iter.Valid(); iter.Next() {
		record := types.NameRecord{}
		if err = m.keeper.cdc.Unmarshal(iter.Value(), &record); err != nil {
			return nil, fmt.Errorf("could not unmarshal the legacy name record at %X: %w", iter.Key(), err)
		}
		records = append(records, record)
	}
	return records, nil
}

// migrateV2Params copies the legacy params entry into the params Item.
func (m Migrator) migrateV2Params(ctx sdk.Context) error {
	store := m.keeper.storeService.OpenKVStore(ctx)

	bz, err := store.Get(LegacyNameParamStoreKey)
	if err != nil {
		return fmt.Errorf("could not read the legacy params: %w", err)
	}
	if bz == nil {
		return nil
	}

	// Start with the default params, not a zero-value Params. Proto3 omits false and 0
	// values, so unmarshalling into a zero-value Params could turn a stored
	// AllowUnrestrictedNames=false into the default true and change the chain's params
	// during the upgrade.
	params := types.DefaultParams()
	if err = m.keeper.cdc.Unmarshal(bz, &params); err != nil {
		return fmt.Errorf("could not unmarshal the legacy params: %w", err)
	}
	if err = m.keeper.paramsStore.Set(ctx, params); err != nil {
		return fmt.Errorf("could not write the params: %w", err)
	}
	return nil
}
