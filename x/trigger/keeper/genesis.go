package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	triggerID, err := k.getTriggerID(ctx)
	if err != nil {
		panic(err)
	}

	queueStartIndex, err := k.getQueueStartIndex(ctx)
	if err != nil {
		panic(err)
	}

	triggers, err := k.GetAllTriggers(ctx)
	if err != nil {
		panic(err)
	}

	queue, err := k.GetAllQueueItems(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(triggerID, queueStartIndex, triggers, queue)
}

// InitGenesis new trigger genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}

	// set trigger ID
	if err := k.setTriggerID(ctx, data.TriggerId); err != nil {
		ctx.Logger().Error(fmt.Sprintf("Failed to set trigger ID: %v", err))
		return
	}

	// set queue start index
	if err := k.setQueueStartIndex(ctx, data.QueueStart); err != nil {
		ctx.Logger().Error(fmt.Sprintf("Failed to set queue start index: %v", err))
		return
	}

	// initialize queue length if needed
	if len(data.QueuedTriggers) == 0 {
		if err := k.setQueueLength(ctx, 0); err != nil {
			ctx.Logger().Error(fmt.Sprintf("Failed to initialize queue length: %v", err))
			return
		}
	}

	// enqueue queued triggers
	for _, queuedTrigger := range data.QueuedTriggers {
		if err := k.Enqueue(ctx, queuedTrigger); err != nil {
			ctx.Logger().Error(fmt.Sprintf("Failed to enqueue trigger: %v", err))
			return
		}
	}

	// set triggers and event listeners
	for _, trigger := range data.Triggers {
		if err := k.SetTrigger(ctx, trigger); err != nil {
			ctx.Logger().Error(fmt.Sprintf(
				"Failed to set trigger %d: %v", trigger.Id, err,
			))
			return
		}

		if err := k.SetEventListener(ctx, trigger); err != nil {
			ctx.Logger().Error(fmt.Sprintf(
				"Failed to set event listener for trigger %d: %v", trigger.Id, err,
			))
			return
		}
	}
}
