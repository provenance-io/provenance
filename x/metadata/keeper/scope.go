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
	// TODO - add events here, deferred instrumentation for new scope?
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
	requiredSignatures := append([]string{}, existing.OwnerAddress...)
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
