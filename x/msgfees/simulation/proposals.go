package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/msgfees/keeper"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

const (
	// OpWeightAddMsgFeesProposal add msg fees proposal
	//nolint:gosec
	OpWeightAddMsgFeesProposal    = "op_weight_add_msg_based_fees_proposal"
	OpWeightRemoveMsgFeesProposal = "op_weight_remove_msg_based_fees_proposal"
)

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightAddMsgFeesProposal,
			simappparams.DefaultWeightAddMsgFeeProposalContent,
			SimulateCreateAddMsgFeesProposal(k),
		),
		simulation.NewWeightedProposalContent(
			OpWeightRemoveMsgFeesProposal,
			simappparams.DefaultWeightRemoveMsgFeeProposalContent,
			SimulateCreateRemoveMsgFeesProposal(k),
		),
	}
}

// SimulateCreateAddMsgFeesProposal generates random additional fee with AddMsgFeesProposal
func SimulateCreateAddMsgFeesProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		msgFeeExists, err := k.GetMsgFee(ctx, sdk.MsgTypeURL(&attributetypes.MsgAddAttributeRequest{}))
		check(err)
		if msgFeeExists == nil {
			return types.NewAddMsgFeeProposal(
				simtypes.RandStringOfLength(r, 10),
				simtypes.RandStringOfLength(r, 100),
				sdk.MsgTypeURL(&attributetypes.MsgAddAttributeRequest{}),
				sdk.NewCoin("hotdog", sdk.NewInt(r.Int63n(100000000))),
			)
		}
		return types.NewUpdateMsgFeeProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			sdk.MsgTypeURL(&attributetypes.MsgAddAttributeRequest{}),
			sdk.NewCoin("hotdog", sdk.NewInt(r.Int63n(100000000))),
		)
	}
}

// SimulateCreateRemoveMsgFeesProposal generates random removal of additional fee with RemoveMsgFeesProposal
func SimulateCreateRemoveMsgFeesProposal(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		msgFeeExists, err := k.GetMsgFee(ctx, sdk.MsgTypeURL(&attributetypes.MsgAddAttributeRequest{}))
		check(err)
		if msgFeeExists != nil {
			return types.NewRemoveMsgFeeProposal(
				simtypes.RandStringOfLength(r, 10),
				simtypes.RandStringOfLength(r, 100),
				sdk.MsgTypeURL(&attributetypes.MsgAddAttributeRequest{}),
			)
		}

		return nil
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
