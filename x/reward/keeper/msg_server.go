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

	rewardProgramID, err := s.Keeper.GetRewardProgramID(ctx)
	if err != nil {
		return &types.MsgCreateRewardProgramResponse{}, err
	}

	if ctx.BlockTime().UTC().After(msg.ProgramStartTime.UTC()) {
		return &types.MsgCreateRewardProgramResponse{},
			fmt.Errorf("start time is before current block time %v : %v ", ctx.BlockTime().UTC(), msg.ProgramStartTime.UTC())
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
		msg.ProgramStartTime,
		claimPeriodDaysInSeconds,
		msg.ClaimPeriodDays,
		experationOffsetInSeconds,
		msg.QualifyingActions,
	)
	err = rewardProgram.Validate()
	if err != nil {
		return nil, err
	}

	s.Keeper.SetRewardProgram(ctx, rewardProgram)
	s.Keeper.SetRewardProgramID(ctx, rewardProgramID+1)

	acc, _ := sdk.AccAddressFromBech32(rewardProgram.DistributeFromAddress)
	err = s.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, acc, types.ModuleName, sdk.NewCoins(rewardProgram.TotalRewardPool))
	if err != nil {
		return nil, fmt.Errorf("unable to send coin to module reward pool: %s", err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardProgramCreated,
			sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", rewardProgramID)),
		),
	)

	return &types.MsgCreateRewardProgramResponse{Id: rewardProgramID}, nil
}

func (s msgServer) ClaimRewards(goCtx context.Context, req *types.MsgClaimRewardRequest) (*types.MsgClaimRewardResponse, error) {
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

	return &types.MsgClaimRewardResponse{
		RewardProgramId:            req.GetRewardProgramId(),
		TotalRewardClaim:           reward,
		ClaimedRewardPeriodDetails: details,
	}, nil
}
