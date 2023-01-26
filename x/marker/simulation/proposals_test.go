package simulation_test

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/simulation"
	"github.com/provenance-io/provenance/x/marker/types"
	"github.com/stretchr/testify/require"
)

func TestProposalContents(t *testing.T) {
	app := simapp.Setup(t)

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(keeper.NewKeeper(app.AppCodec(), app.GetKey(types.ModuleName), app.GetSubspace(types.ModuleName), app.AccountKeeper, app.BankKeeper, app.AuthzKeeper, app.FeeGrantKeeper, app.TransferKeeper, app.GetKey(banktypes.StoreKey)))
	require.Len(t, weightedProposalContent, 6)
}
