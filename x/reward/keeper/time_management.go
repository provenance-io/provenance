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
	err := k.RemoveDeadPrograms(ctx)
	if err != nil {
		ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - error removing dead reward programs: %v ", err))
	}
}

func (k Keeper) StartRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	if rewardProgram == nil {
		ctx.Logger().Error("NOTICE: Attempting to start nil reward program")
		return fmt.Errorf("unable to start nil reward program")
	}

	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Starting reward program: %v ", rewardProgram))
	rewardProgram.State = types.RewardProgram_STARTED
	k.StartRewardProgramClaimPeriod(ctx, rewardProgram)

	return nil
}

func (k Keeper) StartRewardProgramClaimPeriod(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	if rewardProgram == nil {
		ctx.Logger().Error("NOTICE: Attempting to start reward program claim for nil reward program")
		return fmt.Errorf("unable to start reward program claim for nil reward program")
	}

	if rewardProgram.GetClaimPeriods() == 0 {
		ctx.Logger().Error("NOTICE: Attempting to start reward program claim with non positive claim periods")
		return fmt.Errorf("claim periods must be positive")
	}

	blockTime := ctx.BlockTime()
	rewardProgram.ClaimPeriodEndTime = blockTime.Add(time.Duration(rewardProgram.ClaimPeriodSeconds) * time.Second)
	rewardProgram.CurrentClaimPeriod++

	// Get the Claim Period Reward. It should not exceed program balance
	claimPeriodAmount := rewardProgram.GetTotalRewardPool().Amount.Quo(sdk.NewInt(int64(rewardProgram.GetClaimPeriods())))
	claimPeriodPool := sdk.NewCoin(rewardProgram.GetTotalRewardPool().Denom, claimPeriodAmount)
	programBalance, err := k.GetRewardProgramBalance(ctx, rewardProgram.GetId())
	if err != nil || programBalance.GetRewardProgramId() == 0 {
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Missing RewardProgramBalance for RewardProgram %d ", rewardProgram.GetId()))
		return err
	}
	if programBalance.GetBalance().IsLT(claimPeriodPool) {
		claimPeriodPool = programBalance.GetBalance()
	}

	claimPeriodReward := types.NewClaimPeriodRewardDistribution(
		rewardProgram.GetCurrentClaimPeriod(),
		rewardProgram.GetId(),
		claimPeriodPool,
		sdk.NewInt64Coin(claimPeriodPool.Denom, 0),
		0,
		false,
	)
	k.SetClaimPeriodRewardDistribution(ctx, claimPeriodReward)
	return nil
}

func (k Keeper) EndRewardProgramClaimPeriod(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Claim period end hit for reward program %v ", rewardProgram))

	if rewardProgram == nil {
		ctx.Logger().Error("NOTICE: EndRewardProgramClaimPeriod RewardProgram is nil")
		return fmt.Errorf("unable to end reward program claim period for nil reward program")
	}

	programBalance, err := k.GetRewardProgramBalance(ctx, rewardProgram.GetId())
	if err != nil || programBalance.GetRewardProgramId() == 0 {
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Missing RewardProgramBalance for RewardProgram %d ", rewardProgram.GetId()))
		return fmt.Errorf("a program balance does not exist for RewardProgram %d", rewardProgram.GetId())
	}

	claimPeriodReward, err := k.GetClaimPeriodRewardDistribution(ctx, rewardProgram.GetCurrentClaimPeriod(), rewardProgram.GetId())
	if err != nil || claimPeriodReward.GetClaimPeriodId() == 0 {
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Missing ClaimPeriodRewardDistribution for RewardProgram %d ", rewardProgram.GetId()))
		return fmt.Errorf("a ClaimPeriodRewardDistribution does not exist for RewardProgram %d with claim period %d", rewardProgram.GetId(), rewardProgram.GetId())
	}

	totalClaimPeriodRewards, err := k.CalculateRewardClaimPeriodRewards(ctx, rewardProgram.GetMaxRewardByAddress(), claimPeriodReward)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Unable to calculate claim period rewards for RewardProgram %d ", rewardProgram.GetId()))
		return err
	}

	// Update balances
	claimPeriodReward.TotalRewardsPoolForClaimPeriod = claimPeriodReward.TotalRewardsPoolForClaimPeriod.Add(totalClaimPeriodRewards)
	programBalance.Balance = programBalance.Balance.Sub(totalClaimPeriodRewards)
	k.SetClaimPeriodRewardDistribution(ctx, claimPeriodReward)
	k.SetRewardProgramBalance(ctx, programBalance)

	if rewardProgram.IsEnding(ctx, programBalance) {
		k.EndRewardProgram(ctx, rewardProgram)
	} else {
		k.StartRewardProgramClaimPeriod(ctx, rewardProgram)
	}

	return nil
}

func (k Keeper) EndRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	if rewardProgram == nil {
		ctx.Logger().Error("NOTICE: Attempting to end reward program for nil reward program")
		return fmt.Errorf("unable to end reward programfor nil reward program")
	}

	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Ending reward program %v ", rewardProgram))
	blockTime := ctx.BlockTime()
	rewardProgram.State = types.RewardProgram_FINISHED
	rewardProgram.ActualProgramEndTime = blockTime

	return nil
}

func (k Keeper) CalculateRewardClaimPeriodRewards(ctx sdk.Context, maxReward sdk.Coin, claimPeriodReward types.ClaimPeriodRewardDistribution) (sum sdk.Coin, err error) {
	sum = sdk.NewInt64Coin(claimPeriodReward.GetRewardsPool().Denom, 0)

	if maxReward.Denom != claimPeriodReward.RewardsPool.GetDenom() {
		ctx.Logger().Error(fmt.Sprintf("NOTICE: CalculateRewardClaimPeriodRewards denoms don't match %s %s", maxReward.Denom, claimPeriodReward.RewardsPool.GetDenom()))
		return sum, fmt.Errorf("ProgramBalance, MaxReward, and ClaimPeriodReward denoms must match")
	}

	participants, err := k.GetRewardClaimPeriodShares(ctx, claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId())
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("NOTICE: Unable to get shares for reward program %d's claim period %d ", claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId()))
		return sum, fmt.Errorf("unable to get reward claim period shares for reward program %d and claim period %d", claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId())
	}

	for _, participant := range participants {
		reward := k.CalculateParticipantReward(ctx, participant.GetAmount(), claimPeriodReward.GetTotalShares(), claimPeriodReward.GetRewardsPool())
		if maxReward.IsLT(reward) {
			reward = maxReward
		}

		sum = sum.Add(reward)
	}

	return sum, nil
}

func (k Keeper) CalculateParticipantReward(ctx sdk.Context, shares int64, totalShares int64, claimRewardPool sdk.Coin) sdk.Coin {
	numerator := sdk.NewDec(shares)
	denom := sdk.NewDec(totalShares)

	percentage := sdk.NewDec(0)
	if totalShares > 0 {
		percentage = numerator.Quo(denom)
	}

	pool := sdk.NewDec(claimRewardPool.Amount.Int64())
	return sdk.NewInt64Coin(claimRewardPool.Denom, pool.Mul(percentage).TruncateInt64())
}
