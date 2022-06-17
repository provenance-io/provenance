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
