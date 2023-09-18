package simulation_test

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/provenance-io/provenance/x/marker/keeper"

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

func (s *SimTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

// LogOperationMsg logs all fields of the provided operationMsg.
func (s *SimTestSuite) LogOperationMsg(operationMsg simtypes.OperationMsg) {
	msgFmt := "%s"
	if len(bytes.TrimSpace(operationMsg.Msg)) == 0 {
		msgFmt = "    %q"
	}
	fmtLines := []string{
		"operationMsg.Route:   %q",
		"operationMsg.Name:    %q",
		"operationMsg.Comment: %q",
		"operationMsg.OK:      %t",
		"operationMsg.Msg: " + msgFmt,
	}
	s.T().Logf(strings.Join(fmtLines, "\n"),
		operationMsg.Route, operationMsg.Name, operationMsg.Comment, operationMsg.OK, string(operationMsg.Msg),
	)
}

func (s *SimTestSuite) TestWeightedOperations() {
	cdc := s.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, codec.NewProtoCodec(s.app.InterfaceRegistry()), s.app.MarkerKeeper,
		s.app.AccountKeeper, s.app.BankKeeper, s.app.GovKeeper, s.app.AttributeKeeper,
	)

	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accs := s.getTestingAccounts(r, 3)

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
		{simappparams.DefaultWeightMsgFinalize, sdk.MsgTypeURL(&types.MsgFinalizeRequest{}), sdk.MsgTypeURL(&types.MsgFinalizeRequest{})},
		// Possible names: types.TypeAddAccessRequest, fmt.Sprintf("%T", &types.MsgAddAccessRequest{})
		{simappparams.DefaultWeightMsgAddAccess, sdk.MsgTypeURL(&types.MsgAddAccessRequest{}), sdk.MsgTypeURL(&types.MsgAddAccessRequest{})},
		{simappparams.DefaultWeightMsgAddFinalizeActivateMarker, sdk.MsgTypeURL(&types.MsgAddFinalizeActivateMarkerRequest{}), sdk.MsgTypeURL(&types.MsgAddFinalizeActivateMarkerRequest{})},
		{simappparams.DefaultWeightMsgAddMarkerProposal, "gov", sdk.MsgTypeURL(&govtypes.MsgSubmitProposal{})},
		{simappparams.DefaultWeightMsgSetAccountData, sdk.MsgTypeURL(&types.MsgSetAccountDataRequest{}), sdk.MsgTypeURL(&types.MsgSetAccountDataRequest{})},
		{simappparams.DefaultWeightMsgUpdateDenySendList, sdk.MsgTypeURL(&types.MsgUpdateSendDenyListRequest{}), sdk.MsgTypeURL(&types.MsgUpdateSendDenyListRequest{})},
	}

	expNames := make([]string, len(expected))
	for i, exp := range expected {
		expNames[i] = exp.opMsgName
	}

	// Run all the ops and get the operation messages and their names.
	opMsgs := make([]simtypes.OperationMsg, len(weightedOps))
	actualNames := make([]string, len(weightedOps))
	for i, w := range weightedOps {
		opMsgs[i], _, _ = w.Op()(r, s.app.BaseApp, s.ctx, accs, "")
		actualNames[i] = opMsgs[i].Name
	}

	// First, make sure the op names are as expected since a failure there probably means the rest will fail.
	// And it's probably easier to address when you've got a nice list comparison of names and their orderings.
	s.Require().Equal(expNames, actualNames, "operation message names")

	for i := range expected {
		s.Require().Equal(expected[i].weight, weightedOps[i].Weight(), "weightedOps[i].Weight", i)
		s.Require().Equal(expected[i].opMsgRoute, opMsgs[i].Route, "opMsgs[i].Route", i)
		s.Require().Equal(expected[i].opMsgName, opMsgs[i].Name, "opMsgs[i].Name", i)
	}
}

// TestSimulateMsgAddMarker tests the normal scenario of a valid message of type TypeAddMarkerRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (s *SimTestSuite) TestSimulateMsgAddMarker() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgAddMarker(s.app.MarkerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err)

	var msg types.MsgAddMarkerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	s.Require().True(operationMsg.OK, operationMsg.String())
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route)
	s.Require().Len(futureOperations, 0)
}

// TestSimulateMsgAddActivateFinalizeMarker tests the normal scenario of a valid message of type TypeAddActivateFinalizeMarkerRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (s *SimTestSuite) TestSimulateMsgAddActivateFinalizeMarker() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgAddFinalizeActivateMarker(s.app.MarkerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err)

	var msg types.MsgAddFinalizeActivateMarkerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	s.Require().True(operationMsg.OK, operationMsg.String())
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route)
	s.Require().Len(futureOperations, 0)
}

