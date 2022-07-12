package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"

	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

const (
	// OpWeightAddMarkerProposal app params key for add marker proposal
	//nolint:gosec // not credentials
	OpWeightAddMarkerProposal = "op_weight_add_marker_proposal"
	// OpWeightSupplyIncreaseProposal app params key for supply increase proposal
	OpWeightSupplyIncreaseProposal = "op_weight_supply_increase_proposal"
	// OpWeightSupplyDecreaseProposal app params key for supply decrease proposal
	OpWeightSupplyDecreaseProposal = "op_weight_supply_decrease_proposal"
	// OpWeightSetAdministratorProposal app params key for set administrator proposal
	//nolint:gosec // not credentials
	OpWeightSetAdministratorProposal = "op_weight_set_administrator_proposal"
	// OpWeightRemoveAdministratorProposal app params key for remove administrator proposal
	OpWeightRemoveAdministratorProposal = "op_weight_remove_administrator_proposal"
	// OpWeightChangeStatusProposal app params key for change status proposal
	OpWeightChangeStatusProposal = "op_weight_change_status_proposal"
	// OpWeightSetDenomMetadataProposal app params key for change status proposal
	//nolint:gosec // not credentials
	OpWeightSetDenomMetadataProposal = "op_weight_set_denom_metadata"
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
		simulation.NewWeightedProposalContent(
			OpWeightSetDenomMetadataProposal,
			simappparams.DefaultWeightSetDenomMetadataProposalContent,
			SimulateSetDenomMetadataProposalContent(k),
		),
	}
}

// SimulateCreateAddMarkerProposalContent generates random create marker proposal content
func SimulateCreateAddMarkerProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		return types.NewAddMarkerProposal(
			simtypes.RandStringOfLength(r, 5),  // title
			simtypes.RandStringOfLength(r, 10), // description
			randomUnrestrictedDenom(r, k.GetUnrestrictedDenomRegex(ctx)),
			sdk.NewIntFromUint64(randomUint64(r, k.GetMaxTotalSupply(ctx))), // initial supply
			simAccount.Address,              // manager
			types.MarkerStatus(r.Intn(3)+1), // initial status (proposed, finalized, active)
			types.MarkerType(r.Intn(2)+1),   // coin or restricted_coin
			[]types.AccessGrant{{Address: simAccount.Address.String(), Permissions: randomAccessTypes(r)}},
			r.Intn(2) > 0, // fixed supply
			r.Intn(2) > 0, // allow gov
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
		if m == nil || !m.HasGovernanceEnabled() || m.GetStatus() > types.StatusActive {
			return nil
		}
		return types.NewSupplyIncreaseProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			sdk.NewCoin(m.GetDenom(), sdk.NewIntFromUint64(randomUint64(r, k.GetMaxTotalSupply(ctx)-k.CurrentCirculation(ctx, m).Uint64()))),
			dest,
		)
	}
}

// SimulateCreateSupplyDecreaseProposalContent generates random create-root-name proposal content
func SimulateCreateSupplyDecreaseProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		m := randomMarker(r, ctx, k)
		if m == nil || !m.HasGovernanceEnabled() || m.GetStatus() > types.StatusActive {
			return nil
		}
		currentSupply := k.CurrentEscrow(ctx, m).AmountOf(m.GetDenom())
		if currentSupply.LT(sdk.OneInt()) {
			return nil
		}
		burn := sdk.NewCoin(m.GetDenom(), simtypes.RandomAmount(r, currentSupply))
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
		if m == nil || !m.HasGovernanceEnabled() {
			return nil
		}
		return types.NewSetAdministratorProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			m.GetDenom(),
			randomAccessGrants(r, accs, 2),
		)
	}
}

// SimulateCreateRemoveAdministratorProposalContent generates random create-root-name proposal content
func SimulateCreateRemoveAdministratorProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		m := randomMarker(r, ctx, k)
		if m == nil || !m.HasGovernanceEnabled() {
			return nil
		}
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
		if m == nil || !m.HasGovernanceEnabled() || m.GetStatus() > types.StatusCancelled {
			return nil
		}
		return types.NewChangeStatusProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			m.GetDenom(),
			m.GetStatus()+1,
		)
	}
}

// SimulateSetDenomMetadataProposalContent generates random set denom metadata proposal content
func SimulateSetDenomMetadataProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		m := randomMarker(r, ctx, k)
		if m == nil || !m.HasGovernanceEnabled() {
			return nil
		}

		// Create 1 to 4 denom unit entries with 0 to 2 aliases.
		display := m.GetDenom()
		denomUnits := make([]*banktypes.DenomUnit, r.Intn(4)+1)
		for di := range denomUnits {
			du := banktypes.DenomUnit{
				Denom:    simtypes.RandStringOfLength(r, 20),
				Exponent: 0,
				Aliases:  make([]string, r.Intn(3)),
			}
			if di == 0 {
				// First entry needs to have the same denom as the metadata base.
				du.Denom = m.GetDenom()
			} else {
				// First entry needs exponent 0, then they need to go up from there.
				du.Exponent = denomUnits[di-1].Exponent + uint32(r.Intn(4)) + 1
			}
			// Actually set the aliases to something.
			for ai := range du.Aliases {
				du.Aliases[ai] = simtypes.RandStringOfLength(r, 20)
			}
			// Randomly decide to use this one as the display
			if r.Intn(4) == 0 {
				display = du.Denom
			}
			denomUnits[di] = &du
		}

		return types.NewSetDenomMetadataProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			banktypes.Metadata{
				Description: simtypes.RandStringOfLength(r, 100),
				Base:        m.GetDenom(),
				Display:     display,
				Name:        simtypes.RandStringOfLength(r, 20),
				Symbol:      simtypes.RandStringOfLength(r, 5),
				DenomUnits:  denomUnits,
			},
		)
	}
}
