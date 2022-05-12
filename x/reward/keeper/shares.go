package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// Creates a share and generates a unique id that includes the reward program id
// The id will automatically be generated
func (k Keeper) AddShare(ctx sdk.Context, share *types.Share) {
	store := ctx.KVStore(k.storeKey)
	share.Id = k.generateShareId(ctx, share.GetRewardProgramId())
	bz := k.cdc.MustMarshal(share)
	key := types.GetShareKey(share.GetRewardProgramId(), share.GetId())
	store.Set(key, bz)
}

// Removes a share
func (k Keeper) RemoveShare(ctx sdk.Context, rewardProgramId uint64, shareId uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareKey(rewardProgramId, shareId)
	bz := store.Get(key)
	keyExists := store.Has(bz)
	if keyExists {
		store.Delete(bz)
	}
	return keyExists
}

func (k Keeper) GetShare(ctx sdk.Context, rewardProgramId uint64, shareId uint64) (share types.Share, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareKey(rewardProgramId, shareId)
	bz := store.Get(key)
	if len(bz) == 0 {
		return share, nil
	}
	err = k.cdc.Unmarshal(bz, &share)
	return share, err
}

// Iterates over all the shares that belong to a RewardProgram
func (k Keeper) IterateRewardProgramShares(ctx sdk.Context, rewardProgramId uint64, handle func(share types.Share) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardProgramShareKeyPrefix(rewardProgramId))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.Share{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// Iterates over ALL the shares
func (k Keeper) IterateShares(ctx sdk.Context, handle func(share types.Share) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ShareKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.Share{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// Helpers
// Generates a unique id for reward program's share
func (k Keeper) generateShareId(ctx sdk.Context, rewardProgramId uint64) uint64 {
	shareId := k.getShareId(ctx, rewardProgramId)
	k.incrementShareId(ctx, rewardProgramId)
	return shareId
}

// Gets the next id that is available for use.
// The same id will always be returned unless incrementShareId is called after
func (k Keeper) getShareId(ctx sdk.Context, rewardProgramId uint64) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareCounterKey(rewardProgramId)
	bz := store.Get(key)
	if len(bz) == 0 {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// Increments the unique id for a reward program share
func (k Keeper) incrementShareId(ctx sdk.Context, rewardProgramId uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareCounterKey(rewardProgramId)

	shareId := k.getShareId(ctx, rewardProgramId)

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(shareId))
	store.Set(key, bz)
}
