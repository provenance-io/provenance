package simulation_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/trigger/simulation"
	"github.com/provenance-io/provenance/x/trigger/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

func (s *SimTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false)
}

// LogOperationMsg logs all fields of the provided operationMsg.
func (s *SimTestSuite) LogOperationMsg(operationMsg simtypes.OperationMsg, msg string, args ...interface{}) {
	msgFmt := "%s"
	if len(bytes.TrimSpace(operationMsg.Msg)) == 0 {
		msgFmt = "    %q"
	}
	fmtLines := []string{
		fmt.Sprintf(msg, args...),
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

// MakeTestSimState creates a new module.SimulationState struct with the fields needed by the functions being tested.
func (s *SimTestSuite) MakeTestSimState() module.SimulationState {
	return module.SimulationState{
		AppParams: make(simtypes.AppParams),
		Cdc:       s.app.AppCodec(),
		TxConfig:  s.app.GetTxConfig(),
	}
}

func (s *SimTestSuite) TestWeightedOperations() {
	weightedOps := simulation.WeightedOperations(s.MakeTestSimState(), s.app.TriggerKeeper,
		s.app.AccountKeeper, s.app.BankKeeper,
	)

	// setup 3 accounts
	source := rand.NewSource(1)
	r := rand.New(source)
	accs := s.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{weight: simappparams.DefaultWeightSubmitCreateTrigger, opMsgRoute: types.RouterKey, opMsgName: sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{})},
		{weight: simappparams.DefaultWeightSubmitDestroyTrigger, opMsgRoute: types.RouterKey, opMsgName: sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{})},
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

func (s *SimTestSuite) TestSimulateMsgCreateTrigger() {
	// setup 3 accounts
	source := rand.NewSource(1)
	r := rand.New(source)
	accounts := s.getTestingAccounts(r, 3)

	// bad operation
	op := simulation.SimulateMsgCreateTrigger(s.MakeTestSimState(), s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	expBadOp := simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), "cannot choose 2 accounts because there are only 1")
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts[0:1], "")
	s.LogOperationMsg(operationMsg, "bad SimulateMsgCreateTrigger")
	s.Assert().Equal(expBadOp, operationMsg, "bad operationMsg")
	s.Assert().Len(futureOperations, 0, "bad future ops")
	s.Assert().NoError(err, "bad SimulateMsgCreateTrigger op(...) error")

	// execute operation
	op = simulation.SimulateMsgCreateTrigger(s.MakeTestSimState(), s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err = op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgCreateTrigger op(...) error")
	s.LogOperationMsg(operationMsg, "good")

	var msg types.MsgCreateTriggerRequest
	s.Require().NoError(s.app.AppCodec().Unmarshal(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(types.RouterKey, operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) TestSimulateMsgDestroyTrigger() {
	// setup 3 accounts
	source := rand.NewSource(1)
	r := rand.New(source)
	accounts := s.getTestingAccounts(r, 3)

	actions, _ := sdktx.SetMsgs([]sdk.Msg{simulation.NewRandomAction(r, accounts[0].Address.String(), accounts[1].Address.String())})
	anyEvent, _ := codectypes.NewAnyWithValue(simulation.NewRandomEvent(r, s.ctx.BlockTime().UTC()))
	trigger := types.NewTrigger(1000, accounts[0].Address.String(), anyEvent, actions)
	s.app.TriggerKeeper.SetTrigger(s.ctx, trigger)
	s.app.TriggerKeeper.SetEventListener(s.ctx, trigger)
	s.app.TriggerKeeper.SetGasLimit(s.ctx, trigger.GetId(), 1000)

	// execute operation
	op := simulation.SimulateMsgDestroyTrigger(s.MakeTestSimState(), s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgDestroyTrigger op(...) error")
	s.LogOperationMsg(operationMsg, "good")

	var msg types.MsgDestroyTriggerRequest
	s.Require().NoError(s.app.AppCodec().Unmarshal(operationMsg.Msg, &msg), "Unmarshal(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(types.RouterKey, operationMsg.Route, "operationMsg.Route")
	s.Assert().Equal(1000, int(msg.GetId()), "msg.GetId()")
	s.Assert().Equal(accounts[0].Address.String(), msg.GetAuthority(), "msg.GetAuthority()")
	s.Assert().Len(futureOperations, 0, "futureOperations")
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	return testutil.GenerateTestingAccounts(s.T(), s.ctx, s.app, r, n)
}
