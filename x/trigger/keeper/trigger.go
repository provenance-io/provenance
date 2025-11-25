package keeper

import (
	"errors"

	"cosmossdk.io/collections"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// SetTrigger Sets the trigger in the store.
func (k Keeper) SetTrigger(ctx sdk.Context, trigger types.Trigger) error {
	return k.TriggersMap.Set(ctx, trigger.GetId(), trigger)
}

// RemoveTrigger Removes a trigger from the store.
func (k Keeper) RemoveTrigger(ctx sdk.Context, id types.TriggerID) bool {
	exists, err := k.TriggersMap.Has(ctx, id)
	if err != nil {
		return false
	}
	if exists {
		if err := k.TriggersMap.Remove(ctx, id); err != nil {
			return false
		}
		return true
	}
	return false
}

// GetTrigger Gets a trigger from the store by id.
func (k Keeper) GetTrigger(ctx sdk.Context, id types.TriggerID) (trigger types.Trigger, err error) {
	trigger, err = k.TriggersMap.Get(ctx, id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.Trigger{}, types.ErrTriggerNotFound
		}
		return types.Trigger{}, err
	}
	return trigger, nil
}

// IterateTriggers Iterates through all the triggers.
func (k Keeper) IterateTriggers(ctx sdk.Context, handle func(trigger types.Trigger) (stop bool, err error)) error {
	iterator, err := k.TriggersMap.Iterate(ctx, nil) // Already scoped to trigger prefix
	if err != nil {
		return err
	}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			return err
		}
		stop, err := handle(kv.Value)
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

// HasTrigger check trigger id exists.
func (k Keeper) HasTrigger(ctx sdk.Context, id uint64) (bool, error) {
	_, err := k.TriggersMap.Get(ctx, id)
	if errors.Is(err, collections.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// NewTriggerWithID Creates a trigger with the latest ID.
func (k Keeper) NewTriggerWithID(ctx sdk.Context, owner string, event *codectypes.Any, actions []*codectypes.Any) (types.Trigger, error) {
	currentID, err := k.NextTriggerID.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return types.Trigger{}, err
	}

	if errors.Is(err, collections.ErrNotFound) {
		currentID = 0
	}

	nextID := currentID + 1

	if err := k.NextTriggerID.Set(ctx, nextID); err != nil {
		return types.Trigger{}, err
	}

	trigger := types.Trigger{
		Id:      nextID,
		Owner:   owner,
		Event:   event,
		Actions: actions,
	}
	return trigger, nil
}

// setTriggerID Sets the next trigger ID.
func (k Keeper) setTriggerID(ctx sdk.Context, triggerID types.TriggerID) error {
	return k.NextTriggerID.Set(ctx, triggerID)
}

// getNextTriggerID Gets the latest trigger ID and updates the next one.
func (k Keeper) GetNextTriggerID(ctx sdk.Context) (triggerID types.TriggerID, err error) {
	currentID, err := k.NextTriggerID.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return 1, nil // First ID will be 1
		}
		return 0, err
	}
	return currentID + 1, nil
}

// getTriggerID Gets the latest trigger ID.
func (k Keeper) getTriggerID(ctx sdk.Context) (triggerID types.TriggerID, err error) {
	currentID, err := k.NextTriggerID.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return currentID, nil
}
