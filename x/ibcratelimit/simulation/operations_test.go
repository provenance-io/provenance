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
	"github.com/provenance-io/provenance/x/ibcratelimit"

	. "github.com/provenance-io/provenance/x/ibcratelimit/simulation"
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

func (s *SimTestSuite) TestProposalMsgs() {
	expected := []struct {
		key     string
		weight  int
		msgType sdk.Msg
	}{
		{
			key:     OpWeightMsgUpdateParams,
			weight:  simappparams.DefaultWeightIBCRLUpdateParams,
			msgType: &ibcratelimit.MsgUpdateParamsRequest{},
		},
	}

	simState := s.MakeTestSimState()
	var propMsgs []simtypes.WeightedProposalMsg
	testGetPropMsgs := func() {
		propMsgs = ProposalMsgs(simState, s.app.RateLimitingKeeper)
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
			s.Assert().IsType(exp.msgType, msg, "msg")
		})
	}
}

func (s *SimTestSuite) TestSimulatePropMsgUpdateOracle() {
	r := rand.New(rand.NewSource(1))
	expMsg := &ibcratelimit.MsgUpdateParamsRequest{
		Authority: s.app.OracleKeeper.GetAuthority(),
		Params: ibcratelimit.Params{
			ContractAddress: "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3",
		},
	}

	var msgSimFn simtypes.MsgSimulatorFn
	testFnMaker := func() {
		msgSimFn = SimulatePropMsgUpdateParams(s.app.RateLimitingKeeper)
	}
	s.Require().NotPanics(testFnMaker, "SimulatePropMsgUpdateParams")
	s.Require().NotNil(msgSimFn, "SimulatePropMsgUpdateParams resulting MsgSimulationFn")

	var actMsg sdk.Msg
	testSimFn := func() {
		actMsg = msgSimFn(r, s.ctx, nil)
	}
	s.Require().NotPanics(testSimFn, "executing the SimulatePropMsgUpdateParams resulting MsgSimulationFn")
	s.Assert().Equal(expMsg, actMsg)
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
