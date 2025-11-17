package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

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
		if errors.Is(err, collections.ErrNotFound) {
			return nil, types.ErrTriggerNotFound
		}
		return nil, err
	}
	return &types.QueryTriggerByIDResponse{Trigger: &trigger}, nil
}

// Triggers returns the list of triggers.
func (k Keeper) Triggers(ctx context.Context, req *types.QueryTriggersRequest) (*types.QueryTriggersResponse, error) {
	var pagination *query.PageRequest
	if req != nil {
		pagination = req.Pagination
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	triggers, pageRes, err := query.CollectionPaginate[uint64, types.Trigger, *collections.Map[uint64, types.Trigger], types.Trigger](
		sdkCtx,
		&k.TriggersMap,
		pagination,
		func(key uint64, trigger types.Trigger) (types.Trigger, error) {
			trigger.Id = key
			return trigger, nil
		},
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to query triggers: %v", err)
	}

	response := &types.QueryTriggersResponse{
		Triggers:   triggers,
		Pagination: pageRes,
	}

	return response, nil
}
