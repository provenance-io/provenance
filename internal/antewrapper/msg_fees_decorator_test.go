package antewrapper_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"

	"github.com/provenance-io/provenance/internal/antewrapper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// These tests are kicked off by TestAnteTestSuite in testutil_test.go

// checkTx true, high min gas price(high enough so that additional fee in same denom tips it over,
// and this is what sets it apart from MempoolDecorator which has already been run)
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees() {
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)))
	s.Require().NoError(err)

	// Set gas price (1 stake)
	stakePrice := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{stakePrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")
	s.Require().Contains(err.Error(), "Base Fee+additional fee cannot be paid with fee value passed in : \"100000stake\", required: \"100100stake\" = \"100000stake\"(base-fee) +\"100stake\"(additional-fees): insufficient fee", "got wrong message")
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
	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)))
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
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)))
	s.Require().NoError(err)
	// Set high gas price so standard test fee fails
	stakePrice := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(20000))
	highGasPrice := []sdk.DecCoin{stakePrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")
	s.Require().Contains(err.Error(), "Base Fee+additional fee cannot be paid with fee value passed in : \"100000stake\", required: \"2000000100stake\" = \"2000000000stake\"(base-fee) +\"100stake\"(additional-fees): insufficient fee", "got wrong message")
}

func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPass() {
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100100)))
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 stake)
	stakePrice := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{stakePrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on fee payer account not having enough balance.")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"100100stake\"", "got wrong message")

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should not have errored.")
}

func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_AccountBalanceNotEnough() {
	err, antehandler := setUpApp(s, true, "hotdog", 100)
	// fibbing, I don't have hotdog to pay for it right now.
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin("hotdog", 100)))
	s.Require().NoError(err)

	// Set no hotdog balance, only stake balance
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000)))))

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on fee payer account not having enough balance.")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"100hotdog,100000stake\"", "got wrong message")

	// set not enough hotdog balance
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(1)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should have errored on fee payer account not having enough balance.")
	s.Require().Contains(err.Error(), "fee payer account does not have enough balance to pay for \"100hotdog,100000stake\"", "got wrong message")

	// set enough hotdog balance, also stake balance should be enough with prev fund
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(99)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should have errored on fee payer account not having enough balance.")
}

func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant() {
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	s.Require().NoError(err, "should set up test app")

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
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100100))
	gasLimit := s.NewTestGasLimit()

	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	//set fee grant
	// grant fee allowance from `addr2` to `addr1` (plenty to pay)
	err = s.app.FeeGrantKeeper.GrantAllowance(s.ctx, addr2, addr1, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100100)),
	})
	s.Require().NoError(err, "should setup free grant allowance for granted account")
	s.txBuilder.SetFeeGranter(acct2.GetAddress())
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 stake)
	stakePrice := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{stakePrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct2.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should not have errored.")
}

// fee grantee incorrect
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant_1() {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	s.Require().NoError(err, "should set up test app")

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
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100100))
	gasLimit := testdata.NewTestGasLimit()

	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	//set fee grant
	// grant fee allowance from `addr1` to `addr2` // wrong grant
	err = s.app.FeeGrantKeeper.GrantAllowance(s.ctx, addr1, addr2, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100100)),
	})
	s.Require().NoError(err, "should setup free grant allowance for granted account")
	s.txBuilder.SetFeeGranter(acct2.GetAddress())
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 stake)
	stakePrice := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{stakePrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored.")
}

// no grant
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFeesPassFeeGrant_2() {
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	s.Require().NoError(err, "should set up test app")

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
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100100))
	gasLimit := testdata.NewTestGasLimit()

	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)
	s.txBuilder.SetFeeGranter(acct2.GetAddress())
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails, gas price (1 stake)
	stakePrice := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{stakePrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored.")
}

func (s *AnteTestSuite) TestEnsureNonCheckTxPassesAllChecks() {
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	tx, _ := createTestTx(s, err, testdata.NewTestFeeAmount())
	s.Require().NoError(err)

	// Set IsCheckTx to false
	s.ctx = s.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "MsgFeesDecorator did not return error in DeliverTx")
}

