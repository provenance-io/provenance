package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

func (k Keeper) DetectBlockEvents(ctx sdk.Context) {
	k.DetectTransactionEvents(ctx)
	k.DetectBlockHeightEvents(ctx)
	k.DetectTimeEvents(ctx)
}

func (k Keeper) DetectTransactionEvents(ctx sdk.Context) {
	triggers := []types.Trigger{}

	for _, event := range ctx.EventManager().GetABCIEventHistory() {
		matched, err := k.GetMatchingTriggers(ctx, event)
		if err != nil {
			// TODO What to do here? There has to be something bad
			return
		}
		triggers = append(triggers, matched...)
	}

	for _, trigger := range triggers {
		k.QueueTrigger(ctx, trigger)
	}
}

func (k Keeper) DetectBlockHeightEvents(ctx sdk.Context) {

}

func (k Keeper) DetectTimeEvents(ctx sdk.Context) {

}

func (k Keeper) GetMatchingTriggers(ctx sdk.Context, event abci.Event) (triggers []types.Trigger, err error) {
	err = k.IterateEventListeners(ctx, event.GetType(), func(trigger types.Trigger) (stop bool, err error) {

		// This is where we would get the interface type
		tempEvent := trigger.Event.GetCachedValue().(types.TriggerEventI)
		triggerEvent := tempEvent.(*types.TransactionEvent)

		if triggerEvent.Equals(event) {
			triggers = append(triggers, trigger)
		}
		return false, nil
	})
	if err != nil {
		// TODO Return error
		return triggers, err
	}
	return triggers, err
}
