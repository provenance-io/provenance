package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"

	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
)

// OpWeightSubmitCreateRootNameProposal app params key for create root name proposal
const OpWeightSubmitCreateRootNameProposal = "op_weight_submit_create_root_name_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitCreateRootNameProposal,
			simappparams.DefaultWeightCreateRootNameProposal,
			SimulateCreateRootNameProposalContent(k),
		),
	}
}

// SimulateCreateRootNameProposalContent generates random create-root-name proposal content
func SimulateCreateRootNameProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		restricted := simtypes.RandIntBetween(r, 1, 100) > 50

		return types.NewCreateRootNameProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			simtypes.RandStringOfLength(r, 10),
			simAccount.Address,
			restricted,
		)
	}
}
