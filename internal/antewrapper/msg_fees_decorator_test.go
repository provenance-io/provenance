package antewrapper_test

import (
	"fmt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/provenance-io/provenance/internal/antewrapper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// These tests are kicked off by TestAnteTestSuite in testutil_test.go

func (s *AnteTestSuite) TestMsgFeesDecoratorNotEnoughForMsgFee() {
	antehandler := setUpApp(s, true, "atom", 100)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	ctx := s.ctx.WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	// Example error: base fee + additional fee cannot be paid with provided fees: \"1nhash\", required: \"190500000nhash\" = \"190500000nhash\"(base-fee) + \"\"(additional-fees): insufficient fee: insufficient fee
	s.Require().Error(err, "antehandler")
	s.Assert().ErrorContains(err, `base fee + additional fee cannot be paid with provided fees: "100000atom"`)
	s.Assert().ErrorContains(err, `required: "100100atom"`)
	s.Assert().ErrorContains(err, `= "100000atom"(base-fee) + "100atom"(additional-fees)`)
	s.Assert().ErrorContains(err, "insufficient fee")
}

func (s *AnteTestSuite) TestMsgFeesDecoratorIgnoresMinGasPrice() {
	antehandler := setUpApp(s, true, "atom", 0)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))

	// Set gas price to 1,000,000 atom to make sure it's not being used in the handler.
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1_000_000))
	highGasPrice := []sdk.DecCoin{atomPrice}
	ctx := s.ctx.WithMinGasPrices(highGasPrice).WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	s.Require().NoError(err, "antehandler")
}

func (s *AnteTestSuite) TestMsgFeesDecoratorFloorGasPriceNotMet() {
	antehandler := setUpApp(s, true, msgfeestypes.NhashDenom, 0)
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1905)
	reqFee := int64(1905 * s.NewTestGasLimit())
	feeCoins := sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, reqFee-1))
	tx, _ := createTestTx(s, feeCoins)
	ctx := s.ctx.WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	// Example error: base fee cannot be paid with provided fees: \"190499999nhash\", required: \"190500000nhash\" = \"190500000nhash\"(base-fee) + \"\"(additional-fees): insufficient fee: insufficient fee
	s.Require().Error(err, "antehandler")
	s.Assert().ErrorContains(err, fmt.Sprintf(`base fee cannot be paid with provided fees: "%s"`, feeCoins))
	s.Assert().ErrorContains(err, fmt.Sprintf(`required: "%dnhash"`, reqFee))
	s.Assert().ErrorContains(err, fmt.Sprintf(`= "%dnhash"(base-fee) + ""(additional-fees)`, reqFee))
	s.Assert().ErrorContains(err, "insufficient fee")
}

func (s *AnteTestSuite) TestMsgFeesDecoratorFloorGasPriceMet() {
	antehandler := setUpApp(s, true, msgfeestypes.NhashDenom, 0)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	ctx := s.ctx.WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	s.Require().NoError(err, "antehandler")
}

// TODO: Move Make sure this test is being done on the ProvenanceDeductFeeDecorator
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPass() {
	s.T().Skip("Delete this once we know it's being checked where it got moved.")
	antehandler := setUpApp(s, true, "atom", 100)
	tx, acct1 := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin("atom", 100100)))

	// Set high gas price so standard test fee fails, gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	ctx := s.ctx.WithMinGasPrices(highGasPrice).WithChainID("test-chain")

	// antehandler errors with insufficient fees
	_, err := antehandler(ctx, tx, false)
	s.Require().Error(err, "antehandler insufficient funds")
	s.Assert().ErrorContains(err, "TODO")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"100100atom\"", "got wrong message")

	// Fund the account and try again.
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100100)))), "funding account")

	_, err = antehandler(ctx, tx, false)
	s.Require().NoError(err, "antehandler sufficient funds")
}

// TODO: Move Make sure this test is being done on the ProvenanceDeductFeeDecorator
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_AccountBalanceNotEnough() {
	s.T().Skip("Delete this once we know it's being checked where it got moved.")
	antehandler := setUpApp(s, true, "hotdog", 100)
	// fibbing, I don't have hotdog to pay for it right now.
	tx, acct1 := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin("hotdog", 100)))

	// Set no hotdog balance, only atom balance
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100000)))))

	// antehandler errors with insufficient fees
	_, err := antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on fee payer account not having enough balance.")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"100000atom,100hotdog\"", "got wrong message")

	// set not enough hotdog balance
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(1)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on fee payer account not having enough balance.")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"100000atom,100hotdog\"", "got wrong message")

	// set enough hotdog balance, also atom balance should be enough with prev fund
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(99)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should have errored on fee payer account not having enough balance.")
}

