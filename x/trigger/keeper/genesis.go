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

	queueLength, err := k.GetQueueLength(ctx)
	if err == nil && queueLength != uint64(len(queue)) {
		ctx.Logger().Warn("Queue length mismatch during export",
			"stored", queueLength,
			"actual", len(queue))
	}

	return types.NewGenesisState(triggerID, queueStartIndex, triggers, queue)
}

// InitGenesis new trigger genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}

	// Set trigger ID
	if err := k.setTriggerID(ctx, data.TriggerId); err != nil {
		panic(fmt.Sprintf("Failed to set trigger ID: %v", err))
	}

	// Set queue start index
	if err := k.setQueueStartIndex(ctx, data.QueueStart); err != nil {
		panic(fmt.Sprintf("Failed to set queue start index: %v", err))
	}

	// FIX: Initialize queue length BEFORE enqueuing
	if err := k.setQueueLength(ctx, 0); err != nil {
		panic(fmt.Sprintf("Failed to initialize queue length: %v", err))
	}

	// Enqueue queued triggers
	for _, queuedTrigger := range data.QueuedTriggers {
		if err := k.Enqueue(ctx, queuedTrigger); err != nil {
			panic(fmt.Sprintf("Failed to enqueue trigger: %v", err))
		}
	}

	// Set triggers and event listeners
	for _, trigger := range data.Triggers {
		if err := k.SetTrigger(ctx, trigger); err != nil {
			panic(fmt.Sprintf("Failed to set trigger %d: %v", trigger.Id, err))
		}
		if err := k.SetEventListener(ctx, trigger); err != nil {
			panic(fmt.Sprintf("Failed to set event listener for trigger %d: %v", trigger.Id, err))
		}
	}
}
