package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// Removes a share
func (k Keeper) RemoveShare(ctx sdk.Context, rewardProgramID, subPeriod uint64, addr string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareKey(rewardProgramID, subPeriod, []byte(addr))
	if key == nil {
		return false
	}

	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
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
		k.RemoveShare(ctx, share.RewardProgramId, share.ClaimPeriodId, share.Address)
	}

	return nil
}

// Clean up the shared and expired shares
func (k Keeper) RemoveDeadShares(ctx sdk.Context) (err error) {
	err = k.IterateRewardPrograms(ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		err := k.CleanupRewardProgramShares(ctx, &rewardProgram)
		if err != nil {
			ctx.Logger().Info(fmt.Sprintf("NOTICE: RemoveDeadShares - error cleaning up shares for reward program %d: %v ", rewardProgram.Id, err))
		}
		return false
	})
	return err
}

func (k Keeper) GetShare(ctx sdk.Context, rewardProgramID, rewardClaimPeriod uint64, addr string) (share types.Share, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareKey(rewardProgramID, rewardClaimPeriod, []byte(addr))
	bz := store.Get(key)
	if len(bz) == 0 {
		return share, nil
	}
	err = k.cdc.Unmarshal(bz, &share)
	return share, err
}

func (k Keeper) GetRewardClaimPeriodShares(ctx sdk.Context, rewardProgramID, rewardClaimPeriod uint64) (shares []types.Share, err error) {
	err = k.IterateRewardClaimPeriodShares(ctx, rewardProgramID, rewardClaimPeriod, func(share types.Share) (stop bool) {
		shares = append(shares, share)
		return false
	})
	return shares, err
}

func (k Keeper) SetShare(ctx sdk.Context, share *types.Share) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(share)
	key := types.GetShareKey(share.GetRewardProgramId(), share.GetClaimPeriodId(), []byte(share.GetAddress()))
	store.Set(key, bz)
}

func (k Keeper) HasShares(ctx sdk.Context, rewardProgramID uint64) (bool, error) {
	hasShares := false
	err := k.IterateRewardShares(ctx, rewardProgramID, func(share types.Share) (stop bool) {
		hasShares = true
		return true
	})
	return hasShares, err
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

// Iterates over ALL the shares for a reward program's reward claim
func (k Keeper) IterateRewardClaimPeriodShares(ctx sdk.Context, rewardProgramID, rewardClaimPeriod uint64, handle func(share types.Share) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardSubPeriodShareKeyPrefix(rewardProgramID, rewardClaimPeriod))

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
