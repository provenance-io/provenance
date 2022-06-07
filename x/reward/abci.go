package reward

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/keeper"
)

// EndBlocker called every block
/*func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	blockTime := ctx.BlockTime()
	// check if epoch has ended
	ctx.Logger().Info(fmt.Sprintf("NOTICE: Block time: %v Size of events is %d", blockTime, len(ctx.EventManager().GetABCIEventHistory())))
	logEvents(ctx)
	// Step 1
	rewardPrograms, err := k.GetAllActiveRewards(ctx)
	if err != nil {
		// TODO log it imo..we don't want blockchain to stop?
		ctx.Logger().Error(err.Error())
	}

	// only rewards programs who are eligible will be iterated through here
	// Step 2
	for _, rewardProgram := range rewardPrograms {
		epochRewardDistributionForEpoch, err := k.GetEpochRewardDistribution(ctx, "", rewardProgram.Id)
		//epochRewardDistributionForEpoch, err := k.GetEpochRewardDistribution(ctx, rewardProgram.EpochId, rewardProgram.Id)
		if err != nil {
			continue
		}

		// TODO We removed epoch so I am using 0 to just get it to compile
		currentEpoch := uint64(0)
		//currentEpoch := k.EpochKeeper.GetEpochInfo(ctx, rewardProgram.EpochId)
		// epoch reward distribution does it exist till the block has ended, highly unlikely but could happen
		if epochRewardDistributionForEpoch.EpochId == "" {
			epochRewardDistributionForEpoch.EpochId = ""
			//epochRewardDistributionForEpoch.EpochId = rewardProgram.EpochId
			epochRewardDistributionForEpoch.RewardProgramId = rewardProgram.Id
			epochRewardDistributionForEpoch.TotalShares = 0
			epochRewardDistributionForEpoch.TotalRewardsPool = rewardProgram.Coin
			epochRewardDistributionForEpoch.EpochEnded = false
			k.EvaluateRules(ctx, currentEpoch, rewardProgram, epochRewardDistributionForEpoch)
		} else if epochRewardDistributionForEpoch.EpochEnded == false { // if hook epoch end has already been called, this should not get called.
			// end the epoch
			k.EvaluateRules(ctx, currentEpoch, rewardProgram, epochRewardDistributionForEpoch)
		}
	}
}*/

// BeginBlocker called every block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	update(ctx, k)
	cleanup(ctx, k)
}

func update(ctx sdk.Context, k keeper.Keeper) {
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

func cleanup(ctx sdk.Context, k keeper.Keeper) {
	// This is where we want to remove shares
}

// this method is only for testing
func logEvents(ctx sdk.Context) {
	history := ctx.EventManager().GetABCIEventHistory()
	ctx.Logger().Info(fmt.Sprintf("Size of events is %d", len(history)))

	// check if epoch has ended
	for _, s := range ctx.EventManager().GetABCIEventHistory() {
		// ctx.Logger().Info(fmt.Sprintf("events type is %s", s.Type))
		ctx.Logger().Info(fmt.Sprintf("------- %s -------\n", s.Type))
		for _, y := range s.Attributes {
			ctx.Logger().Info(fmt.Sprintf("%s: %s\n", y.Key, y.Value))
			// ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", s.Type, y.Key, y.Value))
			//4:24PM INF events type is coin_spent
			//4:24PM INF event attribute is coin_spent attribute_key:attribute_value  spender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
			//4:24PM INF event attribute is coin_spent attribute_key:attribute_value  amount:76200000000000nhash
			//4:24PM INF events type is coin_received
			//4:24PM INF event attribute is coin_received attribute_key:attribute_value  receiver:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
			//4:24PM INF event attribute is coin_received attribute_key:attribute_value  amount:76200000000000nhash
			//4:24PM INF events type is transfer
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  recipient:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  sender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
			//4:24PM INF event attribute is transfer attribute_key:attribute_value  amount:76200000000000nhash
		}
		ctx.Logger().Info(fmt.Sprintf("------- %s -------\n\n", s.Type))
	}
}

// New Implementation

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Info("NOTICE: -End Blocker-")
	blockTime := ctx.BlockTime()
	ctx.Logger().Info(fmt.Sprintf("NOTICE: Block time: %v Size of events is %d", blockTime, len(ctx.EventManager().GetABCIEventHistory())))
	logEvents(ctx)

	// Get all Active Reward Programs
	rewardPrograms, err := k.GetAllActiveRewardPrograms(ctx)
	if err != nil {
		ctx.Logger().Error(err.Error())
		return
	}

	// Grant shares for qualifying actions
	for _, program := range rewardPrograms {
		// Go through all the reward programs
		program := program
		actions, err := k.DetectQualifyingActions(ctx, &program)
		if err != nil {
			ctx.Logger().Error(err.Error())
			continue
		}

		// Record any results
		err = k.RewardShares(ctx, &program, actions)
		if err != nil {
			ctx.Logger().Error(err.Error())
		}
	}
}
