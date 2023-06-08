package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/trigger/types"
)

// SetTrigger Sets the trigger in the store.
func (k Keeper) SetTrigger(ctx sdk.Context, trigger types.Trigger) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&trigger)
	store.Set(types.GetTriggerKey(trigger.GetId()), bz)
}

// RemoveTrigger Removes a trigger from the store.
func (k Keeper) RemoveTrigger(ctx sdk.Context, id types.TriggerID) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTriggerKey(id)
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
}

// GetTrigger Gets a trigger from the store by id.
func (k Keeper) GetTrigger(ctx sdk.Context, id types.TriggerID) (trigger types.Trigger, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTriggerKey(id)
	bz := store.Get(key)
	if len(bz) == 0 {
		return trigger, types.ErrTriggerNotFound
	}
	err = k.cdc.Unmarshal(bz, &trigger)
	return trigger, err
}

// IterateTriggers Iterates through all the triggers.
func (k Keeper) IterateTriggers(ctx sdk.Context, handle func(trigger types.Trigger) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.TriggerKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.Trigger{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		stop, err := handle(record)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// GetAllTriggers Gets all the triggers within the store.
func (k Keeper) GetAllTriggers(ctx sdk.Context) (triggers []types.Trigger, err error) {
	err = k.IterateTriggers(ctx, func(trigger types.Trigger) (stop bool, err error) {
		triggers = append(triggers, trigger)
		return false, nil
	})
	return
}

// NewTriggerWithID Creates a trigger with the latest ID.
func (k Keeper) NewTriggerWithID(ctx sdk.Context, owner string, event *codectypes.Any, actions []*codectypes.Any) types.Trigger {
	id := k.getNextTriggerID(ctx)
	trigger := types.NewTrigger(id, owner, event, actions)
	return trigger
}

// setTriggerID Sets the next trigger ID.
func (k Keeper) setTriggerID(ctx sdk.Context, triggerID types.TriggerID) {
	store := ctx.KVStore(k.storeKey)
	bz := types.GetTriggerIDBytes(triggerID)
	store.Set(types.GetNextTriggerIDKey(), bz)
}

// getNextTriggerID Gets the latest trigger ID and updates the next one.
func (k Keeper) getNextTriggerID(ctx sdk.Context) (triggerID types.TriggerID) {
	triggerID = k.getTriggerID(ctx)
	k.setTriggerID(ctx, triggerID+1)
	return
}

// getTriggerID Gets the latest trigger ID.
func (k Keeper) getTriggerID(ctx sdk.Context) (triggerID types.TriggerID) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetNextTriggerIDKey())
	if bz == nil {
		return 1
	}
	triggerID = types.GetTriggerIDFromBytes(bz)
	return triggerID
}
