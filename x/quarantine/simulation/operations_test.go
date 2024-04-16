package simulation_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/quarantine"
	"github.com/provenance-io/provenance/x/quarantine/simulation"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(testutil.FundAccount(s.ctx, s.app.BankKeeper, account.Address, initCoins))
	}

	return accounts
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

func (s *SimTestSuite) TestWeightedOperations() {
	expected := []struct {
		weight      int
		opsMsgRoute string
		opsMsgName  string
	}{
		{simulation.WeightMsgOptIn, quarantine.ModuleName, simulation.TypeMsgOptIn},
		{simulation.WeightMsgOptOut, quarantine.ModuleName, simulation.TypeMsgOptOut},
		{simulation.WeightMsgAccept, quarantine.ModuleName, simulation.TypeMsgAccept},
		{simulation.WeightMsgDecline, quarantine.ModuleName, simulation.TypeMsgDecline},
		{simulation.WeightMsgUpdateAutoResponses, quarantine.ModuleName, simulation.TypeMsgUpdateAutoResponses},
	}

	weightedOps := simulation.WeightedOperations(s.MakeTestSimState(), s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)

	s.Require().Len(weightedOps, len(expected), "weighted ops")

	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 10)

	for i, actual := range weightedOps {
		exp := expected[i]
		parts := strings.Split(exp.opsMsgName, ".")
		s.Run(parts[len(parts)-1], func() {
			operationMsg, futureOps, err := actual.Op()(r, s.app.BaseApp, s.ctx, accs, "")
			s.Assert().NoError(err, "op error")
			s.Assert().Equal(exp.weight, actual.Weight(), "op weight")
			s.Assert().Equal(exp.opsMsgRoute, operationMsg.Route, "op route")
			s.Assert().Equal(exp.opsMsgName, operationMsg.Name, "op name")
			s.Assert().Nil(futureOps, "future ops")
		})
	}
}

func (s *SimTestSuite) TestSimulateMsgOptIn() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	op := simulation.SimulateMsgOptIn(s.MakeTestSimState(), s.app.AccountKeeper, s.app.BankKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgOptIn
	err = s.app.AppCodec().Unmarshal(opMsg.Msg, &msg)
	s.Assert().NoError(err, "Unmarshal on opMsg.Msg for MsgOptIn")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Equal(opMsg.Route, quarantine.ModuleName, "opMsg.Route")
	s.Assert().Equal(opMsg.Name, simulation.TypeMsgOptIn, "opMsg.Name")
	s.Assert().Equal(opMsg.Comment, "", "opMsg.Comment")
	s.Assert().Len(futureOps, 0)
}

func (s *SimTestSuite) TestSimulateMsgOptOut() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, accounts[0].Address)
	s.Require().NoError(err, "SetOptIn on accounts[0]")

	op := simulation.SimulateMsgOptOut(s.MakeTestSimState(), s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgOptOut
	err = s.app.AppCodec().Unmarshal(opMsg.Msg, &msg)
	s.Assert().NoError(err, "Unmarshal on opMsg.Msg for MsgOptOut")
	s.Assert().Equal(opMsg.Route, quarantine.ModuleName, "opMsg.Route")
	s.Assert().Equal(opMsg.Name, simulation.TypeMsgOptOut, "opMsg.Name")
	s.Assert().Equal(opMsg.Comment, "", "opMsg.Comment")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Len(futureOps, 0)
}

func (s *SimTestSuite) TestSimulateMsgAccept() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, accounts[0].Address)
	s.Require().NoError(err, "SetOptIn on accounts[0]")
	spendableCoins := s.app.BankKeeper.SpendableCoins(s.ctx, accounts[1].Address)
	toSend, err := simtypes.RandomFees(r, s.ctx, spendableCoins)
	s.Require().NoError(err, "RandomFees(%q)", spendableCoins.String())
	err = s.app.BankKeeper.SendCoins(s.ctx, accounts[1].Address, accounts[0].Address, toSend)
	s.Require().NoError(err, "SendCoins")

	op := simulation.SimulateMsgAccept(s.MakeTestSimState(), s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgAccept
	err = s.app.AppCodec().Unmarshal(opMsg.Msg, &msg)
	s.Assert().NoError(err, "Unmarshal on opMsg.Msg for MsgAccept")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Equal(opMsg.Route, quarantine.ModuleName, "opMsg.Route")
	s.Assert().Equal(opMsg.Name, simulation.TypeMsgAccept, "opMsg.Name")
	s.Assert().Equal(opMsg.Comment, "", "opMsg.Comment")
	s.Assert().Len(futureOps, 0)
}

func (s *SimTestSuite) TestSimulateMsgDecline() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, accounts[0].Address)
	s.Require().NoError(err, "SetOptIn on accounts[0]")
	spendableCoins := s.app.BankKeeper.SpendableCoins(s.ctx, accounts[1].Address)
	toSend, err := simtypes.RandomFees(r, s.ctx, spendableCoins)
	s.Require().NoError(err, "RandomFees(%q)", spendableCoins.String())
	err = s.app.BankKeeper.SendCoins(s.ctx, accounts[1].Address, accounts[0].Address, toSend)
	s.Require().NoError(err, "SendCoins")

	op := simulation.SimulateMsgDecline(s.MakeTestSimState(), s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgDecline
	err = s.app.AppCodec().Unmarshal(opMsg.Msg, &msg)
	s.Assert().NoError(err, "Unmarshal on opMsg.Msg for MsgDecline")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Equal(opMsg.Route, quarantine.ModuleName, "opMsg.Route")
	s.Assert().Equal(opMsg.Name, simulation.TypeMsgDecline, "opMsg.Name")
	s.Assert().Equal(opMsg.Comment, "", "opMsg.Comment")
	s.Assert().Len(futureOps, 0)
}

func (s *SimTestSuite) TestSimulateMsgUpdateAutoResponses() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)

	err := s.app.QuarantineKeeper.SetOptIn(s.ctx, accounts[0].Address)
	s.Require().NoError(err, "SetOptIn on accounts[0]")

	op := simulation.SimulateMsgUpdateAutoResponses(s.MakeTestSimState(), s.app.AccountKeeper, s.app.BankKeeper, s.app.QuarantineKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err, "running SimulateMsgOptIn op")

	var msg quarantine.MsgUpdateAutoResponses
	err = s.app.AppCodec().Unmarshal(opMsg.Msg, &msg)
	s.Assert().NoError(err, "Unmarshal on opMsg.Msg for MsgUpdateAutoResponses")
	s.Assert().True(opMsg.OK, "opMsg.OK")
	s.Assert().Equal(opMsg.Route, quarantine.ModuleName, "opMsg.Route")
	s.Assert().Equal(opMsg.Name, simulation.TypeMsgUpdateAutoResponses, "opMsg.Name")
	s.Assert().Equal(opMsg.Comment, "", "opMsg.Comment")
	s.Assert().Len(futureOps, 0)
	s.Assert().GreaterOrEqual(len(msg.Updates), 1, "number of updates")
}
