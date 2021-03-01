package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// IterateContractSpecs processes all contract specs using a given handler.
func (k Keeper) IterateContractSpecs(ctx sdk.Context, handler func(specification types.ContractSpecification) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.ScopeSpecificationKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var contractSpec types.ContractSpecification
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &contractSpec)
		if handler(contractSpec) {
			break
		}
	}
	return nil
}

// IterateContractSpecsForAddress processes all contract specs associated with an address using a given handler.
func (k Keeper) IterateContractSpecsForAddress(ctx sdk.Context, address sdk.AccAddress, handler func(contractSpecID types.MetadataAddress) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetAddressContractSpecCacheIteratorPrefix(address)
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var contractSpecID types.MetadataAddress
		if err := contractSpecID.Unmarshal(it.Key()[len(prefix):]); err != nil {
			return err
		}
		if handler(contractSpecID) {
			break
		}
	}
	return nil
}

// GetContractSpecification returns the contract spec with the given id.
func (k Keeper) GetContractSpecification(ctx sdk.Context, contractSpecID types.MetadataAddress) (spec types.ContractSpecification, found bool) {
	if !contractSpecID.IsContractSpecificationAddress() {
		return spec, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(contractSpecID)
	if b == nil {
		return types.ContractSpecification{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &spec)
	return spec, true
}

// SetContractSpecification stores a contract specification in the module kv store.
func (k Keeper) SetContractSpecification(ctx sdk.Context, spec types.ContractSpecification) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&spec)

	eventType := types.EventTypeContractSpecificationCreated
	if store.Has(spec.SpecificationId) {
		if oldBytes := store.Get(spec.SpecificationId); oldBytes != nil {
			var oldSpec types.ContractSpecification
			if err := k.cdc.UnmarshalBinaryBare(oldBytes, &oldSpec); err == nil {
				eventType = types.EventTypeContractSpecificationUpdated
				k.clearContractSpecificationIndex(ctx, oldSpec)
			}
		}
	}

	store.Set(spec.SpecificationId, b)
	k.indexContractSpecification(ctx, spec)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyScopeSpecID, spec.SpecificationId.String()),
			sdk.NewAttribute(types.AttributeKeyScopeSpec, spec.String()),
		),
	)
}

// DeleteContractSpecification deletes a contract specification from the module kv store.
func (k Keeper) DeleteContractSpecification(ctx sdk.Context, contractSpecID types.MetadataAddress) {
	store := ctx.KVStore(k.storeKey)

	contractSpec, found := k.GetContractSpecification(ctx, contractSpecID)
	if !found || k.isContractSpecUsed(ctx, contractSpecID) {
		return
	}

	k.clearContractSpecificationIndex(ctx, contractSpec)
	store.Delete(contractSpecID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeContractSpecificationRemoved,
			sdk.NewAttribute(types.AttributeKeyScopeSpecID, contractSpec.SpecificationId.String()),
			sdk.NewAttribute(types.AttributeKeyScopeSpec, contractSpec.String()),
		),
	)
}

// indexContractSpecification adds all desired indexes for a contract specification.
func (k Keeper) indexContractSpecification(ctx sdk.Context, spec types.ContractSpecification) {
	store := ctx.KVStore(k.storeKey)

	// Index all the contract spec owner addresses
	for _, a := range spec.OwnerAddresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Set(types.GetAddressContractSpecCacheKey(addr, spec.SpecificationId), []byte{0x01})
		}
	}
}

// clearContractSpecificationIndex removes all indexes for the given contract spec.
// The provided contract spec must be one that is already stored (as opposed to a new one or updated version of one).
func (k Keeper) clearContractSpecificationIndex(ctx sdk.Context, spec types.ContractSpecification) {
	store := ctx.KVStore(k.storeKey)

	// Delete all owner address + contract spec entries
	for _, a := range spec.OwnerAddresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Delete(types.GetAddressContractSpecCacheKey(addr, spec.SpecificationId))
		}
	}
}

// isContractSpecUsed checks to see if a contract spec is referenced by anything else (e.g. scope spec or session)
func (k Keeper) isContractSpecUsed(ctx sdk.Context, contractSpecID types.MetadataAddress) bool {
	contractSpecReferenceFound := false
	err := k.IterateScopeSpecsForContractSpec(ctx, contractSpecID, func(scopeID types.MetadataAddress) (stop bool) {
		contractSpecReferenceFound = true
		return true
	})
	// If there was an error, that indicates there was probably at least one entry to iterate over.
	// So, to err on the side of caution, return true in that case.
	if err != nil || contractSpecReferenceFound {
		return true
	}

	// TODO: iterate over the sessions to look for one used by this contractSpecID.
	return false
}

