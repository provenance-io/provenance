package simulation_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
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

qfunc (s *SimTestSuite) TestWeightedOperations() {
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

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// bad operation
	op := simulation.SimulateMsgCreateTrigger(s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	expBadOp := simtypes.NoOpMsg(sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), sdk.MsgTypeURL(&types.MsgCreateTriggerRequest{}), "cannot choose 2 accounts because there are only 1")
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts[0:1], "")
	s.LogOperationMsg(operationMsg, "bad SimulateMsgCreateTrigger")
	s.Assert().Equal(expBadOp, operationMsg, "bad operationMsg")
	s.Assert().Len(futureOperations, 0, "bad future ops")
	s.Assert().NoError(err, "bad SimulateMsgCreateTrigger op(...) error")

	// execute operation
	op = simulation.SimulateMsgCreateTrigger(s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err = op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgCreateTrigger op(...) error")
	s.LogOperationMsg(operationMsg, "good")

	var msg types.MsgCreateTriggerRequest
	s.Require().NoError(s.app.AppCodec().UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
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

	// begin a new block
	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgDestroyTrigger(s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "SimulateMsgDestroyTrigger op(...) error")
	s.LogOperationMsg(operationMsg, "good")

	var msg types.MsgDestroyTriggerRequest
	s.Require().NoError(s.app.AppCodec().UnmarshalJSON(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Assert().True(operationMsg.OK, "operationMsg.OK")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name, "operationMsg.Name")
	s.Assert().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route, "operationMsg.Route")
	s.Assert().Equal(1000, int(msg.GetId()), "msg.GetId()")
	s.Assert().Equal(accounts[0].Address.String(), msg.GetAuthority(), "msg.GetAuthority()")
	s.Assert().Len(futureOperations, 0, "futureOperations")
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
