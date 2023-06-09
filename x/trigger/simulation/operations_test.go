package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/trigger/simulation"
	"github.com/provenance-io/provenance/x/trigger/types"
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

func (s *SimTestSuite) TestWeightedOperations() {
	cdc := s.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, s.app.TriggerKeeper,
		s.app.AccountKeeper, s.app.BankKeeper,
	)

	// setup 3 accounts
	source := rand.NewSource(1)
	r := rand.New(source)
	accs := s.getTestingAccounts(r, 3)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simappparams.DefaultWeightSubmitCreateTrigger, sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{})},
		{simappparams.DefaultWeightSubmitDestroyTrigger, sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{}), sdk.MsgTypeURL(&types.MsgDestroyTriggerRequest{})},
	}

	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, s.app.BaseApp, s.ctx, accs, "")
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		s.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		s.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		s.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

func (s *SimTestSuite) TestSimulateMsgCreateTrigger() {

	// setup 3 accounts
	source := rand.NewSource(1)
	r := rand.New(source)
	accounts := s.getTestingAccounts(r, 3)

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// bad operation
	op := simulation.SimulateMsgCreateTrigger(s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, []simtypes.Account{accounts[0]}, "")
	s.Require().Equal(operationMsg, simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), "cannot choose 2 accounts because there are only 1"))
	s.Require().Nil(futureOperations, "should be nil for invalid SimulateMsgCreateTrigger operation")
	s.Require().Nil(err, "should have no error for a invalid SimulateMsgCreateTrigger operation")

	// execute operation
	op = simulation.SimulateMsgCreateTrigger(s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err = op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "should have no error for a valid SimulateMsgCreateTrigger operation")

	var msg types.MsgCreateTriggerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	s.Require().True(operationMsg.OK, operationMsg.String(), "should be an operation.OK for SimulateMsgCreateTrigger")
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "should have correct name for SimulateMsgCreateTrigger")
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "should have correct route for SimulateMsgCreateTrigger")
	s.Require().Len(futureOperations, 0, "should have no future operations for SimulateMsgCreateTrigger")
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

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgDestroyTrigger(s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "should have no error for valid SimulateMsgDestroyTrigger")

	var msg types.MsgDestroyTriggerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	s.Require().True(operationMsg.OK, operationMsg.String(), "should be an operation.OK for SimulateMsgDestroyTrigger")
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "should have the correct message name for SimulateMsgDestroyTrigger")
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "should have the correct message route for SimulateMsgDestroyTrigger")
	s.Require().Equal(int(1000), int(msg.GetId()), "should have the correct id for SimulateMsgDestroyTrigger")
	s.Require().Equal(accounts[0].Address.String(), msg.GetAuthority(), "should have the correct authority for SimulateMsgDestroyTrigger")
	s.Require().Len(futureOperations, 0, "should have no future operations for SimulateMsgDestroyTrigger")
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
			raccs, err := simulation.RandomAccs(r, tc.accs, tc.count)
			if len(tc.err) == 0 {
				s.Require().NoError(err, "should have no error for successful RandomAccs")
				s.Require().Equal(tc.expected, raccs, "should have correct output for successful RandomAccs")
			} else {
				s.Require().EqualError(err, tc.err, "should have correct error message for RandomAccs")
			}
		})
	}
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
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
