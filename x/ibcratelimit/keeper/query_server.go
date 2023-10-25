package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/ibcratelimit/types"
)

var _ types.QueryServer = Keeper{}

// TriggerByID returns a trigger matching the ID.
func (k Keeper) Params(ctx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	params := k.GetParams(ctx)
	return &types.ParamsResponse{Params: params}, nil
}
