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

	rewardprogramID, err := s.Keeper.GetRewardProgramID(ctx)
	if err != nil {
		return &types.MsgCreateRewardProgramResponse{}, err
	}

	// TODO: get next epoch time by taking in day, week, month value and convert it to seconds for creating reward program

	rewardProgram := types.NewRewardProgram(
		msg.Title,
		msg.Description,
		rewardprogramID,
		msg.DistributeFromAddress,
		msg.Coin,
		msg.MaxRewardByAddress,
		msg.ProgramStartTime,
		60, // TODO : add a calculation from a type of day, week, month...
		msg.NumberEpochs,
		msg.EligibilityCriteria,
	)
	err = rewardProgram.Validate()
	if err != nil {
		return nil, err
	}

	s.Keeper.SetRewardProgram(ctx, rewardProgram)
	s.Keeper.SetRewardProgramID(ctx, rewardprogramID+1)

	acc, _ := sdk.AccAddressFromBech32(rewardProgram.DistributeFromAddress)
	err = s.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, acc, types.ModuleName, sdk.NewCoins(rewardProgram.Coin))
	if err != nil {
		return nil, fmt.Errorf("unable to send coin to module reward pool: %s", err)
	}
	//TODO: Add object to track all balances in the module

	ctx.Logger().Info(fmt.Sprintf("NOTICE: Reward Program Proposal %v", rewardProgram))
	// TODO: emit event
	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		types.EventTypeSubmitRewardProgram,
	// 		sdk.NewAttribute(types.AttributeKeyRewardProgramID, fmt.Sprintf("%d", rewardprogramID)),
	// 	),
	// )

	return nil, nil
}
