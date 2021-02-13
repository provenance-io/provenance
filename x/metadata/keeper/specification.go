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
}
