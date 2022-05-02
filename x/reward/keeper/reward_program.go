package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// SetRewardProgram sets the reward program in the keeper
func (k Keeper) SetRewardProgram(ctx sdk.Context, rewardProgram types.RewardProgram) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewardProgram)
	store.Set(types.GetRewardProgramKey(int64(rewardProgram.Id)), bz)
}

// GetRewardProgram returns a RewardProgram by id
func (k Keeper) GetRewardProgram(ctx sdk.Context, id int64) (rewardProgram types.RewardProgram, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetRewardProgramKey(id)
	bz := store.Get(key)
	if len(bz) == 0 {
		return rewardProgram, nil
	}
	err = k.cdc.Unmarshal(bz, &rewardProgram)
	return rewardProgram, err
}

// IterateRewardPrograms iterates all reward programs with the given handler function.
func (k Keeper) IterateRewardPrograms(ctx sdk.Context, handle func(rewardProgram types.RewardProgram) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RewardProgramKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.RewardProgram{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// Gets all the RewardPrograms
func (k Keeper) GetAllRewardPrograms(ctx sdk.Context) ([]types.RewardProgram, error) {
	var rewardPrograms []types.RewardProgram
	err := k.IterateRewardPrograms(ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		rewardPrograms = append(rewardPrograms, rewardProgram)
		return false
	})
	if err != nil {
		return nil, err
	}
	return rewardPrograms, nil
}

// Check if a RewardProgram is valid
func (k Keeper) RewardProgramIsValid(rewardProgram *types.RewardProgram) bool {
	return rewardProgram.Id != 0
}
