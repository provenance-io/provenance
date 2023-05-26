package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	triggerID := k.getTriggerID(ctx)
	queueStartIndex := k.getQueueStartIndex(ctx)

	triggers, err := k.GetAllTriggers(ctx)
	if err != nil {
		panic(err)
	}

	queue, err := k.GetAllQueueItems(ctx)
	if err != nil {
		panic(err)
	}

	gasLimits, err := k.GetAllGasLimits(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(triggerID, queueStartIndex, triggers, gasLimits, queue)
}

// InitGenesis new trigger genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}

	k.setTriggerID(ctx, data.TriggerId)
	k.setQueueStartIndex(ctx, data.QueueStart)

	for _, gasLimit := range data.GasLimits {
		k.SetGasLimit(ctx, gasLimit.TriggerId, gasLimit.Amount)
	}

	for _, queuedTrigger := range data.QueuedTriggers {
		k.Enqueue(ctx, queuedTrigger)
	}

	for _, trigger := range data.Triggers {
		k.SetTrigger(ctx, trigger)
		k.SetEventListener(ctx, trigger)
	}
}
