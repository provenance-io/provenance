package keeper

import (
	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IterateScopes processes all stored scopes with the given handler.
func (k Keeper) IterateScopes(ctx sdk.Context, handler func(types.Scope) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.ScopeKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scope types.Scope
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &scope)
		if handler(scope) {
			break
		}
	}
	return nil
}

// IterateScopesForAddress processes scopes associated with the provided address with the given handler.
func (k Keeper) IterateScopesForAddress(ctx sdk.Context, address sdk.AccAddress, handler func(scopeID types.MetadataAddress) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetAddressCacheIteratorPrefix(address)
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scopeID types.MetadataAddress
		if err := scopeID.Unmarshal(it.Key()[len(prefix):]); err != nil {
			return err
		}
		if handler(scopeID) {
			break
		}
	}
	return nil
}

// GetScope returns the scope with the given id.
func (k Keeper) GetScope(ctx sdk.Context, id types.MetadataAddress) (scope types.Scope, found bool) {
	if !id.IsScopeAddress() {
		return scope, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id.Bytes())
	if b == nil {
		return types.Scope{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &scope)
	return scope, true
}

// SetScope stores a scope in the module kv store.
func (k Keeper) SetScope(ctx sdk.Context, scope types.Scope) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&scope)
	store.Set(scope.ScopeId, b)
}

// RemoveScope removes a scope from the module kv store.
func (k Keeper) RemoveScope(ctx sdk.Context, id types.MetadataAddress) {
	// iterate and remove all records, groups
	store := ctx.KVStore(k.storeKey)

	// Remove all records
	prefix, err := id.ScopeRecordIteratorPrefix()
	if err != nil {
		panic(err)
	}
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.RemoveRecord(ctx, types.MetadataAddress(iter.Key()))
	}

	// Remove all record groups
	prefix, err = id.ScopeGroupIteratorPrefix()
	if err != nil {
		panic(err)
	}
	iter = sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.RemoveRecordGroup(ctx, types.MetadataAddress(iter.Key()))
	}

	// TODO : remove address index records for all OwnerAddress records.
	// TODO : remove value_owner index record

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeScopeRemoved,
			sdk.NewAttribute(types.AttributeKeyScopeID, id.String()),
		),
	)

	store.Delete(id)
}
