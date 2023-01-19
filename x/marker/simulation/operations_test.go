package simulation_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/marker/simulation"
	types "github.com/provenance-io/provenance/x/marker/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
}

func (suite *SimTestSuite) SetupTest() {
	suite.app = app.Setup(suite.T())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (suite *SimTestSuite) TestWeightedOperations() {
	cdc := suite.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, suite.app.MarkerKeeper,
		suite.app.AccountKeeper, suite.app.BankKeeper,
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	// Note: r is now passed around and used in several places, including in SDK functions.
	//	Since we're seeding it, these tests are deterministic. However, if there are changes
	//	made in the SDK or to the operations, these outcomes can change. To further confuse
	//	things, the operation name is sometimes taken from msg.Type(), and sometimes from
	//	fmt.Sprintf("%T", msg), and sometimes hard-coded. The .Type() function is no longer
	//	part of the Msg interface (though it is part of LegacyMsg). But depending on how the
	//	randomness plays out, it can be either of those. If one of these starts failing on
	//	the operation name, and the actual value is one of the other possibilities for that
	//	operation, it's probably just do to a change in the number of times r is used before
	//	getting to that operation.

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		// Possible names: types.TypeAddMarkerRequest, fmt.Sprintf("%T", &types.MsgAddMarkerRequest{})
		{simappparams.DefaultWeightMsgAddMarker, sdk.MsgTypeURL(&types.MsgAddMarkerRequest{}), sdk.MsgTypeURL(&types.MsgAddMarkerRequest{})},
		// Possible names: "ChangeStatus",
		//	types.TypeActivateRequest, fmt.Sprintf("%T", &types.MsgActivateRequest{}),
		//	types.TypeFinalizeRequest, fmt.Sprintf("%T", &types.MsgFinalizeRequest{}),
		//	types.TypeCancelRequest, fmt.Sprintf("%T", &types.MsgCancelRequest{}),
		//	types.TypeDeleteRequest, fmt.Sprintf("%T", &types.MsgDeleteRequest{}),
		{simappparams.DefaultWeightMsgChangeStatus, sdk.MsgTypeURL(&types.MsgActivateRequest{}), sdk.MsgTypeURL(&types.MsgActivateRequest{})},
		// Possible names: types.TypeAddAccessRequest, fmt.Sprintf("%T", &types.MsgAddAccessRequest{})
		{simappparams.DefaultWeightMsgAddAccess, sdk.MsgTypeURL(&types.MsgAddAccessRequest{}), sdk.MsgTypeURL(&types.MsgAddAccessRequest{})},
		{simappparams.DefaultWeightMsgAddFinalizeActivateMarker, sdk.MsgTypeURL(&types.MsgAddFinalizeActivateMarkerRequest{}), sdk.MsgTypeURL(&types.MsgAddFinalizeActivateMarkerRequest{})},
	}
	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, "")
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgAddMarker tests the normal scenario of a valid message of type TypeAddMarkerRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgAddMarker() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgAddMarker(suite.app.MarkerKeeper, suite.app.AccountKeeper, suite.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgAddMarkerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK, operationMsg.String())
	suite.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	suite.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route)
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgAddActivateFinalizeMarker tests the normal scenario of a valid message of type TypeAddActivateFinalizeMarkerRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgAddActivateFinalizeMarker() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgAddFinalizeActivateMarker(suite.app.MarkerKeeper, suite.app.AccountKeeper, suite.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgAddFinalizeActivateMarkerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK, operationMsg.String())
	suite.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	suite.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(1000000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		err := testutil.FundAccount(suite.app.BankKeeper, suite.ctx, account.Address, initCoins)
		suite.Require().NoError(err)
	}

	return accounts
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
