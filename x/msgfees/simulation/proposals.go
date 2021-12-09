package simulation

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simappparams "github.com/provenance-io/provenance/app/params"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/msgfees/keeper"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"math/rand"
)

const (
	// OpWeightAddMsgBasedFeesProposal add msg based fees proposal
	OpWeightAddMsgBasedFeesProposal = "op_weight_add_msg_based_fees_proposal"
	// OpWeightUpdateMsgBasedFeesProposal update msg based fees proposal
	OpWeightUpdateMsgBasedFeesProposal = "op_weight_add_msg_based_fees_proposal"
	// OpWeightRemoveMsgBasedFeesProposal remove msg based fees proposal
	OpWeightRemoveMsgBasedFeesProposal = "op_weight_add_msg_based_fees_proposal"
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightAddMsgBasedFeesProposal,
			simappparams.DefaultWeightAddMarkerProposalContent,
			SimulateCreateAddMsgBasedFeesProposal(k),
		),
		//simulation.NewWeightedProposalContent(
		//	OpWeightUpdateMsgBasedFeesProposal,
		//	simappparams.DefaultWeightAddMarkerProposalContent,
		//	SimulateCreateAddMarkerProposalContent(k),
		//),
		//simulation.NewWeightedProposalContent(
		//	OpWeightRemoveMsgBasedFeesProposal,
		//	simappparams.DefaultWeightAddMarkerProposalContent,
		//	SimulateCreateAddMarkerProposalContent(k),
		//),
	}
}

// SimulateCreateAddMsgBasedFeesProposal generates random additional fee for AddMsgBasedFeesProposal
func SimulateCreateAddMsgBasedFeesProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {

		   msgFeeExists,err := k.GetMsgBasedFee(ctx,sdk.MsgTypeURL(&markertypes.MsgAddMarkerRequest{}))
		   check(err)
			addMarkerRequest, err := cdctypes.NewAnyWithValue(&markertypes.MsgAddMarkerRequest{})
			check(err)
			if msgFeeExists == nil {
				return types.NewAddMsgBasedFeeProposal(
					simtypes.RandStringOfLength(r, 10),
					simtypes.RandStringOfLength(r, 100),
					addMarkerRequest,
					sdk.NewCoin("hotdog", sdk.NewInt(r.Int63n(100000000))),
				)
			}else{
				return types.NewUpdateMsgBasedFeeProposal(
					simtypes.RandStringOfLength(r, 10),
					simtypes.RandStringOfLength(r, 100),
					addMarkerRequest,
					sdk.NewCoin("hotdog", sdk.NewInt(r.Int63n(100000000))),
				)
		}

	}
}

func check(err error){
	if err!=nil{
		panic(err)
	}
}
