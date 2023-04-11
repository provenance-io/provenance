package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/marker/simulation"
	"github.com/provenance-io/provenance/x/marker/types"
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

	weightedOps := simulation.WeightedOperations(appParams, cdc, codec.NewProtoCodec(suite.app.InterfaceRegistry()), suite.app.MarkerKeeper,
		suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.GovKeeper,
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
	//	operation, it's probably just due to a change in the number of times r is used before
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
		{simappparams.DefaultWeightMsgChangeStatus, sdk.MsgTypeURL(&types.MsgCancelRequest{}), sdk.MsgTypeURL(&types.MsgCancelRequest{})},
		// Possible names: types.TypeAddAccessRequest, fmt.Sprintf("%T", &types.MsgAddAccessRequest{})
		{simappparams.DefaultWeightMsgAddAccess, sdk.MsgTypeURL(&types.MsgAddAccessRequest{}), sdk.MsgTypeURL(&types.MsgAddAccessRequest{})},
		{simappparams.DefaultWeightMsgAddFinalizeActivateMarker, sdk.MsgTypeURL(&types.MsgAddFinalizeActivateMarkerRequest{}), sdk.MsgTypeURL(&types.MsgAddFinalizeActivateMarkerRequest{})},
		{simappparams.DefaultWeightMsgAddMarkerProposal, "gov", sdk.MsgTypeURL(&govtypes.MsgSubmitProposal{})},
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

// TestSimulateMsgAddMarkerProposal tests the normal scenario of a valid message of type MsgAddMarkerProposalRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgAddMarkerProposal() {
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 10)
	accounts = append(accounts, suite.createTestingAccountsWithPower(r, 1, 0)...)
	accounts = append(accounts, suite.createTestingAccountsWithPower(r, 1, 1)...)
	acctZero := accounts[len(accounts)-2]
	acctOne := accounts[len(accounts)-1]
	acctOneBalance := suite.app.BankKeeper.SpendableCoins(suite.ctx, acctOne.Address)
	var acctOneBalancePlusOne sdk.Coins
	for _, c := range acctOneBalance {
		acctOneBalancePlusOne = acctOneBalancePlusOne.Add(sdk.NewCoin(c.Denom, c.Amount.AddRaw(1)))
	}

	// Set default deposit params
	govMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 3))
	depositPeriod := 1 * time.Second

	resetParams := func(t *testing.T, ctx sdk.Context) {
		require.NotPanics(suite.T(), func() {
			suite.app.GovKeeper.SetDepositParams(suite.ctx, govtypes.DepositParams{
				MinDeposit:       govMinDep,
				MaxDepositPeriod: &depositPeriod,
			})
		}, "gov SetDepositParams")
	}

	access := types.AccessGrant{
		Address:     acctOne.Address.String(),
		Permissions: types.AccessListByNames("DELETE,MINT,BURN"),
	}

	tests := []struct {
		name            string
		sender          simtypes.Account
		msg             sdk.Msg
		deposit         sdk.Coins
		comment         string
		expSkip         bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
		expInErr        []string
	}{
		{
			name:            "no spendable coins",
			sender:          acctZero,
			msg:             types.NewMsgAddMarkerProposalRequest("test2", sdk.NewInt(100), sdk.AccAddress{}, types.StatusUndefined, types.MarkerType_Coin, []types.AccessGrant{}, true, true, "validAuthority"),
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "should not matter",
			expSkip:         true,
			expOpMsgRoute:   "marker",
			expOpMsgName:    sdk.MsgTypeURL(&types.MsgAddMarkerProposalRequest{}),
			expOpMsgComment: "sender has no spendable coins",
			expInErr:        nil,
		},
		{
			name:            "not enough coins for deposit",
			sender:          acctOne,
			msg:             types.NewMsgAddMarkerProposalRequest("test2", sdk.NewInt(100), sdk.AccAddress{}, types.StatusUndefined, types.MarkerType_Coin, []types.AccessGrant{}, true, true, "validAuthority"),
			deposit:         acctOneBalancePlusOne,
			comment:         "should not be this",
			expSkip:         true,
			expOpMsgRoute:   "marker",
			expOpMsgName:    sdk.MsgTypeURL(&types.MsgAddMarkerProposalRequest{}),
			expOpMsgComment: "sender has insufficient balance to cover deposit",
			expInErr:        nil,
		},
		{
			name:            "nil msg",
			sender:          accounts[0],
			msg:             nil,
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "will not get returned",
			expSkip:         true,
			expOpMsgRoute:   "marker",
			expOpMsgName:    "/",
			expOpMsgComment: "wrapping MsgAddMarkerProposalRequest as Any",
			expInErr:        []string{"Expecting non nil value to create a new Any", "failed packing protobuf message to Any"},
		},
		{
			name: "gen and deliver returns error",
			sender: simtypes.Account{
				PrivKey: accounts[0].PrivKey,
				PubKey:  acctOne.PubKey,
				Address: acctOne.Address,
				ConsKey: accounts[0].ConsKey,
			},
			msg:             types.NewMsgAddMarkerProposalRequest("test2", sdk.NewInt(100), sdk.AccAddress{}, types.StatusUndefined, types.MarkerType_Coin, []types.AccessGrant{}, true, true, "validAuthority"),
			deposit:         acctOneBalance,
			comment:         "this should be ignored",
			expSkip:         true,
			expOpMsgRoute:   "marker",
			expOpMsgName:    sdk.MsgTypeURL(&govtypes.MsgSubmitProposal{}),
			expOpMsgComment: "unable to deliver tx",
			expInErr:        []string{"pubKey does not match signer address", "invalid pubkey"},
		},
		{
			name:            "all good",
			sender:          accounts[1],
			msg:             types.NewMsgAddMarkerProposalRequest("test2", sdk.NewInt(100), sdk.AccAddress{}, types.StatusFinalized, types.MarkerType_Coin, []types.AccessGrant{access}, true, true, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"),
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "this is a test comment",
			expSkip:         false,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govtypes.MsgSubmitProposal{}),
			expOpMsgComment: "this is a test comment",
			expInErr:        nil,
		},
	}

	for _, tc := range tests {
		resetParams(suite.T(), suite.ctx)
		suite.Run(tc.name, func() {

			args := &simulation.SendGovMsgArgs{
				WeightedOpsArgs: suite.getWeightedOpsArgs(),
				R:               rand.New(rand.NewSource(1)),
				App:             suite.app.BaseApp,
				Ctx:             suite.ctx,
				Accs:            accounts,
				ChainID:         "send-gov-test",
				Sender:          tc.sender,
				Msg:             tc.msg,
				Deposit:         tc.deposit,
				Comment:         tc.comment,
			}

			var skip bool
			var opMsg simtypes.OperationMsg

			testFunc := func() {
				skip, opMsg, _ = simulation.SendGovMsg(args)
			}
			suite.Require().NotPanics(testFunc, "SendGovMsg")

			suite.Assert().Equal(tc.expSkip, skip, "SendGovMsg result skip bool")
			suite.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "SendGovMsg result op msg route")
			suite.Assert().Equal(tc.expOpMsgName, opMsg.Name, "SendGovMsg result op msg name")
			suite.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "SendGovMsg result op msg comment")
			if !tc.expSkip && !skip {
				// If we don't expect a skip, and we didn't get one,
				// get the last gov prop and make sure it's the one we just sent.
				expMsgs := []sdk.Msg{tc.msg}
				prop := suite.getLastGovProp()
				if suite.Assert().NotNil(prop, "last gov prop") {
					msgs, err := prop.GetMsgs()
					if suite.Assert().NoError(err, "error from prop.GetMsgs() on the last gov prop") {
						suite.Assert().Equal(expMsgs[0], msgs[0], "messages in the last gov prop")
					}
				}
			}
		})
	}

}

