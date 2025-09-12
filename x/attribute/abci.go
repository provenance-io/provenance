// Package attribute provides ABCI logic for attribute handling.
package attribute

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
)

// MaxExpiredAttributionCount defines the max number of expired attributions to retain.
const MaxExpiredAttributionCount = 100_000

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	deleted := keeper.DeleteExpiredAttributes(ctx, MaxExpiredAttributionCount)
	if deleted > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"beginblock",
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeDeletedExpired),
				sdk.NewAttribute(types.AttributeKeyTotalExpired, strconv.Itoa(deleted)),
			),
		)
	}
}
