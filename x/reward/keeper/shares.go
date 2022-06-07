package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// Removes a share
func (k Keeper) RemoveShare(ctx sdk.Context, rewardProgramID, epochID uint64, addr string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareKey(rewardProgramID, epochID, []byte(addr))
	if key == nil {
		return false
	}

	keyExists := store.Has(key)
	if keyExists {
		bz := store.Get(key)
		store.Delete(bz)
	}
	return keyExists
}

// Clean up the shared and expired shares
func (k Keeper) CleanupRewardProgramShares(ctx sdk.Context, rewardProgram *types.RewardProgram) (err error) {
	blockTime := ctx.BlockTime()
	var deadShares []types.Share

	// Find all the dead shares
	err = k.IterateRewardShares(ctx, rewardProgram.Id, func(share types.Share) (stop bool) {
		// Share has been claimed or expired we can remove it
		if share.Claimed || blockTime.After(share.ExpireTime) || blockTime.Equal(share.ExpireTime) {
			deadShares = append(deadShares, share)
		}
		return false
	})
	if err != nil {
		return err
	}

	// Remove all the dead shares
	for _, share := range deadShares {
		k.RemoveShare(ctx, share.RewardProgramId, share.EpochId, share.Address)
	}

	return nil
}

func (k Keeper) GetShare(ctx sdk.Context, rewardProgramID, epochID uint64, addr string) (share types.Share, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareKey(rewardProgramID, epochID, []byte(addr))
	bz := store.Get(key)
	if len(bz) == 0 {
		return share, nil
	}
	err = k.cdc.Unmarshal(bz, &share)
	return share, err
}

func (k Keeper) SetShare(ctx sdk.Context, share *types.Share) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(share)
	key := types.GetShareKey(share.GetRewardProgramId(), share.GetEpochId(), []byte(share.GetAddress()))
	store.Set(key, bz)
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

// Iterates over ALL the shares for a reward program
func (k Keeper) IterateRewardShares(ctx sdk.Context, rewardProgramID uint64, handle func(share types.Share) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardShareKeyPrefix(rewardProgramID))

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

// Iterates over ALL the shares for a reward program's epoch
func (k Keeper) IterateRewardEpochShares(ctx sdk.Context, rewardProgramID, epochID uint64, handle func(share types.Share) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardEpochShareKeyPrefix(rewardProgramID, epochID))

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
