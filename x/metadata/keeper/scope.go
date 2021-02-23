package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
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
	prefix := types.GetAddressScopeCacheIteratorPrefix(address)
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

// IterateScopesForScopeSpec processes scopes associated with the provided scope specification id with the given handler.
func (k Keeper) IterateScopesForScopeSpec(ctx sdk.Context, scopeSpecID types.MetadataAddress,
	handler func(scopeID types.MetadataAddress) (stop bool),
) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetScopeSpecScopeCacheIteratorPrefix(scopeSpecID)
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

	eventType := types.EventTypeScopeCreated
	if store.Has(scope.ScopeId) {
		if oldScopeBytes := store.Get(scope.ScopeId); oldScopeBytes != nil {
			var oldScope types.Scope
			if err := k.cdc.UnmarshalBinaryBare(oldScopeBytes, &oldScope); err == nil {
				eventType = types.EventTypeScopeUpdated
				k.clearScopeIndex(ctx, oldScope)
			}
		}
	}

	store.Set(scope.ScopeId, b)
	k.indexScope(ctx, scope)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyScopeID, scope.ScopeId.String()),
			sdk.NewAttribute(types.AttributeKeyScope, scope.String()),
		),
	)
}

// DeleteScope removes a scope from the module kv store.
func (k Keeper) DeleteScope(ctx sdk.Context, id types.MetadataAddress) {
	// iterate and remove all records, groups
	store := ctx.KVStore(k.storeKey)

	scope, found := k.GetScope(ctx, id)
	if !found {
		return
	}

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

	// RecordGroups will be removed as the last record in each is deleted.

	k.clearScopeIndex(ctx, scope)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeScopeRemoved,
			sdk.NewAttribute(types.AttributeKeyScopeID, id.String()),
		),
	)

	store.Delete(id)
}

// clearScopeIndex delete any index records for this scope
func (k Keeper) clearScopeIndex(ctx sdk.Context, scope types.Scope) {
	store := ctx.KVStore(k.storeKey)

	// add all party addresses to the list of cache records to remove
	addresses := []string{}
	for _, p := range scope.Owners {
		addresses = append(addresses, p.Address)
	}
	addresses = append(addresses, scope.DataAccess...)
	if len(scope.ValueOwnerAddress) > 0 {
		// Add to list of general addresses to clear the cache of
		addresses = append(addresses, scope.ValueOwnerAddress)
		// Clear out the value owner cache
		addr, err := sdk.AccAddressFromBech32(scope.ValueOwnerAddress)
		if err == nil {
			store.Delete(types.GetValueOwnerScopeCacheKey(addr, scope.ScopeId))
		}
	}
	for _, a := range addresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Delete(types.GetAddressScopeCacheKey(addr, scope.ScopeId))
		}
	}
	store.Delete(types.GetScopeSpecScopeCacheKey(scope.SpecificationId, scope.ScopeId))
}

// indexScope create index records for the given scope
func (k Keeper) indexScope(ctx sdk.Context, scope types.Scope) {
	store := ctx.KVStore(k.storeKey)

	// Index all party addresses on the scope
	addresses := []string{}
	for _, p := range scope.Owners {
		addresses = append(addresses, p.Address)
	}
	addresses = append(addresses, scope.DataAccess...)
	if len(scope.ValueOwnerAddress) > 0 {
		addresses = append(addresses, scope.ValueOwnerAddress)
		// create a value owner cache entry as well.
		addr, err := sdk.AccAddressFromBech32(scope.ValueOwnerAddress)
		if err == nil {
			store.Set(types.GetValueOwnerScopeCacheKey(addr, scope.ScopeId), []byte{0x01})
		}
	}
	for _, a := range addresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Set(types.GetAddressScopeCacheKey(addr, scope.ScopeId), []byte{0x01})
		}
	}
	if len(scope.SpecificationId) > 0 {
		store.Set(types.GetScopeSpecScopeCacheKey(scope.SpecificationId, scope.ScopeId), []byte{0x01})
	}
}

// ValidateScopeUpdate checks the current scope and the proposed scope to determine if the the proposed changes are valid
// based on the existing state
func (k Keeper) ValidateScopeUpdate(ctx sdk.Context, existing, proposed types.Scope, signers []string) error {
	// IDs must match
	if len(existing.ScopeId) > 0 {
		if !proposed.ScopeId.Equals(existing.ScopeId) {
			return fmt.Errorf("cannot update scope identifier. expected %s, got %s", existing.ScopeId, proposed.ScopeId)
		}
	}

	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	// Validate any changes to the ValueOwner property.
	requiredSignatures := []string{}
	for _, p := range existing.Owners {
		requiredSignatures = append(requiredSignatures, p.Address)
	}
	if existing.ValueOwnerAddress != proposed.ValueOwnerAddress {
		// existing value is being changed,
		if len(existing.ValueOwnerAddress) > 0 {
			if k.AccountIsMarker(ctx, existing.ValueOwnerAddress) {
				if !k.HasSignerWithMarkerValueAuthority(ctx, existing.ValueOwnerAddress, signers, markertypes.Access_Withdraw) {
					return fmt.Errorf("missing signature for %s with authority to withdraw/remove existing value owner", existing.ValueOwnerAddress)
				}
			} else {
				// not a marker so require a signature from the existing value owner for this change.
				requiredSignatures = append(requiredSignatures, existing.ValueOwnerAddress)
			}
		}
		// check for a marker account because they have restrictions on adding scopes to them.
		if len(proposed.ValueOwnerAddress) > 0 {
			if k.AccountIsMarker(ctx, proposed.ValueOwnerAddress) {
				if !k.HasSignerWithMarkerValueAuthority(ctx, proposed.ValueOwnerAddress, signers, markertypes.Access_Deposit) {
					return fmt.Errorf("no signatures present with authority to add scope to marker %s", proposed.ValueOwnerAddress)
				}
			}
			// not a marker account, don't care who this new address is...
		}
	}

	// Signatures required of all existing data owners.
	for _, owner := range requiredSignatures {
		found := false
		for _, signer := range signers {
			if owner == signer {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing signature from existing owner %s; required for update", owner)
		}
	}

	return nil
}

// ValidateScopeRemove checks the current scope and the proposed removal scope to determine if the the proposed remove is valid
// based on the existing state
func (k Keeper) ValidateScopeRemove(ctx sdk.Context, existing, proposed types.Scope, signers []string) error {
	// IDs must match
	if len(existing.ScopeId) > 0 {
		if !proposed.ScopeId.Equals(existing.ScopeId) {
			return fmt.Errorf("cannot update scope identifier. expected %s, got %s", existing.ScopeId, proposed.ScopeId)
		}
	}

	// Signatures required of all existing data owners.
	for _, party := range existing.Owners {
		found := false
		for _, signer := range signers {
			if party.Address == signer {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing signature from existing owner %v; required for update", party)
		}
	}

	return nil
}
