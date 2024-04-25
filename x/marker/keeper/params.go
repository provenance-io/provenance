package keeper

import (
	"fmt"
	"regexp"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// GetParams returns the total set of marker parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	params = types.DefaultParams() // Assuming a method that returns default parameters

	// Deserialize parameters if they are set
	if bz := store.Get(types.MarkerParamStoreKey); bz != nil {
		k.cdc.MustUnmarshal(bz, &params)
	}

	return params
}

// SetParams sets the marker parameters to the store.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := params.Validate(); err != nil {
		panic(err)
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.MarkerParamStoreKey, bz)
}

// Deprecated: GetMaxTotalSupply is kept for backwards compatibility.
func (k Keeper) GetMaxTotalSupply(ctx sdk.Context) (max uint64) {
	return k.GetParams(ctx).MaxTotalSupply
}

// GetMaxSupply returns the current parameter value for the max allowed supply.
func (k Keeper) GetMaxSupply(ctx sdk.Context) (max sdkmath.Int) {
	return k.GetParams(ctx).MaxSupply
}

// GetEnableGovernance returns whether governance control is enabled.
func (k Keeper) GetEnableGovernance(ctx sdk.Context) (enabled bool) {
	return k.GetParams(ctx).EnableGovernance
}

// GetUnrestrictedDenomRegex returns the regex for unrestricted denom validation.
func (k Keeper) GetUnrestrictedDenomRegex(ctx sdk.Context) (regex string) {
	return k.GetParams(ctx).UnrestrictedDenomRegex
}

// ValidateUnrestictedDenom checks if the supplied denom is valid based on the module params
func (k Keeper) ValidateUnrestictedDenom(ctx sdk.Context, denom string) error {
	// Anchors are enforced on the denom validation expression.  Similar to how the SDK does hits.
	// https://github.com/cosmos/cosmos-sdk/blob/512b533242d34926972a8fc2f5639e8cf182f5bd/types/coin.go#L625
	exp := k.GetUnrestrictedDenomRegex(ctx)
	if len(exp) == 0 {
		return nil
	}
	r := regexp.MustCompile(fmt.Sprintf(`^%s$`, exp))
	if !r.MatchString(denom) {
		return fmt.Errorf("invalid denom [%s] (fails unrestricted marker denom validation %s)", denom, exp)
	}
	return nil
}
