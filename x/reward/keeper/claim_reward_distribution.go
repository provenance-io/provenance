package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// SetClaimPeriodRewardDistribution sets the ClaimPeriodRewardDistribution in the keeper
func (k Keeper) SetClaimPeriodRewardDistribution(ctx sdk.Context, ClaimPeriodRewardDistribution types.ClaimPeriodRewardDistribution) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&ClaimPeriodRewardDistribution)
	store.Set(types.GetClaimPeriodRewardDistributionKey(ClaimPeriodRewardDistribution.ClaimPeriodId, ClaimPeriodRewardDistribution.RewardProgramId), bz)
}

// GetClaimPeriodRewardDistribution returns a ClaimPeriodRewardDistribution by epoch id and reward id
func (k Keeper) GetClaimPeriodRewardDistribution(ctx sdk.Context, claimPeriodId uint64, rewardID uint64) (ClaimPeriodRewardDistribution types.ClaimPeriodRewardDistribution, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetClaimPeriodRewardDistributionKey(claimPeriodId, rewardID)
	bz := store.Get(key)
	if len(bz) == 0 {
		return ClaimPeriodRewardDistribution, nil
	}
	err = k.cdc.Unmarshal(bz, &ClaimPeriodRewardDistribution)
	return ClaimPeriodRewardDistribution, err
}

// IterateClaimPeriodRewardDistributions  iterates all epoch reward distributions with the given handler function.
func (k Keeper) IterateClaimPeriodRewardDistributions(ctx sdk.Context, handle func(ClaimPeriodRewardDistribution types.ClaimPeriodRewardDistribution) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ClaimPeriodRewardDistributionKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.ClaimPeriodRewardDistribution{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// GetAllClaimPeriodRewardDistributions Gets all the Epoch Reward Distributions
func (k Keeper) GetAllClaimPeriodRewardDistributions(sdkCtx sdk.Context) ([]types.ClaimPeriodRewardDistribution, error) {
	var rewardDistributions []types.ClaimPeriodRewardDistribution
	err := k.IterateClaimPeriodRewardDistributions(sdkCtx, func(rewardDistribution types.ClaimPeriodRewardDistribution) (stop bool) {
		rewardDistributions = append(rewardDistributions, rewardDistribution)
		return false
	})
	if err != nil {
		return nil, err
	}
	return rewardDistributions, nil
}

// ClaimPeriodRewardDistributionIsValid Checks if an Epoch Reward Distribution is valid
func (k Keeper) ClaimPeriodRewardDistributionIsValid(ClaimPeriodReward *types.ClaimPeriodRewardDistribution) bool {
	return ClaimPeriodReward.RewardProgramId != 0
}

// RemoveClaimPeriodRewardDistribution Removes an ClaimPeriodRewardDistribution
func (k Keeper) RemoveClaimPeriodRewardDistribution(ctx sdk.Context, claimPeriodId uint64, rewardID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetClaimPeriodRewardDistributionKey(claimPeriodId, rewardID)
	bz := store.Get(key)
	keyExists := store.Has(bz)
	if keyExists {
		store.Delete(bz)
	}
	return keyExists
}
