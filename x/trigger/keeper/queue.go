package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// QueueTrigger Adds an item to the end of the queue and increases the length
func (k Keeper) QueueTrigger(ctx sdk.Context, trigger types.Trigger) {
	item := types.NewQueuedTrigger(trigger, ctx.BlockTime(), uint64(ctx.BlockHeight()))
	length := k.GetQueueLength(ctx)
	index := k.GetQueueStartIndex(ctx)
	k.SetQueueItem(ctx, index+length, item)
	k.SetQueueLength(ctx, length+1)
}

// Peek Returns the next item to be dequeued
func (k Keeper) Peek(ctx sdk.Context) (types.QueuedTrigger, error) {
	if k.QueueIsEmpty(ctx) {
		// Throw error
		return types.QueuedTrigger{}, nil
	}

	index := k.GetQueueStartIndex(ctx)
	item, err := k.GetQueueItem(ctx, index)
	if err != nil {
		// Throw error
		return types.QueuedTrigger{}, err
	}

	return item, nil
}

// DequeueTrigger Removes the first trigger from the queue and updates the length and indices.
func (k Keeper) DequeueTrigger(ctx sdk.Context) error {
	if k.QueueIsEmpty(ctx) {
		// Throw error
		return nil
	}
	length := k.GetQueueLength(ctx)
	index := k.GetQueueStartIndex(ctx)

	k.RemoveQueueIndex(ctx, index)
	k.SetQueueStartIndex(ctx, index+1)
	k.SetQueueLength(ctx, length-1)
	return nil
}

// QueueIsEmpty Checks if the queue is empty
func (k Keeper) QueueIsEmpty(ctx sdk.Context) bool {
	return k.GetQueueLength(ctx) == 0
}

// GetQueueItem Gets an item from the queue's store
func (k Keeper) GetQueueItem(ctx sdk.Context, index uint64) (trigger types.QueuedTrigger, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetQueueKey(index)
	bz := store.Get(key)
	if len(bz) == 0 {
		return trigger, types.ErrQueueIndexNotFound
	}
	err = k.cdc.Unmarshal(bz, &trigger)
	return trigger, err
}

// SetQueueItem Sets an item in the queue's store
func (k Keeper) SetQueueItem(ctx sdk.Context, index uint64, item types.QueuedTrigger) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&item)
	store.Set(types.GetQueueKey(index), bz)
}

// RemoveQueueIndex Removes the queue's index from the store
func (k Keeper) RemoveQueueIndex(ctx sdk.Context, index uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetQueueKey(index)
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
}

// GetQueueLength Gets the starting index of the queue in the store
func (k Keeper) GetQueueStartIndex(ctx sdk.Context) (index uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetQueueStartIndexKey())
	if bz == nil {
		return 0
	}
	index = types.GetQueueIndexFromBytes(bz)
	return index
}

// SetQueueLength Sets the starting index of the queue in the store
func (k Keeper) SetQueueStartIndex(ctx sdk.Context, index uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := types.GetQueueIndexBytes(index)
	store.Set(types.GetQueueStartIndexKey(), bz)
}

// SetQueueLength Gets the length of the queue in the store
func (k Keeper) GetQueueLength(ctx sdk.Context) (length uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetQueueLengthKey())
	if bz == nil {
		return 0
	}
	length = types.GetQueueIndexFromBytes(bz)
	return length
}

// SetQueueLength Sets the length of the queue in the store
func (k Keeper) SetQueueLength(ctx sdk.Context, length uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := types.GetQueueIndexBytes(length)
	store.Set(types.GetQueueLengthKey(), bz)
}
