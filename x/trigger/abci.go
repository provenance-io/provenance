package trigger

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker runs trigger actions
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
}

// EndBlocker detects tx events for triggers
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
}
