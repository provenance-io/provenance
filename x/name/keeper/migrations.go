package keeper

import (
	"bytes"
	"fmt"
	"slices"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// rekeyedEntry is a store entry that needs to move from oldKey to newKey.
type rekeyedEntry struct {
	oldKey []byte
	newKey []byte
	value  []byte
}

// Migrate2to3 rewrites the name records and the address -> name index entries using keys that
// are a hash of the full name instead of a hash of just the name segments.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	logger := m.keeper.Logger(ctx)
	logger.Info("Migrating name records to keys based on the full name.")

	store := ctx.KVStore(m.keeper.storeKey)

	entries, err := m.rekeyedNameEntries(store)
	if err != nil {
		return err
	}
	indexEntries, err := m.rekeyedAddressIndexEntries(store)
	if err != nil {
		return err
	}
	entries = append(entries, indexEntries...)

	// All of the deletes are done before any of the writes so that an entry keeping its
	// key (any single-segment name) can't be deleted after it has been written again.
	for _, entry := range entries {
		store.Delete(entry.oldKey)
	}
	for _, entry := range entries {
		store.Set(entry.newKey, entry.value)
	}

	logger.Info(fmt.Sprintf("Done migrating name records. Rewrote %d entries.", len(entries)))
	return nil
}

// rekeyedNameEntries returns the name records, [NameKeyPrefix] :: [name key], that need a new key.
func (m Migrator) rekeyedNameEntries(store storetypes.KVStore) ([]rekeyedEntry, error) {
	return m.rekeyedEntries(store, types.NameKeyPrefix, func(_ []byte, nameKey []byte) []byte {
		return nameKey
	})
}

// rekeyedAddressIndexEntries returns the address index entries,
// [AddressKeyPrefix] :: [address length] :: [address] :: [name key], that need a new key.
func (m Migrator) rekeyedAddressIndexEntries(store storetypes.KVStore) ([]rekeyedEntry, error) {
	return m.rekeyedEntries(store, types.AddressKeyPrefix, func(oldKey []byte, nameKey []byte) []byte {
		// The 2nd byte is the length of the address that follows it, and the name key is the rest.
		addrEnd := int(oldKey[1]) + 2
		return append(slices.Clone(oldKey[:addrEnd]), nameKey...)
	})
}

// rekeyedEntries returns each entry under the given prefix whose key changes, using newKey to
// build the new key from the entry's existing key and the new key of the name it holds.
func (m Migrator) rekeyedEntries(store storetypes.KVStore, prefix []byte, newKey func(oldKey []byte, nameKey []byte) []byte) ([]rekeyedEntry, error) {
	var entries []rekeyedEntry

	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close() //nolint:errcheck // close error safe to ignore in this context.

	for ; iter.Valid(); iter.Next() {
		record := types.NameRecord{}
		if err := m.keeper.cdc.Unmarshal(iter.Value(), &record); err != nil {
			return nil, err
		}
		nameKey, err := types.GetNameKeyPrefix(record.Name)
		if err != nil {
			return nil, err
		}
		entry := rekeyedEntry{
			oldKey: slices.Clone(iter.Key()),
			newKey: newKey(iter.Key(), nameKey),
			value:  slices.Clone(iter.Value()),
		}
		if bytes.Equal(entry.oldKey, entry.newKey) {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
