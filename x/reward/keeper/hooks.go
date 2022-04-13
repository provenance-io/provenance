package keeper

import (
	"fmt"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochtypes "github.com/provenance-io/provenance/x/epoch/types"
	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber uint64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber uint64) error {
	// distribute logic goes here, i.e record the number of shares claimable in that epoch and the total rewards pool
	// also unlock the module account?
	ctx.Logger().Info(fmt.Sprintf("In epoch end for %s %d", epochIdentifier, epochNumber))
	rewardPrograms, err := k.GetAllActiveRewardsForEpoch(ctx, epochIdentifier, epochNumber)
	if err != nil {
		return err
	}

	// only rewards programs who are eligible will be iterated through here
	for _, rewardProgram := range rewardPrograms {
		epochRewardDistibutionForEpoch, err := k.GetEpochRewardDistribution(ctx, epochIdentifier, rewardProgram.Id)
		if err != nil {
			return err
		}
		// epoch reward distribution does it exist till the block has ended, highly unlikely but could happen
		if epochRewardDistibutionForEpoch.EpochId == "" {
			epochRewardDistibutionForEpoch.EpochId = epochIdentifier
			epochRewardDistibutionForEpoch.RewardProgramId = rewardProgram.Id
			epochRewardDistibutionForEpoch.TotalShares = 0
			epochRewardDistibutionForEpoch.TotalRewardsPool = rewardProgram.Coin
			if rewardProgram.StartEpoch+rewardProgram.NumberEpochs == epochNumber {
				epochRewardDistibutionForEpoch.EpochEnded = true
			}
			k.EvaluateRules(ctx, epochNumber, rewardProgram, epochRewardDistibutionForEpoch)
			// TODO if shares are still 0 for epochRewardDistibutionForEpoch.TotalShares return all the rewards?
		} else {
			// end the epoch
			// because the start period is also included in the calculation, hence subtracting -1
			// for e.g a reward that begins on epoch 1 and ends in 10 epochs should end in epoch 10
			if (rewardProgram.StartEpoch + rewardProgram.NumberEpochs - 1) == epochNumber {
				epochRewardDistibutionForEpoch.EpochEnded = true
				k.EvaluateRules(ctx, epochNumber, rewardProgram, epochRewardDistibutionForEpoch)
			}
		}
	}

	return nil
}

func (k Keeper) GetAllActiveRewardsForEpoch(ctx sdk.Context, epochIdentifier string, epochNumber uint64) ([]types.RewardProgram, error) {
	var rewardPrograms []types.RewardProgram
	// get all the rewards programs
	err := k.IterateRewardPrograms(ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		// this is epoch that ended, and matches up with the reward program identifier
		// check if any of the events match with any of the reward program running
		// e.g start epoch,current epoch .. start epoch + number of epochs program runs for > current epoch
		// 1,1 .. 1+4 > 1
		// 1,2 .. 1+4 > 2
		// 1,3 .. 1+4 > 3
		// 1,4 .. 1+4 > 4
		if rewardProgram.EpochId == epochIdentifier && epochNumber > rewardProgram.StartEpoch && rewardProgram.StartEpoch+rewardProgram.NumberEpochs > epochNumber {
			rewardPrograms = append(rewardPrograms, rewardProgram)
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return rewardPrograms, nil
}

func (k Keeper) CheckActiveDelegations(ctx sdk.Context, address sdk.AccAddress) []stakingtypes.Delegation {
	return k.stakingKeeper.GetAllDelegatorDelegations(ctx, address)
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

type GovHooks struct {
	k Keeper
}

var _ epochtypes.EpochHooks = Hooks{}
var _ govtypes.GovHooks = GovHooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (k Keeper) GetGovHooks() govtypes.GovHooks {
	return GovHooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber uint64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber uint64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
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

		// TODO: Escrow funds into a pool from distributed_from_address to module
		// rewardPoolPath := fmt.Sprintf("%s/pool/%v", types.ModuleName, rewardProgram.RewardProgramId)

		// if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.AccAddress(rewardProgram.DistributeFromAddress), rewardPoolPath, sdk.NewCoins(rewardProgram.Coin)); err != nil {
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
