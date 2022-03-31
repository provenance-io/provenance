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

	epochInfo := k.epochKeeper.GetEpochInfo(ctx, proposal.RewardProgram.EpochId)
	if (epochInfo == epochtypes.EpochInfo{}) {
		return fmt.Errorf("invalid epoch identifier: %s", proposal.RewardProgram.EpochId)
	}

	k.SetRewardProgram(ctx, *proposal.RewardProgram)

	return nil
}
