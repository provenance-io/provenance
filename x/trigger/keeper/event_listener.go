package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

// SetEventListener Adds the trigger to the event listener store.
func (k Keeper) SetEventListener(ctx sdk.Context, trigger triggertypes.Trigger) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&trigger)
	event, _ := trigger.GetTriggerEventI()
	store.Set(triggertypes.GetEventListenerKey(event.GetEventPrefix(), event.GetEventOrder(), trigger.GetId()), bz)
}

// RemoveEventListener Removes the trigger from the event listener store.
func (k Keeper) RemoveEventListener(ctx sdk.Context, trigger triggertypes.Trigger) bool {
	store := ctx.KVStore(k.storeKey)
	event, _ := trigger.GetTriggerEventI()
	key := triggertypes.GetEventListenerKey(event.GetEventPrefix(), event.GetEventOrder(), trigger.GetId())
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
}

// GetEventListener Gets the event listener from the store.
func (k Keeper) GetEventListener(ctx sdk.Context, eventName string, order uint64, triggerID triggertypes.TriggerID) (trigger triggertypes.Trigger, err error) {
	store := ctx.KVStore(k.storeKey)
	key := triggertypes.GetEventListenerKey(eventName, order, triggerID)
	bz := store.Get(key)
	if len(bz) == 0 {
		return trigger, triggertypes.ErrEventNotFound
	}
	err = k.cdc.Unmarshal(bz, &trigger)
	return trigger, err
}

// IterateEventListeners Iterates through all the event listeners.
func (k Keeper) IterateEventListeners(ctx sdk.Context, eventName string, handle func(trigger triggertypes.Trigger) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, triggertypes.GetEventListenerPrefix(eventName))

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
