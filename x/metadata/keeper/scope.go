package keeper

import (
	"fmt"

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

// IterateGroups processes all stored scopes with the given handler.
func (k Keeper) IterateGroups(ctx sdk.Context, groupID types.MetadataAddress, handler func(types.RecordGroup) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix, err := groupID.ScopeGroupIteratorPrefix()
	if err != nil {
		return err
	}
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var group types.RecordGroup
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &group)
		if handler(group) {
			break
		}
	}
	return nil
}

// IterateRecords processes all records in a scope with the given handler.
func (k Keeper) IterateRecords(ctx sdk.Context, scopeID types.MetadataAddress, handler func(types.Record) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix, err := scopeID.ScopeGroupIteratorPrefix()
	if err != nil {
		return err
	}
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var record types.Record
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &record)
		if handler(record) {
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

	// Remove all record groups
	prefix, err := id.ScopeGroupIteratorPrefix()
	if err != nil {
		panic(err)
	}
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.RemoveRecordGroup(ctx, types.MetadataAddress(iter.Key()))
	}

	// Remove all records
	prefix, err = id.ScopeRecordIteratorPrefix()
	if err != nil {
		panic(err)
	}
	iter = sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.RemoveRecord(ctx, types.MetadataAddress(iter.Key()))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeScopeRemoved,
			sdk.NewAttribute(types.AttributeKeyScopeID, id.String()),
		),
	)

	store.Delete(id)
}

// GetRecordGroup returns the scope with the given id.
func (k Keeper) GetRecordGroup(ctx sdk.Context, id types.MetadataAddress) (group types.RecordGroup, found bool) {
	if !id.IsGroupAddress() {
		return group, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id.Bytes())
	if b == nil {
		return types.RecordGroup{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &group)
	return group, true
}

// SetRecordGroup stores a group in the module kv store.
func (k Keeper) SetRecordGroup(ctx sdk.Context, group types.RecordGroup) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&group)
	eventType := types.EventTypeGroupCreated

	if store.Has(group.GroupId) {
		eventType = types.EventTypeGroupUpdated
	}

	store.Set(group.GroupId, b)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyGroupID, group.GroupId.String()),
		),
	)
}

// RemoveRecordGroup removes a scope from the module kv store.
func (k Keeper) RemoveRecordGroup(ctx sdk.Context, id types.MetadataAddress) {
	if !id.IsGroupAddress() {
		panic(fmt.Errorf("invalid address, address must be for a record group"))
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGroupRemoved,
			sdk.NewAttribute(types.AttributeKeyGroupID, id.String()),
		),
	)
}

// GetRecord returns the record with the given id.
func (k Keeper) GetRecord(ctx sdk.Context, id types.MetadataAddress) (group types.Record, found bool) {
	if !id.IsRecordAddress() {
		return group, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id)
	if b == nil {
		return types.Record{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &group)
	return group, true
}

// SetRecord stores a group in the module kv store.
func (k Keeper) SetRecord(ctx sdk.Context, record types.Record) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&record)
	eventType := types.EventTypeRecordCreated

	recordID := record.GroupId.GetRecordAddress(record.Name)
	if store.Has(recordID) {
		eventType = types.EventTypeRecordUpdated
	}

	store.Set(recordID, b)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyScopeID, recordID.String()),
		),
	)
}

// RemoveRecord removes a scope from the module kv store.
func (k Keeper) RemoveRecord(ctx sdk.Context, id types.MetadataAddress) {
	if !id.IsRecordAddress() {
		panic(fmt.Errorf("invalid address, address must be for a record"))
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecordRemoved,
			sdk.NewAttribute(types.AttributeKeyScopeID, id.String()),
		),
	)
}

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
