package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) Update(ctx sdk.Context) {
	outstanding, err := k.GetOutstandingRewardPrograms(ctx)
	if err != nil {
		ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - error iterating reward programs: %v ", err))
		return
	}

	for _, rewardProgram := range outstanding {
		if rewardProgram.IsStarting(ctx) {
			k.StartRewardProgram(ctx, &rewardProgram)
		} else if rewardProgram.IsEndingClaimPeriod(ctx) {
			k.EndRewardProgramClaimPeriod(ctx, &rewardProgram)
		}

		k.SetRewardProgram(ctx, rewardProgram)
	}
}

func (k Keeper) Cleanup(ctx sdk.Context) {
	err := k.RemoveDeadShares(ctx)
	if err != nil {
		ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - error removing dead shares: %v ", err))
	}

	err = k.RemoveDeadPrograms(ctx)
	if err != nil {
		ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - error removing dead reward programs: %v ", err))
	}
}

func (k Keeper) StartRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Starting reward program: %v ", rewardProgram))
	blockTime := ctx.BlockTime()
	rewardProgram.State = types.RewardProgram_STARTED
	rewardProgram.ClaimPeriodEndTime = blockTime.Add(time.Duration(rewardProgram.ClaimPeriodSeconds) * time.Second)
	rewardProgram.CurrentClaimPeriod = 1

	claimPeriodReward := types.NewClaimPeriodRewardDistribution(
		rewardProgram.GetCurrentClaimPeriod(),
		rewardProgram.GetId(),
		sdk.Coin{},
		0,
		false,
	)
	k.SetClaimPeriodRewardDistribution(ctx, claimPeriodReward)
}

func (k Keeper) EndRewardProgramClaimPeriod(ctx sdk.Context, rewardProgram *types.RewardProgram) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Claim period end hit for reward program %v ", rewardProgram))
	blockTime := ctx.BlockTime()
	rewardProgram.CurrentClaimPeriod++

	// Update the RewardProgramBalance
	programBalance, err := k.GetRewardProgramBalance(ctx, rewardProgram.GetId())
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Missing RewardProgramBalance for RewardProgram %d ", rewardProgram.GetId()))
		//TODO How to handle this. This shouldn't happen
	}
	balance := programBalance.Min(rewardProgram.GetMaxRewardByAddress())
	programBalance.Balance = programBalance.GetBalance().Sub(balance)
	k.SetRewardProgramBalance(ctx, programBalance)

	// Update the ClaimPeriodRewardDistribution
	claimPeriodReward, err := k.GetClaimPeriodRewardDistribution(ctx, rewardProgram.GetCurrentClaimPeriod(), rewardProgram.GetId())
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Missing ClaimPeriodRewardDistribution for RewardProgram %d ", rewardProgram.GetId()))
		//TODO How to handle this. This shouldn't happe
	}
	claimPeriodReward.TotalRewardsPoolForClaimPeriod = balance
	k.SetClaimPeriodRewardDistribution(ctx, claimPeriodReward)

	if rewardProgram.IsEnding(ctx) && programBalance.IsEmpty() {
		k.EndRewardProgram(ctx, rewardProgram)
	} else {
		rewardProgram.ClaimPeriodEndTime = blockTime.Add(time.Duration(rewardProgram.ClaimPeriodSeconds) * time.Second)
	}
}

func (k Keeper) EndRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Ending reward program %v ", rewardProgram))
	blockTime := ctx.BlockTime()
	rewardProgram.State = types.RewardProgram_FINISHED
	rewardProgram.ActualProgramEndTime = blockTime
}
