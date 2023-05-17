package reward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/keeper"
)

// BeginBlocker processes rewards module updates
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger().Error("reward BeginBlocker recovered from panic:", "r", r)
		}
	}()
	k.UpdateUnexpiredRewardsProgram(ctx)
}

// EndBlocker processes events for reward programs
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger().Error("reward EndBlocker recovered from panic:", "r", r)
		}
	}()
	k.ProcessTransactions(ctx)
}
