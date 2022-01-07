package keeper

import (
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
	var goodIndexes keyLookup
	var sessionsToDelete []types.MetadataAddress
	var sessionsToKeep keyLookup
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
			name:   "Reindexing scopes",
			runner: func() error { return reindexScopes(ctx, m.keeper, goodIndexes) },
		},
		{
			name:   "Reindexing scope specs",
			runner: func() error { return reindexScopeSpecs(ctx, m.keeper, goodIndexes) },
		},
		{
			name:   "Reindexing contract specs",
			runner: func() error { return reindexContractSpecs(ctx, m.keeper, goodIndexes) },
		},
		{
			name: "Identifying sessions to keep",
			runner: func() error {
				var err error
				sessionsToKeep, err = identifySessionsWithRecords(ctx, m.keeper)
				return err
			},
		},
		{
			name: "Finding empty sessions",
			runner: func() error {
				var err error
				sessionsToDelete, err = identifySessionsToDelete(ctx, m.keeper, sessionsToKeep)
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

// keyLookup is a map used to identify known keys.
type keyLookup map[string]struct{}

// newKeyLookup creates a new empty keyLookup.
func newKeyLookup() keyLookup {
	return map[string]struct{}{}
}

// add records the given index keys in this keyLookup.
func (m *keyLookup) add(indexKeys ...[]byte) {
	for _, key := range indexKeys {
		k := string(key)
		if _, has := (*m)[k]; !has {
			(*m)[k] = struct{}{}
		}
	}
}

// has returns true if the provided indexKey is known to this keyLookup.
func (m keyLookup) has(indexKey []byte) bool {
	_, rv := m[string(indexKey)]
	return rv
}

// deleteBadIndexes deletes all the bad indexes on Metadata entries, and gets the metadata addresses of things properly indexed.
// This is a function for a migration, not intended for outside use.
func deleteBadIndexes(ctx sdk.Context, mdKeeper Keeper) (keyLookup, error) {
	prefixes := [][]byte{
		types.AddressScopeCacheKeyPrefix, types.ValueOwnerScopeCacheKeyPrefix,
		types.AddressScopeSpecCacheKeyPrefix, types.AddressContractSpecCacheKeyPrefix,
	}
	good := newKeyLookup()
	store := ctx.KVStore(mdKeeper.storeKey)
	for _, pre := range prefixes {
		newGood, err := cleanStore(ctx, store, pre)
		if err != nil {
			return nil, err
		}
		good.add(newGood...)
	}
	ctx.Logger().Info("Done identifying good indexes and deleting bad ones.")
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
	i := 0
	ctx.Logger().Info(fmt.Sprintf("Prefix %X: Identifying good entries and deleting bad ones.", pre))
	for ; iter.Valid(); iter.Next() {
		i++
		key := iter.Key()
		accAddrLen := int(key[0])
		// good key length = 1 length byte + the length of the account address + 17 metadata address bytes.
		if len(key) == accAddrLen+18 {
			good = append(good, append([]byte{pre[0]}, key...))
		} else {
			store.Delete(key)
			deletedCount++
		}
		if i%10000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("Prefix %X: Deleted %d entries, found %d good entries so far.", pre, deletedCount, len(good)))
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Prefix %X: Done. Deleted %d entries, found %d good entries.", pre, deletedCount, len(good)))
	return good, nil
}

// reindexScopes creates all missing scope indexes.
// This is a function for a migration, not intended for outside use.
func reindexScopes(ctx sdk.Context, mdKeeper Keeper, lookup keyLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	i := 0
	ri := 0
	rv := mdKeeper.IterateScopes(ctx, func(scope types.Scope) (stop bool) {
		i++
		didri := false
		for _, key := range getScopeIndexValues(&scope).IndexKeys() {
			if (key[0] == types.AddressScopeCacheKeyPrefix[0] || key[0] == types.ValueOwnerScopeCacheKeyPrefix[0]) && !lookup.has(key) {
				store.Set(key, []byte{0x01})
				didri = true
			}
		}
		if didri {
			ri++
		}
		if i%10000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("Checked %d scopes and reindexed %d of them.", i, ri))
		}
		return false
	})
	ctx.Logger().Info(fmt.Sprintf("Done reindexing %d scopes. All %d are now indexed.", ri, i))
	return rv
}

