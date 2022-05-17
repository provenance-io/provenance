package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/provenance-io/provenance/x/reward/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RewardPrograms(ctx context.Context, req *types.RewardProgramsRequest) (*types.RewardProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	rewardPrograms, err := k.GetAllRewardPrograms(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query all reward programs: %v", err))
	}

	return &types.RewardProgramsResponse{RewardPrograms: rewardPrograms}, nil
}

func (k Keeper) ActiveRewardPrograms(ctx context.Context, req *types.ActiveRewardProgramsRequest) (*types.ActiveRewardProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.ActiveRewardProgramsResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	rewardPrograms, err := k.GetAllActiveRewardPrograms(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to obtain active reward programs: %v", err))
	}

	response.RewardPrograms = rewardPrograms

	return &response, nil
}

func (k Keeper) ModuleAccountBalance(context.Context, *types.QueryModuleAccountBalanceRequest) (*types.QueryModuleAccountBalanceResponse, error) {
	return &types.QueryModuleAccountBalanceResponse{}, nil
}

func (k Keeper) RewardProgramByID(ctx context.Context, req *types.RewardProgramByIDRequest) (*types.RewardProgramByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.RewardProgramByIDResponse{}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rewardProgram, err := k.GetRewardProgram(sdkCtx, req.GetId())
	if err != nil {
		return &response, status.Errorf(codes.Internal, fmt.Sprintf("unable to query for reward program: %v", err))
	}

	if k.RewardProgramIsValid(&rewardProgram) {
		response.RewardProgram = &rewardProgram
	}

	return &response, nil
}

func (k Keeper) RewardClaims(ctx context.Context, req *types.RewardClaimsRequest) (*types.RewardClaimsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.RewardClaimsResponse{}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rewardClaims, err := k.GetAllRewardClaims(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query reward claims: %v", err))
	}
	response.RewardClaims = rewardClaims

	return &response, nil
}

// returns a RewardClaim by id
func (k Keeper) RewardClaimByID(ctx context.Context, req *types.RewardClaimByIDRequest) (*types.RewardClaimByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.RewardClaimByIDResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	claim, err := k.GetRewardClaim(sdkCtx, req.GetId())
	if err != nil {
		return &response, status.Errorf(codes.Internal, fmt.Sprintf("unable to query for reward claim: %v", err))
	}

	if k.RewardClaimIsValid(&claim) {
		response.RewardClaim = &claim
	}

	return &response, nil
}

// returns all EpochRewardDistributions
func (k Keeper) EpochRewardDistributions(ctx context.Context, req *types.EpochRewardDistributionRequest) (*types.EpochRewardDistributionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.EpochRewardDistributionResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rewardDistributions, err := k.GetAllEpochRewardDistributions(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query epoch reward distributions: %v", err))
	}
	response.EpochRewardDistribution = rewardDistributions

	return &response, nil
}

// returns a EpochRewardDistribution by rewardId and epochId
func (k Keeper) EpochRewardDistributionsByID(ctx context.Context, req *types.EpochRewardDistributionByIDRequest) (*types.EpochRewardDistributionByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.EpochRewardDistributionByIDResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	epochReward, err := k.GetEpochRewardDistribution(sdkCtx, req.GetEpochId(), req.GetRewardId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query epoch reward distributions: %v", err))
	}

	if k.EpochRewardDistributionIsValid(&epochReward) {
		response.EpochRewardDistribution = &epochReward
	}

	return &response, nil
}

// returns all EligibilityCriterias
func (k Keeper) EligibilityCriteria(ctx context.Context, req *types.EligibilityCriteriaRequest) (*types.EligibilityCriteriaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.EligibilityCriteriaResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	criteria, err := k.GetAllEligibilityCriteria(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query all eligibility criteria: %v", err))
	}
	response.EligibilityCriteria = criteria

	return &response, nil
}

// returns a EligibilityCriteria by name
func (k Keeper) EligibilityCriteriaByName(ctx context.Context, req *types.EligibilityCriteriaRequestByNameRequest) (*types.EligibilityCriteriaRequestByNameResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.EligibilityCriteriaRequestByNameResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	criteria, err := k.GetEligibilityCriteria(sdkCtx, req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query eligibility criteria: %v", err))
	}

	if k.EligibilityCriteriaIsValid(&criteria) {
		response.EligibilityCriteria = &criteria
	}

	return &response, nil
}
