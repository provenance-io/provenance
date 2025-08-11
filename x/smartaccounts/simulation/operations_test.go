package simulation_test

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/smartaccounts/simulation"
	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
	cdc *codec.ProtoCodec
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

func (s *SimTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false)
	s.cdc = codec.NewProtoCodec(s.app.InterfaceRegistry())
}

// LogOperationMsg logs all fields of the provided operationMsg.
func (s *SimTestSuite) LogOperationMsg(operationMsg simtypes.OperationMsg) {
	msgFmt := "%s"
	if len(bytes.TrimSpace(operationMsg.Msg)) == 0 {
		msgFmt = "    %q"
	}
	fmtLines := []string{
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
	weightedOps := simulation.WeightedOperations(
		s.MakeTestSimState(),
		s.app.SmartAccountKeeper,
		s.app.BankKeeper,
		codec.NewProtoCodec(s.app.InterfaceRegistry()),
	)

	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accs := s.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{weight: simappparams.DefaultWeightMsgRegisterWebAuthnCredential, opMsgRoute: types.ModuleName, opMsgName: sdk.MsgTypeURL(&types.MsgRegisterFido2Credential{})},
		{weight: simappparams.DefaultWeightMsgRegisterCosmosCredential, opMsgRoute: types.ModuleName, opMsgName: sdk.MsgTypeURL(&types.MsgRegisterCosmosCredential{})},
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
	s.Require().Equal(expNames, actualNames, "operation message names")

	for i := range expected {
		s.Require().Equal(expected[i].weight, weightedOps[i].Weight(), "weightedOps[%d].Weight", i)
		s.Require().Equal(expected[i].opMsgRoute, opMsgs[i].Route, "opMsgs[%d].Route", i)
		s.Require().Equal(expected[i].opMsgName, opMsgs[i].Name, "opMsgs[%d].Name", i)
	}
}

// TestSimulateMsgRegisterCosmosCredential tests the normal scenario of a valid message of type MsgRegisterCosmosCredential.
func (s *SimTestSuite) TestSimulateMsgRegisterCosmosCredential() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	// Create the operation for MsgRegisterCosmosCredential
	args := &simulation.WeightedOpsArgs{
		SimState:           s.MakeTestSimState(),
		ProtoCodec:         codec.NewProtoCodec(s.app.InterfaceRegistry()),
		Smartaccountkeeper: s.app.SmartAccountKeeper,
		Bankkeeper:         s.app.BankKeeper,
	}
	op := simulation.SimulateMsgRegisterCosmosCredential(s.app.SmartAccountKeeper, args)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err)

	var msg types.MsgRegisterCosmosCredential
	s.Require().NoError(s.app.AppCodec().Unmarshal(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Require().True(operationMsg.OK, operationMsg.String())
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	s.Require().Equal(types.ModuleName, operationMsg.Route)
	s.Require().Len(futureOperations, 0)
	s.Require().NotNil(msg.Pubkey, "Expected a non-nil public key")
	s.Require().NotEmpty(msg.Sender, "Expected a non-empty sender address")
}

// TestSimulateMsgRegisterWebAuthnAccount tests the normal scenario of a valid message of type MsgRegisterFido2Credential.
func (s *SimTestSuite) TestSimulateMsgRegisterWebAuthnAccount() {
	// setup 3 accounts
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 3)

	// Create the operation for MsgRegisterFido2Credential
	args := &simulation.WeightedOpsArgs{
		SimState:           s.MakeTestSimState(),
		ProtoCodec:         codec.NewProtoCodec(s.app.InterfaceRegistry()),
		Smartaccountkeeper: s.app.SmartAccountKeeper,
		Bankkeeper:         s.app.BankKeeper,
	}
	op := simulation.SimulateMsgRegisterWebAuthnAccount(s.app.SmartAccountKeeper, args)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accounts, "")
	s.Require().NoError(err)

	var msg types.MsgRegisterFido2Credential
	s.Require().NoError(s.app.AppCodec().Unmarshal(operationMsg.Msg, &msg), "UnmarshalJSON(operationMsg.Msg)")

	s.Require().True(operationMsg.OK, operationMsg.String())
	s.Require().Equal(sdk.MsgTypeURL(&msg), operationMsg.Name)
	s.Require().Equal(types.ModuleName, operationMsg.Route)
	s.Require().Len(futureOperations, 0)
	s.Require().NotEmpty(msg.EncodedAttestation, "Expected non-empty encoded attestation")
	s.Require().NotEmpty(msg.UserIdentifier, "Expected non-empty user identifier")
	s.Require().NotEmpty(msg.Sender, "Expected a non-empty sender address")
}

func (s *SimTestSuite) TestDispatch() {
	// setup test account
	src := rand.NewSource(1)
	r := rand.New(src)
	accounts := s.getTestingAccounts(r, 1)
	sender := accounts[0]

	// Generate random public key for credential
	credPubKey, err := codectypes.NewAnyWithValue(simulation.GenRandomSecp256k1PubKey())
	s.Require().NoError(err)

	// Create the message
	msg := &types.MsgRegisterCosmosCredential{
		Sender: sender.Address.String(),
		Pubkey: credPubKey,
	}

	// Execute dispatch
	opMsg, fops, err := simulation.Dispatch(r, s.app.BaseApp, s.ctx, s.MakeTestSimState(),
		sender, "", msg, s.app.AccountKeeper, s.app.BankKeeper, "test dispatch")

	s.Require().NoError(err)
	s.Require().Equal("test dispatch", opMsg.Comment)
	s.Require().Len(fops, 0)
	s.Require().True(opMsg.OK)
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	return testutil.GenerateTestingAccounts(s.T(), s.ctx, s.app, r, n)
}
