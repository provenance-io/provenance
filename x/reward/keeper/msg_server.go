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

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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
	err = rewardProgram.ValidateBasic()
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
	response := types.MsgClaimRewardResponse{}

	rewardProgram, err := s.Keeper.GetRewardProgram(ctx, req.GetRewardProgramId())
	if err != nil || rewardProgram.ValidateBasic() != nil {
		return nil, fmt.Errorf("reward program %d does not exist", req.GetRewardProgramId())
	}

	// 1.) gathers all the claimable awards for address and completed claim period
	response.RewardProgramId = req.GetRewardProgramId()
	rewards := s.Keeper.ClaimRewards(ctx, rewardProgram, req.GetDistributeToAddress())

	// 2.) sums the total reward coins from claim periods
	for i := 0; i < len(rewards); i++ {
		reward := rewards[i]
		response.TotalRewardClaim.Denom = reward.GetClaimPeriodReward().Denom
		response.TotalRewardClaim = response.TotalRewardClaim.Add(reward.GetClaimPeriodReward())
		response.ClaimedRewardPeriodDetails = append(response.ClaimedRewardPeriodDetails, &reward)
	}

	// 3.) sends total coins from module escrow to address
	acc, err := sdk.AccAddressFromBech32(rewardProgram.DistributeFromAddress)
	if err != nil {
		return nil, fmt.Errorf("not a valid address: %s", err)
	}
	err = s.Keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acc, sdk.NewCoins(rewardProgram.ClaimedAmount))
	if err != nil {
		return nil, fmt.Errorf("unable to send coin from module to account: %s", err)
	}

	// 		a.) will need to update reward program with total claimed funds
	rewardProgram.ClaimedAmount = rewardProgram.ClaimedAmount.Add(response.TotalRewardClaim)
	s.Keeper.SetRewardProgram(ctx, rewardProgram)

	// 5.) emit event of claims, possibly a typed event
	if len(rewards) > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeClaimRewards,
				sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", req.RewardProgramId)),
				sdk.NewAttribute(types.AttributeKeyRewardsClaimAddress, req.DistributeToAddress),
			),
		)
	}

	// 4.) returns details of claim periods and total funds to be populated in MsgClaimRewardResponse
	return &response, nil
}
