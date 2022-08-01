package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the account MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateRewardProgram creates new reward program from msg
func (s msgServer) CreateRewardProgram(goCtx context.Context, msg *types.MsgCreateRewardProgramRequest) (*types.MsgCreateRewardProgramResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rewardProgramID, err := s.Keeper.GetNextRewardProgramID(ctx)
	if err != nil {
		return &types.MsgCreateRewardProgramResponse{}, err
	}

	claimPeriodDaysInSeconds := uint64(types.DayInSeconds) * msg.GetClaimPeriodDays()
	experationOffsetInSeconds := uint64(types.DayInSeconds) * msg.GetExpireDays()

	rewardProgram := types.NewRewardProgram(
		msg.Title,
		msg.Description,
		rewardProgramID,
		msg.DistributeFromAddress,
		msg.TotalRewardPool,
		msg.MaxRewardPerClaimAddress,
		msg.ProgramStartTime.UTC(),
		claimPeriodDaysInSeconds,
		msg.ClaimPeriods,
		msg.MaxRolloverClaimPeriods,
		experationOffsetInSeconds,
		msg.QualifyingActions,
	)
	err = s.Keeper.CreateRewardProgram(ctx, rewardProgram)
	if err != nil {
		return &types.MsgCreateRewardProgramResponse{}, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardProgramCreated,
			sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", rewardProgramID)),
		),
	)

	return &types.MsgCreateRewardProgramResponse{Id: rewardProgramID}, nil
}

// EndRewardProgram ends reward program from msg
func (s msgServer) EndRewardProgram(goCtx context.Context, msg *types.MsgEndRewardProgramRequest) (*types.MsgEndRewardProgramResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rewardProgram, err := s.Keeper.GetRewardProgram(ctx, msg.RewardProgramId)
	if err != nil {
		return &types.MsgEndRewardProgramResponse{}, err
	}
	if rewardProgram.DistributeFromAddress != msg.ProgramOwnerAddress {
		return &types.MsgEndRewardProgramResponse{}, fmt.Errorf("%v not authorized is not program owner", msg.ProgramOwnerAddress)
	}
	if rewardProgram.State != types.RewardProgram_STATE_PENDING && rewardProgram.State != types.RewardProgram_STATE_STARTED {
		return &types.MsgEndRewardProgramResponse{}, fmt.Errorf("cannot end program in state ended or expired state")
	}

	s.Keeper.EndingRewardProgram(ctx, rewardProgram)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardProgramEnded,
			sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", rewardProgram.Id)),
		),
	)

	return &types.MsgEndRewardProgramResponse{}, nil
}

func (s msgServer) ClaimRewards(goCtx context.Context, req *types.MsgClaimRewardsRequest) (*types.MsgClaimRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	details, reward, err := s.Keeper.ClaimRewards(ctx, req.GetRewardProgramId(), req.GetRewardAddress())
	if err != nil {
		return nil, err
	}

	if len(details) > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeClaimRewards,
				sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", req.RewardProgramId)),
				sdk.NewAttribute(types.AttributeKeyRewardsClaimAddress, req.GetRewardAddress()),
			),
		)
	}

	return &types.MsgClaimRewardsResponse{
		ClaimDetails: types.RewardProgramClaimDetail{
			RewardProgramId:            req.GetRewardProgramId(),
			TotalRewardClaim:           reward,
			ClaimedRewardPeriodDetails: details,
		},
	}, nil
}

func (s msgServer) ClaimAllRewards(goCtx context.Context, req *types.MsgClaimAllRewardsRequest) (*types.MsgClaimAllRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	details, reward, err := s.Keeper.ClaimAllRewards(ctx, req.GetRewardAddress())
	if err != nil {
		return nil, err
	}

	var programIDs []uint64
	for _, detail := range details {
		programIDs = append(programIDs, detail.GetRewardProgramId())
	}

	if len(details) > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeClaimAllRewards,
				sdk.NewAttribute(types.AttributeKeyRewardProgramIDs, fmt.Sprintf("%v", programIDs)),
				sdk.NewAttribute(types.AttributeKeyRewardsClaimAddress, req.GetRewardAddress()),
			),
		)
	}

	return &types.MsgClaimAllRewardsResponse{
		TotalRewardClaim: reward,
		ClaimDetails:     details,
	}, nil
}
