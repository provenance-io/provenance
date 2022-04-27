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

func (k Keeper) ActiveRewardPrograms(ctx context.Context, req *types.ActiveRewardProgramsRequest) (*types.ActiveRewardProgramsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	response := types.ActiveRewardProgramsResponse{}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	rewardPrograms, err := k.GetAllActiveRewards(sdkCtx)
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
	rewardProgram, err := k.GetRewardProgram(sdkCtx, int64(req.GetId()))
	if err != nil {
		return &response, err
	}

	// 0 is not a valid id. This means the program was not found
	// TODO - Since we are using the Null Object Pattern we may want extract the logic into an IsValid
	// TODO - !keeper.RewardIsValid(rewardProgram)
	if rewardProgram.Id != 0 {
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
	err := k.IterateRewardClaims(sdkCtx, func(rewardClaim types.RewardClaim) (stop bool) {
		response.RewardClaims = append(response.RewardClaims, rewardClaim)
		return false
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to iterate reward programs: %v", err))
	}

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
		return &response, err
	}

	// "" is not a valid address. This means the program was not found
	// TODO - Since we are using the Null Object Pattern we may want extract the logic into an IsValid
	// TODO - !keeper.ClaimIsValid(claim)
	if claim.Address != "" {
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
	rewardDistributions := []types.EpochRewardDistribution{}
	err := k.IterateEpochRewardDistributions(sdkCtx, func(rewardDistribution types.EpochRewardDistribution) (stop bool) {
		rewardDistributions = append(rewardDistributions, rewardDistribution)
		return false
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to obtain epoch reward distributions: %v", err))
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
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to obtain epoch reward distributions: %v", err))
	}

	// 0 is not a valid id. This means the program was not found
	// TODO - Since we are using the Null Object Pattern we may want extract the logic into an IsValid
	// TODO - !keeper.EpochRewardIsValid(epochReward)
	if epochReward.RewardProgramId != 0 {
		response.EpochRewardDistribution = &epochReward
	}

	return &response, nil
}

// returns all EligibilityCriterias
func (k Keeper) EligibilityCriteria(context.Context, *types.EligibilityCriteriaRequest) (*types.EligibilityCriteriaResponse, error) {
	response := types.EligibilityCriteriaResponse{}
	return &response, nil
}

// returns a EligibilityCriteria by name
func (k Keeper) EligibilityCriteriaByName(context.Context, *types.EligibilityCriteriaRequestByNameRequest) (*types.EligibilityCriteriaRequestByNameResponse, error) {
	response := types.EligibilityCriteriaRequestByNameResponse{}
	return &response, nil
}
