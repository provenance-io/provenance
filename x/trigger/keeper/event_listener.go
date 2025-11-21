package keeper

import (
	"crypto/sha256"
	"errors"
	"strings"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

// SetEventListener Adds the trigger to the event listener store.
func (k Keeper) SetEventListener(ctx sdk.Context, trigger triggertypes.Trigger) error {
	event, err := trigger.GetTriggerEventI()
	if err != nil {
		return err
	}
	eventHash := getEventNameHash(event.GetEventPrefix())
	key := collections.Join3(eventHash, event.GetEventOrder(), trigger.GetId())
	return k.EventListeners.Set(ctx, key)
}

// RemoveEventListener Removes the trigger from the event listener store.
func (k Keeper) RemoveEventListener(ctx sdk.Context, trigger triggertypes.Trigger) (bool, error) {
	event, err := trigger.GetTriggerEventI()
	if err != nil {
		return false, err
	}

	eventHash := getEventNameHash(event.GetEventPrefix())
	key := collections.Join3(eventHash, event.GetEventOrder(), trigger.GetId())

	exists, err := k.EventListeners.Has(ctx, key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	err = k.EventListeners.Remove(ctx, key)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetEventListener Gets the event listener from the store.
func (k Keeper) GetEventListener(ctx sdk.Context, eventName string, order uint64, triggerID triggertypes.TriggerID) (trigger triggertypes.Trigger, err error) {
	eventHash := getEventNameHash(eventName)
	key := collections.Join3(eventHash, order, triggerID)

	exists, err := k.EventListeners.Has(ctx, key)
	if err != nil {
		return triggertypes.Trigger{}, err
	}
	if !exists {
		return triggertypes.Trigger{}, triggertypes.ErrEventNotFound
	}

	// Fetch the actual trigger
	return k.GetTrigger(ctx, triggerID)
}

// IterateEventListeners Iterates through all the event listeners.
func (k Keeper) IterateEventListeners(ctx sdk.Context, eventName string, handle func(trigger triggertypes.Trigger) (stop bool, err error)) error {
	eventHash := getEventNameHash(eventName)

	rng := collections.NewPrefixedTripleRange[[]byte, uint64, uint64](eventHash)

	iter, err := k.EventListeners.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		triggerID := key.K3()
		trigger, err := k.GetTrigger(ctx, triggerID)
		if err != nil {
			if errors.Is(err, triggertypes.ErrTriggerNotFound) {
				continue
			}
			return err
		}
		stop, err := handle(trigger)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// HasEventListener checks if a specific event listener exists
func (k Keeper) HasEventListener(ctx sdk.Context, eventName string, order uint64, triggerID uint64) (bool, error) {
	eventHash := getEventNameHash(eventName)
	key := collections.Join3(eventHash, order, triggerID)
	return k.EventListeners.Has(ctx, key)
}

// RemoveAllEventListenersForTrigger removes all event listeners for a specific trigger
func (k Keeper) RemoveAllEventListenersForTrigger(ctx sdk.Context, triggerID uint64) error {
	trigger, err := k.GetTrigger(ctx, triggerID)
	if err != nil {
		return err
	}
	_, err = k.RemoveEventListener(ctx, trigger)
	return err
}

// GetEventListenerCount returns the number of event listeners for a specific event
func (k Keeper) GetEventListenerCount(ctx sdk.Context, eventName string) (uint64, error) {
	count := uint64(0)
	err := k.IterateEventListeners(ctx, eventName, func(trigger triggertypes.Trigger) (bool, error) { //nolint:revive // safe conversion
		count++
		return false, nil
	})
	return count, err
}

// getEventNameHash returns a 32-byte hash of the event name
func getEventNameHash(name string) []byte {
	eventName := strings.ToLower(strings.TrimSpace(name))
	if len(eventName) == 0 {
		panic("invalid event name: cannot be empty")
	}
	hash := sha256.Sum256([]byte(eventName))
	return hash[:]
}
