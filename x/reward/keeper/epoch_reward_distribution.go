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
func (k Keeper) GetEpochRewardDistribution(ctx sdk.Context, epochId string, rewardId uint64) (epochRewardDistribution types.EpochRewardDistribution, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEpochRewardDistributionKey(epochId, fmt.Sprintf("%d", rewardId))
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

func (k Keeper) EpochRewardDistributionIsValid(epochReward *types.EpochRewardDistribution) bool {
	return epochReward.RewardProgramId != 0
}
