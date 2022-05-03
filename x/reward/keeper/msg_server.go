package keeper

import (
	"context"
	"fmt"
	"time"

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

func (s msgServer) CreateRewardProgram(goCtx context.Context, msg *types.MsgCreateRewardProgramRequest) (*types.MsgCreateRewardProgramResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	rewardProgram := types.NewRewardProgram(msg.RewardProgramId,
		msg.DistributeFromAddress,
		msg.Coin,
		msg.MaxRewardByAddress,
		msg.ProgramStartTime,
		time.Now(), //TODO: Calculate end for next epoch .  Programstart time + epoch type
		msg.EpochType,
		msg.NumberEpochs,
		msg.EligibilityCriteria,
		false,
	)
	err := rewardProgram.ValidateBasic()
	if err != nil {
		return nil, err
	}

	s.Keeper.SetRewardProgram(ctx, rewardProgram)

	acc, _ := sdk.AccAddressFromBech32(rewardProgram.DistributeFromAddress)
	err = s.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, acc, types.ModuleName, sdk.NewCoins(rewardProgram.Coin))
	if err != nil {
		return nil, fmt.Errorf("unable to send coin to module reward pool: %s", err)
	}
	//TODO: Add object to track all balances in the module

	ctx.Logger().Info(fmt.Sprintf("NOTICE: Reward Program Proposal %v", rewardProgram))
	return nil, nil
}