func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_1() {
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	tx, acct1 := createTestTx(s, err, NewTestFeeAmountMultiple())

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("steak", sdk.NewInt(100)))), "should fund account for test setup")
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(150)))), "should fund account for test setup")
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NoError(err, "MsgFeesDecorator returned error in DeliverTx")
}

// wrong denom passed in, errors with insufficient fee
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_2() {
	err, antehandler := setUpApp(s, false, msgfeestypes.NhashDenom, 100)
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)

	_, acct1 := createTestTx(s, err, NewTestFeeAmountMultiple())

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("steak", sdk.NewInt(10000)))), "should fund account for test setup")

	tx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().EqualError(err, "not enough fees; after deducting fees required,got: \"-100nhash,10000steak\", required additional fee: \"100nhash\": insufficient fee", "wrong error message")

	hashPrice := sdk.NewDecCoinFromDec(msgfeestypes.NhashDenom, sdk.NewDec(100))
	lowGasPrice := []sdk.DecCoin{hashPrice}
	s.ctx = s.ctx.WithMinGasPrices(lowGasPrice)

	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	s.Require().EqualError(err, "not enough fees; after deducting fees required,got: \"-100nhash,10000steak\", required additional fee: \"100nhash\": insufficient fee", "wrong error message")
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
	err, antehandler := setUpApp(s, false, sdk.DefaultBondDenom, 100)
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500200)))
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(190500100)))))
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	s.Require().Contains(err.Error(), "not enough fees; after deducting fees required,got: \"190500200nhash,-100stake\", required additional fee: \"100stake\"", "wrong error message")
}

