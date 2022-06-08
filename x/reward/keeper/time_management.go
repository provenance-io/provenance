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
		} else if rewardProgram.IsEndingEpoch(ctx) {
			k.EndRewardProgramEpoch(ctx, &rewardProgram)
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
	rewardProgram.Started = true
	rewardProgram.EpochEndTime = blockTime.Add(time.Duration(rewardProgram.EpochSeconds) * time.Second)
	rewardProgram.CurrentEpoch = 1
}

func (k Keeper) EndRewardProgramEpoch(ctx sdk.Context, rewardProgram *types.RewardProgram) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Epoch end hit for reward program %v ", rewardProgram))
	blockTime := ctx.BlockTime()
	rewardProgram.CurrentEpoch++
	if rewardProgram.IsEnding(ctx) {
		k.EndRewardProgram(ctx, rewardProgram)
	} else {
		rewardProgram.EpochEndTime = blockTime.Add(time.Duration(rewardProgram.EpochSeconds) * time.Second)
	}
}

func (k Keeper) EndRewardProgram(ctx sdk.Context, rewardProgram *types.RewardProgram) {
	ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Ending reward program %v ", rewardProgram))
	blockTime := ctx.BlockTime()
	rewardProgram.Finished = true
	rewardProgram.FinishedTime = blockTime
}
