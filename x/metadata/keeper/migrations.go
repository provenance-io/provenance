package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v042 "github.com/provenance-io/provenance/x/metadata/legacy/v042"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2 to convert attribute keys from 20 to 32 length
func (m *Migrator) Migrate1to2(ctx sdk.Context) error {
	ctx.Logger().Info("Migrating Metadata Module from Version 1 to 2")
	err := v042.MigrateAddresses(ctx, m.keeper.storeKey)
	ctx.Logger().Info("Finished Migrating Metadata Module from Version 1 to 2")
	return err
}

// Migrate2to3 migrates from version 2 to 3 to fix the indexes.
func (m *Migrator) Migrate2to3(ctx sdk.Context) error {
	ctx.Logger().Info("Migrating Metadata Module from Version 2 to 3")
	var goodIndexes *indexLookup
	var sessionsToDelete []types.MetadataAddress
	steps := []struct {
		name   string
		runner func() error
	}{
		{
			name: "Deleting bad indexes and getting good",
			runner: func() error {
				var err error
				goodIndexes, err = deleteBadIndexes(ctx, m.keeper)
				return err
			},
		},
		{
			name:   "Reindex scopes",
			runner: func() error { return reindexScopes(ctx, m.keeper, goodIndexes) },
		},
		{
			name:   "Reindex scope specs",
			runner: func() error { return reindexScopeSpecs(ctx, m.keeper, goodIndexes) },
		},
		{
			name:   "Reindex contract specs",
			runner: func() error { return reindexContractSpecs(ctx, m.keeper, goodIndexes) },
		},
		{
			name: "Finding empty sessions",
			runner: func() error {
				var err error
				sessionsToDelete, err = getEmptySessions(ctx, m.keeper)
				return err
			},
		},
		{
			name:   "Deleting empty sessions",
			runner: func() error { return deleteSessions(ctx, m.keeper, sessionsToDelete) },
		},
	}

	for i, step := range steps {
		logHeader := fmt.Sprintf("Metadata step %d of %d: %s", i+1, len(steps), step.name)
		ctx.Logger().Info(logHeader)
		if err := step.runner(); err != nil {
			ctx.Logger().Error(logHeader, "error", err)
			return err
		}
	}

	ctx.Logger().Info("Finished Migrating Metadata Module from Version 2 to 3")
	return nil
}

// indexLookup is a struct holding index byte slices in a way that makes them easier to find.
type indexLookup struct {
	// entries is a map of the first keyLen bytes to all indexes that start with those bytes.
	entries map[string][][]byte
	// pkLen is the number of bytes from index to use as the key for entries.
	pkLen int
}

// newIndexLookup creates a new empty IndexLookup.
func newIndexLookup() *indexLookup {
	return &indexLookup{
		entries: map[string][][]byte{},
		// All the keys we have so far have length 20.
		// Add 1 for the type byte, and 1 for length byte. Then also use the first 2 of metadata address.
		// 20 + 1 + 1 + 2 = 24.
		pkLen: 24,
	}
}

// getPK creates a primary key string for the provided index key.
func (i indexLookup) getPK(indexKey []byte) string {
	if len(indexKey) <= i.pkLen {
		return string(indexKey)
	}
	return string(indexKey[:i.pkLen])
}

// add records the given index keys in this indexLookup.
func (i *indexLookup) add(indexKeys ...[]byte) {
	for _, indexKey := range indexKeys {
		pk := i.getPK(indexKey)
		i.entries[pk] = append(i.entries[pk], indexKey)
	}
}

// has returns true if the provided indexKey is known to this indexLookup.
func (i indexLookup) has(indexKey []byte) bool {
	if len(indexKey) == 0 {
		return false
	}
	l, ok := i.entries[i.getPK(indexKey)]
	if !ok {
		return false
	}
	for _, k := range l {
		if bytes.Equal(indexKey, k) {
			return true
		}
	}
	return false
}

// deleteBadIndexes deletes all the bad indexes on Metadata entries, and gets the metadata addresses of things de-indexed.
// This is a function for a migration, not intended for outside use.
func deleteBadIndexes(ctx sdk.Context, mdKeeper Keeper) (*indexLookup, error) {
	prefixes := [][]byte{
		types.AddressScopeCacheKeyPrefix, types.ValueOwnerScopeCacheKeyPrefix,
		types.AddressScopeSpecCacheKeyPrefix, types.AddressContractSpecCacheKeyPrefix,
	}
	good := newIndexLookup()
	store := ctx.KVStore(mdKeeper.storeKey)
	for _, pre := range prefixes {
		newGood, err := cleanStore(ctx, store, pre)
		if err != nil {
			return nil, err
		}
		good.add(newGood...)
	}
	return good, nil
}

