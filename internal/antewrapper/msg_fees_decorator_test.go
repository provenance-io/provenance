package antewrapper_test

import (
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

// checkTx true, high min gas price(high enough so that additional fee in same denom tips it over,
// and this is what sets it apart from MempoolDecorator which has already been run)
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees() {
	err, antehandler := setUpApp(s, true, "atom", 100)
	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	s.Require().NoError(err)

	// Set gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")
	s.Require().Contains(err.Error(), "Base Fee+additional fee cannot be paid with fee value passed in : \"100000atom\", required: \"100100atom\" = \"100000atom\"(base-fee) +\"100atom\"(additional-fees): insufficient fee", "got wrong message")
}

// checkTx true, fees supplied that do not meet floor gas price requirement (floor gas price * gas wanted)
func (s *AnteTestSuite) TestEnsureFloorGasPriceNotMet() {
	err, antehandler := setUpApp(s, true, msgfeestypes.NhashDenom, 0)
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1905)
	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 100000)))
	s.Require().NoError(err)
	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)
	s.ctx = s.ctx.WithChainID("test-chain")

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")
	s.Require().Contains(err.Error(), "not enough fees based on floor gas price: \"1905nhash\"; required base fees >=\"190500000nhash\": Supplied fee was \"100000nhash\": insufficient fee", "got wrong message")
}

// checkTx true, fees supplied that do meet floor gas price requirement (floor gas price * gas wanted), and should not error.
func (s *AnteTestSuite) TestEnsureFloorGasPriceMet() {
	err, antehandler := setUpApp(s, true, msgfeestypes.NhashDenom, 0)
	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	s.Require().NoError(err)
	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)
	s.ctx = s.ctx.WithChainID("test-chain")

	gasPrice := []sdk.DecCoin{sdk.NewDecCoinFromDec(msgfeestypes.NhashDenom, sdk.NewDec(1905))}
	s.ctx = s.ctx.WithMinGasPrices(gasPrice)
	// antehandler does not error.
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Should not have errored")
}

// checkTx true, high min gas price(high enough so that additional fee in same denom tips it over,
// and this is what sets it apart from MempoolDecorator which has already been run)
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesNoAdditionalFeesLowGas() {
	err, antehandler := setUpApp(s, true, msgfeestypes.NhashDenom, 0)
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1905)
	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 100000)))
	s.Require().NoError(err)
	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)
	s.ctx = s.ctx.WithChainID("test-chain")

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")
	s.Require().Contains(err.Error(), "not enough fees based on floor gas price: \"1905nhash\"; required base fees >=\"190500000nhash\": Supplied fee was \"100000nhash\": insufficient fee", "got wrong message")
}

// checkTx true, high min gas price irrespective of additional fees
func (s *AnteTestSuite) TestEnsureMempoolHighMinGasPrice() {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	err, antehandler := setUpApp(s, true, "atom", 100)
	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	s.Require().NoError(err)
	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(20000))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")
	s.Require().Contains(err.Error(), "Base Fee+additional fee cannot be paid with fee value passed in : \"100000atom\", required: \"2000000100atom\" = \"2000000000atom\"(base-fee) +\"100atom\"(additional-fees): insufficient fee", "got wrong message")
}

func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPass() {
	err, antehandler := setUpApp(s, true, "atom", 100)
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100100)))
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on fee payer account not having enough balance.")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"100100atom\"", "got wrong message")

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should not have errored.")
}

func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_AccountBalanceNotEnough() {
	err, antehandler := setUpApp(s, true, "hotdog", 100)
	// fibbing, I don't have hotdog to pay for it right now.
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin("hotdog", 100)))
	s.Require().NoError(err)

	// Set no hotdog balance, only atom balance
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100000)))))

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
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

func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant() {
	err, antehandler := setUpApp(s, true, "atom", 100)

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
	err = s.app.FeeGrantKeeper.GrantAllowance(s.ctx, addr2, addr1, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 100100)),
	})
	s.txBuilder.SetFeeGranter(acct2.GetAddress())
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct2.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should not have errored.")
}

// fee grantee incorrect
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant_1() {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	err, antehandler := setUpApp(s, true, "atom", 100)

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
	gasLimit := testdata.NewTestGasLimit()

	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	//set fee grant
	// grant fee allowance from `addr1` to `addr2` // wrong grant
	err = s.app.FeeGrantKeeper.GrantAllowance(s.ctx, addr1, addr2, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 100100)),
	})
	s.txBuilder.SetFeeGranter(acct2.GetAddress())
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored.")
}

// no grant
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant_2() {
	err, antehandler := setUpApp(s, true, "atom", 100)

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
	gasLimit := testdata.NewTestGasLimit()

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
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored.")
}

