package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// UpdateUnexpiredRewardsProgram called from begin blocker, starts/ends or expires rewards programs.
func (k Keeper) UpdateUnexpiredRewardsProgram(origCtx sdk.Context) {
	rewardPrograms, err := k.GetAllUnexpiredRewardPrograms(origCtx)
	if err != nil {
		origCtx.Logger().Error(fmt.Sprintf("error iterating reward programs: %v ", err))
		// called from the beginblocker, not much we can do here but return
		return
	}

	for index := range rewardPrograms {
		// Because this is designed for the BeginBlocker, we don't always have auto-rollback.
		// We don't partial state recorded if an error is encountered in the middle.
		// So use a cache context and only write it if there wasn't an error.
		ctx, writeCtx := origCtx.CacheContext()
		switch {
		case rewardPrograms[index].IsStarting(ctx):
			err = k.StartRewardProgram(ctx, &rewardPrograms[index])
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("cannot start program because of error %v ", err))
				continue
			}
		case rewardPrograms[index].IsEndingClaimPeriod(ctx):
			err = k.EndRewardProgramClaimPeriod(ctx, &rewardPrograms[index])
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("cannot end reward program claim period because of error %v ", err))
				continue
			}
		case rewardPrograms[index].IsExpiring(ctx):
			err = k.ExpireRewardProgram(ctx, &rewardPrograms[index])
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("cannot expire reward program because of error %v ", err))
				continue
			}
		}
		k.SetRewardProgram(ctx, rewardPrograms[index])
		writeCtx()
	}
}

// StartRewardProgram transition reward program to STARTED
func (k Keeper) StartRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	if rewardProgram == nil {
		ctx.Logger().Error("Attempting to start nil reward program")
		return fmt.Errorf("unable to start nil reward program")
	}

	if rewardProgram.GetTotalRewardPool().IsZero() {
		ctx.Logger().Error("Attempting to start reward program with no balance")
		return fmt.Errorf("unable to start reward program with no balance")
	}

	ctx.Logger().Info(fmt.Sprintf("Starting reward program: %v ", rewardProgram))
	rewardProgram.State = types.RewardProgram_STATE_STARTED
	err := k.StartRewardProgramClaimPeriod(ctx, rewardProgram)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardProgramStarted,
			sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", rewardProgram.GetId())),
		),
	)

	return nil
}

// StartRewardProgramClaimPeriod Start ClaimPeriod on a given reward program.
func (k Keeper) StartRewardProgramClaimPeriod(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	if rewardProgram == nil {
		ctx.Logger().Error("Attempting to start reward program claim for nil reward program")
		return fmt.Errorf("unable to start reward program claim for nil reward program")
	}

	if rewardProgram.GetClaimPeriods() == 0 {
		ctx.Logger().Error("Attempting to start reward program claim with non positive claim periods")
		return fmt.Errorf("claim periods must be positive")
	}

	blockTime := ctx.BlockTime()
	rewardProgram.ClaimPeriodEndTime = blockTime.Add(time.Duration(rewardProgram.ClaimPeriodSeconds) * time.Second)
	rewardProgram.CurrentClaimPeriod++
	if rewardProgram.CurrentClaimPeriod > rewardProgram.ClaimPeriods {
		rewardProgram.ExpectedProgramEndTime = rewardProgram.ExpectedProgramEndTime.Add(time.Duration(rewardProgram.ClaimPeriodSeconds) * time.Second)
	}

	// Get the Claim Period Reward. It should not exceed program balance
	claimPeriodAmount := rewardProgram.GetTotalRewardPool().Amount.Quo(sdk.NewInt(int64(rewardProgram.GetClaimPeriods())))
	claimPeriodPool := sdk.NewCoin(rewardProgram.GetTotalRewardPool().Denom, claimPeriodAmount)
	if rewardProgram.RemainingPoolBalance.IsLT(claimPeriodPool) {
		claimPeriodPool = rewardProgram.RemainingPoolBalance
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

// EndRewardProgramClaimPeriod end the claim period of a given reward program.
func (k Keeper) EndRewardProgramClaimPeriod(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	ctx.Logger().Info(fmt.Sprintf("BeginBlocker - Claim period end hit for reward program %v ", rewardProgram))

	if rewardProgram == nil {
		ctx.Logger().Error("EndRewardProgramClaimPeriod RewardProgram is nil")
		return fmt.Errorf("unable to end reward program claim period for nil reward program")
	}

	claimPeriodReward, err := k.GetClaimPeriodRewardDistribution(ctx, rewardProgram.GetCurrentClaimPeriod(), rewardProgram.GetId())
	if err != nil || claimPeriodReward.GetClaimPeriodId() == 0 {
		ctx.Logger().Error(fmt.Sprintf("Missing ClaimPeriodRewardDistribution for RewardProgram %d ", rewardProgram.GetId()))
		return fmt.Errorf("a ClaimPeriodRewardDistribution does not exist for RewardProgram %d with claim period %d", rewardProgram.GetId(), rewardProgram.GetId())
	}

	totalClaimPeriodRewards, err := k.CalculateRewardClaimPeriodRewards(ctx, rewardProgram.GetMaxRewardByAddress(), claimPeriodReward)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to calculate claim period rewards for RewardProgram %d ", rewardProgram.GetId()), "err", err)
		return err
	}

	err = k.MakeRewardClaimsClaimableForPeriod(ctx, rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod())
	if err != nil {
		return err
	}
	// Update balances
	claimPeriodReward.TotalRewardsPoolForClaimPeriod = claimPeriodReward.TotalRewardsPoolForClaimPeriod.Add(totalClaimPeriodRewards)
	claimPeriodReward.ClaimPeriodEnded = true
	rewardProgram.RemainingPoolBalance = rewardProgram.RemainingPoolBalance.Sub(totalClaimPeriodRewards)
	k.SetClaimPeriodRewardDistribution(ctx, claimPeriodReward)
	k.SetRewardProgram(ctx, *rewardProgram)

	if rewardProgram.IsEnding(ctx, rewardProgram.RemainingPoolBalance) {
		err = k.EndRewardProgram(ctx, rewardProgram)
		if err != nil {
			return err
		}
	} else {
		err = k.StartRewardProgramClaimPeriod(ctx, rewardProgram)
		if err != nil {
			return err
		}
	}

	return nil
}

// EndRewardProgram ActualProgramEndTime is updated and program ENDED
func (k Keeper) EndRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	if rewardProgram == nil {
		ctx.Logger().Error("Attempting to end reward program for nil reward program")
		return fmt.Errorf("unable to end reward program for nil reward program")
	}

	ctx.Logger().Info(fmt.Sprintf("BeginBlocker - Ending reward program %v ", rewardProgram))
	blockTime := ctx.BlockTime()
	rewardProgram.State = types.RewardProgram_STATE_FINISHED
	rewardProgram.ActualProgramEndTime = blockTime

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardProgramFinished,
			sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", rewardProgram.GetId())),
		),
	)

	return nil
}

