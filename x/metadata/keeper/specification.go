package keeper

import (
	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetGroupSpecification returns the record with the given id.
func (k Keeper) GetGroupSpecification(ctx sdk.Context, id types.MetadataAddress) (spec types.GroupSpecification, found bool) {
	if !id.IsGroupSpecificationAddress() {
		return spec, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id)
	if b == nil {
		return types.GroupSpecification{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &spec)
	return spec, true
}

// SetGroupSpecification stores a group specification in the module kv store.
func (k Keeper) SetGroupSpecification(ctx sdk.Context, spec types.GroupSpecification) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&spec)
	store.Set(spec.SpecificationId, b)
}

// GetScopeSpecification returns the record with the given id.
func (k Keeper) GetScopeSpecification(ctx sdk.Context, id types.MetadataAddress) (spec types.ScopeSpecification, found bool) {
	if !id.IsScopeSpecificationAddress() {
		return spec, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id)
	if b == nil {
		return types.ScopeSpecification{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &spec)
	return spec, true
}

// SetScopeSpecification stores a group specification in the module kv store.
func (k Keeper) SetScopeSpecification(ctx sdk.Context, spec types.ScopeSpecification) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&spec)
	store.Set(spec.SpecificationId, b)

	eventType := types.EventTypeScopeSpecificationCreated
	if store.Has(spec.SpecificationId) {
		if oldBytes := store.Get(spec.SpecificationId); oldBytes != nil {
			var oldSpec types.ScopeSpecification
			if err := k.cdc.UnmarshalBinaryBare(oldBytes, &oldSpec); err == nil {
				eventType = types.EventTypeScopeUpdated
				k.clearScopeSpecificationIndex(ctx, oldSpec)
			}
		}
	}

	store.Set(spec.SpecificationId, b)
	k.indexScopeSpecification(ctx, spec)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyScopeSpecID, spec.SpecificationId.String()),
			sdk.NewAttribute(types.AttributeKeyScopeSpec, spec.String()),
		),
	)
}

func (k Keeper) indexScopeSpecification(ctx sdk.Context, scopeSpec types.ScopeSpecification) {
	// TODO: Clean this up
	/*
	store := ctx.KVStore(k.storeKey)

	// Index all the scope spec owner addresses
	for _, a := range scopeSpec.OwnerAddresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Set(types.GetAddressScopeCacheKey(addr, scopeSpec.SpecificationId), []byte{0x01})
		}
	}

	// Index all the session spec ids
	for _, groupSpecId := range scopeSpec.GroupSpecIds {
		store.Set(types.GetScopeSpecScopeCacheKey(scopeSpec.SpecificationId, groupSpecId), []byte{0x01})
	}
	 */
}

func (k Keeper) clearScopeSpecificationIndex(ctx sdk.Context, scopeSpec types.ScopeSpecification) {
	// TODO: finish this up
	/*
	store := ctx.KVStore(k.storeKey)

	// Delete all scope spec owner address entries
	 */
}

func (k Keeper) isScopeSpecUsed(ctx sdk.Context, id types.MetadataAddress) bool {
	scopeSpecReferenceFound := false
	k.IterateScopesForScopeSpec(ctx, id, func(scopeID types.MetadataAddress) (stop bool) {
		scopeSpecReferenceFound = true
		return true
	})
	return scopeSpecReferenceFound
}
