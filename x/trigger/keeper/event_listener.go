package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"

	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

// SetEventListener Adds the trigger to the event listener store.
func (k Keeper) SetEventListener(ctx sdk.Context, trigger triggertypes.Trigger) {
	store := ctx.KVStore(k.storeKey)
	event, _ := trigger.GetTriggerEventI()
	store.Set(triggertypes.GetEventListenerKey(event.GetEventPrefix(), event.GetEventOrder(), trigger.GetId()), []byte{})
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
	if !store.Has(key) {
		return trigger, triggertypes.ErrEventNotFound
	}
	return k.GetTrigger(ctx, triggerID)
}

// IterateEventListeners Iterates through all the event listeners.
func (k Keeper) IterateEventListeners(ctx sdk.Context, eventName string, handle func(trigger triggertypes.Trigger) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, triggertypes.GetEventListenerPrefix(eventName))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		triggerID := binary.BigEndian.Uint64(iterator.Key()[41:49])
		record, err := k.GetTrigger(ctx, triggerID)
		if err != nil {
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
