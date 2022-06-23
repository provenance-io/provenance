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
	rewardProgram.State = types.RewardProgram_STARTED
	k.StartRewardProgramClaimPeriod(ctx, rewardProgram)

}

func (k Keeper) StartRewardProgramClaimPeriod(ctx sdk.Context, rewardProgram *types.RewardProgram) {
	blockTime := ctx.BlockTime()
	rewardProgram.ClaimPeriodEndTime = blockTime.Add(time.Duration(rewardProgram.ClaimPeriodSeconds) * time.Second)
	rewardProgram.CurrentClaimPeriod++
	claim_period_amount := rewardProgram.GetTotalRewardPool().Amount.Quo(sdk.NewInt(int64(rewardProgram.GetClaimPeriods())))
	claim_period_pool := sdk.NewCoin(rewardProgram.GetTotalRewardPool().Denom, claim_period_amount)
	claimPeriodReward := types.NewClaimPeriodRewardDistribution(
		rewardProgram.GetCurrentClaimPeriod(),
		rewardProgram.GetId(),
		claim_period_pool,
		sdk.NewInt64Coin(claim_period_pool.Denom, 0),
		0,
		false,
	)
	k.SetClaimPeriodRewardDistribution(ctx, claimPeriodReward)
}

func (k Keeper) EndRewardProgramClaimPeriod(ctx sdk.Context, rewardProgram *types.RewardProgram) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Claim period end hit for reward program %v ", rewardProgram))

	programBalance, err := k.GetRewardProgramBalance(ctx, rewardProgram.GetId())
	if err != nil {
		//TODO How to handle this. This shouldn't happen unless we are in a bad state
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Missing RewardProgramBalance for RewardProgram %d ", rewardProgram.GetId()))
		return
	}

	claimPeriodReward, err := k.GetClaimPeriodRewardDistribution(ctx, rewardProgram.GetCurrentClaimPeriod(), rewardProgram.GetId())
	if err != nil {
		//TODO How to handle this. This shouldn't happen unless we are in a bad state
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Missing ClaimPeriodRewardDistribution for RewardProgram %d ", rewardProgram.GetId()))
		return
	}

	totalClaimPeriodRewards, err := k.SumRewardClaimPeriodRewards(ctx, programBalance.GetBalance(), claimPeriodReward)
	if err != nil {
		//TODO How to handle this. This shouldn't happen unless we are in a bad state
		ctx.Logger().Error(fmt.Sprintf("%v", err))
		return
	}

	// Update balances
	claimPeriodReward.TotalRewardsPoolForClaimPeriod = claimPeriodReward.TotalRewardsPoolForClaimPeriod.Add(totalClaimPeriodRewards)
	programBalance.Balance = programBalance.Balance.Sub(totalClaimPeriodRewards)
	k.SetClaimPeriodRewardDistribution(ctx, claimPeriodReward)
	k.SetRewardProgramBalance(ctx, programBalance)

	if rewardProgram.IsEnding(ctx) && programBalance.IsEmpty() {
		k.EndRewardProgram(ctx, rewardProgram)
	} else {
		k.StartRewardProgramClaimPeriod(ctx, rewardProgram)
	}
}

func (k Keeper) EndRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Ending reward program %v ", rewardProgram))
	blockTime := ctx.BlockTime()
	rewardProgram.State = types.RewardProgram_FINISHED
	rewardProgram.ActualProgramEndTime = blockTime
}

func (k Keeper) SumRewardClaimPeriodRewards(ctx sdk.Context, programBalance sdk.Coin, claimPeriodReward types.ClaimPeriodRewardDistribution) (sum sdk.Coin, err error) {
	totalShares := claimPeriodReward.GetTotalShares()
	claimRewardPool := claimPeriodReward.GetRewardsPool().Amount

	participants, err := k.GetRewardClaimPeriodShares(ctx, claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId())
	if err != nil {
		// TODO How to handle this. This shouldn't happen unless we are in a bad state
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Unable to get shares for reward program %d's claim period %d ", claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId()))
		return sum, fmt.Errorf("unable to get reward claim period shares for reward program %d and claim period %d", claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId())
	}

	if claimPeriodReward.GetTotalShares() == 0 {
		return sum, nil
	}

	for _, participant := range participants {
		percentage := sdk.NewInt(participant.GetAmount()).Quo(sdk.NewInt(totalShares))
		reward := sdk.NewCoin(claimPeriodReward.GetRewardsPool().Denom, percentage.Mul(claimRewardPool))

		if programBalance.IsLT(reward) {
			reward = programBalance
		}
		programBalance = programBalance.Sub(reward)
		sum = sum.Add(reward)
	}

	return sum, nil
}
