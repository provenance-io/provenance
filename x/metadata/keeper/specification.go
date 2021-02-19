package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
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

// IterateScopeSpecs processes all scope specs using a given handler.
func (k Keeper) IterateScopeSpecs(ctx sdk.Context, handler func(specification types.ScopeSpecification) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.ScopeSpecificationPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scopeSpec types.ScopeSpecification
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &scopeSpec)
		if handler(scopeSpec) {
			break
		}
	}
	return nil
}

// IterateScopeSpecsForAddress processes all scope specs associated with an address using a given handler.
func (k Keeper) IterateScopeSpecsForAddress(ctx sdk.Context, address sdk.AccAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetAddressScopeSpecCacheIteratorPrefix(address)
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scopeSpecID types.MetadataAddress
		if err := scopeSpecID.Unmarshal(it.Key()[len(prefix):]); err != nil {
			return err
		}
		if handler(scopeSpecID) {
			break
		}
	}
	return nil
}

// IterateScopeSpecsForContractSpec processes all scope specs associated with a contract spec id using a given handler.
func (k Keeper) IterateScopeSpecsForContractSpec(ctx sdk.Context, contractSpecID types.MetadataAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetContractSpecScopeSpecCacheIteratorPrefix(contractSpecID)
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scopeSpecID types.MetadataAddress
		if err := scopeSpecID.Unmarshal(it.Key()[len(prefix):]); err != nil {
			return err
		}
		if handler(scopeSpecID) {
			break
		}
	}
	return nil
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

// SetScopeSpecification stores a scope specification in the module kv store.
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

// DeleteScopeSpecification deletes a scope specification from the module kv store.
func (k Keeper) DeleteScopeSpecification(ctx sdk.Context, id types.MetadataAddress) {
	store := ctx.KVStore(k.storeKey)

	scopeSpec, found := k.GetScopeSpecification(ctx, id)
	if !found || k.isScopeSpecUsed(ctx, id) {
		return
	}

	k.clearScopeSpecificationIndex(ctx, scopeSpec)

	store.Delete(id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeScopeSpecificationRemoved,
			sdk.NewAttribute(types.AttributeKeyScopeSpecID, scopeSpec.SpecificationId.String()),
			sdk.NewAttribute(types.AttributeKeyScopeSpec, scopeSpec.String()),
		),
	)
}

// indexScopeSpecification adds all desired indexes for a scope specification.
func (k Keeper) indexScopeSpecification(ctx sdk.Context, scopeSpec types.ScopeSpecification) {
	store := ctx.KVStore(k.storeKey)

	// Index all the scope spec owner addresses
	for _, a := range scopeSpec.OwnerAddresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Set(types.GetAddressScopeSpecCacheKey(addr, scopeSpec.SpecificationId), []byte{0x01})
		}
	}

	// Index all the session spec ids
	for _, contractSpecID := range scopeSpec.ContractSpecIds {
		store.Set(types.GetContractSpecScopeSpecCacheKey(contractSpecID, scopeSpec.SpecificationId), []byte{0x01})
	}
}

// clearScopeSpecificationIndex removes all indexes for the given scope spec.
// The provided scope spec must be one that is already stored (as opposed to a new one or updated version of one).
func (k Keeper) clearScopeSpecificationIndex(ctx sdk.Context, scopeSpec types.ScopeSpecification) {
	store := ctx.KVStore(k.storeKey)

	// Delete all owner address + scope spec entries
	for _, a := range scopeSpec.OwnerAddresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Delete(types.GetAddressScopeSpecCacheKey(addr, scopeSpec.SpecificationId))
		}
	}

	// Delete all contract spec + scope spec entries
	for _, contractSpecID := range scopeSpec.ContractSpecIds {
		store.Delete(types.GetContractSpecScopeSpecCacheKey(contractSpecID, scopeSpec.SpecificationId))
	}
}

// isScopeSpecUsed checks to see if a scope exists that is defined by this scope spec.
func (k Keeper) isScopeSpecUsed(ctx sdk.Context, id types.MetadataAddress) bool {
	scopeSpecReferenceFound := false
	err := k.IterateScopesForScopeSpec(ctx, id, func(scopeID types.MetadataAddress) (stop bool) {
		scopeSpecReferenceFound = true
		return true
	})
	// If there was an error, that means there was an entry, so return true.
	return err != nil || scopeSpecReferenceFound
}

// ValidateScopeSpecUpdate - full validation of a scope specification.
func (k Keeper) ValidateScopeSpecUpdate(ctx sdk.Context, existing, proposed types.ScopeSpecification, signers []string) error {
	// IDS must match
	if len(existing.SpecificationId) > 0 {
		if !proposed.SpecificationId.Equals(existing.SpecificationId) {
			return fmt.Errorf("cannot update scope spec identifier. expected %s, got %s",
				existing.SpecificationId, proposed.SpecificationId)
		}
	}

	// Must pass basic validation.
	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	// Signatures required of all existing data owners.
	if err := k.ValidateScopeSpecAllOwnersAreSigners(existing, signers); err != nil {
		return err
	}

	// Validate the proposed contract spec ids.
	for _, contractSpecID := range proposed.ContractSpecIds {
		contractSpec, found := k.GetGroupSpecification(ctx, contractSpecID)
		// Make sure that all contract spec ids are valid and exist
		if !found {
			return fmt.Errorf("no contract spec exists with id %s", contractSpecID)
		}
		// Also make sure that the parties in each contract spec are also in the scope spec.
		for _, contractSpecParty := range contractSpec.PartiesInvolved {
			found := false
			for _, scopeSpecParty := range proposed.PartiesInvolved {
				if contractSpecParty == scopeSpecParty {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("contract specification party involved missing from from scope specification parties involved: (%d) %s",
					contractSpecParty, contractSpecParty.String())
			}
		}
	}

	return nil
}

// ValidateScopeSpecAllOwnersAreSigners validates that all entries in the scopeSpec.OwnerAddresses list are contained in the provided signers list.
func (k Keeper) ValidateScopeSpecAllOwnersAreSigners(scopeSpec types.ScopeSpecification, signers []string) error {
	for _, owner := range scopeSpec.OwnerAddresses {
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
