package antewrapper_test

import (
	"fmt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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
