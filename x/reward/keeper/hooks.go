package keeper

import (
	"fmt"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochtypes "github.com/provenance-io/provenance/x/epoch/types"
	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber uint64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber uint64) error {
	// distribute logic goes here, i.e record the number of shares claimable in that epoch and the total rewards pool
	// also unlock the module account?
	ctx.Logger().Info(fmt.Sprintf("In epoch end for %s %d", epochIdentifier, epochNumber))
	rewardPrograms, err := k.GetAllActiveRewardsForEpoch(ctx, epochIdentifier, epochNumber)
	if err != nil {
		return err
	}

	// only rewards programs who are eligible will be iterated through here
	for _, rewardProgram := range rewardPrograms {
		epochRewardDistibutionForEpoch, err := k.GetEpochRewardDistribution(ctx, epochIdentifier, rewardProgram.Id)
		if err != nil {
			return err
		}
		// epoch reward distribution does it exist till the block has ended, highly unlikely but could happen
		if epochRewardDistibutionForEpoch.EpochId == "" {
			epochRewardDistibutionForEpoch.EpochId = epochIdentifier
			epochRewardDistibutionForEpoch.RewardProgramId = rewardProgram.Id
			epochRewardDistibutionForEpoch.TotalShares = 0
			epochRewardDistibutionForEpoch.EpochEnded = true
			k.EvaluateRules(ctx, epochNumber, rewardProgram, epochRewardDistibutionForEpoch)
			// TODO if shares are still 0 for epochRewardDistibutionForEpoch.TotalShares return all the rewards?
		} else {
			// end the epoch
			epochRewardDistibutionForEpoch.EpochEnded = true
			k.EvaluateRules(ctx, epochNumber, rewardProgram, epochRewardDistibutionForEpoch)
		}
	}

	return nil
}

func (k Keeper) GetAllActiveRewardsForEpoch(ctx sdk.Context, epochIdentifier string, epochNumber uint64) ([]types.RewardProgram, error) {
	var rewardPrograms []types.RewardProgram
	// get all the rewards programs
	err := k.IterateRewardPrograms(ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		// this is epoch that ended, and matches up with the reward program identifier
		// check if any of the events match with any of the reward program running
		// e.g start epoch,current epoch .. start epoch + number of epochs program runs for > current epoch
		// 1,1 .. 1+4 > 1
		// 1,2 .. 1+4 > 2
		// 1,3 .. 1+4 > 3
		// 1,4 .. 1+4 > 4
		if rewardProgram.EpochId == epochIdentifier && rewardProgram.StartEpoch+rewardProgram.NumberEpochs > epochNumber {
			rewardPrograms = append(rewardPrograms, rewardProgram)
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return rewardPrograms, nil
}

func (k Keeper) CheckActiveDelegations(ctx sdk.Context, address sdk.AccAddress) []stakingtypes.Delegation {
	return k.stakingKeeper.GetAllDelegatorDelegations(ctx, address)
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochtypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber uint64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber uint64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
