package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) Update(ctx sdk.Context) {
	blockTime := ctx.BlockTime()
	// ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Block time: %v ", blockTime))

	outstanding, err := k.GetOutstandingRewardPrograms(ctx)
	if err != nil {
		ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - error iterating reward programs: %v ", err))
		return
	}

	for _, rewardProgram := range outstanding {
		if rewardProgram.IsStarting(ctx) {
			ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Starting reward program: %v ", rewardProgram))
			rewardProgram.Started = true
			rewardProgram.EpochEndTime = blockTime.Add(time.Duration(rewardProgram.EpochSeconds) * time.Second)
			rewardProgram.CurrentEpoch = 1
		} else if rewardProgram.IsEndingEpoch(ctx) {
			ctx.Logger().Info(fmt.Sprintf("NOTICE: BeginBlocker - Epoch end hit for reward program %v ", rewardProgram))
			rewardProgram.CurrentEpoch++
			if rewardProgram.IsEnding(ctx) {
				rewardProgram.Finished = true
				rewardProgram.FinishedTime = blockTime
			} else {
				rewardProgram.EpochEndTime = blockTime.Add(time.Duration(rewardProgram.EpochSeconds) * time.Second)
			}
		}
	}

	for _, rewardProgram := range outstanding {
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