// ExpireRewardProgram return unclaimed rewards to reward creator, and expire the reward program.
func (k Keeper) ExpireRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) error {
	if rewardProgram == nil {
		ctx.Logger().Error("Attempting to expire reward program for nil reward program")
		return fmt.Errorf("unable to expire reward program for nil reward program")
	}
	ctx.Logger().Info(fmt.Sprintf("BeginBlocker - Expiring reward program %v ", rewardProgram))

	rewardProgram.State = types.RewardProgram_STATE_EXPIRED
	err := k.ExpireRewardClaimsForRewardProgram(ctx, rewardProgram.GetId())
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Failed to expire reward claims for reward program. %v", err))
	}
	err = k.RefundRewardClaims(ctx, *rewardProgram)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Failed to refund reward claims. %v", err))
	}
	err = k.RefundRemainingBalance(ctx, rewardProgram)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Failed to refund remaining balance. %v", err))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardProgramExpired,
			sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", rewardProgram.GetId())),
		),
	)

	return err
}

// CalculateRewardClaimPeriodRewards calculate reward accrued for a claim period for each participant.
func (k Keeper) CalculateRewardClaimPeriodRewards(ctx sdk.Context, maxReward sdk.Coin, claimPeriodReward types.ClaimPeriodRewardDistribution) (sum sdk.Coin, err error) {
	sum = sdk.NewInt64Coin(claimPeriodReward.GetRewardsPool().Denom, 0)

	if maxReward.Denom != claimPeriodReward.RewardsPool.GetDenom() {
		ctx.Logger().Error(fmt.Sprintf("CalculateRewardClaimPeriodRewards denoms don't match %s %s", maxReward.Denom, claimPeriodReward.RewardsPool.GetDenom()))
		return sum, fmt.Errorf("ProgramBalance, MaxReward, and ClaimPeriodReward denoms must match")
	}

	participants, err := k.GetRewardAccountStatesForClaimPeriod(ctx, claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId())
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to get shares for reward program %d's claim period %d ", claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId()))
		return sum, fmt.Errorf("unable to get reward claim period shares for reward program %d and claim period %d", claimPeriodReward.GetRewardProgramId(), claimPeriodReward.GetClaimPeriodId())
	}

	for _, participant := range participants {
		reward := k.CalculateParticipantReward(ctx, int64(participant.GetSharesEarned()), claimPeriodReward.GetTotalShares(), claimPeriodReward.GetRewardsPool(), maxReward)

		sum = sum.Add(reward)
	}

	return sum, nil
}

// CalculateParticipantReward for each address/participant
func (k Keeper) CalculateParticipantReward(ctx sdk.Context, shares int64, totalShares int64, claimRewardPool sdk.Coin, maxReward sdk.Coin) sdk.Coin {
	numerator := sdk.NewDec(shares)
	denom := sdk.NewDec(totalShares)

	percentage := sdk.NewDec(0)
	if totalShares > 0 {
		percentage = numerator.Quo(denom)
	}

	pool := sdk.NewDec(claimRewardPool.Amount.Int64())
	reward := sdk.NewInt64Coin(claimRewardPool.Denom, pool.Mul(percentage).TruncateInt64())

	if maxReward.IsLT(reward) {
		reward = maxReward
	}

	return reward
}
