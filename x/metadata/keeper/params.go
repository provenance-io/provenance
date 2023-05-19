package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetParams returns the total set of metadata parameters.
func (k Keeper) GetParams(_ sdk.Context) (params types.Params) {
	return types.Params{} // there are currently no params so no further action required here.
}

// SetParams sets the metadata parameters to the param space.
func (k Keeper) SetParams(_ sdk.Context, _ types.Params) {} // currently no params are supported