// TODO: Move Make sure this test is being done on the ProvenanceDeductFeeDecorator
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant() {
	s.T().Skip("Delete this once we know it's being checked where it got moved.")
	antehandler := setUpApp(s, true, "atom", 100)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acct1)

	// fee granter account
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct2 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr2)
	s.app.AccountKeeper.SetAccount(s.ctx, acct2)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	//add additional fee
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin("atom", 100100))
	gasLimit := s.NewTestGasLimit()

	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	//set fee grant
	// grant fee allowance from `addr2` to `addr1` (plenty to pay)
	err := s.app.FeeGrantKeeper.GrantAllowance(s.ctx, addr2, addr1, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 100100)),
	})
	s.txBuilder.SetFeeGranter(acct2.GetAddress())
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	ctx := s.ctx.WithMinGasPrices(highGasPrice)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, ctx, acct2.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100100)))))
	_, err = antehandler(ctx, tx, false)
	s.Require().Nil(err, "Decorator should not have errored.")
}

// TODO: I'm not entirely sure what this test is supposed to be doing
//		 But it looks like it's accidentally passing now because the thing it's
//		 checking was moved.
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant_1() {
	s.T().Skip("Delete this once we know it's being checked where it got moved.")
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	antehandler := setUpApp(s, true, "atom", 100)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acct1)

	// fee granter account
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct2 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr2)
	s.app.AccountKeeper.SetAccount(s.ctx, acct2)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	gasLimit := testdata.NewTestGasLimit()
	feeCoins := sdk.NewCoins(sdk.NewInt64Coin("atom", int64(gasLimit)+100))

	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeCoins)
	s.txBuilder.SetGasLimit(gasLimit)
	//set fee grant
	err := s.app.FeeGrantKeeper.GrantAllowance(s.ctx, addr1, addr2, &feegrant.BasicAllowance{
		SpendLimit: feeCoins,
	})
	s.Require().NoError(err, "GrantAllowance")
	s.txBuilder.SetFeeGranter(acct2.GetAddress())
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err, "CreateTestTx")

	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	ctx := s.ctx.WithMinGasPrices(highGasPrice).WithChainID("test-chain")

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100100)))))
	_, err = antehandler(ctx, tx, false)
	s.Require().NoError(err, "Decorator should not have errored.")
}

// TODO: I'm not entirely sure what this test is supposed to be doing
//		 But it looks like it's accidentally passing now because the thing it's
//		 checking was moved.
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant_2() {
	s.T().Skip("Delete this once we know it's being checked where it got moved.")
	antehandler := setUpApp(s, true, "atom", 100)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acct1)

	// fee granter account
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct2 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr2)
	s.app.AccountKeeper.SetAccount(s.ctx, acct2)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	//add additional fee
	gasLimit := testdata.NewTestGasLimit()
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin("atom", int64(gasLimit)+100))

	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	s.txBuilder.SetFeeGranter(acct2.GetAddress())
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	ctx := s.ctx.WithMinGasPrices(highGasPrice)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100100)))))
	_, err = antehandler(ctx, tx, false)
	s.Require().Error(err, "Decorator should not have errored.")
}

func (s *AnteTestSuite) TestMsgFeesDecoratorNonCheckTxPassesAllChecks() {
	antehandler := setUpApp(s, false, "bananas", 100)
	tx, _ := createTestTx(s, testdata.NewTestFeeAmount())

	// antehandler should not error since we do not check anything in DeliverTx
	_, err := antehandler(s.ctx, tx, false)
	s.Require().NoError(err, "antehandler")
}

func (s *AnteTestSuite) TestMsgFeesDecoratorSimulatingPassesAllChecks() {
	antehandler := setUpApp(s, false, "bananas", 100)
	tx, _ := createTestTx(s, testdata.NewTestFeeAmount())

	// antehandler should not error since we do not check anything in DeliverTx
	_, err := antehandler(s.ctx, tx, true)
	s.Require().NoError(err, "antehandler")
}

