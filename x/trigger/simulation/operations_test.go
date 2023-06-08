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

	// execute operation
	op := simulation.SimulateMsgCreateTrigger(s.app.TriggerKeeper, s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err)

	var msg types.MsgCreateTriggerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	s.Require().True(operationMsg.OK, operationMsg.String())
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route)
	s.Require().Len(futureOperations, 0)
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
	s.Require().NoError(err)

	var msg types.MsgDestroyTriggerRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	s.Require().True(operationMsg.OK, operationMsg.String())
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Route)
	s.Require().Equal(uint64(1000), msg.GetId())
	s.Require().Equal(accounts[0].Address.String(), msg.GetAuthority())
	s.Require().Len(futureOperations, 0)
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
