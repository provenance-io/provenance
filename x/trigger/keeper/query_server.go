package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

var _ types.QueryServer = Keeper{}

// TriggerByID returns a trigger matching the ID.
func (k Keeper) TriggerByID(ctx context.Context, req *types.QueryTriggerByIDRequest) (*types.QueryTriggerByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	trigger, err := k.GetTrigger(sdkCtx, req.GetId())
	if err != nil {
		// TODO Return an error here
		return nil, nil
	}
	return &types.QueryTriggerByIDResponse{Trigger: &trigger}, nil
}
