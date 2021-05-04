package marker

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// BeginBlocker returns the begin blocker for the marker module.
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper, bk bankkeeper.Keeper) {
	// Iterate through all marker accounts and check for supply above or below expected targets.
	var err error
	k.IterateMarkers(ctx, func(record types.MarkerAccountI) bool {
		// Supply checks are only done against active markers with a fixed supply.
		if record.GetStatus() == types.StatusActive && record.HasFixedSupply() {
			requiredSupply := record.GetSupply()
			currentSupply := getCurrentSupply(ctx, record.GetDenom(), bk)

			// If the current amount of marker coin in circulation doesn't match configured supply, make adjustments
			if !requiredSupply.IsEqual(currentSupply) {
				err = k.AdjustCirculation(ctx, record, requiredSupply)
			}
			// else supply is equal, nothing to do here.
		}
		// Clear out markers that are in the destroyed status
		if record.GetStatus() == types.StatusDestroyed {
			k.RemoveMarker(ctx, record)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"beginblock",
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeDestroy),
					sdk.NewAttribute(types.EventAttributeDenomKey, record.GetDenom()),
				),
			)
		}
		return err != nil
	})
	// We have no way of dealing with this and the invariant will fail soon from mismatch halting the chain.
	if err != nil {
		panic(err)
	}
}

// Iterator over all coins and find the any matching our target marker denom, add their amounts to the returned total.
func getCurrentSupply(ctx sdk.Context, denom string, bk bankkeeper.Keeper) sdk.Coin {
	sup := bk.GetSupply(ctx)
	for _, coin := range sup.GetTotal() {
		if coin.Denom == denom {
			return coin
		}
	}
	return sdk.NewInt64Coin(denom, 0)
}
