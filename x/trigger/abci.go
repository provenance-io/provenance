package trigger

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/keeper"
)

// BeginBlocker Runs trigger actions.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.ProcessTriggers(ctx)
}

// EndBlocker Detects tx events for triggers.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.DetectBlockEvents(ctx)
}