func (s *AnteTestSuite) TestEnsureNonCheckTxPassesAllChecks() {
	err, antehandler := setUpApp(s, true, "atom", 100)
	tx, _ := createTestTx(s, err, testdata.NewTestFeeAmount())
	s.Require().NoError(err)

	// Set IsCheckTx to false
	s.ctx = s.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "MsgFeesDecorator did not return error in DeliverTx")
}

func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_1() {
	err, antehandler := setUpApp(s, true, "atom", 100)
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	tx, acct1 := createTestTx(s, err, NewTestFeeAmountMultiple())

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("steak", sdk.NewInt(100)))))
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(150)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "MsgFeesDecorator returned error in DeliverTx")
}

// wrong denom passed in, errors with insufficient fee
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_2() {
	err, antehandler := setUpApp(s, false, msgfeestypes.NhashDenom, 100)
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	tx, acct1 := createTestTx(s, err, testdata.NewTestFeeAmount())

	tx, acct1 = createTestTx(s, err, NewTestFeeAmountMultiple())

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("steak", sdk.NewInt(10000)))))

	tx, acct1 = createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	_, err = antehandler(s.ctx, tx, false)

	hashPrice := sdk.NewDecCoinFromDec(msgfeestypes.NhashDenom, sdk.NewDec(100))
	lowGasPrice := []sdk.DecCoin{hashPrice}
	s.ctx = s.ctx.WithMinGasPrices(lowGasPrice)

	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	s.Require().Contains(err.Error(), "not enough fees; after deducting fees required,got: \"-100nhash,10000steak\", required additional fee: \"100nhash\"", "wrong error message")
}

// additional fee denom as default base fee denom, fails because gas passed in * floor gas price (module param) exceeds fees passed in.
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_3() {
	err, antehandler := setUpApp(s, false, msgfeestypes.NhashDenom, 100)
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 10000)))
	s.ctx = s.ctx.WithChainID("test-chain")
	params := s.app.MsgFeesKeeper.GetParams(s.ctx)
	params.FloorGasPrice = sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1905)
	s.app.MsgFeesKeeper.SetParams(s.ctx, params)
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(10000)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	// TODO revisit this
	s.Require().Contains(err.Error(), "not enough fees based on floor gas price: \"1905nhash\"; after deducting (total fee supplied fees - additional fees(\"100nhash\")) required base fees >=\"190500000nhash\": Supplied fee was \"9900nhash\": insufficient fee", "wrong error message")
}

// additional fee same denom as default base fee denom, fails because of insufficient fee and then passes when enough fee is present.
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_4() {
	err, antehandler := setUpApp(s, false, msgfeestypes.NhashDenom, 100)
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500200)))
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(190500100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"190500200nhash\"", "wrong error message")
	// add some nhash
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should not have errored for insufficient additional fee")
}

// additional fee same denom as default base fee denom, fails because of insufficient fee and then passes when enough fee is present.
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_5() {
	err, antehandler := setUpApp(s, false, "atom", 100)
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500200)))
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(190500100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	s.Require().Contains(err.Error(), "not enough fees; after deducting fees required,got: \"-100atom,190500200nhash\", required additional fee: \"100atom\"", "wrong error message")
}

// checkTx true, but additional fee not provided
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_6() {
	err, antehandler := setUpApp(s, true, "atom", 100)
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500200)))
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(190500100)))))
	hashPrice := sdk.NewDecCoinFromDec(msgfeestypes.NhashDenom, sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{hashPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	s.Require().Contains(err.Error(), "not enough fees; after deducting fees required,got: \"-100atom,190500200nhash\", required fees: \"100atom,100000nhash\" = \"100000nhash\"(base-fee) +\"100atom\"(additional-fees): insufficient fee", "wrong error message")

}

func createTestTx(suite *AnteTestSuite, err error, feeAmount sdk.Coins) (signing.Tx, types.AccountI) {
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
	return tx, acct1
}

func setUpApp(suite *AnteTestSuite, checkTx bool, additionalFeeCoinDenom string, additionalFeeCoinAmt int64) (error, sdk.AnteHandler) {
	suite.SetupTest(checkTx) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	// create fee in stake
	newCoin := sdk.NewInt64Coin(additionalFeeCoinDenom, additionalFeeCoinAmt)
	if additionalFeeCoinAmt != 0 {
		err := suite.CreateMsgFee(newCoin, &testdata.TestMsg{})
		if err != nil {
			panic(err)
		}
	}
	// setup NewMsgFeesDecorator
	app := suite.app
	mfd := antewrapper.NewMsgFeesDecorator(app.MsgFeesKeeper)
	antehandler := sdk.ChainAnteDecorators(mfd)
	return nil, antehandler
}
