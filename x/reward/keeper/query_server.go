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

func (k Keeper) QueryRewardDistributionsByAddress(ctx context.Context, request *types.QueryRewardsByAddressRequest) (*types.QueryAccountByAddressResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	address, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var states []types.RewardAccountState
	err = k.IterateAllRewardAccountStates(sdkCtx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 && state.Address == address.String() && (request.ClaimStatus == types.QueryRewardsByAddressRequest_ALL || request.ClaimStatus.String() == state.ClaimStatus.String()) {
			states = append(states, state)
			return true
		}
		return false
	})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	rewardAccountResponses := k.convertRewardAccountStateToRewardAccountResponse(sdkCtx, states)
	rewardAccountByAddressResponse := types.QueryAccountByAddressResponse{
		Address:            request.Address,
		RewardAccountState: rewardAccountResponses,
	}
	return &rewardAccountByAddressResponse, nil
}

func (k Keeper) convertRewardAccountStateToRewardAccountResponse(ctx sdk.Context, states []types.RewardAccountState) []types.RewardAccountResponse {
	var rewardAccountResponse []types.RewardAccountResponse
	for _, state := range states {
		rewardProgram, err := k.GetRewardProgram(ctx, state.GetRewardProgramId())
		distribution, err := k.GetClaimPeriodRewardDistribution(ctx, state.ClaimPeriodId, state.RewardProgramId)
		if err != nil {
			continue
		}

		participantReward := k.CalculateParticipantReward(ctx, int64(state.GetSharesEarned()), distribution.GetTotalShares(), distribution.GetRewardsPool(), rewardProgram.MaxRewardByAddress)
		accountResponse := types.RewardAccountResponse{
			RewardProgramId:  state.RewardProgramId,
			TotalRewardClaim: participantReward,
			ClaimStatus:      state.ClaimStatus,
		}
		rewardAccountResponse = append(rewardAccountResponse, accountResponse)
	}

	return rewardAccountResponse
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
