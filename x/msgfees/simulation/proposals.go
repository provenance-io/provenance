package simulation

// import (
// 	"math/rand"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
// 	"github.com/cosmos/cosmos-sdk/x/simulation"

// 	simappparams "github.com/provenance-io/provenance/app/params"

// 	"github.com/provenance-io/provenance/x/msgfees/types"
// 	"github.com/provenance-io/provenance/x/msgfees/keeper"
// )

// const (
// 	// OpWeightAddMsgBasedFeesProposal add msg based fees proposal
// 	OpWeightAddMsgBasedFeesProposal = "op_weight_add_msg_based_fees_proposal"
// 	// OpWeightUpdateMsgBasedFeesProposal update msg based fees proposal
// 	OpWeightUpdateMsgBasedFeesProposal = "op_weight_add_msg_based_fees_proposal"
// 	// OpWeightRemoveMsgBasedFeesProposal remove msg based fees proposal
// 	OpWeightRemoveMsgBasedFeesProposal = "op_weight_add_msg_based_fees_proposal"
// )

// // ProposalContents defines the module weighted proposals' contents
// func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
// 	return []simtypes.WeightedProposalContent{
// 		simulation.NewWeightedProposalContent(
// 			OpWeightAddMsgBasedFeesProposal,
// 			simappparams.DefaultWeightAddMarkerProposalContent,
// 			SimulateCreateAddMsgBasedFeesProposal(k),
// 		),
// 		simulation.NewWeightedProposalContent(
// 			OpWeightUpdateMsgBasedFeesProposal,
// 			simappparams.DefaultWeightAddMarkerProposalContent,
// 			SimulateCreateAddMarkerProposalContent(k),
// 		),
// 		simulation.NewWeightedProposalContent(
// 			OpWeightRemoveMsgBasedFeesProposal,
// 			simappparams.DefaultWeightAddMarkerProposalContent,
// 			SimulateCreateAddMarkerProposalContent(k),
// 		),
// 	}
// }

// // SimulateCreateSupplyIncreaseProposalContent generates random increase marker supply proposal content
// func SimulateCreateAddMsgBasedFeesProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
// 	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
// 		dest := ""
// 		if r.Intn(100) < 40 {
// 			acc, _ := simtypes.RandomAcc(r, accs)
// 			dest = acc.Address.String()
// 		}
// 		// m := randomMarker(r, ctx, k)
// 		// if m == nil || !m.HasGovernanceEnabled() || m.GetStatus() > types.StatusActive {
// 		// 	return nil
// 		// }
// 		return types.NewAddMsgBasedFeesProposal(
// 			simtypes.RandStringOfLength(r, 10),
// 			simtypes.RandStringOfLength(r, 100),
// 			sdk.NewCoin(m.GetDenom(), sdk.NewIntFromUint64(randomUint64(r, k.GetMaxTotalSupply(ctx)-k.CurrentCirculation(ctx, m).Uint64()))),
// 			dest,
// 		)
// 	}
// }