// checkTx true, but additional fee not provided
func (s *AnteTestSuite) TestEnsureMempoolAndMsgFees_6() {
	err, antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	tx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500200)))
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(190500100)))))
	hashPrice := sdk.NewDecCoinFromDec(msgfeestypes.NhashDenom, sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{hashPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NotNil(err, "Decorator should not have errored for insufficient additional fee")
	s.Require().Contains(err.Error(), "not enough fees; after deducting fees required,got: \"190500200nhash,-100stake\", required fees: \"100000nhash,100stake\" = \"100000nhash\"(base-fee) +\"100stake\"(additional-fees): insufficient fee: insufficient fee", "wrong error message")

}

func (s *AnteTestSuite) TestCalculateAdditionalFeesToBePaid() {
	err, _ := setUpApp(s, false, sdk.DefaultBondDenom, 100)
	s.Require().NoError(err)

	someAddress := s.clientCtx.FromAddress
	recipient1 := "recipient1"
	recipient2 := "recipient2"
	sendTypeURL := sdk.MsgTypeURL(&banktypes.MsgSend{})
	assessFeeTypeURL := sdk.MsgTypeURL(&msgfeestypes.MsgAssessCustomMsgFeeRequest{})
	oneHash := sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_000_000_000)
	twoHash := sdk.NewInt64Coin(msgfeestypes.NhashDenom, 2_000_000_000)
	msgSend := banktypes.NewMsgSend(someAddress, someAddress, sdk.NewCoins(oneHash))
	assessFeeWithRecipient1 := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("", oneHash, recipient1, someAddress.String())
	assessFeeWithRecipient2 := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("", oneHash, recipient2, someAddress.String())
	assessFeeWithoutRecipient := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("", oneHash, "", someAddress.String())
	s.app.MsgFeesKeeper.SetMsgFee(s.ctx, msgfeestypes.NewMsgFee(sendTypeURL, oneHash, "", msgfeestypes.DefaultMsgFeeBips))

	type RecipientDistribution struct {
		recipient string
		amount    sdk.Coin
	}

	testCases := []struct {
		name                         string
		msgs                         []sdk.Msg
		expectedTotalAdditionalFees  sdk.Coins
		expectedAdditionalModuleFees sdk.Coins
		expectedRecipients           []RecipientDistribution
		addAdditionalMsgFee          msgfeestypes.MsgFee
		expectErrMsg                 string
	}{
		{
			"should calculate single msg with fee and no recipients",
			[]sdk.Msg{msgSend},
			sdk.NewCoins(oneHash),
			sdk.NewCoins(oneHash),
			[]RecipientDistribution{},
			msgfeestypes.MsgFee{},
			"",
		},
		{
			"should calculate two msgs with fee and no recipients",
			[]sdk.Msg{msgSend, msgSend},
			sdk.NewCoins(twoHash),
			sdk.NewCoins(twoHash),
			[]RecipientDistribution{},
			msgfeestypes.MsgFee{},
			"",
		},
		{
			"should calculate single msg with fee and an assess msg with no recipients (all fees should go to fee module)",
			[]sdk.Msg{
				msgSend,
				&assessFeeWithoutRecipient,
			},
			sdk.NewCoins(twoHash),
			sdk.NewCoins(twoHash),
			[]RecipientDistribution{},
			msgfeestypes.MsgFee{},
			"",
		},
		{
			"should calculate single msg with fee and single recipient which claims half the fee",
			[]sdk.Msg{
				msgSend,
				&assessFeeWithRecipient1,
			},
			sdk.NewCoins(twoHash),
			sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_500_000_000)),
			[]RecipientDistribution{
				{
					amount:    sdk.NewInt64Coin(msgfeestypes.NhashDenom, 500_000_000),
					recipient: recipient1,
				},
			},
			msgfeestypes.MsgFee{},
			"",
		},
		{
			"should calculate single msg with fee and assess fee msgs with two recipients which claims half the fee each",
			[]sdk.Msg{
				msgSend,
				&assessFeeWithRecipient1,
				&assessFeeWithRecipient2,
			},
			sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 3_000_000_000)),
			sdk.NewCoins(twoHash),
			[]RecipientDistribution{
				{
					amount:    sdk.NewInt64Coin(msgfeestypes.NhashDenom, 500_000_000),
					recipient: recipient1,
				},
				{
					amount:    sdk.NewInt64Coin(msgfeestypes.NhashDenom, 500_000_000),
					recipient: recipient2,
				},
			},
			msgfeestypes.MsgFee{},
			"",
		},
		{
			"should calculate single msg with fee, a assess fee msg that has a msg fee without a recipient",
			[]sdk.Msg{
				msgSend,
				&assessFeeWithoutRecipient,
			},
			sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 3_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 3_000_000_000)),
			[]RecipientDistribution{},
			msgfeestypes.NewMsgFee(assessFeeTypeURL, oneHash, "", msgfeestypes.DefaultMsgFeeBips),
			"",
		},
		{
			"should calculate single msg with fee, a assess fee msg that has a msg fee with a recipient which claims half the fee",
			[]sdk.Msg{
				msgSend,
				&assessFeeWithRecipient1,
			},
			sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 3_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 2_500_000_000)),
			[]RecipientDistribution{
				{
					amount:    sdk.NewInt64Coin(msgfeestypes.NhashDenom, 500_000_000),
					recipient: recipient1,
				},
			},
			msgfeestypes.NewMsgFee(assessFeeTypeURL, oneHash, "", msgfeestypes.DefaultMsgFeeBips),
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			if len(tc.addAdditionalMsgFee.MsgTypeUrl) != 0 {
				s.app.MsgFeesKeeper.SetMsgFee(s.ctx, tc.addAdditionalMsgFee)
			}
			result, err := antewrapper.CalculateAdditionalFeesToBePaid(s.ctx, s.app.MsgFeesKeeper, tc.msgs...)
			if len(tc.expectErrMsg) == 0 {
				s.Require().NoError(err, "should calculate additional fee charges")
				s.Assert().Equal(tc.expectedTotalAdditionalFees, result.TotalAdditionalFees, "should have total additional fee amount equal to sum of msg fees and assess fee")
				s.Assert().Equal(tc.expectedAdditionalModuleFees, result.AdditionalModuleFees, "should send total of additional fee amount that is distributed to fee module account")
				s.Assert().Equal(len(tc.expectedRecipients), len(result.RecipientDistributions), "should contain all recipients that get a split of fee")
				for _, rd := range tc.expectedRecipients {
					s.Assert().Equal(rd.amount, result.RecipientDistributions[rd.recipient], "should have allocated proper funds to recipient")
				}
			}
		})
	}

}

