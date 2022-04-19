package keeper

import (
	"fmt"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/reward/types"
)

type GovHooks struct {
	k Keeper
}

var _ govtypes.GovHooks = GovHooks{}

func (k Keeper) GetGovHooks() govtypes.GovHooks {
	return GovHooks{k}
}

// AfterProposalSubmission - call hook if registered
func (gh GovHooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	gh.k.AfterProposalSubmission(ctx, proposalID)
}

func (k Keeper) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	ctx.Logger().Info(fmt.Sprintf("AfterProposalSubmission: %v ... moving reward coins to escrow account", proposalID))
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if found {
		ctx.Logger().Info(fmt.Sprintf("AfterProposalSubmission: %v Content: %v", proposalID, string(proposal.Content.Value)))
		var rewardProgram types.AddRewardProgramProposal
		if err := k.cdc.Unmarshal(proposal.Content.Value, &rewardProgram); err != nil {
			ctx.Logger().Info(fmt.Sprintf("AfterProposalSubmission: %v was not an AddRewardProgramProposal: %v ", proposalID, proposal.Content.TypeUrl))
			return
		}

		// rewardPoolPath := fmt.Sprintf("%s/pool/%v", types.ModuleName, rewardProgram.RewardProgramId)
		// acc := k.authkeeper.GetModuleAccount(ctx, rewardPoolPath)
		// if acc == nil {
		// 	acc = authtypes.NewEmptyModuleAccount(rewardPoolPath)
		// 	k.authkeeper.SetModuleAccount(ctx, authtypes.NewEmptyModuleAccount(rewardPoolPath))
		// }
		// if err := k.bankKeeper.SendCoins(ctx, sdk.AccAddress(rewardProgram.DistributeFromAddress), acc.GetAddress(), sdk.NewCoins(rewardProgram.Coin)); err != nil {
		// 	return
		// }
	}
}

// AfterProposalDeposit - call hook if registered
func (gh GovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
	ctx.Logger().Info("AfterProposalDeposit")
}

// AfterProposalVote - call hook if registered
func (gh GovHooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	ctx.Logger().Info("AfterProposalVote")
}

// AfterProposalFailedMinDeposit - call hook if registered
func (gh GovHooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
	ctx.Logger().Info("AfterProposalFailedMinDeposit")
}

// AfterProposalVotingPeriodEnded - call hook if registered
func (gh GovHooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	gh.k.AfterProposalVotingPeriodEnded(ctx, proposalID)
}

// AfterProposalVotingPeriodEnded - call hook if registered
func (k Keeper) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	ctx.Logger().Info(fmt.Sprintf("AfterProposalVotingPeriodEnded: %v ...", proposalID))
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if found {
		ctx.Logger().Info(fmt.Sprintf("AfterProposalVotingPeriodEnded: %v Content: %v", proposalID, string(proposal.Content.Value)))
		var rewardProgram types.AddRewardProgramProposal
		if err := k.cdc.Unmarshal(proposal.Content.Value, &rewardProgram); err != nil {
			ctx.Logger().Info(fmt.Sprintf("AfterProposalVotingPeriodEnded: %v was not an AddRewardProgramProposal: %v ", proposalID, proposal.Content.TypeUrl))
			return
		}

		//TODO: check if proposal passed, if not return funds to distributed_from_address
	}
}
