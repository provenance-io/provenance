package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/provenance-io/provenance/x/ibcratelimit/keeper"
)

// Simulation operation weights constants
const (
	//nolint:gosec // not credentials
	OpWeightMsgUpdateParams = "op_weight_msg_update_params"
)

// ProposalMsgs returns all the governance proposal messages.
func ProposalMsgs(simState module.SimulationState, k *keeper.Keeper) []simtypes.WeightedProposalMsg {
	var wMsgUpdateParams int

	simState.AppParams.GetOrGenerate(OpWeightMsgUpdateParams, &wMsgUpdateParams, nil,
		func(_ *rand.Rand) { wMsgUpdateParams = simappparams.DefaultWeightIBCRLUpdateParams })

	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(OpWeightMsgUpdateParams, wMsgUpdateParams, SimulatePropMsgUpdateParams(k)),
	}
}

func SimulatePropMsgUpdateParams(k *keeper.Keeper) simtypes.MsgSimulatorFn {
	return func(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
		// change it to a new random account.
		raccs := simtypes.RandomAccounts(r, 1)
		return ibcratelimit.NewMsgUpdateParamsRequest(k.GetAuthority(), raccs[0].Address.String())
	}
}
