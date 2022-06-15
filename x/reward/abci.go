package reward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/keeper"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.Update(ctx)
	k.Cleanup(ctx)
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Info("NOTICE: -End Blocker-")
	k.ProcessTransactions(ctx)
}
