package reward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/keeper"
)

// BeginBlocker for rewards module updates the
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Debug("NOTICE: -Begin Blocker of rewards module-")
	k.UpdateUnexpiredRewardsProgram(ctx)
}
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Debug("NOTICE: -End Blocker of rewards module-")
	k.ProcessTransactions(ctx)
}
