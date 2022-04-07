package keeper

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/provenance-io/provenance/x/epoch/types"
	"github.com/provenance-io/provenance/x/reward/types"
)

// HandleAddMsgFeeProposal handles an Add msg fees governance proposal request
func HandleAddMsgFeeProposal(ctx sdk.Context, k Keeper, proposal *types.AddRewardProgramProposal, registry codectypes.InterfaceRegistry) error {
	if err := proposal.ValidateBasic(); err != nil {
		return err
	}
	epochInfo := k.EpochKeeper.GetEpochInfo(ctx, proposal.EpochId)
	if (epochInfo == epochtypes.EpochInfo{}) {
		return fmt.Errorf("invalid epoch identifier: %s", proposal.EpochId)
	}

	// calculate the start epoch height from current heigh + proposal offset height
	startEpoch := uint64(ctx.BlockHeight()) + proposal.EpochStartOffset

	rewardProgram := types.NewRewardProgram(proposal.RewardProgramId,
		proposal.DistributeFromAddress,
		proposal.Coin,
		proposal.Coin,
		proposal.EpochId,
		startEpoch,
		proposal.NumberEpochs,
		*proposal.EligibilityCriteria,
		false,
		1,
		2,
	)
	k.SetRewardProgram(ctx, rewardProgram)

	return nil
}
