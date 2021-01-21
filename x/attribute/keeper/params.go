package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// GetParams returns the total set of account parameters.
func (k Keeper) GetParams(clientCtx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(clientCtx, &params)
	return params
}

// SetParams sets the account parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetMaxValueLength returns the current distribution community tax.
func (k Keeper) GetMaxValueLength(ctx sdk.Context) (maxValueLength uint32) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyMaxValueLength, &maxValueLength)
	return maxValueLength
}
