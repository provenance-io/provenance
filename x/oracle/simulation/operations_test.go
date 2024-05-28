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
