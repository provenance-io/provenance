package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/gogoproto/proto"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// Migrate3To4 will update the metadata store from version 3 to version 4. This should be part of the viridian upgrade.
func (m Migrator) Migrate3To4(ctx sdk.Context) error {
	logger := m.keeper.Logger(ctx)
	logger.Info("Starting migration of x/metadata from 3 to 4.")
	if err := migrateValueOwners(ctx, newKeeper3To4(m.keeper)); err != nil {
		logger.Error("Error migrating scope value owners.", "error", err)
		return err
	}
	logger.Info("Done migrating x/metadata from 3 to 4.")
	return nil
}

// keeper3To4I is an interface with keeper-related stuff needed to migrate metadata from 3 to 4.
type keeper3To4I interface {
	Logger(ctx sdk.Context) log.Logger
	GetStore(ctx sdk.Context) storetypes.KVStore
	Unmarshal(bz []byte, ptr proto.Message) error
	SetScopeValueOwner(ctx sdk.Context, scopeID types.MetadataAddress, newValueOwner string) error
}

// keeper3To4 is a wrapper on the metadata Keeper with a few extra things exposed to satisfy the keeper3To4I interface.
type keeper3To4 struct {
	Keeper
}

// newKeeper3To4 wraps the provided metadata Keeper as a keeper3To4.
func newKeeper3To4(kpr Keeper) keeper3To4 {
	return keeper3To4{Keeper: kpr}
}

// GetStore returns a store for the metadata stuff.
func (k keeper3To4) GetStore(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.Keeper.storeKey)
}

// Unmarshal uses the Keeper's codec to unmarshal something.
func (k keeper3To4) Unmarshal(bz []byte, ptr proto.Message) error {
	return k.Keeper.cdc.Unmarshal(bz, ptr)
}

// V3WriteNewScope writes a new scope to state in the way v3 of the metadata module did.
// Deprecated: Only exists to facilitate testing of the migration of the metadata module from v3 to v4.
func (k Keeper) V3WriteNewScope(ctx sdk.Context, scope types.Scope) error {
	store := ctx.KVStore(k.storeKey)
	if store.Has(scope.ScopeId) {
		return fmt.Errorf("scope %s already exists", scope.ScopeId)
	}
	bz, err := k.cdc.Marshal(&scope)
	if err != nil {
		return fmt.Errorf("could not marshal scope %s: %w", scope.ScopeId, err)
	}
	store.Set(scope.ScopeId, bz)
	k.indexScope(store, &scope, nil)
	// indexScope no longer does anything with the value owner address, but
	// we used to have a couple index entries for it.
	if len(scope.ValueOwnerAddress) > 0 {
		vo := sdk.MustAccAddressFromBech32(scope.ValueOwnerAddress)
		k1 := types.GetAddressScopeCacheKey(vo, scope.ScopeId)
		k2 := GetValueOwnerScopeCacheKey(vo, scope.ScopeId)
		store.Set(k1, []byte{0x01})
		store.Set(k2, []byte{0x01})
	}
	return nil
}

// migrateValueOwners will loop through all scopes and move the value owner info into the bank module.
func migrateValueOwners(ctx sdk.Context, kpr keeper3To4I) error {
	logger := kpr.Logger(ctx)
	logger.Info("Moving scope value owner data into x/bank ledger.")
	store := kpr.GetStore(ctx)
	it := storetypes.KVStorePrefixIterator(store, types.ScopeKeyPrefix)
	defer it.Close()

	// If a scope's value owner is a marker, someone had the required deposit permission.
	// But we don't have that permission here, and have no way to get it again. So, we
	// bypass the marker send restrictions under the assumption that if a scope has a
	// marker for a value owner, it was set that way by someone with proper permissions.
	// We do NOT bypass the quarantine send restrictions though because we don't
	// actually know that the value owner wanted to be the value owner of the scope.
	ctx = markertypes.WithBypass(ctx)

	scopeCount := 0
	valueOwnerCount := 0
	for ; it.Valid(); it.Next() {
		scopeCount++
		scopeBz := it.Value()
		var scope types.Scope
		if err := kpr.Unmarshal(scopeBz, &scope); err != nil {
			scopeID := types.MetadataAddress(it.Key())
			logger.Error(fmt.Sprintf("[%d]: ScopeID=%q", scopeCount, scopeID), "bytes", scopeBz)
			return fmt.Errorf("error reading scope %s from state: %w", scopeID, err)
		}

		if len(scope.ValueOwnerAddress) > 0 {
			valueOwnerCount++
			if err := migrateValueOwnerToBank(ctx, kpr, store, scope); err != nil {
				return err
			}
		}

		if scopeCount%10_000 == 0 {
			logger.Info("Progress update:", "scopes", scopeCount, "value owners", valueOwnerCount)
		}
	}
	logger.Info("Done moving scope value owners into bank module.", "scopes", scopeCount, "value owners", valueOwnerCount)
	return nil
}

// migrateValueOwnerToBank will switch a scope's value owner to be maintained by the bank module instead of in the scope.
func migrateValueOwnerToBank(ctx sdk.Context, kpr keeper3To4I, store storetypes.KVStore, scope types.Scope) error {
	if err := kpr.SetScopeValueOwner(ctx, scope.ScopeId, scope.ValueOwnerAddress); err != nil {
		return fmt.Errorf("could not migrate scope %s value owner %q to bank module: %w",
			scope.ScopeId, scope.ValueOwnerAddress, err)
	}
	deleteValueOwnerIndexEntries(store, scope)
	return nil
}

// valueOwnerScopeCacheKeyPrefix is the prefix key that we used to use for a value owner -> scope index.
var valueOwnerScopeCacheKeyPrefix = []byte{0x18}

// getValueOwnerScopeCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func getValueOwnerScopeCacheIteratorPrefix(addr sdk.AccAddress) []byte {
	return append(valueOwnerScopeCacheKeyPrefix, address.MustLengthPrefix(addr.Bytes())...)
}

// GetValueOwnerScopeCacheKey returns the store key for an address cache entry
func GetValueOwnerScopeCacheKey(addr sdk.AccAddress, scopeID types.MetadataAddress) []byte {
	return append(getValueOwnerScopeCacheIteratorPrefix(addr), scopeID.Bytes()...)
}

// deleteValueOwnerIndexEntries will delete the index entries involving a scope's value owner.
func deleteValueOwnerIndexEntries(store storetypes.KVStore, scope types.Scope) {
	// Don't do anything without a valid value owner.
	vo, err := sdk.AccAddressFromBech32(scope.ValueOwnerAddress)
	if err != nil || len(vo) == 0 {
		return
	}

	// Delete the value owner -> scope index entry (that's now a denom owners thing).
	key := GetValueOwnerScopeCacheKey(vo, scope.ScopeId)
	store.Delete(key)

	// The address -> scope index no longer associates a value owner with a scope; it only applies to the Owners.
	// So, if the value owner is also in the list of owners, we keep the entry, otherwise, we delete it.
	for _, owner := range scope.Owners {
		if owner.Address == scope.ValueOwnerAddress {
			return // The value owner is also an owner. Nothing more to do.
		}
	}
	key = types.GetAddressScopeCacheKey(vo, scope.ScopeId)
	store.Delete(key)
}
