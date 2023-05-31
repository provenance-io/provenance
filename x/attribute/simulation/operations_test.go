package simulation_test

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/attribute/simulation"
	"github.com/provenance-io/provenance/x/attribute/types"
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
// Use this if there's a possible error that we probably don't care about (but might).
func (s *SimTestSuite) LogIfError(err error, format string, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		s.T().Logf(format+": %v", args)
	}
}

func (s *SimTestSuite) TestWeightedOperations() {
	cdc := s.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, s.app.AttributeKeeper,
		s.app.AccountKeeper, s.app.BankKeeper, s.app.NameKeeper,
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
		{simappparams.DefaultWeightMsgAddAttribute, sdk.MsgTypeURL(&types.MsgAddAttributeRequest{}), sdk.MsgTypeURL(&types.MsgAddAttributeRequest{})},
		{simappparams.DefaultWeightMsgUpdateAttribute, sdk.MsgTypeURL(&types.MsgUpdateAttributeRequest{}), sdk.MsgTypeURL(&types.MsgUpdateAttributeRequest{})},
		{simappparams.DefaultWeightMsgDeleteAttribute, sdk.MsgTypeURL(&types.MsgDeleteAttributeRequest{}), sdk.MsgTypeURL(&types.MsgDeleteAttributeRequest{})},
		{simappparams.DefaultWeightMsgDeleteDistinctAttribute, sdk.MsgTypeURL(&types.MsgDeleteDistinctAttributeRequest{}), sdk.MsgTypeURL(&types.MsgDeleteDistinctAttributeRequest{})},
		{simappparams.DefaultWeightMsgSetAccountData, sdk.MsgTypeURL(&types.MsgSetAccountDataRequest{}), sdk.MsgTypeURL(&types.MsgSetAccountDataRequest{})},
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

// TestSimulateMsgAddAttribute tests the normal scenario of a valid message of type TypeMsgAddAttribute.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (s *SimTestSuite) TestSimulateMsgAddAttribute() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	name := "example.provenance"
	s.LogIfError(s.app.NameKeeper.SetNameRecord(s.ctx, name, accounts[0].Address, false), "SetNameRecord(%q) error", name)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgAddAttribute(s.app.AttributeKeeper, s.app.AccountKeeper, s.app.BankKeeper, s.app.NameKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgAddAttribute op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgAddAttributeRequest
	s.Require().NoError(types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal("cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Account, "msg.Account")
	s.Assert().Equal("cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Owner, "msg.Owner")
	s.Assert().Equal(name, msg.Name, "msg.Name")
	s.Assert().Equal(types.AttributeType_Uri, msg.AttributeType, "msg.AttributeType")
	s.Assert().Equal([]byte("http://www.example.com/"), msg.Value, "msg.Value")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) TestSimulateMsgUpdateAttribute() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	name := "example.provenance"
	s.LogIfError(s.app.NameKeeper.SetNameRecord(s.ctx, name, accounts[0].Address, false), "SetNameRecord(%q) error", name)
	expireTime := GenerateRandomTime(1)
	attr := types.NewAttribute(name, accounts[1].Address.String(), types.AttributeType_String, []byte("test"), &expireTime)
	s.LogIfError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, accounts[0].Address), "SetAttribute(%q) error", name)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash, Time: time.Now()}})

	// execute operation
	op := simulation.SimulateMsgUpdateAttribute(s.app.AttributeKeeper, s.app.AccountKeeper, s.app.BankKeeper, s.app.NameKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgUpdateAttribute op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgUpdateAttributeRequest
	s.Require().NoError(types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(name, msg.Name, "msg.Name")
	s.Assert().Equal(accounts[0].Address.String(), msg.Owner, "msg.Owner")
	s.Assert().Equal(accounts[1].Address.String(), msg.Account, "msg.Account")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) TestSimulateMsgDeleteAttribute() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)
	expireTime := GenerateRandomTime(1)

	name := "example.provenance"
	s.LogIfError(s.app.NameKeeper.SetNameRecord(s.ctx, name, accounts[0].Address, false), "SetNameRecord(%q) error", name)
	attr := types.NewAttribute(name, accounts[1].Address.String(), types.AttributeType_String, []byte("test"), &expireTime)
	s.LogIfError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, accounts[0].Address), "SetAttribute(%q) error", name)
	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash, Time: time.Now()}})

	// execute operation
	op := simulation.SimulateMsgDeleteAttribute(s.app.AttributeKeeper, s.app.AccountKeeper, s.app.BankKeeper, s.app.NameKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgDeleteAttribute op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgDeleteAttributeRequest
	s.Require().NoError(types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(name, msg.Name, "msg.Name")
	s.Assert().Equal(accounts[0].Address.String(), msg.Owner, "msg.Owner")
	s.Assert().Equal(accounts[1].Address.String(), msg.Account, "msg.Account")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) TestSimulateMsgDeleteDistinctAttribute() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	name := "example.provenance"
	expireTime := GenerateRandomTime(1)
	s.LogIfError(s.app.NameKeeper.SetNameRecord(s.ctx, name, accounts[0].Address, false), "SetNameRecord(%q) error", name)
	attr := types.NewAttribute(name, accounts[1].Address.String(), types.AttributeType_String, []byte("test"), &expireTime)
	s.LogIfError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, accounts[0].Address), "SetAttribute(%q) error", name)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash, Time: time.Now()}})

	// execute operation
	op := simulation.SimulateMsgDeleteDistinctAttribute(s.app.AttributeKeeper, s.app.AccountKeeper, s.app.BankKeeper, s.app.NameKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgDeleteDistinctAttribute op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgDeleteDistinctAttributeRequest
	s.Require().NoError(types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(name, msg.Name, "msg.Name")
	s.Assert().Equal(accounts[0].Address.String(), msg.Owner, "msg.Owner")
	s.Assert().Equal(accounts[1].Address.String(), msg.Account, "msg.Account")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) TestSimulateMsgSetAccountData() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash, Time: time.Now()}})

	// execute operation
	op := simulation.SimulateMsgSetAccountData(s.app.AttributeKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgDeleteDistinctAttribute op(...) error")
	s.LogOperationMsg(operationMsg)

	var msg types.MsgSetAccountDataRequest
	s.Require().NoError(types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal("", msg.Value, "msg.Value")
	s.Assert().Equal(accounts[1].Address.String(), msg.Account, "msg.Account")
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

func GenerateRandomTime(minHours int) time.Time {
	currentTime := time.Now()

	// Generate a random duration between minHours and 2*minHours
	randomDuration := time.Duration(rand.Intn(minHours) + minHours)

	// Generate a random time by adding the random duration to the current time
	randomTime := currentTime.Add(randomDuration)

	return randomTime
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
