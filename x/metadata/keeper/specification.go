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
		err := k.cdc.Unmarshal(it.Value(), &recordSpec)
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
	var recordItErr error
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
	k.cdc.MustUnmarshal(b, &spec)
	return spec, true
}

// SetRecordSpecification stores a record specification in the module kv store.
func (k Keeper) SetRecordSpecification(ctx sdk.Context, spec types.RecordSpecification) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&spec)

	var event proto.Message = types.NewEventRecordSpecificationCreated(spec.SpecificationId)
	action := types.TLAction_Created
	if store.Has(spec.SpecificationId) {
		event = types.NewEventRecordSpecificationUpdated(spec.SpecificationId)
		action = types.TLAction_Updated
	}

	store.Set(spec.SpecificationId, b)
	k.EmitEvent(ctx, event)
	defer types.GetIncObjFunc(types.TLType_RecordSpec, action)
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
	defer types.GetIncObjFunc(types.TLType_RecordSpec, types.TLAction_Deleted)
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
		err := k.cdc.Unmarshal(it.Value(), &contractSpec)
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
	k.cdc.MustUnmarshal(b, &spec)
	return spec, true
}

// SetContractSpecification stores a contract specification in the module kv store.
func (k Keeper) SetContractSpecification(ctx sdk.Context, spec types.ContractSpecification) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&spec)

	var oldSpec *types.ContractSpecification
	var event proto.Message = types.NewEventContractSpecificationCreated(spec.SpecificationId)
	action := types.TLAction_Created
	if store.Has(spec.SpecificationId) {
		event = types.NewEventContractSpecificationUpdated(spec.SpecificationId)
		action = types.TLAction_Updated
		if oldBytes := store.Get(spec.SpecificationId); oldBytes != nil {
			oldSpec = &types.ContractSpecification{}
			if err := k.cdc.Unmarshal(oldBytes, oldSpec); err != nil {
				k.Logger(ctx).Error("could not unmarshal old contract spec", "err", err, "specId", spec.SpecificationId.String(), "oldBytes", oldBytes)
				oldSpec = nil
			}
		}
	}

	store.Set(spec.SpecificationId, b)
	k.indexContractSpecification(ctx, &spec, oldSpec)
	k.EmitEvent(ctx, event)
	defer types.GetIncObjFunc(types.TLType_ContractSpec, action)
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

	k.indexContractSpecification(ctx, nil, &contractSpec)
	store.Delete(contractSpecID)
	k.EmitEvent(ctx, types.NewEventContractSpecificationDeleted(contractSpecID))
	defer types.GetIncObjFunc(types.TLType_ContractSpec, types.TLAction_Deleted)
	return nil
}

// contractSpecIndexValues is a struct containing the values used to index a contract specification.
type contractSpecIndexValues struct {
	SpecificationID types.MetadataAddress
	OwnerAddresses  []string
}

// getContractSpecIndexValues extracts the values used to index a contract specification.
func getContractSpecIndexValues(spec *types.ContractSpecification) *contractSpecIndexValues {
	if spec == nil {
		return nil
	}
	return &contractSpecIndexValues{
		SpecificationID: spec.SpecificationId,
		OwnerAddresses:  spec.OwnerAddresses,
	}
}

// getMissingContractSpecIndexValues extracts the index values in the required set that are not in the found set.
func getMissingContractSpecIndexValues(required, found *contractSpecIndexValues) contractSpecIndexValues {
	rv := contractSpecIndexValues{}
	if required == nil {
		return rv
	}
	if found == nil {
		return *required
	}
	rv.SpecificationID = required.SpecificationID
	rv.OwnerAddresses = FindMissing(required.OwnerAddresses, found.OwnerAddresses)
	return rv
}

// IndexKeys creates all of the index key byte arrays that this contractSpecIndexValues represents.
func (v contractSpecIndexValues) IndexKeys() [][]byte {
	rv := make([][]byte, 0)
	if v.SpecificationID.Empty() {
		return rv
	}
	for _, addrStr := range v.OwnerAddresses {
		if addr, err := sdk.AccAddressFromBech32(addrStr); err == nil {
			rv = append(rv, types.GetAddressContractSpecCacheKey(addr, v.SpecificationID))
		}
	}
	return rv
}

