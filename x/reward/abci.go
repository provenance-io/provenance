package reward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/keeper"
)

// BeginBlocker called every block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Info("NOTICE: -Begin Blocker of rewards module-")
	k.Update(ctx)
}
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Info("NOTICE: -End Blocker of rewards module-")
	k.ProcessTransactions(ctx)
}
