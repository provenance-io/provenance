package antewrapper_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/antewrapper"
)

func (suite *AnteTestSuite) TestEnsureMempoolAndMsgBasedFees() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	newCoin := sdk.NewInt64Coin("steak", 100)
	suite.CreateMsgFee(newCoin,&testdata.TestMsg{})

	// setup
	app := suite.app
	mfd := antewrapper.NewMsgBasedFeeDecorator(app.BankKeeper,app.AccountKeeper,app.FeeGrantKeeper,app.MsgBasedFeeKeeper)
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acct1)
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
	suite.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")

	// Set IsCheckTx to false
	suite.ctx = suite.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NotNil(err, "MsgBasedFeeDecorator did not return error in DeliverTx")

	simapp.FundAccount(app, suite.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))))
	suite.txBuilder.SetFeeAmount(NewTestFeeAmountMultiple())
	tx, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Nil(err, "MsgBasedFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	suite.ctx = suite.ctx.WithIsCheckTx(true)

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	suite.ctx = suite.ctx.WithMinGasPrices(lowGasPrice)

	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Nil(err, "Decorator should not have errored on fee higher than local gasPrice")


	suite.txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("nhash", 150), sdk.NewInt64Coin("steak",100)))
	tx, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	tx, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	_, err = antehandler(suite.ctx, tx, false)

	newCoin = sdk.NewInt64Coin("nhash", 100)
	suite.CreateMsgFee(newCoin,&testdata.TestMsg{})
	hashPrice := sdk.NewDecCoinFromDec("nhash", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice = []sdk.DecCoin{hashPrice}
	suite.ctx = suite.ctx.WithMinGasPrices(lowGasPrice)
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NotNil(err, "Decorator should not have errored on fee higher than local gasPrice")
	simapp.FundAccount(app, suite.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("nhash", sdk.NewInt(190500000))))
	suite.txBuilder.SetFeeAmount( sdk.NewCoins(sdk.NewInt64Coin("nhash", 100)))
	// Set IsCheckTx to true
	suite.ctx = suite.ctx.WithIsCheckTx(false)
	tx, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NotNil(err, "Decorator should not have errored on fee higher than local gasPrice")

	suite.txBuilder.SetFeeAmount( sdk.NewCoins(sdk.NewInt64Coin("nhash", 190500200)))
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NotNil(err, "Decorator should have errored on fee higher than what the account has")
	simapp.FundAccount(app, suite.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("nhash", sdk.NewInt(200))))
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Nil(err, "Decorator should not have errored on fee higher than what the account has")

}
