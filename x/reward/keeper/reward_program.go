package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) CreateRewardProgram(ctx sdk.Context, rewardProgram types.RewardProgram) (err error) {
	err = rewardProgram.Validate()
	if err != nil {
		return err
	}

	if ctx.BlockTime().UTC().After(rewardProgram.ProgramStartTime.UTC()) {
		return fmt.Errorf("start time is before current block time %v : %v ", ctx.BlockTime().UTC(), rewardProgram.ProgramStartTime.UTC())
	}

	// error check done in reward Validate()
	acc, _ := sdk.AccAddressFromBech32(rewardProgram.DistributeFromAddress)
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, acc, types.ModuleName, sdk.NewCoins(rewardProgram.TotalRewardPool))
	if err != nil {
		return fmt.Errorf("unable to send coin to module reward pool: %s", err)
	}
	return nil
}

// SetRewardProgram sets the reward program in the keeper
func (k Keeper) SetRewardProgram(ctx sdk.Context, rewardProgram types.RewardProgram) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewardProgram)
	store.Set(types.GetRewardProgramKey(rewardProgram.Id), bz)
}

// Removes a reward program in the keeper
func (k Keeper) RemoveRewardProgram(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetRewardProgramKey(id)
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
}

// GetRewardProgram returns a RewardProgram by id
func (k Keeper) GetRewardProgram(ctx sdk.Context, id uint64) (rewardProgram types.RewardProgram, err error) {
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

// Gets all RewardPrograms that have not expired
func (k Keeper) GetOutstandingRewardPrograms(ctx sdk.Context) ([]types.RewardProgram, error) {
	return k.getRewardProgramByState(ctx, types.RewardProgram_PENDING, types.RewardProgram_STARTED)
}

// GetAllActiveRewardPrograms gets all RewardPrograms that have started
func (k Keeper) GetAllActiveRewardPrograms(ctx sdk.Context) ([]types.RewardProgram, error) {
	return k.getRewardProgramByState(ctx, types.RewardProgram_STARTED)
}

// GetAllCompletedRewardPrograms gets all completed the RewardPrograms
func (k Keeper) GetAllCompletedRewardPrograms(ctx sdk.Context) ([]types.RewardProgram, error) {
	return k.getRewardProgramByState(ctx, types.RewardProgram_FINISHED, types.RewardProgram_EXPIRED)
}

// GetAllPendingRewardPrograms gets all pending the RewardPrograms
func (k Keeper) GetAllPendingRewardPrograms(ctx sdk.Context) ([]types.RewardProgram, error) {
	return k.getRewardProgramByState(ctx, types.RewardProgram_PENDING)
}

// GetAllExpiredRewardPrograms gets all expired RewardPrograms
func (k Keeper) GetAllExpiredRewardPrograms(ctx sdk.Context) ([]types.RewardProgram, error) {
	return k.getRewardProgramByState(ctx, types.RewardProgram_EXPIRED)
}

// GetAllExpiredRewardPrograms gets all RewardPrograms that are not expired
func (k Keeper) GetUnexpiredRewardPrograms(ctx sdk.Context) ([]types.RewardProgram, error) {
	return k.getRewardProgramByState(ctx, types.RewardProgram_PENDING, types.RewardProgram_STARTED, types.RewardProgram_FINISHED)
}

// getRewardProgramByState gets rewards based on state
func (k Keeper) getRewardProgramByState(ctx sdk.Context, states ...types.RewardProgram_State) ([]types.RewardProgram, error) {
	var rewardPrograms []types.RewardProgram
	// get all the rewards programs by state
	err := k.IterateRewardPrograms(ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		for _, state := range states {
			if rewardProgram.GetState() == state {
				rewardPrograms = append(rewardPrograms, rewardProgram)
				break
			}
		}

		return false
	})
	return rewardPrograms, err
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

// GetRewardProgramID gets the highest rewardprogram ID
func (k Keeper) GetRewardProgramID(ctx sdk.Context) (rewardprogramID uint64, err error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.RewardProgramIDKey)
	if bz == nil {
		return 0, errors.New("initial rewardprogram ID hasn't been set")
	}

	rewardprogramID = types.GetRewardProgramIDFromBytes(bz)
	return rewardprogramID, nil
}

// setRewardProgramID sets the new rewardprogram ID to the store
func (k Keeper) setRewardProgramID(ctx sdk.Context, rewardprogramID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.RewardProgramIDKey, types.GetRewardProgramIDBytes(rewardprogramID))
}

// GetNextRewardProgramID returns the next available reward program ID and increments keeper with next reward program ID
func (k Keeper) GetNextRewardProgramID(ctx sdk.Context) (rewardProgramID uint64, err error) {
	rewardProgramID, err = k.GetRewardProgramID(ctx)
	if err == nil {
		k.setRewardProgramID(ctx, rewardProgramID+1)
	}
	return rewardProgramID, err
}

// RefundRemainingBalance returns the remaining pool balance to the reward program creator
func (k Keeper) RefundRemainingBalance(ctx sdk.Context, rewardProgram types.RewardProgram) error {
	acc, _ := sdk.AccAddressFromBech32(rewardProgram.DistributeFromAddress)
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acc, sdk.NewCoins(rewardProgram.RemainingPoolBalance))
	rewardProgram.RemainingPoolBalance = sdk.NewInt64Coin(rewardProgram.RemainingPoolBalance.GetDenom(), 0)
	k.SetRewardProgram(ctx, rewardProgram)
	return err
}