// getWeightedOpsArgs creates a standard WeightedOpsArgs.
func (suite *SimTestSuite) getWeightedOpsArgs() simulation.WeightedOpsArgs {
	return simulation.WeightedOpsArgs{
		AppParams:  make(simtypes.AppParams),
		JSONCodec:  suite.app.AppCodec(),
		ProtoCodec: codec.NewProtoCodec(suite.app.InterfaceRegistry()),
		AK:         suite.app.AccountKeeper,
		BK:         suite.app.BankKeeper,
		GK:         suite.app.GovKeeper,
	}
}

// getLastGovProp gets the last gov prop to be submitted.
func (suite *SimTestSuite) getLastGovProp() *govtypes.Proposal {
	props := suite.app.GovKeeper.GetProposals(suite.ctx)
	if len(props) == 0 {
		return nil
	}
	return props[len(props)-1]
}

// freshCtx creates a new context and sets it to this SimTestSuite's ctx field.
func (suite *SimTestSuite) freshCtx() {
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
}

// createTestingAccountsWithPower creates new accounts with the specified power (coins amount).
func (suite *SimTestSuite) createTestingAccountsWithPower(r *rand.Rand, count int, power int64) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, count)

	initAmt := sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, account.Address, initCoins))
	}

	return accounts
}
