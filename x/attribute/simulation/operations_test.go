package simulation_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"

	"github.com/provenance-io/provenance/x/attribute/simulation"
	types "github.com/provenance-io/provenance/x/attribute/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
}

func (suite *SimTestSuite) SetupTest() {
	checkTx := false
	app := app.Setup(checkTx)
	suite.app = app
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{})
}

func (suite *SimTestSuite) TestWeightedOperations() {
	cdc := suite.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams, cdc, suite.app.AttributeKeeper,
		suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.NameKeeper,
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simappparams.DefaultWeightMsgAddAttribute, types.ModuleName, types.TypeMsgAddAttribute},
		{simappparams.DefaultWeightMsgUpdateAttribute, types.ModuleName, types.TypeMsgUpdateAttribute},
		{simappparams.DefaultWeightMsgDeleteAttribute, types.ModuleName, types.TypeMsgDeleteAttribute},
		{simappparams.DefaultWeightMsgDeleteDistinctAttribute, types.ModuleName, types.TypeMsgDeleteDistinctAttribute},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, "")
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgAddAttribute tests the normal scenario of a valid message of type TypeMsgAddAttribute.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgAddAttribute() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)
	suite.app.NameKeeper.SetNameRecord(suite.ctx, "example.provenance", accounts[0].Address, false)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgAddAttribute(suite.app.AttributeKeeper, suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.NameKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgAddAttributeRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	fmt.Println(msg.Account)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(accounts[0].Address.String(), msg.Account)
	suite.Require().Equal(accounts[0].Address.String(), msg.Owner)
	suite.Require().Equal("example.provenance", msg.Name)
	suite.Require().Equal(types.AttributeType_Uri, msg.AttributeType)
	suite.Require().Equal([]byte("http://www.example.com/"), msg.Value)
	suite.Require().Equal(types.TypeMsgAddAttribute, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateMsgUpdateAttribute() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)
	suite.app.NameKeeper.SetNameRecord(suite.ctx, "example.provenance", accounts[0].Address, false)
	suite.app.AttributeKeeper.SetAttribute(suite.ctx, types.NewAttribute("example.provenance", accounts[1].Address, types.AttributeType_String, []byte("test")), accounts[0].Address)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgUpdateAttribute(suite.app.AttributeKeeper, suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.NameKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgUpdateAttributeRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.TypeMsgUpdateAttribute, msg.Type())
	suite.Require().Equal("example.provenance", msg.Name)
	suite.Require().Equal(accounts[0].Address.String(), msg.Owner)
	suite.Require().Equal(accounts[1].Address.String(), msg.Account)
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateMsgDeleteAttribute() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)
	suite.app.NameKeeper.SetNameRecord(suite.ctx, "example.provenance", accounts[0].Address, false)
	suite.app.AttributeKeeper.SetAttribute(suite.ctx, types.NewAttribute("example.provenance", accounts[1].Address, types.AttributeType_String, []byte("test")), accounts[0].Address)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgDeleteAttribute(suite.app.AttributeKeeper, suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.NameKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgDeleteAttributeRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.TypeMsgDeleteAttribute, msg.Type())
	suite.Require().Equal("example.provenance", msg.Name)
	suite.Require().Equal(accounts[0].Address.String(), msg.Owner)
	suite.Require().Equal(accounts[1].Address.String(), msg.Account)
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateMsgDeleteDistinctAttribute() {

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// generate attr address
	privKey, _ := secp256r1.GenPrivKey()
	attrAddr, _ := sdk.AccAddressFromHex(privKey.PubKey().Address().String())

	suite.app.NameKeeper.SetNameRecord(suite.ctx, "example.provenance", accounts[0].Address, false)
	suite.app.AttributeKeeper.SetAttribute(suite.ctx, types.NewAttribute("example.provenance", attrAddr, types.AttributeType_String, []byte("test")), accounts[0].Address)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgDeleteDistinctAttribute(suite.app.AttributeKeeper, suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.NameKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgDeleteDistinctAttributeRequest
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.TypeMsgDeleteDistinctAttribute, msg.Type())
	suite.Require().Equal("example.provenance", msg.Name)
	suite.Require().Equal(accounts[0].Address.String(), msg.Owner)
	suite.Require().Equal(attrAddr.String(), msg.Account)
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(300, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		app.FundAccount(suite.app, suite.ctx, account.Address, initCoins)
		err := app.FundAccount(suite.app, suite.ctx, account.Address, initCoins)
		suite.Require().NoError(err)
	}

	return accounts
}

// RandomAccounts generates n random accounts
func RandomAccounts(r *rand.Rand, n int) []simtypes.Account {
	accs := make([]simtypes.Account, n)

	for i := 0; i < n; i++ {
		// don't need that much entropy for simulation
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)

		accs[i].PrivKey, _ = secp256r1.GenPrivKey()
		accs[i].PubKey = accs[i].PrivKey.PubKey()
		accs[i].Address = sdk.AccAddress(accs[i].PubKey.Address())

		// accs[i].ConsKey = ed25519.GenPrivKeyFromSecret(privkeySeed)
	}

	return accs
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