func (s *AnteTestSuite) TestCalculateDistributions() {
	testCases := []struct {
		name                                string
		recipient                           string
		additionalFee                       sdk.Coin
		basisPoints                         uint32
		msgFeesDistribution                 antewrapper.MsgFeesDistribution
		expectedAdditionalModuleFees        sdk.Coins
		expectedTotalAdditionalFees         sdk.Coins
		expectedTotalRecipientDistributions map[string]sdk.Coin
		expectErrMsg                        string
	}{
		{
			"should calculate distributions without a recipient",
			"",
			sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000),
			5_000,
			antewrapper.MsgFeesDistribution{},
			sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
			sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
			make(map[string]sdk.Coin),
			"",
		},
		{
			"should fail to calculate distributions, recipient with invalid bip amount passed in (0 - 10,000 are valid)",
			"cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
			sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000),
			10_001,
			antewrapper.MsgFeesDistribution{},
			sdk.NewCoins(),
			sdk.NewCoins(),
			make(map[string]sdk.Coin),
			"invalid: 10001: invalid bips amount",
		},
		{
			"should calculate distributions with a recipient and valid bips",
			"cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
			sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000),
			5_000,
			antewrapper.MsgFeesDistribution{
				RecipientDistributions: make(map[string]sdk.Coin),
			},
			sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500)),
			sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000)),
			map[string]sdk.Coin{"cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27": sdk.NewInt64Coin(sdk.DefaultBondDenom, 500)},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			err := antewrapper.CalculateDistributions(tc.recipient, tc.additionalFee, tc.basisPoints, &tc.msgFeesDistribution)
			if len(tc.expectErrMsg) == 0 {
				s.Require().NoError(err, "should calculate additional fee charges")
				s.Assert().True(tc.expectedAdditionalModuleFees.IsEqual(tc.msgFeesDistribution.AdditionalModuleFees), "Additional Module Fees should be equal %v : %v", tc.expectedAdditionalModuleFees, tc.msgFeesDistribution.AdditionalModuleFees)
				s.Assert().True(tc.expectedTotalAdditionalFees.IsEqual(tc.msgFeesDistribution.TotalAdditionalFees), "Total Additional Fees should be equal %v : %v", tc.expectedTotalAdditionalFees, tc.msgFeesDistribution.TotalAdditionalFees)
				s.Assert().True(tc.expectedTotalRecipientDistributions[tc.recipient].IsEqual(tc.msgFeesDistribution.RecipientDistributions[tc.recipient]), "Recipient Distributions are not equal %v : %v", tc.expectedTotalRecipientDistributions[tc.recipient], tc.msgFeesDistribution.RecipientDistributions[tc.recipient])
			} else {
				s.Require().EqualError(err, tc.expectErrMsg)
			}
		})
	}

}

func createTestTx(s *AnteTestSuite, err error, feeAmount sdk.Coins) (signing.Tx, types.AccountI) {
	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acct1)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	gasLimit := s.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	return tx, acct1
}

func setUpApp(s *AnteTestSuite, checkTx bool, additionalFeeCoinDenom string, additionalFeeCoinAmt int64) (error, sdk.AnteHandler) {
	s.SetupTest(checkTx) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
	// create fee in stake
	newCoin := sdk.NewInt64Coin(additionalFeeCoinDenom, additionalFeeCoinAmt)
	if additionalFeeCoinAmt != 0 {
		err := s.CreateMsgFee(newCoin, &testdata.TestMsg{})
		if err != nil {
			panic(err)
		}
	}
	// setup NewMsgFeesDecorator
	app := s.app
	mfd := antewrapper.NewMsgFeesDecorator(app.BankKeeper, app.AccountKeeper, app.FeeGrantKeeper, app.MsgFeesKeeper)
	antehandler := sdk.ChainAnteDecorators(mfd)
	return nil, antehandler
}
