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
	steps := []struct {
		name   string
		runner func() error
	}{
		{
			name:   "Deleting indexes",
			runner: func() error { return deleteAllIndexes(ctx, m.keeper.storeKey) },
		},
		{
			name:   "Reindexing scopes",
			runner: func() error { return reindexScopes(ctx, m.keeper) },
		},
		{
			name:   "Reindexing scope specs",
			runner: func() error { return reindexScopeSpecs(ctx, m.keeper) },
		},
		{
			name:   "Reindexing contract specs",
			runner: func() error { return reindexContractSpecs(ctx, m.keeper) },
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

// deleteAllIndexes deletes all the indexes on Metadata entries.
func deleteAllIndexes(ctx sdk.Context, storeKey sdk.StoreKey) error {
	prefixes := [][]byte{
		types.AddressScopeCacheKeyPrefix, types.ScopeSpecScopeCacheKeyPrefix,
		types.ValueOwnerScopeCacheKeyPrefix, types.AddressScopeSpecCacheKeyPrefix,
		types.ContractSpecScopeSpecCacheKeyPrefix, types.AddressContractSpecCacheKeyPrefix,
	}

	store := ctx.KVStore(storeKey)
	for _, pre := range prefixes {
		if err := clearStore(prefix.NewStore(store, pre)); err != nil {
			return err
		}
	}
	return nil
}

// clearStore deletes all the entries in a store.
func clearStore(store sdk.KVStore) (err error) {
	iter := store.Iterator(nil, nil)
	defer func() {
		err = iter.Close()
	}()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
	return nil
}

// reindexScopes indexes all scopes.
func reindexScopes(ctx sdk.Context, mdKeeper Keeper) error {
	return mdKeeper.IterateScopes(ctx, func(scope types.Scope) (stop bool) {
		mdKeeper.indexScope(ctx, scope)
		return false
	})
}

// reindexScopeSpecs indexes all scope specifications.
func reindexScopeSpecs(ctx sdk.Context, mdKeeper Keeper) error {
	return mdKeeper.IterateScopeSpecs(ctx, func(spec types.ScopeSpecification) (stop bool) {
		mdKeeper.indexScopeSpecification(ctx, spec)
		return false
	})
}

// reindexContractSpecs indexes all contract specifications.
func reindexContractSpecs(ctx sdk.Context, mdKeeper Keeper) error {
	return mdKeeper.IterateContractSpecs(ctx, func(spec types.ContractSpecification) (stop bool) {
		mdKeeper.indexContractSpecification(ctx, spec)
		return false
	})
}
