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
	var err error
	var goodIndexes *IndexLookup
	steps := []struct {
		name   string
		runner func() error
	}{
		{
			name: "Deleting bad indexes and getting good",
			runner: func() error {
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

type IndexLookup struct {
	Entries map[byte][][]byte
}

func NewIndexLookup() *IndexLookup {
	return &IndexLookup{
		Entries: map[byte][][]byte{},
	}
}

func (i *IndexLookup) AddEntries(pk byte, keys ...[]byte) {
	i.Entries[pk] = append(i.Entries[pk], keys...)
}

func (i IndexLookup) Has(key []byte) bool {
	if len(key) == 0 {
		return false
	}
	l, ok := i.Entries[key[0]]
	if !ok {
		return false
	}
	for _, k := range l {
		if bytes.Equal(key, k) {
			return true
		}
	}
	return false
}

// deleteBadIndexes deletes all the bad indexes on Metadata entries, and gets the metadata addresses of things de-indexed.
func deleteBadIndexes(ctx sdk.Context, mdKeeper Keeper) (*IndexLookup, error) {
	prefixes := [][]byte{
		types.AddressScopeCacheKeyPrefix, types.ValueOwnerScopeCacheKeyPrefix,
		types.AddressScopeSpecCacheKeyPrefix, types.AddressContractSpecCacheKeyPrefix,
	}
	good := NewIndexLookup()
	store := ctx.KVStore(mdKeeper.storeKey)
	for _, pre := range prefixes {
		newGood, err := cleanStore(ctx, store, pre)
		if err != nil {
			return nil, err
		}
		good.AddEntries(pre[0], newGood...)
	}
	return good, nil
}

// cleanStore deletes all the bad index entries with the given prefix, and returns all the good (remaining) entries with that prefix.
func cleanStore(ctx sdk.Context, baseStore sdk.KVStore, pre []byte) (good [][]byte, err error) {
	// All of the prefix stores being given to this are prefix stores for index entries involving account addresses.
	// All of those index entries will have the format {type}{addr length}{addr}{metadata addr}.
	// The {type} byte will be part of the store provided, so all the iterator keys will just have {addr length}{addr}{metadata addr}.
	// Incorrectly migrated indexes will be 1 byte shorter overall than the corrected indexes.
	// Basically, what was identified as the last byte of the account addr, was actually the first byte of the metadata address.
	// Then the rest of the metadata address was appended to finish off the key.
	// Metadata addresses involved in these indexes have length 33.
	// The correct length of the iterator keys is 33 + {addr length} + 1.
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
		// good key length = 1 length byte + the length of the account address + 33 metadata address bytes.
		if len(key) == 1+accAddrLen+33 {
			good = append(good, append([]byte{pre[0]}, key...))
		} else {
			store.Delete(key)
			deletedCount++
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Found %d good index entries and %d were deleted.", len(good), deletedCount))
	return good, nil
}

// reindexScopes creates all missing scope indexes.
func reindexScopes(ctx sdk.Context, mdKeeper Keeper, lookup *IndexLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	return mdKeeper.IterateScopes(ctx, func(scope types.Scope) (stop bool) {
		for _, key := range getScopeIndexValues(&scope).IndexKeys() {
			if (key[0] == types.AddressScopeCacheKeyPrefix[0] || key[0] == types.ValueOwnerScopeCacheKeyPrefix[0]) && !lookup.Has(key) {
				store.Set(key, []byte{0x01})
			}
		}
		return false
	})
}

// reindexScopeSpecs creates all missing scope specification indexes.
func reindexScopeSpecs(ctx sdk.Context, mdKeeper Keeper, lookup *IndexLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	return mdKeeper.IterateScopeSpecs(ctx, func(scopeSpec types.ScopeSpecification) (stop bool) {
		for _, key := range getScopeSpecIndexValues(&scopeSpec).IndexKeys() {
			if key[0] == types.AddressScopeSpecCacheKeyPrefix[0] && !lookup.Has(key) {
				store.Set(key, []byte{0x01})
			}
		}
		return false
	})
}

// reindexContractSpecs creates all missing contract specification indexes.
func reindexContractSpecs(ctx sdk.Context, mdKeeper Keeper, lookup *IndexLookup) error {
	store := ctx.KVStore(mdKeeper.storeKey)
	return mdKeeper.IterateContractSpecs(ctx, func(contractSpec types.ContractSpecification) (stop bool) {
		for _, key := range getContractSpecIndexValues(&contractSpec).IndexKeys() {
			if !lookup.Has(key) {
				store.Set(key, []byte{0x01})
			}
		}
		return false
	})
}
