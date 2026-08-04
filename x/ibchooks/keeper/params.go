package keeper

import (
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ibchooks/types"
)

// GetParams returns the total set of the module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	params, err := k.params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.DefaultParams()
		}
		panic(err)
	}
	return params
}

// SetParams sets the module's parameters with the provided parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := k.params.Set(ctx, params); err != nil {
		panic(err)
	}
}
