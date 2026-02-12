package keeper

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

// SetEventListener Adds the trigger to the event listener store.
func (k Keeper) SetEventListener(ctx sdk.Context, trigger triggertypes.Trigger) error {
	event, err := trigger.GetTriggerEventI()
	if err != nil {
		return fmt.Errorf("could not get trigger event for trigger %d: %w", trigger.Id, err)
	}

	eventHash, err := getEventNameHash(event.GetEventPrefix())
	if err != nil {
		return fmt.Errorf("could not compute event hash for trigger %d: %w", trigger.Id, err)
	}

	key := triggertypes.GetEventListenerKey(eventHash, event.GetEventOrder(), trigger.GetId())

	if err := k.EventListeners.Set(ctx, key); err != nil {
		return fmt.Errorf("failed to store event listener for trigger %d: %w", trigger.Id, err)
	}

	return nil
}

// RemoveEventListener Removes the trigger from the event listener store.
func (k Keeper) RemoveEventListener(ctx sdk.Context, trigger triggertypes.Trigger) (bool, error) {
	event, err := trigger.GetTriggerEventI()
	if err != nil {
		return false, err
	}

	eventHash, err := getEventNameHash(event.GetEventPrefix())
	if err != nil {
		return false, err
	}
	key := triggertypes.GetEventListenerKey(eventHash, event.GetEventOrder(), trigger.GetId())

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
	eventHash, err := getEventNameHash(eventName)
	if err != nil {
		return triggertypes.Trigger{}, err
	}
	key := triggertypes.GetEventListenerKey(eventHash, order, triggerID)
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
	eventHash, err := getEventNameHash(eventName)
	if err != nil {
		return err
	}
	rng := collections.NewPrefixedTripleRange[[]byte, uint64, uint64](eventHash)

	iter, err := k.EventListeners.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iter.Close() //nolint:errcheck // ignoring close error on iterator: not critical for this context.

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
	eventHash, err := getEventNameHash(eventName)
	if err != nil {
		return false, err
	}
	key := triggertypes.GetEventListenerKey(eventHash, order, triggerID)
	return k.EventListeners.Has(ctx, key)
}

// RemoveEventListenerForTriggerID removes all event listeners for a specific trigger
func (k Keeper) RemoveEventListenerForTriggerID(ctx sdk.Context, triggerID uint64) error {
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
func getEventNameHash(name string) ([]byte, error) {
	eventName := strings.ToLower(strings.TrimSpace(name))
	if len(eventName) == 0 {
		return nil, triggertypes.ErrEventNotFound
	}
	hash := sha256.Sum256([]byte(eventName))
	return hash[:], nil
}
