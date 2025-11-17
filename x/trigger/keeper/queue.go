package keeper

import (
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// QueueTrigger Creates a QueuedTrigger and Enqueues it
func (k Keeper) QueueTrigger(ctx sdk.Context, trigger types.Trigger) {
	item := types.NewQueuedTrigger(trigger, ctx.BlockTime().UTC(), uint64(ctx.BlockHeight())) //nolint:gosec // safe: block height always ≥ 0
	k.Enqueue(ctx, item)
}

// QueuePeek Returns the next item to be dequeued.
func (k Keeper) QueuePeek(ctx sdk.Context) *types.QueuedTrigger {
	isEmpty := k.QueueIsEmpty(ctx)

	if isEmpty {
		return nil
	}
	startIndex, err := k.QueueStartIndex.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil
		}
		return nil
	}
	item, err := k.Queue.Get(ctx, startIndex)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil
		}
		return nil
	}
	return &item
}

// Enqueue Adds an item to the end of the queue and adjusts the internal counters.
func (k Keeper) Enqueue(ctx sdk.Context, item types.QueuedTrigger) error {
	startIndex, err := k.QueueStartIndex.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	if errors.Is(err, collections.ErrNotFound) {
		startIndex = 0
	}

	length, err := k.QueueLength.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	if errors.Is(err, collections.ErrNotFound) {
		length = 0
	}

	// Calculate position for new item
	position := startIndex + length

	// Store the item
	if err := k.Queue.Set(ctx, position, item); err != nil {
		return err
	}

	// Update queue length
	return k.QueueLength.Set(ctx, length+1)
}

// Dequeue Removes the first item from the queue and updates the internal counters.
func (k Keeper) Dequeue(ctx sdk.Context) {
	if k.QueueIsEmpty(ctx) {
		panic("unable to dequeue from empty queue.")
	}

	// Get the start index
	startIndex, err := k.QueueStartIndex.Get(ctx)
	if err != nil {
		panic(err)
	}

	// Get the queue length
	length, err := k.QueueLength.Get(ctx)
	if err != nil {
		panic(err)
	}

	// Remove the first item
	if err := k.Queue.Remove(ctx, startIndex); err != nil {
		panic(err)
	}

	// Update queue metadata
	newStart := startIndex + 1
	newLength := length - 1

	if err := k.QueueStartIndex.Set(ctx, newStart); err != nil {
		panic(err)
	}

	if err := k.QueueLength.Set(ctx, newLength); err != nil {
		panic(err)
	}
}

// QueueIsEmpty Checks if the queue is empty.
func (k Keeper) QueueIsEmpty(ctx sdk.Context) bool {
	length, err := k.QueueLength.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return true
		}
		return false
	}
	return length == 0
}

// GetAllQueueItems Gets all the queue items within the store.
func (k Keeper) GetAllQueueItems(ctx sdk.Context) (items []types.QueuedTrigger, err error) {
	iter, err := k.Queue.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		_, err := iter.Key()
		if err != nil {
			return nil, err
		}
		value, err := iter.Value()
		if err != nil {
			return nil, err
		}

		items = append(items, value)
	}

	return items, nil
}

// getQueueItem Gets an item from the queue's store.
func (k Keeper) GetQueueItem(ctx sdk.Context, index uint64) (trigger types.QueuedTrigger, err error) {
	item, err := k.Queue.Get(ctx, index)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.QueuedTrigger{}, types.ErrQueueIndexNotFound
		}
		return types.QueuedTrigger{}, err
	}
	return item, nil
}

// setQueueItem Sets an item in the queue's store.
func (k Keeper) SetQueueItem(ctx sdk.Context, index uint64, item types.QueuedTrigger) error {
	return k.Queue.Set(ctx, index, item)
}

// removeQueueIndex Removes the queue's index from the store.
func (k Keeper) removeQueueIndex(ctx sdk.Context, index uint64) (bool, error) {
	exists, err := k.Queue.Has(ctx, index)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	err = k.Queue.Remove(ctx, index)
	if err != nil {
		return false, err
	}

	return true, nil
}

// iterateQueue Iterates through all the queue items.
func (k Keeper) iterateQueue(ctx sdk.Context, handle func(trigger types.QueuedTrigger) (stop bool, err error)) error {
	iter, err := k.Queue.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		value, err := iter.Value()
		if err != nil {
			return err
		}

		stop, err := handle(value)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}

	return nil
}

// getQueueStartIndex Gets the starting index of the queue in the store.
func (k Keeper) getQueueStartIndex(ctx sdk.Context) (uint64, error) {
	startIndex, err := k.QueueStartIndex.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return 0, nil // Default to 0 if not set
		}
		return 0, err
	}
	return startIndex, nil
}

// setQueueStartIndex Sets the starting index of the queue in the store.
func (k Keeper) setQueueStartIndex(ctx sdk.Context, index uint64) error {
	return k.QueueStartIndex.Set(ctx, index)
}

// getQueueLength Gets the length of the queue in the store.
func (k Keeper) GetQueueLength(ctx sdk.Context) (uint64, error) {
	length, err := k.QueueLength.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return 0, nil // Default to 0 if not set
		}
		return 0, err
	}
	return length, nil
}

// setQueueLength Sets the length of the queue in the store.
func (k Keeper) setQueueLength(ctx sdk.Context, length uint64) error {
	return k.QueueLength.Set(ctx, length)
}
