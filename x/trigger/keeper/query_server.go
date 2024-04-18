package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"

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
		return &types.QueryTriggerByIDResponse{}, err
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

	response := types.QueryTriggersResponse{}
	kvStore := sdkCtx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.TriggerKeyPrefix)
	pageResponse, err := query.FilteredPaginate(prefixStore, pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var trigger types.Trigger
		vErr := trigger.Unmarshal(value)

		if vErr != nil {
			return false, vErr
		}

		if accumulate {
			response.Triggers = append(response.Triggers, trigger)
		}

		return true, nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to query all triggers: %v", err)
	}
	response.Pagination = pageResponse

	return &response, nil
}
