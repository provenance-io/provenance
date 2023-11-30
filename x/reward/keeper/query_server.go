package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/reward/types"
)

var _ types.QueryServer = Keeper{}

// RewardPrograms returns a list of reward programs matching the query type.
func (k Keeper) RewardPrograms(ctx context.Context, req *types.QueryRewardProgramsRequest) (*types.QueryRewardProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var err error

	rewardProgramStates := []types.RewardProgram_State{}
	switch req.QueryType {
	case types.QueryRewardProgramsRequest_QUERY_TYPE_PENDING:
		rewardProgramStates = []types.RewardProgram_State{types.RewardProgram_STATE_PENDING}
	case types.QueryRewardProgramsRequest_QUERY_TYPE_ACTIVE:
		rewardProgramStates = []types.RewardProgram_State{types.RewardProgram_STATE_STARTED}
	case types.QueryRewardProgramsRequest_QUERY_TYPE_FINISHED:
		rewardProgramStates = []types.RewardProgram_State{types.RewardProgram_STATE_FINISHED, types.RewardProgram_STATE_EXPIRED}
	case types.QueryRewardProgramsRequest_QUERY_TYPE_OUTSTANDING:
		rewardProgramStates = []types.RewardProgram_State{types.RewardProgram_STATE_PENDING, types.RewardProgram_STATE_STARTED}
	}

	response := types.QueryRewardProgramsResponse{}
	kvStore := sdkCtx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.RewardProgramKeyPrefix)
	pageResponse, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var rewardProgram types.RewardProgram
		vErr := rewardProgram.Unmarshal(value)

		if vErr != nil {
			return false, vErr
		}

		matched := rewardProgram.MatchesState(rewardProgramStates)
		if accumulate && matched {
			response.RewardPrograms = append(response.RewardPrograms, rewardProgram)
		}

		return matched, nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query all reward programs: %v", err))
	}
	response.Pagination = pageResponse
	return &response, nil
}

// RewardProgramByID returns a reward program matching the ID.
func (k Keeper) RewardProgramByID(ctx context.Context, req *types.QueryRewardProgramByIDRequest) (*types.QueryRewardProgramByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rewardProgram, err := k.GetRewardProgram(sdkCtx, req.GetId())
	if err != nil {
		return &types.QueryRewardProgramByIDResponse{}, status.Errorf(codes.Internal, fmt.Sprintf("unable to query for reward program by ID: %v", err))
	}
	return &types.QueryRewardProgramByIDResponse{RewardProgram: &rewardProgram}, nil
}

// ClaimPeriodRewardDistributions returns a list of claim period reward distributions matching the claim_status.
func (k Keeper) ClaimPeriodRewardDistributions(ctx context.Context, req *types.QueryClaimPeriodRewardDistributionsRequest) (*types.QueryClaimPeriodRewardDistributionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.QueryClaimPeriodRewardDistributionsResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	kvStore := sdkCtx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.ClaimPeriodRewardDistributionKeyPrefix)
	pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var claimPeriodRewardDist types.ClaimPeriodRewardDistribution
		vErr := claimPeriodRewardDist.Unmarshal(value)

		if vErr != nil {
			return false, vErr
		}

		if accumulate {
			response.ClaimPeriodRewardDistributions = append(response.ClaimPeriodRewardDistributions, claimPeriodRewardDist)
		}

		return true, nil
	})
	if err != nil {
		return &response, status.Error(codes.Unavailable, err.Error())
	}
	response.Pagination = pageRes
	return &response, nil
}

// ClaimPeriodRewardDistributionsByID returns a claim period reward distribution matching the ID.
func (k Keeper) ClaimPeriodRewardDistributionsByID(ctx context.Context, req *types.QueryClaimPeriodRewardDistributionsByIDRequest) (*types.QueryClaimPeriodRewardDistributionsByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.QueryClaimPeriodRewardDistributionsByIDResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ClaimPeriodReward, err := k.GetClaimPeriodRewardDistribution(sdkCtx, req.GetClaimPeriodId(), req.GetRewardId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("unable to query claim period reward distributions by ID: %v", err))
	}

	if ClaimPeriodReward.Validate() == nil {
		response.ClaimPeriodRewardDistribution = &ClaimPeriodReward
	}

	return &response, nil
}

// RewardDistributionsByAddress returns a list of reward claims belonging to the account and matching the claim status.
func (k Keeper) RewardDistributionsByAddress(ctx context.Context, request *types.QueryRewardDistributionsByAddressRequest) (*types.QueryRewardDistributionsByAddressResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	address, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrap(err.Error())
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var states []types.RewardAccountState
	getAllRewardAccountStore := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.GetAllRewardAccountByAddressPartialKey(address))

	pageRes, err := query.FilteredPaginate(getAllRewardAccountStore, request.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		lookupVal, errFromParsingKey := types.ParseFilterLookUpKey(key, address)

		if errFromParsingKey != nil {
			return false, err
		}
		result, errFromGetRewardAccount := k.GetRewardAccountState(sdkCtx, lookupVal.RewardID, lookupVal.ClaimID, lookupVal.Addr.String())
		// think ignoring the error maybe ok here since it's just another lookup
		if errFromGetRewardAccount != nil {
			return false, errFromGetRewardAccount
		}
		if result.GetSharesEarned() == 0 || (request.ClaimStatus != result.ClaimStatus && request.ClaimStatus != types.RewardAccountState_CLAIM_STATUS_UNSPECIFIED) {
			return false, nil
		}

		if accumulate {
			states = append(states, result)
		}

		return true, nil
	})

	if err != nil {
		return nil, types.ErrIterateAllRewardAccountStates.Wrap(err.Error())
	}

	rewardAccountResponses := k.convertRewardAccountStateToRewardAccountResponse(sdkCtx, states)
	rewardAccountByAddressResponse := types.QueryRewardDistributionsByAddressResponse{
		Address:            request.Address,
		RewardAccountState: rewardAccountResponses,
		Pagination:         pageRes,
	}

	return &rewardAccountByAddressResponse, nil
}

func (k Keeper) convertRewardAccountStateToRewardAccountResponse(ctx sdk.Context, states []types.RewardAccountState) []types.RewardAccountResponse {
	rewardAccountResponse := make([]types.RewardAccountResponse, 0)
	for _, state := range states {
		rewardProgram, err := k.GetRewardProgram(ctx, state.GetRewardProgramId())
		if err != nil {
			continue
		}
		distribution, err := k.GetClaimPeriodRewardDistribution(ctx, state.ClaimPeriodId, state.RewardProgramId)
		if err != nil {
			continue
		}

		participantReward := k.CalculateParticipantReward(ctx, int64(state.GetSharesEarned()), distribution.GetTotalShares(), distribution.GetRewardsPool(), rewardProgram.MaxRewardByAddress)
		accountResponse := types.RewardAccountResponse{
			RewardProgramId:  state.RewardProgramId,
			TotalRewardClaim: participantReward,
			ClaimStatus:      state.ClaimStatus,
			ClaimId:          state.ClaimPeriodId,
		}
		rewardAccountResponse = append(rewardAccountResponse, accountResponse)
	}

	return rewardAccountResponse
}
