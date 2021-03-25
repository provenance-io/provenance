package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// GetParams returns the total set of distribution parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	return types.Params{
		MaxTotalSupply:         k.GetMaxTotalSupply(ctx),
		EnableGovernance:       k.GetEnableGovernance(ctx),
		UnrestrictedDenomRegex: k.GetUnrestrictedDenomRegex(ctx),
	}
}

// SetParams sets the distribution parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetMaxTotalSupply return the current parameter value for the max allowed total supply (or default if unset)
func (k Keeper) GetMaxTotalSupply(ctx sdk.Context) (max uint64) {
	max = types.DefaultMaxTotalSupply
	if k.paramSpace.Has(ctx, types.ParamStoreKeyMaxTotalSupply) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyMaxTotalSupply, &max)
	}
	return
}

// GetEnableGovernance returns the current parameter value for enabling governance control (or default if unset)
func (k Keeper) GetEnableGovernance(ctx sdk.Context) (enabled bool) {
	enabled = types.DefaultEnableGovernance
	if k.paramSpace.Has(ctx, types.ParamStoreKeyEnableGovernance) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyEnableGovernance, &enabled)
	}
	return
}

// GetUnrestrictedDenomRegex returns the current parameter value for enabling governance control (or default if unset)
func (k Keeper) GetUnrestrictedDenomRegex(ctx sdk.Context) (regex string) {
	regex = types.DefaultUnrestrictedDenomRegex
	if k.paramSpace.Has(ctx, types.ParamStoreKeyUnrestrictedDenomRegex) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyUnrestrictedDenomRegex, &regex)
	}
	return
}
