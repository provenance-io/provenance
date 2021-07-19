package marker

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// BeginBlocker returns the begin blocker for the marker module.
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper, bk bankkeeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	// Iterate through all marker accounts and check for supply above or below expected targets.
	var err error
	k.IterateMarkers(ctx, func(record types.MarkerAccountI) bool {
		// Supply checks are only done against active markers with a fixed supply.
		if record.GetStatus() == types.StatusActive && record.HasFixedSupply() {
			requiredSupply := record.GetSupply()
			currentSupply := bk.GetSupply(ctx, record.GetDenom())

			// If the current amount of marker coin in circulation doesn't match configured supply, make adjustments
			if !requiredSupply.IsEqual(currentSupply) {
				ctx.Logger().Error(
					fmt.Sprintf("Current %s supply is NOT at the required amount, adjusting %s to required supply level",
						record.GetDenom(), currentSupply))
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
