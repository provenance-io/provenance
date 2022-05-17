package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// SetEpochRewardDistribution sets the EpochRewardDistribution in the keeper
func (k Keeper) SetEpochRewardDistribution(ctx sdk.Context, epochRewardDistribution types.EpochRewardDistribution) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&epochRewardDistribution)
	store.Set(types.GetEpochRewardDistributionKey(epochRewardDistribution.EpochId, fmt.Sprintf("%d", epochRewardDistribution.RewardProgramId)), bz)
}

// GetEpochRewardDistribution returns a EpochRewardDistribution by epoch id and reward id
func (k Keeper) GetEpochRewardDistribution(ctx sdk.Context, epochID string, rewardID uint64) (epochRewardDistribution types.EpochRewardDistribution, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEpochRewardDistributionKey(epochID, fmt.Sprintf("%d", rewardID))
	bz := store.Get(key)
	if len(bz) == 0 {
		return epochRewardDistribution, nil
	}
	err = k.cdc.Unmarshal(bz, &epochRewardDistribution)
	return epochRewardDistribution, err
}

// IterateEpochRewardDistributions  iterates all epoch reward distributions with the given handler function.
func (k Keeper) IterateEpochRewardDistributions(ctx sdk.Context, handle func(epochRewardDistribution types.EpochRewardDistribution) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.EpochRewardDistributionKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.EpochRewardDistribution{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// Gets all the Epoch Reward Distributions
func (k Keeper) GetAllEpochRewardDistributions(sdkCtx sdk.Context) ([]types.EpochRewardDistribution, error) {
	var rewardDistributions []types.EpochRewardDistribution
	err := k.IterateEpochRewardDistributions(sdkCtx, func(rewardDistribution types.EpochRewardDistribution) (stop bool) {
		rewardDistributions = append(rewardDistributions, rewardDistribution)
		return false
	})
	if err != nil {
		return nil, err
	}
	return rewardDistributions, nil
}

// Checks if an Epoch Reward Distribution is valid
func (k Keeper) EpochRewardDistributionIsValid(epochReward *types.EpochRewardDistribution) bool {
	return epochReward.RewardProgramId != 0
}

// Removes an EpochRewardDistribution
func (k Keeper) RemoveEpochRewardDistribution(ctx sdk.Context, epochID string, rewardID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEpochRewardDistributionKey(epochID, fmt.Sprintf("%d", rewardID))
	bz := store.Get(key)
	keyExists := store.Has(bz)
	if keyExists {
		store.Delete(bz)
	}
	return keyExists
}
