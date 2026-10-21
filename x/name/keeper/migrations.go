package keeper

import (
	"crypto/sha256"
	"fmt"
	"strings"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/provenance-io/provenance/x/name/types"
)

// The pre-collections (v2) store layout.
var (
	legacyNameKeyPrefix     = []byte{0x03}
	legacyAddressKeyPrefix  = []byte{0x05}
	legacyNameParamStoreKey = []byte{0x06}
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

	if err := m.deleteLegacyData(ctx); err != nil {
		return err
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

	iter, err := store.Iterator(legacyNameKeyPrefix, storetypes.PrefixEndBytes(legacyNameKeyPrefix))
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

	bz, err := store.Get(legacyNameParamStoreKey)
	if err != nil {
		return fmt.Errorf("could not read the legacy params: %w", err)
	}
	if bz == nil {
		return nil
	}

	// The codec resets the message before unmarshalling, so the stored values are exactly what we get here.
	var params types.Params
	if err = m.keeper.cdc.Unmarshal(bz, &params); err != nil {
		return fmt.Errorf("could not unmarshal the legacy params: %w", err)
	}
	if err = m.keeper.paramsStore.Set(ctx, params); err != nil {
		return fmt.Errorf("could not write the params: %w", err)
	}
	return nil
}

// deleteLegacyData removes all entries from the pre-collections layout:
// name records, the address index, and the legacy params entry copied by migrateV2Params.
func (m Migrator) deleteLegacyData(ctx sdk.Context) error {
	kvStore := m.keeper.storeService.OpenKVStore(ctx)

	for _, prefix := range [][]byte{legacyNameKeyPrefix, legacyAddressKeyPrefix} {
		iter, err := kvStore.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return fmt.Errorf("could not iterate legacy prefix %X: %w", prefix, err)
		}
		var keys [][]byte
		for ; iter.Valid(); iter.Next() {
			keys = append(keys, append([]byte{}, iter.Key()...))
		}
		// Close right away (not deferred) so the iterator isn't open while deleting.
		if err = iter.Close(); err != nil {
			return fmt.Errorf("could not close the iterator for legacy prefix %X: %w", prefix, err)
		}

		for _, key := range keys {
			if err = kvStore.Delete(key); err != nil {
				return fmt.Errorf("could not delete legacy key %X: %w", key, err)
			}
		}
	}

	if err := kvStore.Delete(legacyNameParamStoreKey); err != nil {
		return fmt.Errorf("could not delete the legacy params: %w", err)
	}
	return nil
}

// LegacyComputeNameHash reproduces the old name-key hash by hashing segments in reverse order.
//
// Deprecated: used only by the 3->4 migration and its tests.
func LegacyComputeNameHash(name string) ([]byte, error) {
	comps := strings.Split(name, ".")
	hsh := sha256.New()
	for i := len(comps) - 1; i >= 0; i-- {
		comp := strings.TrimSpace(comps[i])
		if len(comp) == 0 {
			return nil, fmt.Errorf("name segment cannot be empty: %w", types.ErrNameInvalid)
		}
		if _, err := hsh.Write([]byte(comp)); err != nil {
			return nil, err
		}
	}
	return hsh.Sum(nil), nil
}

// LegacyGetNameKeyBytes returns the full store key a name occupied before the 2->3 migration.
func LegacyGetNameKeyBytes(name string) ([]byte, error) {
	hash, err := LegacyComputeNameHash(name)
	if err != nil {
		return nil, err
	}
	return append(append([]byte{}, legacyNameKeyPrefix...), hash...), nil
}
