package simulation_test

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/name/simulation"
	"github.com/provenance-io/provenance/x/name/types"
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

// LogIfError logs an error if it's not nil.
// The error is automatically added to the format and args.
func (s *SimTestSuite) LogIfError(err error, format string, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		s.T().Logf(format+": %v", args)
	}
}

func (s *SimTestSuite) TestWeightedOperations() {
	cdc := s.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, s.app.NameKeeper,
		s.app.AccountKeeper, s.app.BankKeeper,
	)

	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accs := s.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simappparams.DefaultWeightMsgBindName, sdk.MsgTypeURL(&types.MsgBindNameRequest{}), sdk.MsgTypeURL(&types.MsgBindNameRequest{})},
		{simappparams.DefaultWeightMsgDeleteName, sdk.MsgTypeURL(&types.MsgDeleteNameRequest{}), sdk.MsgTypeURL(&types.MsgDeleteNameRequest{})},
		{simappparams.DefaultWeightMsgModifyName, sdk.MsgTypeURL(&types.MsgModifyNameRequest{}), sdk.MsgTypeURL(&types.MsgModifyNameRequest{})},
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

	// Now assert that each entry was as expected.
	for i := range expected {
		s.Assert().Equal(expected[i].weight, weightedOps[i].Weight(), "weightedOps[%d].Weight", i)
		s.Assert().Equal(expected[i].opMsgRoute, opMsgs[i].Route, "weightedOps[%d] operationMsg.Route", i)
		s.Assert().Equal(expected[i].opMsgName, opMsgs[i].Name, "weightedOps[%d] operationMsg.Name", i)
	}
}

// TestSimulateMsgBindName tests the normal scenario of a valid message of type TypeMsgBindName.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (s *SimTestSuite) TestSimulateMsgBindName() {

	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	name := "provenance"
	s.LogIfError(s.app.NameKeeper.SetNameRecord(s.ctx, name, accounts[0].Address, false), "SetNameRecord(%q)", name)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgBindName(s.app.NameKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgBindName op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgBindNameRequest
	s.Require().NoError(types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Record.Address, "msg.Record.Address")
	s.Assert().Equal("cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Parent.Address, "msg.Parent.Address")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

// TestSimulateMsgDeleteName tests the normal scenario of a valid message of type MsgDeleteName.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (s *SimTestSuite) TestSimulateMsgDeleteName() {

	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 1)

	name := "deleteme"
	s.LogIfError(s.app.NameKeeper.SetNameRecord(s.ctx, name, accounts[0].Address, false), "SetNameRecord(%q)", name)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgDeleteName(s.app.NameKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgDeleteName op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgDeleteNameRequest
	s.Require().NoError(types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal("cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Record.Address, "msg.Record.Address")
	s.Assert().Equal(name, msg.Record.Name, "msg.Record.Name")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

// simAccount.Address tests the normal scenario of a valid message of type MsgModifyName.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (s *SimTestSuite) TestSimulateMsgModifyName() {

	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 1)

	name := "modifyme"
	s.LogIfError(s.app.NameKeeper.SetNameRecord(s.ctx, name, accounts[0].Address, false), "SetNameRecord(%q)", name)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgModifyName(s.app.NameKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgModifyName op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgModifyNameRequest
	s.Require().NoError(types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal("cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Record.Address, "msg.Record.Address")
	s.Assert().Equal(name, msg.Record.Name, "msg.Record.Name")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for i, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		err := testutil.FundAccount(s.app.BankKeeper, s.ctx, account.Address, initCoins)
		s.Require().NoError(err, "[%d]: FundAccount", i)
	}

	return accounts
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
