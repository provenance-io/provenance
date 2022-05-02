package keeper

// import (
// 	"fmt"

// 	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
// 	sdk "github.com/cosmos/cosmos-sdk/types"

// 	epochtypes "github.com/provenance-io/provenance/x/epoch/types"
// 	"github.com/provenance-io/provenance/x/reward/types"
// )

// // HandleAddRewardProposal handles an Add reward program governance proposal request
// func HandleAddRewardProgramProposal(ctx sdk.Context, k Keeper, proposal *types.AddRewardProgramProposal, registry codectypes.InterfaceRegistry) error {
// 	if err := proposal.ValidateBasic(); err != nil {
// 		return err
// 	}
// 	epochInfo := k.EpochKeeper.GetEpochInfo(ctx, proposal.EpochId)
// 	if (epochInfo == epochtypes.EpochInfo{}) {
// 		return fmt.Errorf("invalid epoch identifier: %s", proposal.EpochId)
// 	}

// 	// calculate the start epoch from current epoch + proposal offset height
// 	startEpoch := uint64(epochInfo.CurrentEpoch) + proposal.EpochStartOffset

// 	rewardProgram := types.NewRewardProgram(proposal.RewardProgramId,
// 		proposal.DistributeFromAddress,
// 		proposal.Coin,
// 		proposal.Coin,
// 		proposal.EpochId,
// 		startEpoch,
// 		proposal.NumberEpochs,
// 		proposal.EligibilityCriteria,
// 		false,
// 		proposal.Minimum,
// 		proposal.Maximum,
// 	)
// 	err := rewardProgram.ValidateBasic()
// 	if err != nil {
// 		return err
// 	}

// 	k.SetRewardProgram(ctx, rewardProgram)

// 	acc, _ := sdk.AccAddressFromBech32(rewardProgram.DistributeFromAddress)
// 	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, acc, types.ModuleName, sdk.NewCoins(rewardProgram.Coin))
// 	if err != nil {
// 		return fmt.Errorf("unable to send coin to module reward pool: %s", err)
// 	}
// 	//TODO: Add object to track all balances in the module

// 	ctx.Logger().Info(fmt.Sprintf("NOTICE: Reward Program Proposal %v", rewardProgram))
// 	return nil
// }
