package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

func (k Keeper) SetEventListener(ctx sdk.Context, trigger triggertypes.Trigger) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&trigger)

	// This is where we would unmarshal into the interface type
	event := trigger.Event.GetCachedValue().(triggertypes.TriggerEventI)

	store.Set(triggertypes.GetEventRegistryKey(event.GetEventPrefix(), trigger.GetId()), bz)
}

// RemoveTrigger Removes a trigger from the store
func (k Keeper) RemoveEventListener(ctx sdk.Context, trigger triggertypes.Trigger) bool {
	store := ctx.KVStore(k.storeKey)

	// This is where we would unmarshal into the interface type
	event := trigger.Event.GetCachedValue().(triggertypes.TriggerEventI)

	key := triggertypes.GetEventRegistryKey(event.GetEventPrefix(), trigger.GetId())
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
}

// GetTrigger Gets a trigger by id
func (k Keeper) GetEventListener(ctx sdk.Context, eventName string, triggerID triggertypes.TriggerID) (trigger triggertypes.Trigger, err error) {
	store := ctx.KVStore(k.storeKey)
	key := triggertypes.GetEventRegistryKey(eventName, triggerID)
	bz := store.Get(key)
	if len(bz) == 0 {
		return trigger, triggertypes.ErrEventNotFound
	}
	err = k.cdc.Unmarshal(bz, &trigger)
	return trigger, err
}

// IterateTriggers Iterates through all the triggers
func (k Keeper) IterateEventListeners(ctx sdk.Context, eventName string, handle func(trigger triggertypes.Trigger) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, triggertypes.GetEventRegistryPrefix(eventName))

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
