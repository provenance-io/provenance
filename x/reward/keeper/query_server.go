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

// returns all ClaimPeriodRewardDistributions
func (k Keeper) ClaimPeriodRewardDistributions(ctx context.Context, req *types.ClaimPeriodRewardDistributionRequest) (*types.ClaimPeriodRewardDistributionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.ClaimPeriodRewardDistributionResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rewardDistributions, err := k.GetAllClaimPeriodRewardDistributions(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to query epoch reward distributions: %v", err))
	}
	response.ClaimPeriodRewardDistribution = rewardDistributions

	return &response, nil
}

// returns a ClaimPeriodRewardDistribution by rewardId and epochId
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
