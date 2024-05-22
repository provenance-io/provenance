package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/testutil"
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

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	return testutil.GenerateTestingAccounts(s.T(), s.ctx, s.app, r, n)
}

// MakeTestSimState creates a new module.SimulationState struct with the fields needed by the functions being tested.
func (s *SimTestSuite) MakeTestSimState() module.SimulationState {
	return module.SimulationState{
		AppParams: make(simtypes.AppParams),
		Cdc:       s.app.AppCodec(),
		TxConfig:  s.app.GetTxConfig(),
	}
}

func (s *SimTestSuite) NewSimTestHelper(accCount int) *testutil.SimTestHelper {
	return testutil.NewSimTestHelper(s.T(), s.ctx, rand.New(rand.NewSource(1)), s.app).
		WithTestingAccounts(accCount)
}

func (s *SimTestSuite) TestWeightedOperations() {
	expected := []testutil.ExpectedWeightedOp{
		{Weight: simappparams.DefaultWeightSendOracleQuery, Route: types.ModuleName, MsgType: &types.MsgSendQueryOracleRequest{}},
	}

	simState := s.MakeTestSimState()
	wOpsFn := func() simulation.WeightedOperations {
		return WeightedOperations(simState, s.app.OracleKeeper,
			s.app.AccountKeeper, s.app.BankKeeper, s.app.IBCKeeper.ChannelKeeper)
	}
	s.NewSimTestHelper(3).AssertWeightedOperations(expected, wOpsFn)
}

func (s *SimTestSuite) TestProposalMsgs() {
	expected := []testutil.ExpectedProposalMsg{
		{Key: OpWeightMsgUpdateOracle, Weight: simappparams.DefaultWeightUpdateOracle, MsgType: &types.MsgUpdateOracleRequest{}},
	}

	simState := s.MakeTestSimState()
	proposalMsgsFn := func() []simtypes.WeightedProposalMsg {
		return ProposalMsgs(simState, s.app.OracleKeeper)
	}
	s.NewSimTestHelper(10).AssertProposalMsgs(expected, proposalMsgsFn)
}

func (s *SimTestSuite) TestSimulatePropMsgUpdateOracle() {
	// This expected Address might change if use of the randomizer changes (e.g. generating more accounts).
	expMsg := &types.MsgUpdateOracleRequest{
		Address:   "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3",
		Authority: s.app.OracleKeeper.GetAuthority(),
	}

	simFnMaker := func() simtypes.MsgSimulatorFn {
		return SimulatePropMsgUpdateOracle(s.app.OracleKeeper)
	}
	s.NewSimTestHelper(0).AssertMsgSimulatorFn(expMsg, simFnMaker)
}

func (s *SimTestSuite) TestSimulateMsgSendQueryOracle() {
	expected := testutil.ExpectedOp{
		Route:    types.ModuleName,
		EmptyMsg: &types.MsgSendQueryOracleRequest{},
		Comment:  "cannot get random channel because none exist",
		OK:       false,
	}

	simState := s.MakeTestSimState()
	opMaker := func() simtypes.Operation {
		return SimulateMsgSendQueryOracle(simState, s.app.OracleKeeper,
			s.app.AccountKeeper, s.app.BankKeeper, s.app.IBCKeeper.ChannelKeeper)
	}
	s.NewSimTestHelper(3).AssertSimOp(expected, opMaker, "no channel")
}

func (s *SimTestSuite) TestRandomAccs() {
	source := rand.NewSource(1)
	r := rand.New(source)
	accounts := testutil.GenerateTestingAccounts(s.T(), s.ctx, s.app, r, 3)

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
