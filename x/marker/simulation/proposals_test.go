package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	simapp "github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/simulation"
	"github.com/provenance-io/provenance/x/marker/types"
)

func TestProposalContents(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(
		keeper.NewKeeper(
			app.AppCodec(),
			app.GetKey(types.ModuleName),
			app.GetSubspace(types.ModuleName),
			app.AccountKeeper,
			app.BankKeeper,
			app.AuthzKeeper,
			app.FeeGrantKeeper,
			app.AttributeKeeper,
			app.NameKeeper,
			app.TransferKeeper,
		),
	)
	require.Len(t, weightedProposalContent, 6)

	w0 := weightedProposalContent[0]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightSupplyIncreaseProposal, w0.AppParamsKey())
	require.Equal(t, simappparams.DefaultWeightSupplyIncreaseProposalContent, w0.DefaultWeight())

	addTestMarker(t, ctx, app, r, accounts)

	content := w0.ContentSimulatorFn()(r, ctx, accounts)

	require.NotNil(t, content, "content")
	assert.Equal(t, "weXhSUkMhPjMaxKlMIJMOXcnQfyzeOcbWwNbeHVIkPZBSpYuLyYggwexjxusrBqDOTtGTOWeLrQKjLxzIivHSlcxgdXhhuTSkuxK", content.GetDescription(), "GetDescription")
	assert.Equal(t, "yNhYFmBZHe", content.GetTitle(), "GetTitle")

	assert.Equal(t, "marker", content.ProposalRoute(), "ProposalRoute")
	assert.Equal(t, "IncreaseSupply", content.ProposalType(), "ProposalType")
}

func addTestMarker(t *testing.T, ctx sdk.Context, app *simapp.App, r *rand.Rand, accs []simtypes.Account) {
	simAcc, _ := simtypes.RandomAcc(r, accs)

	server := keeper.NewMsgServerImpl(app.MarkerKeeper)
	_, err := server.AddMarker(sdk.WrapSDKContext(ctx), &types.MsgAddMarkerRequest{
		Amount:      sdk.NewInt64Coin("simtestcoin", 100),
		Manager:     "",
		FromAddress: app.MarkerKeeper.GetAuthority(),
		Status:      types.StatusActive,
		MarkerType:  types.MarkerType_Coin,
		AccessList: []types.AccessGrant{{
			Address: simAcc.Address.String(),
			Permissions: []types.Access{
				types.Access_Mint, types.Access_Burn,
				types.Access_Deposit, types.Access_Withdraw,
				types.Access_Delete, types.Access_Admin,
			},
		}},
		SupplyFixed:            false,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
		RequiredAttributes:     nil,
	})
	require.NoError(t, err, "AddMarker simtestcoin")
}
