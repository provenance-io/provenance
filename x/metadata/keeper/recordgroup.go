package keeper

import (
	"fmt"

	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
