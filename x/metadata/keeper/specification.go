package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// IterateRecordSpecs processes all record specs using a given handler.
func (k Keeper) IterateRecordSpecs(ctx sdk.Context, handler func(specification types.RecordSpecification) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.RecordSpecificationKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var recordSpec types.RecordSpecification
		err := k.cdc.UnmarshalBinaryBare(it.Value(), &recordSpec)
		if err != nil {
			return err
		}
		if handler(recordSpec) {
			break
		}
	}
	return nil
}

// IterateRecordSpecsForOwner processes all record specs owned by an address using a given handler.
func (k Keeper) IterateRecordSpecsForOwner(ctx sdk.Context, ownerAddress sdk.AccAddress, handler func(recordSpecID types.MetadataAddress) (stop bool)) error {
	var recordItErr error = nil
	contractItErr := k.IterateContractSpecsForOwner(ctx, ownerAddress, func(contractSpecID types.MetadataAddress) bool {
		needToStop := false
		recordItErr = k.IterateRecordSpecsForContractSpec(ctx, contractSpecID, func(recordSpecID types.MetadataAddress) bool {
			needToStop = handler(recordSpecID)
			return needToStop
		})
		if recordItErr != nil {
			return true
		}
		return needToStop
	})
	if recordItErr != nil {
		return recordItErr
	}
	if contractItErr != nil {
		return contractItErr
	}
	return nil
}

// IterateRecordSpecsForContractSpec processes all record specs for a contract spec using a given handler.
func (k Keeper) IterateRecordSpecsForContractSpec(ctx sdk.Context, contractSpecID types.MetadataAddress, handler func(recordSpecID types.MetadataAddress) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix, err := contractSpecID.ContractSpecRecordSpecIteratorPrefix()
	if err != nil {
		return err
	}
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var recordSpecID types.MetadataAddress
		if err := recordSpecID.Unmarshal(it.Key()); err != nil {
			return err
		}
		if handler(recordSpecID) {
			break
		}
	}
	return nil
}

// GetRecordSpecificationsForContractSpecificationID returns all the record specifications associated with given contractSpecID
func (k Keeper) GetRecordSpecificationsForContractSpecificationID(ctx sdk.Context, contractSpecID types.MetadataAddress) ([]*types.RecordSpecification, error) {
	retval := []*types.RecordSpecification{}
	err := k.IterateRecordSpecsForContractSpec(ctx, contractSpecID, func(recordSpecID types.MetadataAddress) bool {
		recordSpec, found := k.GetRecordSpecification(ctx, recordSpecID)
		if found {
			retval = append(retval, &recordSpec)
		} else {
			k.Logger(ctx).Error(fmt.Sprintf("iterator found record spec id %s but no record spec was found with that id", recordSpecID))
		}
		return false
	})
	return retval, err
}

// GetRecordSpecification returns the record specification with the given id.
func (k Keeper) GetRecordSpecification(ctx sdk.Context, recordSpecID types.MetadataAddress) (spec types.RecordSpecification, found bool) {
	if !recordSpecID.IsRecordSpecificationAddress() {
		return spec, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(recordSpecID)
	if b == nil {
		return types.RecordSpecification{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &spec)
	return spec, true
}

// SetRecordSpecification stores a record specification in the module kv store.
func (k Keeper) SetRecordSpecification(ctx sdk.Context, spec types.RecordSpecification) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&spec)

	var event proto.Message = types.NewEventRecordSpecificationCreated(spec.SpecificationId)
	action := types.TLActionCreated
	if store.Has(spec.SpecificationId) {
		event = types.NewEventRecordSpecificationUpdated(spec.SpecificationId)
		action = types.TLActionUpdated
	}

	store.Set(spec.SpecificationId, b)
	k.EmitEvent(ctx, event)
	defer types.GetIncObjFunc(types.TLTypeRecordSpec, action)
}

