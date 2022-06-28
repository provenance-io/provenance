package keeper

import (
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
