package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ibcratelimit/types"
)

var _ types.QueryServer = Keeper{}

// Params returns the params used by the module
func (k Keeper) Params(ctx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, err := k.GetParams(sdkCtx)
	if err != nil {
		return nil, err
	}

	return &types.ParamsResponse{Params: params}, nil
}
