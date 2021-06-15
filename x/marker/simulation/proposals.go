package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simappparams "github.com/provenance-io/provenance/app/params"

	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

const (
	// OpWeightAddMarkerProposal app params key for add marker proposal
	OpWeightAddMarkerProposal = "op_weight_add_marker_proposal"
	// OpWeightSupplyIncreaseProposal app params key for supply increase proposal
	OpWeightSupplyIncreaseProposal = "op_weight_supply_increase_proposal"
	// OpWeightSupplyDecreaseProposal app params key for supply decrease proposal
	OpWeightSupplyDecreaseProposal = "op_weight_supply_decrease_proposal"
	// OpWeightSetAdministratorProposal app params key for set administrator proposal
	OpWeightSetAdministratorProposal = "op_weight_set_administrator_proposal"
	// OpWeightRemoveAdministratorProposal app params key for remove administrator proposal
	OpWeightRemoveAdministratorProposal = "op_weight_remove_administrator_proposal"
	// OpWeightChangeStatusProposal app params key for change status proposal
	OpWeightChangeStatusProposal = "op_weight_change_status_proposal"
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightAddMarkerProposal,
			simappparams.DefaultWeightAddMarkerProposalContent,
			SimulateCreateAddMarkerProposalContent(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSupplyIncreaseProposal,
			simappparams.DefaultWeightSupplyIncreaseProposalContent,
			SimulateCreateSupplyIncreaseProposalContent(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSupplyDecreaseProposal,
			simappparams.DefaultWeightSupplyDecreaseProposalContent,
			SimulateCreateSupplyDecreaseProposalContent(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightSetAdministratorProposal,
			simappparams.DefaultWeightSetAdministratorProposalContent,
			SimulateCreateSetAdministratorProposalContent(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightRemoveAdministratorProposal,
			simappparams.DefaultWeightRemoveAdministratorProposalContent,
			SimulateCreateRemoveAdministratorProposalContent(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightChangeStatusProposal,
			simappparams.DefaultWeightChangeStatusProposalContent,
			SimulateCreateChangeStatusProposalContent(k),
		),
	}
}

// SimulateCreateAddMarkerProposalContent generates random create marker proposal content
func SimulateCreateAddMarkerProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		return types.NewAddMarkerProposal(
			simtypes.RandStringOfLength(r, 10),  // title
			simtypes.RandStringOfLength(r, 100), // description
			randomUnrestrictedDenom(r, k.GetUnrestrictedDenomRegex(ctx)),
			sdk.NewInt(r.Int63()),         // initial supply
			simAccount.Address,            // manager
			types.StatusActive,            // status
			types.MarkerType(r.Intn(1)+1), // coin or restricted_coin
			randomAccessGrants(r, accs),
			r.Intn(1) > 0, // fixed supply
			r.Intn(1) > 0, // allow gov
		)
	}
}

// SimulateCreateSupplyIncreaseProposalContent generates random increase marker supply proposal content
func SimulateCreateSupplyIncreaseProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		dest := ""
		if r.Intn(100) < 40 {
			acc, _ := simtypes.RandomAcc(r, accs)
			dest = acc.Address.String()
		}
		m := randomMarker(r, ctx, k)
		return types.NewSupplyIncreaseProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			sdk.NewCoin(m.GetDenom(), sdk.NewInt(r.Int63())),
			dest,
		)
	}
}

// SimulateCreateSupplyDecreaseProposalContent generates random create-root-name proposal content
func SimulateCreateSupplyDecreaseProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		m := randomMarker(r, ctx, k)
		burn := sdk.NewCoin(m.GetDenom(), sdk.NewInt(r.Int63()))
		if m.GetSupply().Amount.GT(sdk.ZeroInt()) {
			burn.Amount = simtypes.RandomAmount(r, m.GetSupply().Amount)

		}
		return types.NewSupplyDecreaseProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			burn,
		)
	}
}

// SimulateCreateSetAdministratorProposalContent generates random create-root-name proposal content
func SimulateCreateSetAdministratorProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		m := randomMarker(r, ctx, k)

		return types.NewSetAdministratorProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			m.GetDenom(),
			randomAccessGrants(r, accs),
		)
	}
}

// SimulateCreateRemoveAdministratorProposalContent generates random create-root-name proposal content
func SimulateCreateRemoveAdministratorProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		m := randomMarker(r, ctx, k)

		return types.NewRemoveAdministratorProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			m.GetDenom(),
			[]string{simAccount.Address.String()},
		)
	}
}

// SimulateCreateChangeStatusProposalContent generates random create-root-name proposal content
func SimulateCreateChangeStatusProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		m := randomMarker(r, ctx, k)

		return types.NewChangeStatusProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			m.GetDenom(),
			types.MarkerStatus(r.Intn(6)+1),
		)
	}
}