// RemoveRecordSpecification removes a record specification from the module kv store.
func (k Keeper) RemoveRecordSpecification(ctx sdk.Context, recordSpecID types.MetadataAddress) error {
	if k.isRecordSpecUsed(ctx, recordSpecID) {
		return fmt.Errorf("record specification with id %s still in use", recordSpecID)
	}

	store := ctx.KVStore(k.storeKey)

	if !store.Has(recordSpecID) {
		return fmt.Errorf("record specification with id %s not found", recordSpecID)
	}

	store.Delete(recordSpecID)
	k.EmitEvent(ctx, types.NewEventRecordSpecificationDeleted(recordSpecID))
	defer types.GetIncObjFunc(types.TLTypeRecordSpec, types.TLActionDeleted)
	return nil
}

// ValidateRecordSpecUpdate full validation of a proposed record spec possibly against an existing one.
func (k Keeper) ValidateRecordSpecUpdate(ctx sdk.Context, existing *types.RecordSpecification, proposed types.RecordSpecification) error {
	// Must pass basic validation.
	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	if existing != nil {
		// IDs must match
		if !proposed.SpecificationId.Equals(existing.SpecificationId) {
			return fmt.Errorf("cannot update record spec identifier. expected %s, got %s",
				existing.SpecificationId, proposed.SpecificationId)
		}
		// Names must match
		if proposed.Name != existing.Name {
			return fmt.Errorf("cannot update record spec name. expected %s, got %s",
				existing.Name, proposed.Name)
		}
	}

	return nil
}

func (k Keeper) isRecordSpecUsed(ctx sdk.Context, recordSpecID types.MetadataAddress) bool {
	// TODO: Check for records created from this spec.
	return false
}

// IterateContractSpecs processes all contract specs using a given handler.
func (k Keeper) IterateContractSpecs(ctx sdk.Context, handler func(specification types.ContractSpecification) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.ContractSpecificationKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var contractSpec types.ContractSpecification
		err := k.cdc.UnmarshalBinaryBare(it.Value(), &contractSpec)
		if err != nil {
			return err
		}
		if handler(contractSpec) {
			break
		}
	}
	return nil
}

// IterateContractSpecsForOwner processes all contract specs owned by an address using a given handler.
func (k Keeper) IterateContractSpecsForOwner(ctx sdk.Context, ownerAddress sdk.AccAddress, handler func(contractSpecID types.MetadataAddress) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetAddressContractSpecCacheIteratorPrefix(ownerAddress)
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

// GetContractSpecification returns the contract specification with the given id.
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

	var event proto.Message = types.NewEventContractSpecificationCreated(spec.SpecificationId)
	action := types.TLActionCreated
	if store.Has(spec.SpecificationId) {
		event = types.NewEventContractSpecificationUpdated(spec.SpecificationId)
		action = types.TLActionUpdated
		if oldBytes := store.Get(spec.SpecificationId); oldBytes != nil {
			var oldSpec types.ContractSpecification
			if err := k.cdc.UnmarshalBinaryBare(oldBytes, &oldSpec); err == nil {
				k.clearContractSpecificationIndex(ctx, oldSpec)
			}
		}
	}

	store.Set(spec.SpecificationId, b)
	k.indexContractSpecification(ctx, spec)
	k.EmitEvent(ctx, event)
	defer types.GetIncObjFunc(types.TLTypeContractSpec, action)
}

// RemoveContractSpecification removes a contract specification from the module kv store.
func (k Keeper) RemoveContractSpecification(ctx sdk.Context, contractSpecID types.MetadataAddress) error {
	if k.isContractSpecUsed(ctx, contractSpecID) {
		return fmt.Errorf("contract specification with id %s still in use", contractSpecID)
	}

	store := ctx.KVStore(k.storeKey)

	contractSpec, found := k.GetContractSpecification(ctx, contractSpecID)
	if !found {
		return fmt.Errorf("contract specification with id %s not found", contractSpecID)
	}

	k.clearContractSpecificationIndex(ctx, contractSpec)
	store.Delete(contractSpecID)
	k.EmitEvent(ctx, types.NewEventContractSpecificationDeleted(contractSpecID))
	defer types.GetIncObjFunc(types.TLTypeContractSpec, types.TLActionDeleted)
	return nil
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
	itScopeSpecErr := k.IterateScopeSpecsForContractSpec(ctx, contractSpecID, func(scopeID types.MetadataAddress) (stop bool) {
		contractSpecReferenceFound = true
		return true
	})
	// If there was an error, that indicates there was probably at least one entry to iterate over.
	// So, to err on the side of caution, return true in that case.
	if itScopeSpecErr != nil || contractSpecReferenceFound {
		return true
	}

	// TODO: Look for sessions used by this contractSpecID.

	// Look for a used record spec that is part of this contract spec
	hasUsedRecordSpec := false
	itRecSpecErr := k.IterateRecordSpecsForContractSpec(ctx, contractSpecID, func(recordSpecID types.MetadataAddress) bool {
		hasUsedRecordSpec = k.isRecordSpecUsed(ctx, recordSpecID)
		return hasUsedRecordSpec
	})

	return itRecSpecErr != nil || hasUsedRecordSpec
}

