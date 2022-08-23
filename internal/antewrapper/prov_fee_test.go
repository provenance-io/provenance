package antewrapper_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	pioante "github.com/provenance-io/provenance/internal/antewrapper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// These tests are kicked off by TestAnteTestSuite in testutil_test.go

func (suite *AnteTestSuite) TestEnsureMempoolFees() {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.FeeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	testaccs := suite.CreateTestAccounts(1)
	priv1 := testaccs[0].priv
	addr1 := testaccs[0].acc.GetAddress()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	highGasPrice := []sdk.DecCoin{atomPrice}
	suite.ctx = suite.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	suite.ctx = suite.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().ErrorContains(err, "insufficient fees", "Decorator should have errored on too low fee for local gasPrice")

	// Set IsCheckTx to false
	suite.ctx = suite.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Nil(err, "MempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	suite.ctx = suite.ctx.WithIsCheckTx(true)

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	suite.ctx = suite.ctx.WithMinGasPrices(lowGasPrice)

	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Nil(err, "Decorator should not have errored on fee higher than local gasPrice")
}

func (suite *AnteTestSuite) TestDeductFees() {
	suite.SetupTest(false) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin("atom", 200000))
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)
	// Set account with insufficient funds
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10)))
	err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, coins)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin("atom", 10)), suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1), "should have the new balance after funding account")

	decorators := []sdk.AnteDecorator{pioante.NewFeeMeterContextDecorator(), pioante.NewProvenanceDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, nil, suite.app.MsgFeesKeeper)}
	antehandler := sdk.ChainAnteDecorators(decorators...)

	_, err = antehandler(suite.ctx, tx, false)

	suite.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200_000))))
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin("atom", 200_010)), suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1), "Balance before tx")
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NoError(err, "Tx errored after account has been set with sufficient funds")
	suite.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin("atom", 10)), suite.app.BankKeeper.GetAllBalances(suite.ctx, addr1), "Balance after tx")
}

func (suite *AnteTestSuite) TestEnsureAdditionalFeesPaid() {
	// given
	suite.SetupTest(true)
	newCoin := sdk.NewInt64Coin("steak", 100)
	suite.CreateMsgFee(newCoin, &testdata.TestMsg{})

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// when
	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	suite.ctx.ChainID()
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	// then
	// Set the account with insufficient funds (base fee coin insufficient)
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10)))
	err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, coins)
	suite.Require().NoError(err)

	decorators := []sdk.AnteDecorator{pioante.NewFeeMeterContextDecorator(), pioante.NewProvenanceDeductFeeDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, nil, suite.app.MsgFeesKeeper)}

	antehandler := sdk.ChainAnteDecorators(decorators...)

	_, err = antehandler(suite.ctx, tx, false)

	suite.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds for base fees and but not additional fees
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200_000))))
	suite.Require().NoError(err)

	_, err = antehandler(suite.ctx, tx, false)

	suite.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// valid case
	// set gas fee and msg fees (steak)
	// Set account with sufficient funds
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err = testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(sdk.NewCoin("steak", sdk.NewInt(100))))
	suite.Require().NoError(err)

	suite.txBuilder.SetFeeAmount(NewTestFeeAmountMultiple())
	suite.txBuilder.SetGasLimit(gasLimit)

	tx, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	_, err = antehandler(suite.ctx, tx, false)

	suite.Require().Nil(err, "Tx did not error when fee payer had insufficient funds")
}

// NewTestFeeAmount is a test fee amount with multiple coins.
func NewTestFeeAmountMultiple() sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin("steak", 100))
}
