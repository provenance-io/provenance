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