// indexContractSpecification updates the index entries for a contract specification.
//
// When adding a new contract spec:  indexContractSpecification(ctx, spec, nil)
//
// When deleting a contract spec:  indexContractSpecification(ctx, nil, spec)
//
// When updating a contract spec:  indexContractSpecification(ctx, newSpec, oldSpec)
//
// If both newSpec and oldSpec are not nil, it is assumed that they have the same SpecificationId.
// Failure to meet this assumption will result in strange and bad behavior.
func (k Keeper) indexContractSpecification(ctx sdk.Context, newSpec, oldSpec *types.ContractSpecification) {
	if newSpec == nil && oldSpec == nil {
		return
	}

	newSpecIndexValues := getContractSpecIndexValues(newSpec)
	oldSpecIndexValues := getContractSpecIndexValues(oldSpec)

	toAdd := getMissingContractSpecIndexValues(newSpecIndexValues, oldSpecIndexValues)
	toRemove := getMissingContractSpecIndexValues(oldSpecIndexValues, newSpecIndexValues)

	store := ctx.KVStore(k.storeKey)
	for _, indexKey := range toAdd.IndexKeys() {
		store.Set(indexKey, []byte{0x01})
	}
	for _, indexKey := range toRemove.IndexKeys() {
		store.Delete(indexKey)
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
		err := k.cdc.Unmarshal(it.Value(), &scopeSpec)
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
	k.cdc.MustUnmarshal(b, &spec)
	return spec, true
}

// SetScopeSpecification stores a scope specification in the module kv store.
func (k Keeper) SetScopeSpecification(ctx sdk.Context, spec types.ScopeSpecification) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&spec)

	var oldSpec *types.ScopeSpecification
	var event proto.Message = types.NewEventScopeSpecificationCreated(spec.SpecificationId)
	action := types.TLAction_Created
	if store.Has(spec.SpecificationId) {
		event = types.NewEventScopeSpecificationUpdated(spec.SpecificationId)
		action = types.TLAction_Updated
		if oldBytes := store.Get(spec.SpecificationId); oldBytes != nil {
			oldSpec = &types.ScopeSpecification{}
			if err := k.cdc.Unmarshal(oldBytes, oldSpec); err != nil {
				k.Logger(ctx).Error("could not unmarshal old scope spec", "err", err, "specId", spec.SpecificationId.String(), "oldBytes", oldBytes)
				oldSpec = nil
			}
		}
	}

	store.Set(spec.SpecificationId, b)
	k.indexScopeSpecification(ctx, &spec, oldSpec)
	k.EmitEvent(ctx, event)
	defer types.GetIncObjFunc(types.TLType_ScopeSpec, action)
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

	k.indexScopeSpecification(ctx, nil, &scopeSpec)
	store.Delete(scopeSpecID)
	k.EmitEvent(ctx, types.NewEventScopeSpecificationDeleted(scopeSpecID))
	defer types.GetIncObjFunc(types.TLType_ScopeSpec, types.TLAction_Deleted)
	return nil
}

// scopeSpecIndexValues is a struct containing the values used to index a scope specification.
type scopeSpecIndexValues struct {
	SpecificationID types.MetadataAddress
	OwnerAddresses  []string
	ContractSpecIDs []types.MetadataAddress
}

// getScopeSpecIndexValues extracts the values used to index a scope specification.
func getScopeSpecIndexValues(spec *types.ScopeSpecification) *scopeSpecIndexValues {
	if spec == nil {
		return nil
	}
	return &scopeSpecIndexValues{
		SpecificationID: spec.SpecificationId,
		OwnerAddresses:  spec.OwnerAddresses,
		ContractSpecIDs: spec.ContractSpecIds,
	}
}

// getMissingScopeSpecIndexValues extracts the index values in the required set that are not in the found set.
func getMissingScopeSpecIndexValues(required, found *scopeSpecIndexValues) scopeSpecIndexValues {
	rv := scopeSpecIndexValues{}
	if required == nil {
		return rv
	}
	if found == nil {
		return *required
	}
	rv.SpecificationID = required.SpecificationID
	rv.OwnerAddresses = FindMissing(required.OwnerAddresses, found.OwnerAddresses)
	rv.ContractSpecIDs = FindMissingMdAddr(required.ContractSpecIDs, found.ContractSpecIDs)
	return rv
}

// IndexKeys creates all of the index key byte arrays that this scopeSpecIndexValues represents.
func (v scopeSpecIndexValues) IndexKeys() [][]byte {
	rv := make([][]byte, 0)
	if v.SpecificationID.Empty() {
		return rv
	}
	for _, addrStr := range v.OwnerAddresses {
		if addr, err := sdk.AccAddressFromBech32(addrStr); err == nil {
			rv = append(rv, types.GetAddressScopeSpecCacheKey(addr, v.SpecificationID))
		}
	}
	for _, specID := range v.ContractSpecIDs {
		rv = append(rv, types.GetContractSpecScopeSpecCacheKey(specID, v.SpecificationID))
	}
	return rv
}

// indexScopeSpecification updates the index entries for a scope specification.
//
// When adding a new scope spec:  indexScopeSpecification(ctx, scope, nil)
//
// When deleting a scope spec:  indexScopeSpecification(ctx, nil, scope)
//
// When updating a scope spec:  indexScopeSpecification(ctx, newScope, oldScope)
//
// If both newSpec and oldSpec are not nil, it is assumed that they have the same SpecificationID.
// Failure to meet this assumption will result in strange and bad behavior.
func (k Keeper) indexScopeSpecification(ctx sdk.Context, newSpec, oldSpec *types.ScopeSpecification) {
	if newSpec == nil && oldSpec == nil {
		return
	}

	newSpecIndexValues := getScopeSpecIndexValues(newSpec)
	oldSpecIndexValues := getScopeSpecIndexValues(oldSpec)

	toAdd := getMissingScopeSpecIndexValues(newSpecIndexValues, oldSpecIndexValues)
	toRemove := getMissingScopeSpecIndexValues(oldSpecIndexValues, newSpecIndexValues)

	store := ctx.KVStore(k.storeKey)
	for _, indexKey := range toAdd.IndexKeys() {
		store.Set(indexKey, []byte{0x01})
	}
	for _, indexKey := range toRemove.IndexKeys() {
		store.Delete(indexKey)
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
