package keeper

import (
	"context"
	"encoding/binary"
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

func (k Keeper) RewardPrograms(ctx context.Context, req *types.QueryRewardProgramsRequest) (*types.QueryRewardProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var rewardPrograms []types.RewardProgram
	var err error

	switch req.QueryType {
	case types.QueryRewardProgramsRequest_ALL:
		rewardPrograms, err = k.GetAllRewardPrograms(sdkCtx)
	case types.QueryRewardProgramsRequest_PENDING:
		rewardPrograms, err = k.GetAllPendingRewardPrograms(sdkCtx)
	case types.QueryRewardProgramsRequest_ACTIVE:
		rewardPrograms, err = k.GetAllActiveRewardPrograms(sdkCtx)
	case types.QueryRewardProgramsRequest_FINISHED:
		rewardPrograms, err = k.GetAllCompletedRewardPrograms(sdkCtx)
	case types.QueryRewardProgramsRequest_OUTSTANDING:
		rewardPrograms, err = k.GetOutstandingRewardPrograms(sdkCtx)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query all reward programs: %v", err))
	}

	return &types.QueryRewardProgramsResponse{RewardPrograms: rewardPrograms}, nil
}

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

// returns paginated ClaimPeriodRewardDistributions
func (k Keeper) ClaimPeriodRewardDistributions(ctx context.Context, req *types.QueryClaimPeriodRewardDistributionsRequest) (*types.QueryClaimPeriodRewardDistributionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	pageRequest := getPageRequest(req)

	response := types.QueryClaimPeriodRewardDistributionsResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	kvStore := sdkCtx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.ClaimPeriodRewardDistributionKeyPrefix)
	pageRes, err := query.Paginate(prefixStore, pageRequest, func(key, value []byte) error {
		var claimPeriodRewardDist types.ClaimPeriodRewardDistribution
		vErr := claimPeriodRewardDist.Unmarshal(value)
		if vErr == nil {
			response.ClaimPeriodRewardDistributions = append(response.ClaimPeriodRewardDistributions, claimPeriodRewardDist)
		}
		return nil
	})
	if err != nil {
		return &response, status.Error(codes.Unavailable, err.Error())
	}
	response.Pagination = pageRes
	return &response, nil
}

// ClaimPeriodRewardDistributionsByID returns a ClaimPeriodRewardDistribution by rewardId and epochId
func (k Keeper) ClaimPeriodRewardDistributionsByID(ctx context.Context, req *types.QueryClaimPeriodRewardDistributionByIDRequest) (*types.QueryClaimPeriodRewardDistributionByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.QueryClaimPeriodRewardDistributionByIDResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ClaimPeriodReward, err := k.GetClaimPeriodRewardDistribution(sdkCtx, req.GetClaimPeriodId(), req.GetRewardId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("unable to query claim period reward distributions by ID: %v", err))
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
	pageRequest := getPageRequest(request)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var states []types.RewardAccountState
	getAllRewardAccountStore := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.GetAllRewardAccountByAddressPartialKey([]byte(address.String())))

	pageRes, err := query.FilteredPaginate(getAllRewardAccountStore, pageRequest, func(key []byte, value []byte, accumulate bool) (bool, error) {
		lookupVal, errFromParsingKey := ParseFilterLookUpKey(key, address)
		if errFromParsingKey != nil {
			return false, err
		}
		result, errFromGetRewardAccount := k.GetRewardAccountState(sdkCtx, lookupVal.rewardId, lookupVal.claimId, lookupVal.addr.String())
		// think ignoring the error maybe ok here since it's just another lookup
		if errFromGetRewardAccount != nil {
			return false, nil
		}
		if !(result.GetSharesEarned() > 0 && (request.ClaimStatus == types.QueryRewardsByAddressRequest_ALL || request.ClaimStatus.String() == result.ClaimStatus.String())) {
			return false, nil
		}

		if accumulate {
			states = append(states, result)
		}
		return true, nil
	})

	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrIterateAllRewardAccountStates, err.Error())
	}

	rewardAccountResponses := k.convertRewardAccountStateToRewardAccountResponse(sdkCtx, states)
	rewardAccountByAddressResponse := types.QueryAccountByAddressResponse{
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

// hasPageRequest is just for use with the getPageRequest func below.
type hasPageRequest interface {
	GetPagination() *query.PageRequest
}

// Gets the query.PageRequest from the provided request if there is one.
// Also sets the default limit if it's not already set yet.
func getPageRequest(req hasPageRequest) *query.PageRequest {
	var pageRequest *query.PageRequest
	if req.GetPagination() != nil {
		pageRequest = req.GetPagination()
		// enforce max/min page limit
		enforceMaxMinPageLimit(pageRequest)
		return pageRequest
	}
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}
	enforceMaxMinPageLimit(pageRequest)
	return pageRequest
}

func enforceMaxMinPageLimit(pageRequest *query.PageRequest) {
	if pageRequest.Limit == 0 || pageRequest.Limit > defaultPerPageLimit {
		pageRequest.Limit = defaultPerPageLimit
	}
}

func ParseFilterLookUpKey(accountStateAddressLookupKey []byte, addr sdk.AccAddress) (RewardAccountLookup, error) {
	rewardId := binary.BigEndian.Uint64(accountStateAddressLookupKey[0:8])
	claimId := binary.BigEndian.Uint64(accountStateAddressLookupKey[8:16])
	return RewardAccountLookup{
		addr:     addr,
		rewardId: rewardId,
		claimId:  claimId,
	}, nil
}

type RewardAccountLookup struct {
	addr     sdk.Address
	rewardId uint64
	claimId  uint64
}
