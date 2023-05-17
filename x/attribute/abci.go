package attribute

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/keeper"
)

const MaxExpiredAttributionCount = 100_000

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	keeper.DeleteExpiredAttributes(ctx, MaxExpiredAttributionCount)
}
