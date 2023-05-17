package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// QueueTrigger Adds an item to the end of the queue and adjusts the internal counters.
func (k Keeper) QueueTrigger(ctx sdk.Context, trigger types.Trigger) {
	item := types.NewQueuedTrigger(trigger, ctx.BlockTime(), uint64(ctx.BlockHeight()))
	length := k.getQueueLength(ctx)
	index := k.getQueueStartIndex(ctx)
	k.SetQueueItem(ctx, index+length, item)
	k.setQueueLength(ctx, length+1)
}

// QueuePeek Returns the next item to be dequeued.
func (k Keeper) QueuePeek(ctx sdk.Context) (item types.QueuedTrigger) {
	if k.QueueIsEmpty(ctx) {
		panic("unable to peek empty queue")
	}
	index := k.getQueueStartIndex(ctx)
	item = k.GetQueueItem(ctx, index)
	return
}

// DequeueTrigger Removes the first trigger from the queue and updates the internal counters.
func (k Keeper) DequeueTrigger(ctx sdk.Context) {
	if k.QueueIsEmpty(ctx) {
		panic("unable to dequeue from empty queue.")
	}
	length := k.getQueueLength(ctx)
	index := k.getQueueStartIndex(ctx)
	k.RemoveQueueIndex(ctx, index)
	k.setQueueStartIndex(ctx, index+1)
	k.setQueueLength(ctx, length-1)
}

// QueueIsEmpty Checks if the queue is empty.
func (k Keeper) QueueIsEmpty(ctx sdk.Context) bool {
	return k.getQueueLength(ctx) == 0
}

// GetQueueItem Gets an item from the queue's store.
func (k Keeper) GetQueueItem(ctx sdk.Context, index uint64) (trigger types.QueuedTrigger) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetQueueKey(index)
	bz := store.Get(key)
	if len(bz) == 0 {
		panic("queue index not found")
	}
	k.cdc.MustUnmarshal(bz, &trigger)
	return
}

// SetQueueItem Sets an item in the queue's store.
func (k Keeper) SetQueueItem(ctx sdk.Context, index uint64, item types.QueuedTrigger) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&item)
	store.Set(types.GetQueueKey(index), bz)
}

// RemoveQueueIndex Removes the queue's index from the store.
func (k Keeper) RemoveQueueIndex(ctx sdk.Context, index uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetQueueKey(index)
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
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