// ValidateContractSpecUpdate full validation of a proposed contract spec possibly against an existing one.
func (k Keeper) ValidateContractSpecUpdate(ctx sdk.Context, existing *types.ContractSpecification, proposed types.ContractSpecification) error {
	// IDS must match if there's an existing entry
	if existing != nil && !proposed.SpecificationId.Equals(existing.SpecificationId) {
		return fmt.Errorf("cannot update contract spec identifier. expected %s, got %s",
			existing.SpecificationId, proposed.SpecificationId)
	}

	// Must pass basic validation.
	if err := proposed.ValidateBasic(); err != nil {
		return err
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
		err := k.cdc.UnmarshalBinaryBare(it.Value(), &scopeSpec)
		if err != nil {
			return err
		}
		if handler(scopeSpec) {
			break
		}
	}
	return nil
}

// IterateScopeSpecsForOwner processes all scope specs owned by an address using a given handler.
func (k Keeper) IterateScopeSpecsForOwner(ctx sdk.Context, ownerAddress sdk.AccAddress, handler func(scopeSpecID types.MetadataAddress) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetAddressScopeSpecCacheIteratorPrefix(ownerAddress)
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

// GetScopeSpecification returns the scope specification with the given id.
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

	var event proto.Message = types.NewEventScopeSpecificationCreated(spec.SpecificationId)
	action := types.TLActionCreated
	if store.Has(spec.SpecificationId) {
		event = types.NewEventScopeSpecificationUpdated(spec.SpecificationId)
		action = types.TLActionUpdated
		if oldBytes := store.Get(spec.SpecificationId); oldBytes != nil {
			var oldSpec types.ScopeSpecification
			if err := k.cdc.UnmarshalBinaryBare(oldBytes, &oldSpec); err == nil {
				k.clearScopeSpecificationIndex(ctx, oldSpec)
			}
		}
	}

	store.Set(spec.SpecificationId, b)
	k.indexScopeSpecification(ctx, spec)
	k.EmitEvent(ctx, event)
	defer types.GetIncObjFunc(types.TLTypeScopeSpec, action)
}

// RemoveScopeSpecification removes a scope specification from the module kv store.
func (k Keeper) RemoveScopeSpecification(ctx sdk.Context, scopeSpecID types.MetadataAddress) error {
	if k.isScopeSpecUsed(ctx, scopeSpecID) {
		return fmt.Errorf("scope specification with id %s still in use", scopeSpecID)
	}

	store := ctx.KVStore(k.storeKey)

	scopeSpec, found := k.GetScopeSpecification(ctx, scopeSpecID)
	if !found {
		return fmt.Errorf("scope specification with id %s not found", scopeSpecID)
	}

	k.clearScopeSpecificationIndex(ctx, scopeSpec)
	store.Delete(scopeSpecID)
	k.EmitEvent(ctx, types.NewEventScopeSpecificationDeleted(scopeSpecID))
	defer types.GetIncObjFunc(types.TLTypeScopeSpec, types.TLActionDeleted)
	return nil
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
func (k Keeper) ValidateScopeSpecUpdate(ctx sdk.Context, existing *types.ScopeSpecification, proposed types.ScopeSpecification) error {
	// IDS must match if there's an existing entry
	if existing != nil && !proposed.SpecificationId.Equals(existing.SpecificationId) {
		return fmt.Errorf("cannot update scope spec identifier. expected %s, got %s",
			existing.SpecificationId, proposed.SpecificationId)
	}

	// Must pass basic validation.
	if err := proposed.ValidateBasic(); err != nil {
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
