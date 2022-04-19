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

	var rewardPrograms []types.RewardProgram
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := k.IterateRewardPrograms(sdkCtx, func(rewardProgram types.RewardProgram) (stop bool) {
		rewardPrograms = append(rewardPrograms, rewardProgram)
		return false
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to iterate reward programs: %v", err))
	}

	return &types.RewardProgramsResponse{RewardPrograms: rewardPrograms}, nil
}

func (k Keeper) ActiveRewardPrograms(context.Context, *types.ActiveRewardProgramsRequest) (*types.ActiveRewardProgramsResponse, error) {
	return &types.ActiveRewardProgramsResponse{}, nil
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
	rewardProgram, err := k.GetRewardProgram(sdkCtx, int64(req.GetId()))
	if err != nil {
		return &response, err
	}

	// 0 is not a valid id. This means the program was not found
	if rewardProgram.Id != 0 {
		response.RewardProgram = &rewardProgram
	}

	return &response, nil
}
