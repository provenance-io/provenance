package keeper

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

func (k Keeper) SetTrigger(ctx sdk.Context, trigger triggertypes.Trigger) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&trigger)
	store.Set(triggertypes.GetTriggerKey(trigger.GetId()), bz)
}

// RemoveTrigger Removes a trigger from the store
func (k Keeper) RemoveTrigger(ctx sdk.Context, id triggertypes.TriggerID) bool {
	store := ctx.KVStore(k.storeKey)
	key := triggertypes.GetTriggerKey(id)
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
}

// RemoveTrigger Gets a trigger by id
func (k Keeper) GetTrigger(ctx sdk.Context, id triggertypes.TriggerID) (trigger *triggertypes.Trigger, err error) {
	store := ctx.KVStore(k.storeKey)
	key := triggertypes.GetTriggerKey(id)
	bz := store.Get(key)
	if len(bz) == 0 {
		return trigger, triggertypes.ErrTriggerNotFound
	}
	err = k.cdc.Unmarshal(bz, trigger)
	return trigger, err
}

// IterateTriggers Iterates through all the triggers
func (k Keeper) IterateTriggers(ctx sdk.Context, handle func(trigger triggertypes.Trigger) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, triggertypes.TriggerKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := triggertypes.Trigger{}
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

func (k Keeper) GetTriggerID(ctx sdk.Context) (triggerID triggertypes.TriggerID) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(triggertypes.GetTriggerIDKey())
	if bz == nil {
		return 1
	}

	triggerID = triggertypes.GetTriggerIDFromBytes(bz)
	return triggerID
}

func (k Keeper) NewTriggerWithID(ctx sdk.Context, owner string, event triggertypes.Event, action *types.Any) (*triggertypes.Trigger, error) {
	id := k.getNextTriggerID(ctx)
	trigger := triggertypes.NewTrigger(id, owner, event, action)
	return &trigger, nil
}

func (k Keeper) setTriggerID(ctx sdk.Context, triggerID triggertypes.TriggerID) {
	store := ctx.KVStore(k.storeKey)
	bz := triggertypes.GetTriggerIDBytes(triggerID)
	store.Set(triggertypes.GetTriggerIDKey(), bz)
}

func (k Keeper) getNextTriggerID(ctx sdk.Context) (triggerID triggertypes.TriggerID) {
	triggerID = k.GetTriggerID(ctx)
	k.setTriggerID(ctx, triggerID+1)
	return
}
