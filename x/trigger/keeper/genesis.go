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

	must := func(err error, msg string) {
		if err != nil {
			panic(fmt.Errorf("%s: %w", msg, err))
		}
	}

	must(k.setTriggerID(ctx, data.TriggerId), "failed to set trigger ID")
	must(k.setQueueStartIndex(ctx, data.QueueStart), "failed to set queue start index")

	if len(data.QueuedTriggers) == 0 {
		must(k.setQueueLength(ctx, 0), "failed to initialize queue length")
	}

	for _, queuedTrigger := range data.QueuedTriggers {
		must(k.Enqueue(ctx, queuedTrigger), "failed to enqueue trigger")
	}

	for _, trigger := range data.Triggers {
		must(k.SetTrigger(ctx, trigger), fmt.Sprintf("failed to set trigger %d", trigger.Id))
		must(k.SetEventListener(ctx, trigger), fmt.Sprintf("failed to set event listener for trigger %d", trigger.Id))
	}
}
