package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/provenance-io/provenance/x/reward/types"
)

var _ types.QueryServer = Keeper{}

const defaultPerPageLimit = 100

func (k Keeper) RewardPrograms(ctx context.Context, req *types.RewardProgramsRequest) (*types.RewardProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var rewardPrograms []types.RewardProgram
	var err error

	switch req.QueryType {
	case types.RewardProgramsRequest_ALL:
		rewardPrograms, err = k.GetAllRewardPrograms(sdkCtx)
	case types.RewardProgramsRequest_PENDING:
		rewardPrograms, err = k.GetAllPendingRewardPrograms(sdkCtx)
	case types.RewardProgramsRequest_ACTIVE:
		rewardPrograms, err = k.GetAllActiveRewardPrograms(sdkCtx)
	case types.RewardProgramsRequest_FINISHED:
		rewardPrograms, err = k.GetAllCompletedRewardPrograms(sdkCtx)
	case types.RewardProgramsRequest_OUTSTANDING:
		rewardPrograms, err = k.GetOutstandingRewardPrograms(sdkCtx)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query all reward programs: %v", err))
	}

	return &types.RewardProgramsResponse{RewardPrograms: rewardPrograms}, nil
}

func (k Keeper) RewardProgramByID(ctx context.Context, req *types.RewardProgramByIDRequest) (*types.RewardProgramByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rewardProgram, err := k.GetRewardProgram(sdkCtx, req.GetId())
	if err != nil {
		return &types.RewardProgramByIDResponse{}, status.Errorf(codes.Internal, fmt.Sprintf("unable to query for reward program: %v", err))
	}
	return &types.RewardProgramByIDResponse{RewardProgram: &rewardProgram}, nil
}

// returns paginated ClaimPeriodRewardDistributions
func (k Keeper) ClaimPeriodRewardDistributions(ctx context.Context, req *types.ClaimPeriodRewardDistributionRequest) (*types.ClaimPeriodRewardDistributionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	pageRequest := getPageRequest(req)

	response := types.ClaimPeriodRewardDistributionResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	kvStore := sdkCtx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.ClaimPeriodRewardDistributionKeyPrefix)
	pageRes, err := query.Paginate(prefixStore, pageRequest, func(key, value []byte) error {
		var claimPeriodRewardDist types.ClaimPeriodRewardDistribution
		vErr := claimPeriodRewardDist.Unmarshal(value)
		if vErr == nil {
			response.ClaimPeriodRewardDistribution = append(response.ClaimPeriodRewardDistribution, claimPeriodRewardDist)
		}
		// move on for now, even if error
		return nil
	})
	if err != nil {
		return &response, status.Error(codes.Unavailable, err.Error())
	}
	response.Pagination = pageRes
	return &response, nil
}

// ClaimPeriodRewardDistributionsByID returns a ClaimPeriodRewardDistribution by rewardId and epochId
func (k Keeper) ClaimPeriodRewardDistributionsByID(ctx context.Context, req *types.ClaimPeriodRewardDistributionByIDRequest) (*types.ClaimPeriodRewardDistributionByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.ClaimPeriodRewardDistributionByIDResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ClaimPeriodReward, err := k.GetClaimPeriodRewardDistribution(sdkCtx, req.GetClaimPeriodId(), req.GetRewardId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query epoch reward distributions: %v", err))
	}

	if k.ClaimPeriodRewardDistributionIsValid(&ClaimPeriodReward) {
		response.ClaimPeriodRewardDistribution = &ClaimPeriodReward
	}

	return &response, nil
}

// ClaimPeriodRewardDistributionsByAddress returns ClaimPeriodRewardDistributionByIDResponse for the given address
func (k Keeper) ClaimPeriodRewardDistributionsByAddress(ctx context.Context, request *types.ClaimPeriodRewardDistributionByAddressRequest) (*types.ClaimPeriodRewardDistributionByIDResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	_, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	panic("implement me")
}

// hasPageRequest is just for use with the getPageRequest func below.
type hasPageRequest interface {
	GetPagination() *query.PageRequest
}

// Gets the query.PageRequest from the provided request if there is one.
// Also sets the default limit if it's not already set yet.
func getPageRequest(req hasPageRequest) *query.PageRequest {
	var pageRequest *query.PageRequest
	if req != nil {
		pageRequest = req.GetPagination()
	}
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}
	if pageRequest.Limit == 0 {
		pageRequest.Limit = defaultPerPageLimit
	}
	return pageRequest
}
