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

	for _, queuedTrigger := range data.QueuedTriggers {
		k.Enqueue(ctx, queuedTrigger)
	}

	for i := range data.Triggers {
		trigger := data.Triggers[i]
		gasLimit := data.GasLimits[i]
		k.SetTrigger(ctx, trigger)
		k.SetGasLimit(ctx, trigger.GetId(), gasLimit)
	}
}
