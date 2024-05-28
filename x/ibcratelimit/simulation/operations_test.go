package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/testutil"
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

func (s *SimTestSuite) TestProposalMsgs() {
	expected := []testutil.ExpectedProposalMsg{
		{
			Key:     OpWeightMsgUpdateParams,
			Weight:  simappparams.DefaultWeightIBCRLUpdateParams,
			MsgType: &ibcratelimit.MsgUpdateParamsRequest{},
		},
	}

	simState := s.MakeTestSimState()
	proposalMsgsFn := func() []simtypes.WeightedProposalMsg {
		return ProposalMsgs(simState, s.app.RateLimitingKeeper)
	}
	s.NewSimTestHelper(10).AssertProposalMsgs(expected, proposalMsgsFn)
}

func (s *SimTestSuite) TestSimulatePropMsgUpdateOracle() {
	expMsg := &ibcratelimit.MsgUpdateParamsRequest{
		Authority: s.app.OracleKeeper.GetAuthority(),
		Params: ibcratelimit.Params{
			ContractAddress: "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3",
		},
	}

	simFnMaker := func() simtypes.MsgSimulatorFn {
		return SimulatePropMsgUpdateParams(s.app.RateLimitingKeeper)
	}
	s.NewSimTestHelper(0).AssertMsgSimulatorFn(expMsg, simFnMaker)
}
