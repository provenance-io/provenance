package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetParams returns the total set of metadata parameters.
func (k Keeper) GetOSLocatorParams(ctx sdk.Context) (osLocatorParams types.OSLocatorParams) {
	k.paramSpace.GetParamSet(ctx, &osLocatorParams)
	return osLocatorParams
}