// TestSimulateMsgAddMarkerProposal tests the normal scenario of a valid message of type MsgAddMarkerProposalRequest.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (s *SimTestSuite) TestSimulateMsgAddMarkerProposal() {
	NewMsgAddMarker := func(denom string, totalSupply sdkmath.Int, manager sdk.AccAddress, status types.MarkerStatus,
		markerType types.MarkerType, access []types.AccessGrant, fixed bool, allowGov bool, authority string,
	) *types.MsgAddMarkerRequest {
		return &types.MsgAddMarkerRequest{
			Amount: sdk.Coin{
				Denom:  denom,
				Amount: totalSupply,
			},
			Manager:                manager.String(),
			FromAddress:            authority,
			Status:                 status,
			MarkerType:             markerType,
			AccessList:             access,
			SupplyFixed:            fixed,
			AllowGovernanceControl: allowGov,
		}
	}

	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 10)
	accounts = append(accounts, s.createTestingAccountsWithPower(r, 1, 0)...)
	accounts = append(accounts, s.createTestingAccountsWithPower(r, 1, 1)...)
	acctZero := accounts[len(accounts)-2]
	acctOne := accounts[len(accounts)-1]
	acctOneBalance := s.app.BankKeeper.SpendableCoins(s.ctx, acctOne.Address)
	var acctOneBalancePlusOne sdk.Coins
	for _, c := range acctOneBalance {
		acctOneBalancePlusOne = acctOneBalancePlusOne.Add(sdk.NewCoin(c.Denom, c.Amount.AddRaw(1)))
	}

	// Set default deposit params
	govMinDep := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 3))
	depositPeriod := 1 * time.Second

	resetParams := func(t *testing.T, ctx sdk.Context) {
		require.NotPanics(s.T(), func() {
			s.app.GovKeeper.SetDepositParams(s.ctx, govtypes.DepositParams{
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
			msg:             NewMsgAddMarker("test2", sdk.NewInt(100), sdk.AccAddress{}, types.StatusUndefined, types.MarkerType_Coin, []types.AccessGrant{}, true, true, "validAuthority"),
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "should not matter",
			expSkip:         true,
			expOpMsgRoute:   "marker",
			expOpMsgName:    sdk.MsgTypeURL(&types.MsgAddMarkerRequest{}),
			expOpMsgComment: "sender has no spendable coins",
			expInErr:        nil,
		},
		{
			name:            "not enough coins for deposit",
			sender:          acctOne,
			msg:             NewMsgAddMarker("test2", sdk.NewInt(100), sdk.AccAddress{}, types.StatusUndefined, types.MarkerType_Coin, []types.AccessGrant{}, true, true, "validAuthority"),
			deposit:         acctOneBalancePlusOne,
			comment:         "should not be this",
			expSkip:         true,
			expOpMsgRoute:   "marker",
			expOpMsgName:    sdk.MsgTypeURL(&types.MsgAddMarkerRequest{}),
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
			msg:             NewMsgAddMarker("test2", sdk.NewInt(100), sdk.AccAddress{}, types.StatusUndefined, types.MarkerType_Coin, []types.AccessGrant{}, true, true, "validAuthority"),
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
			msg:             NewMsgAddMarker("test2", sdk.NewInt(100), sdk.AccAddress{}, types.StatusFinalized, types.MarkerType_Coin, []types.AccessGrant{access}, true, true, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"),
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
		resetParams(s.T(), s.ctx)
		s.Run(tc.name, func() {

			args := &simulation.SendGovMsgArgs{
				WeightedOpsArgs: s.getWeightedOpsArgs(),
				R:               rand.New(rand.NewSource(1)),
				App:             s.app.BaseApp,
				Ctx:             s.ctx,
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
			s.Require().NotPanics(testFunc, "SendGovMsg")

			s.Assert().Equal(tc.expSkip, skip, "SendGovMsg result skip bool")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "SendGovMsg result op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "SendGovMsg result op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "SendGovMsg result op msg comment")
			if !tc.expSkip && !skip {
				// If we don't expect a skip, and we didn't get one,
				// get the last gov prop and make sure it's the one we just sent.
				expMsgs := []sdk.Msg{tc.msg}
				prop := s.getLastGovProp()
				if s.Assert().NotNil(prop, "last gov prop") {
					msgs, err := prop.GetMsgs()
					if s.Assert().NoError(err, "error from prop.GetMsgs() on the last gov prop") {
						s.Assert().Equal(expMsgs[0], msgs[0], "messages in the last gov prop")
					}
				}
			}
		})
	}
}

func (s *SimTestSuite) TestSimulateMsgSetAccountData() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	// Add a marker with deposit permissions so that it can be found by the sim.
	newMarker := &types.MsgAddFinalizeActivateMarkerRequest{
		Amount:      sdk.NewInt64Coin("simcoin", 1000),
		Manager:     accounts[1].Address.String(),
		FromAddress: accounts[1].Address.String(),
		MarkerType:  types.MarkerType_RestrictedCoin,
		AccessList: []types.AccessGrant{
			{
				Address: accounts[1].Address.String(),
				Permissions: types.AccessList{
					types.Access_Mint, types.Access_Burn, types.Access_Deposit, types.Access_Withdraw,
					types.Access_Delete, types.Access_Admin, types.Access_Transfer,
				},
			},
		},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
		RequiredAttributes:     nil,
	}
	markerMsgServer := keeper.NewMsgServerImpl(s.app.MarkerKeeper)
	_, err := markerMsgServer.AddFinalizeActivateMarker(s.ctx, newMarker)
	s.Require().NoError(err, "AddFinalizeActivateMarker")

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	args := s.getWeightedOpsArgs()
	// execute operation
	op := simulation.SimulateMsgSetAccountData(s.app.MarkerKeeper, &args)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgSetAccountData op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgSetAccountDataRequest
	s.Require().NoError(s.app.AppCodec().UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal("simcoin", msg.Denom, "msg.Denom")
	s.Assert().Equal("", msg.Value, "msg.Value")
	s.Assert().Equal(accounts[1].Address.String(), msg.Signer, "msg.Signer")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) TestSimulateMsgUpdateSendDenyList() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	// Add a marker with deposit permissions so that it can be found by the sim.
	newMarker := &types.MsgAddFinalizeActivateMarkerRequest{
		Amount:      sdk.NewInt64Coin("simcoin", 1000),
		Manager:     accounts[1].Address.String(),
		FromAddress: accounts[1].Address.String(),
		MarkerType:  types.MarkerType_RestrictedCoin,
		AccessList: []types.AccessGrant{
			{
				Address: accounts[1].Address.String(),
				Permissions: types.AccessList{
					types.Access_Mint, types.Access_Burn, types.Access_Deposit, types.Access_Withdraw,
					types.Access_Delete, types.Access_Admin, types.Access_Transfer,
				},
			},
		},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
		RequiredAttributes:     nil,
	}
	markerMsgServer := keeper.NewMsgServerImpl(s.app.MarkerKeeper)
	_, err := markerMsgServer.AddFinalizeActivateMarker(s.ctx, newMarker)
	s.Require().NoError(err, "AddFinalizeActivateMarker")

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	args := s.getWeightedOpsArgs()
	// execute operation
	op := simulation.SimulateMsgUpdateSendDenyList(s.app.MarkerKeeper, &args)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgUpdateSendDenyList op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgUpdateSendDenyListRequest
	s.Require().NoError(s.app.AppCodec().UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal("simcoin", msg.Denom, "msg.Denom")
	s.Assert().Len(msg.AddDeniedAddresses, 10, "msg.AddDeniedAddresses")
	s.Assert().Equal(accounts[1].Address.String(), msg.Authority, "msg.Authority")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(1000000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		err := testutil.FundAccount(s.app.BankKeeper, s.ctx, account.Address, initCoins)
		s.Require().NoError(err)
	}

	return accounts
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

// getWeightedOpsArgs creates a standard WeightedOpsArgs.
func (s *SimTestSuite) getWeightedOpsArgs() simulation.WeightedOpsArgs {
	return simulation.WeightedOpsArgs{
		AppParams:  make(simtypes.AppParams),
		JSONCodec:  s.app.AppCodec(),
		ProtoCodec: codec.NewProtoCodec(s.app.InterfaceRegistry()),
		AK:         s.app.AccountKeeper,
		BK:         s.app.BankKeeper,
		GK:         s.app.GovKeeper,
		AttrK:      s.app.AttributeKeeper,
	}
}

// getLastGovProp gets the last gov prop to be submitted.
func (s *SimTestSuite) getLastGovProp() *govtypes.Proposal {
	props := s.app.GovKeeper.GetProposals(s.ctx)
	if len(props) == 0 {
		return nil
	}
	return props[len(props)-1]
}

// freshCtx creates a new context and sets it to this SimTestSuite's ctx field.
func (s *SimTestSuite) freshCtx() {
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

// createTestingAccountsWithPower creates new accounts with the specified power (coins amount).
func (s *SimTestSuite) createTestingAccountsWithPower(r *rand.Rand, count int, power int64) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, count)

	initAmt := sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, account.Address, initCoins))
	}

	return accounts
}