func (k Keeper) ValidateContractSpecUpdate(ctx sdk.Context, existing, proposed types.ContractSpecification, signers []string) error {
	// IDS must match
	if len(existing.SpecificationId) > 0 {
		if !proposed.SpecificationId.Equals(existing.SpecificationId) {
			return fmt.Errorf("cannot update contract spec identifier. expected %s, got %s",
				existing.SpecificationId, proposed.SpecificationId)
		}
	}

	// Must pass basic validation.
	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	// Signatures required of all existing data owners.
	if err := k.ValidateAllOwnersAreSigners(existing.OwnerAddresses, signers); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)

	// Validate the proposed record spec ids.
	for _, recordSpecID := range proposed.RecordSpecIds {
		// Make sure that all record spec ids exist
		if !store.Has(recordSpecID) {
			return fmt.Errorf("no record spec exists with id %s", recordSpecID)
		}
	}

	return nil
}

// IterateScopeSpecs processes all scope specs using a given handler.
func (k Keeper) IterateScopeSpecs(ctx sdk.Context, handler func(specification types.ScopeSpecification) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.ScopeSpecificationKeyPrefix)
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
func (k Keeper) GetScopeSpecification(ctx sdk.Context, scopeSpecID types.MetadataAddress) (spec types.ScopeSpecification, found bool) {
	if !scopeSpecID.IsScopeSpecificationAddress() {
		return spec, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(scopeSpecID)
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

	eventType := types.EventTypeScopeSpecificationCreated
	if store.Has(spec.SpecificationId) {
		if oldBytes := store.Get(spec.SpecificationId); oldBytes != nil {
			var oldSpec types.ScopeSpecification
			if err := k.cdc.UnmarshalBinaryBare(oldBytes, &oldSpec); err == nil {
				eventType = types.EventTypeScopeSpecificationUpdated
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

// XXXXXXXXScopeSpecification removes a scope specification from the module kv store.
func (k Keeper) XXXXXXXXScopeSpecification(ctx sdk.Context, scopeSpecID types.MetadataAddress) {
	store := ctx.KVStore(k.storeKey)

	scopeSpec, found := k.GetScopeSpecification(ctx, scopeSpecID)
	if !found || k.isScopeSpecUsed(ctx, scopeSpecID) {
		return
	}

	k.clearScopeSpecificationIndex(ctx, scopeSpec)

	store.Delete(scopeSpecID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeScopeSpecificationRemoved,
			sdk.NewAttribute(types.AttributeKeyScopeSpecID, scopeSpec.SpecificationId.String()),
			sdk.NewAttribute(types.AttributeKeyScopeSpec, scopeSpec.String()),
		),
	)
}

// indexScopeSpecification adds all desired indexes for a scope specification.
func (k Keeper) indexScopeSpecification(ctx sdk.Context, spec types.ScopeSpecification) {
	store := ctx.KVStore(k.storeKey)

	// Index all the scope spec owner addresses
	for _, a := range spec.OwnerAddresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Set(types.GetAddressScopeSpecCacheKey(addr, spec.SpecificationId), []byte{0x01})
		}
	}

	// Index all the contract spec ids
	for _, contractSpecID := range spec.ContractSpecIds {
		store.Set(types.GetContractSpecScopeSpecCacheKey(contractSpecID, spec.SpecificationId), []byte{0x01})
	}
}

// clearScopeSpecificationIndex removes all indexes for the given scope spec.
// The provided scope spec must be one that is already stored (as opposed to a new one or updated version of one).
func (k Keeper) clearScopeSpecificationIndex(ctx sdk.Context, spec types.ScopeSpecification) {
	store := ctx.KVStore(k.storeKey)

	// Delete all owner address + scope spec entries
	for _, a := range spec.OwnerAddresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err == nil {
			store.Delete(types.GetAddressScopeSpecCacheKey(addr, spec.SpecificationId))
		}
	}

	// Delete all contract spec + scope spec entries
	for _, contractSpecID := range spec.ContractSpecIds {
		store.Delete(types.GetContractSpecScopeSpecCacheKey(contractSpecID, spec.SpecificationId))
	}
}

// isScopeSpecUsed checks to see if a scope exists that is defined by this scope spec.
func (k Keeper) isScopeSpecUsed(ctx sdk.Context, scopeSpecID types.MetadataAddress) bool {
	scopeSpecReferenceFound := false
	err := k.IterateScopesForScopeSpec(ctx, scopeSpecID, func(scopeID types.MetadataAddress) (stop bool) {
		scopeSpecReferenceFound = true
		return true
	})
	// If there was an error, that indicates there was probably at least one entry to iterate over.
	// So, to err on the side of caution, return true in that case.
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
	if err := k.ValidateAllOwnersAreSigners(existing.OwnerAddresses, signers); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)

	// Validate the proposed contract spec ids.
	for _, contractSpecID := range proposed.ContractSpecIds {
		// Make sure that all contract spec ids exist
		if !store.Has(contractSpecID) {
			return fmt.Errorf("no contract spec exists with id %s", contractSpecID)
		}
	}

	return nil
}

// ValidateAllOwnersAreSigners makes sure that all entries in the existingOwners list are contained in the signers list.
func (k Keeper) ValidateAllOwnersAreSigners(existingOwners []string, signers []string) error {
	for _, owner := range existingOwners {
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
