package simulation_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/oracle/types"

	. "github.com/provenance-io/provenance/x/oracle/simulation"
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
	weightedOps := WeightedOperations(s.MakeTestSimState(), s.app.OracleKeeper,
		s.app.AccountKeeper, s.app.BankKeeper, s.app.IBCKeeper.ChannelKeeper,
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
		{weight: simappparams.DefaultWeightSendOracleQuery, opMsgRoute: types.ModuleName, opMsgName: sdk.MsgTypeURL(&types.MsgSendQueryOracleRequest{})},
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

func (s *SimTestSuite) TestProposalMsgs() {
	expected := []struct {
		key     string
		weight  int
		msgType sdk.Msg
	}{
		{
			key:     OpWeightMsgUpdateOracle,
			weight:  simappparams.DefaultWeightUpdateOracle,
			msgType: &types.MsgUpdateOracleRequest{},
		},
	}

	simState := s.MakeTestSimState()
	var propMsgs []simtypes.WeightedProposalMsg
	testGetPropMsgs := func() {
		propMsgs = ProposalMsgs(simState, s.app.OracleKeeper)
	}
	s.Require().NotPanics(testGetPropMsgs, "ProposalMsgs")
	s.Require().Len(propMsgs, len(expected), "ProposalMsgs")

	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 10)

	for i, propMsg := range propMsgs {
		exp := expected[i]
		s.Run(exp.key, func() {
			expMsgType := fmt.Sprintf("%T", exp.msgType)

			s.Assert().Equal(exp.key, propMsg.AppParamsKey(), "AppParamsKey()")
			s.Assert().Equal(exp.weight, propMsg.DefaultWeight(), "DefaultWeight()")
			s.Require().NotNil(propMsg.MsgSimulatorFn(), "MsgSimulatorFn()")

			var msg sdk.Msg
			testPropMsg := func() {
				msg = propMsg.MsgSimulatorFn()(r, s.ctx, accs)
			}
			s.Require().NotPanics(testPropMsg, "calling the propMsg.MsgSimulatorFn()")
			s.Require().NotNil(msg, "msg result")
			actMsgType := fmt.Sprintf("%T", msg)
			s.Assert().Equal(expMsgType, actMsgType, "msg result")
		})
	}
}

func (s *SimTestSuite) TestSimulatePropMsgUpdateOracle() {
	r := rand.New(rand.NewSource(1))
	expMsg := &types.MsgUpdateOracleRequest{
		Address:   "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3",
		Authority: s.app.OracleKeeper.GetAuthority(),
	}

	var msgSimFn simtypes.MsgSimulatorFn
	testFnMaker := func() {
		msgSimFn = SimulatePropMsgUpdateOracle(s.app.OracleKeeper)
	}
	s.Require().NotPanics(testFnMaker, "SimulatePropMsgUpdateOracle")
	s.Require().NotNil(msgSimFn, "SimulatePropMsgUpdateOracle resulting MsgSimulationFn")

	var actMsg sdk.Msg
	testSimFn := func() {
		actMsg = msgSimFn(r, s.ctx, nil)
	}
	s.Require().NotPanics(testSimFn, "executing the SimulatePropMsgUpdateOracle resulting MsgSimulationFn")
	s.Assert().Equal(expMsg, actMsg)
}

func (s *SimTestSuite) TestSimulateMsgSendQueryOracle() {
	// setup 3 accounts
	source := rand.NewSource(1)
	r := rand.New(source)
	accounts := s.getTestingAccounts(r, 3)

	// execute operation
	op := SimulateMsgSendQueryOracle(s.MakeTestSimState(), s.app.OracleKeeper, s.app.AccountKeeper, s.app.BankKeeper, s.app.IBCKeeper.ChannelKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgSendQueryOracle op(...) error")
	s.LogOperationMsg(operationMsg, "no channel")

	var msg types.MsgSendQueryOracleRequest
	s.Require().NoError(s.app.AppCodec().Unmarshal(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().False(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(types.ModuleName, operationMsg.Route, "operationMsg.Route")
	s.Assert().Len(futureOperations, 0, "futureOperations")
	s.Assert().Equal("cannot get random channel because none exist", operationMsg.Comment, "operationMsg.Comment")
}

func (s *SimTestSuite) TestRandomAccs() {
	source := rand.NewSource(1)
	r := rand.New(source)
	accounts := s.getTestingAccounts(r, 3)

	tests := []struct {
		name     string
		accs     []simtypes.Account
		expected []simtypes.Account
		count    uint64
		err      string
	}{
		{
			name:     "valid - return nothing when count is 0",
			accs:     []simtypes.Account{},
			expected: []simtypes.Account{},
			count:    0,
		},
		{
			name:     "valid - return 1 when count is 1",
			accs:     []simtypes.Account{accounts[0]},
			expected: []simtypes.Account{accounts[0]},
			count:    1,
		},
		{
			name:     "valid - return multiple when count greater than 1",
			accs:     []simtypes.Account{accounts[0], accounts[1]},
			expected: []simtypes.Account{accounts[1], accounts[0]},
			count:    2,
		},
		{
			name:     "valid - return is limited by count",
			accs:     []simtypes.Account{accounts[0], accounts[1], accounts[2]},
			expected: []simtypes.Account{accounts[1]},
			count:    1,
		},
		{
			name:     "invalid - return error when count is greater than length",
			accs:     []simtypes.Account{accounts[0], accounts[1]},
			expected: []simtypes.Account{},
			count:    3,
			err:      "cannot choose 3 accounts because there are only 2",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			raccs, err := RandomAccs(r, tc.accs, tc.count)
			if len(tc.err) == 0 {
				s.Require().NoError(err, "should have no error for successful RandomAccs")
				s.Require().Equal(tc.expected, raccs, "should have correct output for successful RandomAccs")
			} else {
				s.Require().EqualError(err, tc.err, "should have correct error message for RandomAccs")
			}
		})
	}
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(1000000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		err := testutil.FundAccount(s.ctx, s.app.BankKeeper, account.Address, initCoins)
		s.Require().NoError(err)
	}

	return accounts
}
