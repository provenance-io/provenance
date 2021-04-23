package keeper

import (
	"fmt"
	"regexp"

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
		// use the default value for empty regex expressions.
		if len(regex) == 0 {
			regex = types.DefaultUnrestrictedDenomRegex
		}
	}
	return
}

// ValidateUnrestictedDenom checks if the supplied denom is valid based on the module params
func (k Keeper) ValidateUnrestictedDenom(ctx sdk.Context, denom string) error {
	// Anchors are enforced on the denom validation expression.  Similar to how the SDK does hits.
	// https://github.com/cosmos/cosmos-sdk/blob/512b533242d34926972a8fc2f5639e8cf182f5bd/types/coin.go#L625
	exp := k.GetUnrestrictedDenomRegex(ctx)

	// Safe to use must compile here because the regular expression is validated on parameter set.
	r := regexp.MustCompile(fmt.Sprintf(`^%s$`, exp))
	if !r.MatchString(denom) {
		return fmt.Errorf("invalid denom [%s] (fails unrestricted marker denom validation %s)", denom, exp)
	}
	return nil
}
