package keeper

import (
	"bytes"
	"fmt"
	"sort"

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
	var mdAddrsToRedo []types.MetadataAddress
	steps := []struct {
		name   string
		runner func() error
	}{
		{
			name:   "Deleting bad indexes",
			runner: func() error {
				mdAddrsToRedo, err = deleteBadIndexes(ctx, m.keeper.storeKey)
				return err
			},
		},
		{
			name:   "Reindex scopes",
			runner: func() error { return reindexScopes(ctx, m.keeper, mdAddrsToRedo) },
		},
		{
			name:   "Reindex scope specs",
			runner: func() error { return reindexScopeSpecs(ctx, m.keeper, mdAddrsToRedo) },
		},
		{
			name:   "Reindex contract specs",
			runner: func() error { return reindexContractSpecs(ctx, m.keeper, mdAddrsToRedo) },
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

// deleteBadIndexes deletes all the bad indexes on Metadata entries, and gets the metadata addresses of things de-indexed.
func deleteBadIndexes(ctx sdk.Context, storeKey sdk.StoreKey) ([]types.MetadataAddress, error) {
	prefixes := [][]byte{
		types.AddressScopeCacheKeyPrefix, types.ValueOwnerScopeCacheKeyPrefix,
		types.AddressScopeSpecCacheKeyPrefix, types.AddressContractSpecCacheKeyPrefix,
	}
	deleted := []types.MetadataAddress{}

	store := ctx.KVStore(storeKey)
	for _, pre := range prefixes {
		newDels, err := cleanStore(ctx, prefix.NewStore(store, pre), pre)
		if err != nil {
			return nil, err
		}
		deleted = append(deleted, newDels...)
	}
	// Get only unique entries.
	sort.Slice(deleted, func(i, j int) bool {
		return bytes.Compare(deleted[i], deleted[j]) < 0
	})
	rv := []types.MetadataAddress{}
	for i, v := range deleted {
		if i == 0 || !v.Equals(deleted[i-1]) {
			rv = append(rv, v)
		}
	}
	return rv, nil
}

// cleanStore deletes all the bad index entries in a store, and returns the Metadata Addresses of things deleted.
func cleanStore(ctx sdk.Context, baseStore sdk.KVStore, pre []byte) (deleted []types.MetadataAddress, err error) {
	// All of the prefix stores being given to this are prefix stores for index entries involving account addresses.
	// All of those index entries will have the format {type}{addr length}{addr}{metadata addr}.
	// The {type} byte will be part of the store provided, so all the iterator keys will just have {addr length}{addr}{metadata addr}.
	// Incorrectly migrated indexes will be 1 byte shorter overall than the corrected indexes.
	// Basically, what was identified as the last byte of the account addr, was actually the first byte of the metadata address.
	// Then the rest of the metadata address was appended to finish off the key.
	// By identifying the incorrect entries and extracting the metadata addresses from them, we can identify what exactly needs reindexing.
	// Metadata addresses involved in these indexes have length 33.
	// The correct length of the iterator keys is 33 + {addr length} + 1.
	deleted = []types.MetadataAddress{}
	store := prefix.NewStore(baseStore, pre)
	iter := store.Iterator(nil, nil)
	defer func() {
		err = iter.Close()
	}()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		// If the key has less than 34 bytes, something is really wrong, just delete it.
		if len(key) < 34 {
			ctx.Logger().Error("Deleting unknown key with length %d (less than 34): %X", len(key), key)
			store.Delete(key)
			continue
		}
		mdAddr := types.MetadataAddress(key[len(key)-33:])
		if mdErr := mdAddr.Validate(); mdErr != nil {
			ctx.Logger().Error("Deleting unknown key with invalid metadata address: %X:%v", mdAddr, mdErr)
			store.Delete(key)
			continue
		}
		accAddrLen := int(key[0])
		accAddr := key[1:len(key)-33]
		if len(accAddr) != accAddrLen {
			store.Delete(key)
			deleted = append(deleted, mdAddr)
		}
	}
	return deleted, nil
}

// reindexScopes indexes all scopes.
func reindexScopes(ctx sdk.Context, mdKeeper Keeper, mdAddrs []types.MetadataAddress) error {
	notFound := []types.MetadataAddress{}
	for _, mdAddr := range mdAddrs {
		if !mdAddr.IsScopeAddress() {
			continue
		}
		scope, found := mdKeeper.GetScope(ctx, mdAddr)
		if !found {
			notFound = append(notFound, mdAddr)
			continue
		}
		// Only need to reindex the owners, data access, and value owner addresses (not the spec).
		// Make an "old" scope without those so that they're identified as changed and get recreated.
		oldScope := types.Scope{
			ScopeId:           scope.ScopeId,
			SpecificationId:   scope.SpecificationId,
			Owners:            nil,
			DataAccess:        nil,
			ValueOwnerAddress: "",
		}
		mdKeeper.indexScope(ctx, &scope, &oldScope)
	}
	if len(notFound) > 0 {
		return fmt.Errorf("scope(s) not found (%d): %q", len(notFound), notFound)
	}
	return nil
}

// reindexScopeSpecs indexes all scope specifications.
func reindexScopeSpecs(ctx sdk.Context, mdKeeper Keeper, mdAddrs []types.MetadataAddress) error {
	notFound := []types.MetadataAddress{}
	for _, mdAddr := range mdAddrs {
		if !mdAddr.IsScopeSpecificationAddress() {
			continue
		}
		scopeSpec, found := mdKeeper.GetScopeSpecification(ctx, mdAddr)
		if !found {
			notFound = append(notFound, mdAddr)
			continue
		}
		// Only need to reindex the owner addresses (not the contract specs).
		// Make an "old" scope spec without the addresses so that they're identified as changed and get recreated.
		oldScopeSpec := types.ScopeSpecification{
			SpecificationId: scopeSpec.SpecificationId,
			Description:     scopeSpec.Description,
			OwnerAddresses:  nil,
			PartiesInvolved: scopeSpec.PartiesInvolved,
			ContractSpecIds: scopeSpec.ContractSpecIds,
		}
		mdKeeper.indexScopeSpecification(ctx, &scopeSpec, &oldScopeSpec)
	}
	if len(notFound) > 0 {
		return fmt.Errorf("scope specification(s) not found (%d): %q", len(notFound), notFound)
	}
	return nil
}

// reindexContractSpecs indexes all contract specifications.
func reindexContractSpecs(ctx sdk.Context, mdKeeper Keeper, mdAddrs []types.MetadataAddress) error {
	notFound := []types.MetadataAddress{}
	for _, mdAddr := range mdAddrs {
		if !mdAddr.IsContractSpecificationAddress() {
			continue
		}
		contractSpec, found := mdKeeper.GetContractSpecification(ctx, mdAddr)
		if !found {
			notFound = append(notFound, mdAddr)
			continue
		}
		// Only thing indexed here is the owner addresses, so we can pass in nil and redo all for it.
		mdKeeper.indexContractSpecification(ctx, &contractSpec, nil)
	}
	if len(notFound) > 0 {
		return fmt.Errorf("contract specification(s) not found (%d): %q", len(notFound), notFound)
	}
	return nil
}