func (s *AnteTestSuite) TestMsgFeesDecoratorWrongDenomOnlyMsg() {
	antehandler := setUpApp(s, true, msgfeestypes.NhashDenom, 100)
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	ctx := s.ctx.WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	// Example error: base fee + additional fee cannot be paid with provided fees: "10000steak", required: "100nhash" = ""(base-fee) + "100nhash"(additional-fees): insufficient fee: insufficient fee
	s.Require().Error(err, "antehandler")
	s.Assert().ErrorContains(err, `base fee + additional fee cannot be paid with provided fees: "10000steak"`)
	s.Assert().ErrorContains(err, `required: "100nhash"`)
	s.Assert().ErrorContains(err, `= ""(base-fee) + "100nhash"(additional-fees)`)
	s.Assert().ErrorContains(err, `insufficient fee`)
}

func (s *AnteTestSuite) TestMsgFeesDecoratorFloorFromParams() {
	antehandler := setUpApp(s, true, msgfeestypes.NhashDenom, 100)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 10000)))
	ctx := s.ctx.WithChainID("test-chain")
	params := s.app.MsgFeesKeeper.GetParams(ctx)
	params.FloorGasPrice = sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1905)
	s.app.MsgFeesKeeper.SetParams(ctx, params)

	_, err := antehandler(ctx, tx, false)
	// Example error: base fee + additional fee cannot be paid with provided fees: "10000nhash", required: "190500100nhash" = "190500000nhash"(base-fee) + "100nhash"(additional-fees): insufficient fee: insufficient fee
	s.Require().Error(err, "antehandler")
	s.Assert().ErrorContains(err, `base fee + additional fee cannot be paid with provided fees: "10000nhash"`)
	s.Assert().ErrorContains(err, `required: "190500100nhash"`)
	s.Assert().ErrorContains(err, `= "190500000nhash"(base-fee) + "100nhash"(additional-fees)`)
	s.Assert().ErrorContains(err, `insufficient fee`)
}

// TODO: Make sure this check is being done on the ProvenanceDeductFeeDecorator instead.
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_4() {
	s.T().Skip("Delete this once we know it's being checked where it got moved.")
	antehandler := setUpApp(s, false, msgfeestypes.NhashDenom, 100)
	tx, acct1 := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500200)))
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(190500100)))))
	_, err := antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"190500200nhash\"", "wrong error message")

	_, err = antehandler(s.ctx, tx, false)
	s.Require().NoError(err, "antehandler after funding")
}

func (s *AnteTestSuite) TestMsgFeesDecoratorWrongDenom() {
	antehandler := setUpApp(s, true, "atom", 100)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500200)))
	ctx := s.ctx.WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	// Example error: base fee + additional fee cannot be paid with provided fees: "190500200nhash", required: "100100atom" = "100000atom"(base-fee) + "100atom"(additional-fees): insufficient fee: insufficient fee
	s.Require().Error(err, "antehandler")
	s.Assert().ErrorContains(err, `base fee + additional fee cannot be paid with provided fees: "190500200nhash"`)
	s.Assert().ErrorContains(err, `required: "100100atom"`)
	s.Assert().ErrorContains(err, `= "100000atom"(base-fee) + "100atom"(additional-fees)`)
	s.Assert().ErrorContains(err, `insufficient fee`)
}

func createTestTx(suite *AnteTestSuite, feeAmount sdk.Coins) (signing.Tx, types.AccountI) {
	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acct1)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	gasLimit := suite.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err, "CreateTestTx")
	return tx, acct1
}

func setUpApp(suite *AnteTestSuite, checkTx bool, additionalFeeCoinDenom string, additionalFeeCoinAmt int64) sdk.AnteHandler {
	suite.SetupTest(checkTx) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	// create fee in stake
	newCoin := sdk.NewInt64Coin(additionalFeeCoinDenom, additionalFeeCoinAmt)
	if additionalFeeCoinAmt != 0 {
		err := suite.CreateMsgFee(newCoin, &testdata.TestMsg{})
		suite.Require().NoError(err, "CreateMsgFee")
	}
	// setup NewMsgFeesDecorator
	mfd := antewrapper.NewMsgFeesDecorator(suite.app.MsgFeesKeeper)
	antehandler := sdk.ChainAnteDecorators(mfd)
	return antehandler
}
