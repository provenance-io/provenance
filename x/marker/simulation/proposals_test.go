package simulation_test

import (
	"math/rand"
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	simapp "github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/simulation"
	"github.com/provenance-io/provenance/x/marker/types"
)

func TestProposalContents(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(keeper.NewKeeper(app.AppCodec(), app.GetKey(types.ModuleName), app.GetSubspace(types.ModuleName), app.AccountKeeper, app.BankKeeper, app.AuthzKeeper, app.FeeGrantKeeper, app.GetKey(banktypes.StoreKey)))
	require.Len(t, weightedProposalContent, 7)

	w0 := weightedProposalContent[0]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightAddMarkerProposal, w0.AppParamsKey())
	require.Equal(t, simappparams.DefaultWeightAddMarkerProposalContent, w0.DefaultWeight())

	content := w0.ContentSimulatorFn()(r, ctx, accounts)

	require.Equal(t, "eAerqyNEUz", content.GetDescription())
	require.Equal(t, "GkqEG", content.GetTitle())

	require.Equal(t, "marker", content.ProposalRoute())
	require.Equal(t, "AddMarker", content.ProposalType())
}
