package keeper

import (
	"fmt"

	"github.com/provenance-io/provenance/x/marker/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// The name of the marker supply invariant
const invariantName = "required-marker-supply"

// RegisterInvariants registers module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, mk Keeper, bk bankkeeper.Keeper) {
	ir.RegisterRoute(types.ModuleName, invariantName, supplyInvariant(mk, bk))
}

// AllInvariants runs all invariants of the marker module.
func AllInvariants(k Keeper, bk bankkeeper.Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return supplyInvariant(k, bk)(ctx)
	}
}

// Checks that all of the marker supply values match the expected system totals.
func supplyInvariant(mk Keeper, bk bankkeeper.Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		statusMessage := ""
		isBroken := false
		mk.IterateMarkers(ctx, func(record types.MarkerAccountI) bool {
			// Invariant checks are only done against active markers.
			if record.GetStatus() == types.StatusActive && record.HasFixedSupply() {
				requiredSupply := record.GetSupply()
				currentSupply := bk.GetSupply(ctx, requiredSupply.Denom)

				// Just log the supply status
				if !requiredSupply.IsEqual(currentSupply) {
					ctx.Logger().Error(
						fmt.Sprintf("Current %s supply is NOT at the required amount",
							requiredSupply.Denom), invariantName, currentSupply)
					isBroken = true
				} else {
					ctx.Logger().Info(fmt.Sprintf("Current %s supply is at the required amount",
						requiredSupply.Denom), invariantName, currentSupply)
				}
				msg := fmt.Sprintf("invalid %s supply: required (%+v) current (%+v)\n",
					requiredSupply.Denom, requiredSupply.Amount, currentSupply)
				statusMessage += sdk.FormatInvariant(types.ModuleName, invariantName, msg)
			}
			return false
		})
		if isBroken {
			statusMessage = fmt.Sprintf("failed to assess invariant: %s", statusMessage)
		}

		return statusMessage, isBroken
	}
}