// cleanStore deletes all the bad index entries with the given prefix, and returns all the good (remaining) entries with that prefix.
// This is a function for a migration, not intended for outside use.
func cleanStore(ctx sdk.Context, baseStore sdk.KVStore, pre []byte) (good [][]byte, err error) {
	// All of the prefix stores being given to this are prefix stores for index entries involving account addresses.
	// All of those index entries will have the format {type}{addr length}{addr}{metadata addr}.
	// The {type} byte will be part of the store provided, so all the iterator keys will just have {addr length}{addr}{metadata addr}.
	// Incorrectly migrated indexes will be 1 byte shorter overall than the corrected indexes.
	// Basically, what was identified as the last byte of the account addr, was actually the first byte of the metadata address.
	// Then the rest of the metadata address was appended to finish off the key.
	// Metadata addresses involved in these indexes have length 17 (one type byte, and 16 uuid bytes).
	// The correct length of the iterator keys is 17 + {addr length} + 1.
	good = [][]byte{}
	deletedCount := 0
	store := prefix.NewStore(baseStore, pre)
	iter := store.Iterator(nil, nil)
	defer func() {
		err = iter.Close()
	}()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		accAddrLen := int(key[0])
		// good key length = 1 length byte + the length of the account address + 17 metadata address bytes.
		if len(key) == accAddrLen+18 {
			good = append(good, append([]byte{pre[0]}, key...))
		} else {
			store.Delete(key)
			deletedCount++
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Prefix %X: Deleted %d entries, found %d good entries.", pre, deletedCount, len(good)))
	return good, nil
}

// reindexScopes creates all missing scope indexes.
// This is a function for a migration, not intended for outside use.
func reindexScopes(ctx sdk.Context, mdKeeper Keeper, lookup *indexLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	return mdKeeper.IterateScopes(ctx, func(scope types.Scope) (stop bool) {
		for _, key := range getScopeIndexValues(&scope).IndexKeys() {
			if (key[0] == types.AddressScopeCacheKeyPrefix[0] || key[0] == types.ValueOwnerScopeCacheKeyPrefix[0]) && !lookup.has(key) {
				store.Set(key, []byte{0x01})
			}
		}
		return false
	})
}

// reindexScopeSpecs creates all missing scope specification indexes.
// This is a function for a migration, not intended for outside use.
func reindexScopeSpecs(ctx sdk.Context, mdKeeper Keeper, lookup *indexLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	return mdKeeper.IterateScopeSpecs(ctx, func(scopeSpec types.ScopeSpecification) (stop bool) {
		for _, key := range getScopeSpecIndexValues(&scopeSpec).IndexKeys() {
			if key[0] == types.AddressScopeSpecCacheKeyPrefix[0] && !lookup.has(key) {
				store.Set(key, []byte{0x01})
			}
		}
		return false
	})
}

// reindexContractSpecs creates all missing contract specification indexes.
// This is a function for a migration, not intended for outside use.
func reindexContractSpecs(ctx sdk.Context, mdKeeper Keeper, lookup *indexLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	return mdKeeper.IterateContractSpecs(ctx, func(contractSpec types.ContractSpecification) (stop bool) {
		for _, key := range getContractSpecIndexValues(&contractSpec).IndexKeys() {
			if !lookup.has(key) {
				store.Set(key, []byte{0x01})
			}
		}
		return false
	})
}

// getEmptySessions finds all sessions that don't have any records.
// This is a function for a migration, not intended for outside use.
func getEmptySessions(ctx sdk.Context, mdKeeper Keeper) (rv []types.MetadataAddress, err error) {
	store := ctx.KVStore(mdKeeper.storeKey)
	sPre := types.SessionKeyPrefix
	iter := sdk.KVStorePrefixIterator(store, sPre)
	defer func() {
		err = iter.Close()
	}()
	for ; iter.Valid(); iter.Next() {
		if !mdKeeper.sessionHasRecords(ctx, iter.Key()) {
			rv = appendMDIfNew(rv, iter.Key())
		}
	}
	return
}

// deleteSessions is a migration function that deletes the provided sessions.
// This is a function for a migration, not intended for outside use.
func deleteSessions(ctx sdk.Context, mdKeeper Keeper, sessionsToDelete []types.MetadataAddress) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	for _, sessionID := range sessionsToDelete {
		store.Delete(sessionID)
	}
	return nil
}

// appendMDIfNew appends elements to a slice that aren't already in the slice.
func appendMDIfNew(slice []types.MetadataAddress, elems ...types.MetadataAddress) []types.MetadataAddress {
	for _, elem := range elems {
		if !containsMD(slice, elem) {
			slice = append(slice, elem)
		}
	}
	return slice
}

// containsMD returns true if the slice contains the provided elem.
func containsMD(slice []types.MetadataAddress, elem types.MetadataAddress) bool {
	for _, md := range slice {
		if elem.Equals(md) {
			return true
		}
	}
	return false
}