// reindexScopeSpecs creates all missing scope specification indexes.
// This is a function for a migration, not intended for outside use.
func reindexScopeSpecs(ctx sdk.Context, mdKeeper Keeper, lookup keyLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	i := 0
	ri := 0
	rv := mdKeeper.IterateScopeSpecs(ctx, func(scopeSpec types.ScopeSpecification) (stop bool) {
		i++
		didri := false
		for _, key := range getScopeSpecIndexValues(&scopeSpec).IndexKeys() {
			if key[0] == types.AddressScopeSpecCacheKeyPrefix[0] && !lookup.has(key) {
				store.Set(key, []byte{0x01})
				didri = true
			}
		}
		if didri {
			ri++
		}
		if i%10000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("Checked %d scope specs and reindexed %d of them.", i, ri))
		}
		return false
	})
	ctx.Logger().Info(fmt.Sprintf("Done reindexing %d scope specs. All %d are now indexed.", ri, i))
	return rv
}

// reindexContractSpecs creates all missing contract specification indexes.
// This is a function for a migration, not intended for outside use.
func reindexContractSpecs(ctx sdk.Context, mdKeeper Keeper, lookup keyLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	i := 0
	ri := 0
	rv := mdKeeper.IterateContractSpecs(ctx, func(contractSpec types.ContractSpecification) (stop bool) {
		i++
		didri := false
		for _, key := range getContractSpecIndexValues(&contractSpec).IndexKeys() {
			if !lookup.has(key) {
				store.Set(key, []byte{0x01})
				didri = true
			}
		}
		if didri {
			ri++
		}
		if i%10000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("Checked %d contract specs and reindexed %d of them.", i, ri))
		}
		return false
	})
	ctx.Logger().Info(fmt.Sprintf("Done reindexing %d contract specs. All %d are now indexed.", ri, i))
	return rv
}

// identifySessionsWithRecords identifies all sessions that have records associated with them.
// This is a function for a migration, not intended for outside use.
func identifySessionsWithRecords(ctx sdk.Context, mdKeeper Keeper) (rv keyLookup, err error) {
	rv = newKeyLookup()
	pre := types.RecordKeyPrefix
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(mdKeeper.storeKey), pre)
	defer func() {
		err = iter.Close()
	}()
	i := 0
	for ; iter.Valid(); iter.Next() {
		i++
		var record types.Record
		merr := mdKeeper.cdc.Unmarshal(iter.Value(), &record)
		if merr == nil {
			rv.add(record.SessionId)
		}
		if i%10000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("Checked %d records and found %d session ids so far.", i, len(rv)))
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Done identifying a total of %d session ids used by %d records.", len(rv), i))
	return
}

// identifySessionsToDelete identifies all sessions that are not in the provided keyLookup.
// This is a function for a migration, not intended for outside use.
func identifySessionsToDelete(ctx sdk.Context, mdKeeper Keeper, sessionsToKeep keyLookup) (rv []types.MetadataAddress, err error) {
	pre := types.SessionKeyPrefix
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(mdKeeper.storeKey), pre)
	defer func() {
		err = iter.Close()
	}()
	i := 0
	var key []byte
	for ; iter.Valid(); iter.Next() {
		i++
		key = iter.Key()
		if !sessionsToKeep.has(key) {
			rv = append(rv, key)
		}
		if i%10000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("Checked %d sessions and found %d without records so far.", i, len(rv)))
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Done identifying a total of %d sessions without records (out of %d sessions).", len(rv), i))
	return
}

// deleteSessions is a migration function that deletes the provided sessions.
// This is a function for a migration, not intended for outside use.
func deleteSessions(ctx sdk.Context, mdKeeper Keeper, sessionsToDelete []types.MetadataAddress) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	for i, sessionID := range sessionsToDelete {
		store.Delete(sessionID)
		if i%10000 == 9999 {
			ctx.Logger().Info(fmt.Sprintf("Deleted %d empty sessions, %d to go.", i+1, len(sessionsToDelete)-i-1))
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Done deleting %d empty sessions.", len(sessionsToDelete)))
	return nil
}
