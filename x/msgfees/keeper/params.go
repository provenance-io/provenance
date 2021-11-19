package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// GetParams returns the total set of distribution parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	return types.Params{
		EnableGovernance: k.GetEnableGovernance(ctx),
		MinGasPrice: k.GetMinGasPrice(ctx),
	}
}

// GetEnableGovernance returns the current parameter value for enabling governance control (or default if unset)
func (k Keeper) GetEnableGovernance(ctx sdk.Context) (enabled bool) {
	enabled = types.DefaultEnableGovernance
	if k.paramSpace.Has(ctx, types.ParamStoreKeyEnableGovernance) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyEnableGovernance, &enabled)
	}
	return
}
