package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

var _ ibcratelimit.QueryServer = Keeper{}

// Params returns the params used by the module
func (k Keeper) Params(ctx context.Context, _ *ibcratelimit.ParamsRequest) (*ibcratelimit.ParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, err := k.GetParams(sdkCtx)
	if err != nil {
		return nil, err
	}

	return &ibcratelimit.ParamsResponse{Params: params}, nil
}
