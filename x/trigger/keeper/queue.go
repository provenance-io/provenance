package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// QueueTrigger Creates a QueuedTrigger and Enqueues it
func (k Keeper) QueueTrigger(ctx sdk.Context, trigger types.Trigger) {
	item := types.NewQueuedTrigger(trigger, ctx.BlockTime().UTC(), uint64(ctx.BlockHeight()))
	k.Enqueue(ctx, item)
}

// QueuePeek Returns the next item to be dequeued.
func (k Keeper) QueuePeek(ctx sdk.Context) *types.QueuedTrigger {
	if k.QueueIsEmpty(ctx) {
		return nil
	}
	index := k.getQueueStartIndex(ctx)
	item := k.getQueueItem(ctx, index)
	return &item
}

// Enqueue Adds an item to the end of the queue and adjusts the internal counters.
func (k Keeper) Enqueue(ctx sdk.Context, item types.QueuedTrigger) {
	length := k.getQueueLength(ctx)
	index := k.getQueueStartIndex(ctx)
	k.setQueueItem(ctx, index+length, item)
	k.setQueueLength(ctx, length+1)
}

// DequeueTrigger Removes the first item from the queue and updates the internal counters.
func (k Keeper) Dequeue(ctx sdk.Context) {
	if k.QueueIsEmpty(ctx) {
		panic("unable to dequeue from empty queue.")
	}
	length := k.getQueueLength(ctx)
	index := k.getQueueStartIndex(ctx)
	k.removeQueueIndex(ctx, index)
	k.setQueueStartIndex(ctx, index+1)
	k.setQueueLength(ctx, length-1)
}

// QueueIsEmpty Checks if the queue is empty.
func (k Keeper) QueueIsEmpty(ctx sdk.Context) bool {
	return k.getQueueLength(ctx) == 0
}

// GetAllQueueItems Gets all the queue items within the store.
func (k Keeper) GetAllQueueItems(ctx sdk.Context) (items []types.QueuedTrigger, err error) {
	err = k.iterateQueue(ctx, func(item types.QueuedTrigger) (stop bool, err error) {
		items = append(items, item)
		return false, nil
	})
	return
}

// getQueueItem Gets an item from the queue's store.
func (k Keeper) getQueueItem(ctx sdk.Context, index uint64) (trigger types.QueuedTrigger) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetQueueKey(index)
	bz := store.Get(key)
	if len(bz) == 0 {
		panic("queue index not found")
	}
	k.cdc.MustUnmarshal(bz, &trigger)
	return
}

// setQueueItem Sets an item in the queue's store.
func (k Keeper) setQueueItem(ctx sdk.Context, index uint64, item types.QueuedTrigger) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&item)
	store.Set(types.GetQueueKey(index), bz)
}

// removeQueueIndex Removes the queue's index from the store.
func (k Keeper) removeQueueIndex(ctx sdk.Context, index uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetQueueKey(index)
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
}

// iterateQueue Iterates through all the queue items.
func (k Keeper) iterateQueue(ctx sdk.Context, handle func(trigger types.QueuedTrigger) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.QueueKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.QueuedTrigger{}
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

// getQueueStartIndex Gets the starting index of the queue in the store.
func (k Keeper) getQueueStartIndex(ctx sdk.Context) (index uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetQueueStartIndexKey())
	if bz == nil {
		return 0
	}
	index = types.GetQueueIndexFromBytes(bz)
	return index
}

// setQueueStartIndex Sets the starting index of the queue in the store.
func (k Keeper) setQueueStartIndex(ctx sdk.Context, index uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := types.GetQueueIndexBytes(index)
	store.Set(types.GetQueueStartIndexKey(), bz)
}

// getQueueLength Gets the length of the queue in the store.
func (k Keeper) getQueueLength(ctx sdk.Context) (length uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetQueueLengthKey())
	if bz == nil {
		return 0
	}
	length = types.GetQueueIndexFromBytes(bz)
	return length
}

// setQueueLength Sets the length of the queue in the store.
func (k Keeper) setQueueLength(ctx sdk.Context, length uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := types.GetQueueIndexBytes(length)
	store.Set(types.GetQueueLengthKey(), bz)
}
