package antewrapper_test

import (
	sdkmath "cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
)

const (
	NHash = "nhash"
)

// These tests are kicked off by TestAnteTestSuite in testutil_test.go

func (s *AnteTestSuite) TestFlatFeeSetupDecorator_NotEnoughForMsgFee() {
	antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 99)))
	ctx := s.ctxWithFlatFeeGasMeter().WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	// Example error: "fee required: \"100stake\", fee provided \"99stake\": insufficient fee"
	s.Require().Error(err, "antehandler")
	s.Assert().ErrorContains(err, `fee required: "100stake"`)
	s.Assert().ErrorContains(err, `fee provided: "99stake"`)
	s.Assert().ErrorContains(err, "insufficient fee")
}

func (s *AnteTestSuite) TestFlatFeeSetupDecorator_IgnoresMinGasPrice() {
	antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 0)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)))

	// Set gas price to 1,000,000 stake to make sure it's not being used in the handler.
	stakePrice := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdkmath.LegacyNewDec(1_000_000))
	highGasPrice := []sdk.DecCoin{stakePrice}
	ctx := s.ctxWithFlatFeeGasMeter().WithMinGasPrices(highGasPrice).WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	s.Require().NoError(err, "antehandler")
}

func (s *AnteTestSuite) TestFlatFeeSetupDecorator_NonCheckTxPassesAllChecks() {
	antehandler := setUpApp(s, false, "bananas", 100)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150), sdk.NewInt64Coin(NHash, 88)))

	// antehandler should not error since we do not check anything in DeliverTx
	_, err := antehandler(s.ctxWithFlatFeeGasMeter(), tx, false)
	s.Require().NoError(err, "antehandler")
}

func (s *AnteTestSuite) TestFlatFeeSetupDecorator_SimulatingPassesAllChecks() {
	antehandler := setUpApp(s, true, "bananas", 100)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150), sdk.NewInt64Coin(NHash, 88)))

	// antehandler should not error since we do not check anything in Simulate
	_, err := antehandler(s.ctxWithFlatFeeGasMeter(), tx, true)
	s.Require().NoError(err, "antehandler")
}

func (s *AnteTestSuite) TestFlatFeeSetupDecorator_WrongDenomOnlyMsg() {
	antehandler := setUpApp(s, true, NHash, 100)
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 0)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	ctx := s.ctxWithFlatFeeGasMeter().WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	// Example error: "fee required: \"100nhash\", fee provided: \"10000steak\": insufficient fee"
	s.Require().Error(err, "antehandler")
	s.Assert().ErrorContains(err, `fee required: "100nhash"`)
	s.Assert().ErrorContains(err, `fee provided: "10000steak"`)
	s.Assert().ErrorContains(err, `insufficient fee`)
}

func (s *AnteTestSuite) TestFlatFeeSetupDecorator_WrongDenom() {
	antehandler := setUpApp(s, true, sdk.DefaultBondDenom, 100)
	tx, _ := createTestTx(s, sdk.NewCoins(sdk.NewInt64Coin(NHash, 190500200)))
	ctx := s.ctxWithFlatFeeGasMeter().WithChainID("test-chain")

	_, err := antehandler(ctx, tx, false)
	// Example error: "fee required: \"100stake\", fee provided: \"190500200nhash\": insufficient fee"
	s.Require().Error(err, "antehandler")
	s.Assert().ErrorContains(err, `fee required: "100stake"`)
	s.Assert().ErrorContains(err, `fee provided: "190500200nhash"`)
	s.Assert().ErrorContains(err, `insufficient fee`)
}

func createTestTx(s *AnteTestSuite, feeAmount sdk.Coins) (signing.Tx, sdk.AccountI) {
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
	s.Require().NoError(err, "CreateTestTx")
	return tx, acct1
}

func setUpApp(s *AnteTestSuite, checkTx bool, additionalFeeCoinDenom string, additionalFeeCoinAmt int64) sdk.AnteHandler {
	s.SetupTest(checkTx) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
	// create fee in stake
	if additionalFeeCoinAmt != 0 {
		newCoin := sdk.NewInt64Coin(additionalFeeCoinDenom, additionalFeeCoinAmt)
		err := s.CreateMsgFee(newCoin, &testdata.TestMsg{})
		s.Require().NoError(err, "CreateMsgFee")
	}

	// Set the x/flatfees params (for the default costs).
	err := s.app.FlatFeesKeeper.SetParams(s.ctx, flatfeestypes.Params{
		DefaultCost: sdk.NewInt64Coin(additionalFeeCoinDenom, 0),
		ConversionFactor: flatfeestypes.ConversionFactor{
			BaseAmount:      sdk.NewInt64Coin(additionalFeeCoinDenom, 1),
			ConvertedAmount: sdk.NewInt64Coin(additionalFeeCoinDenom, 1),
		},
	})

	// setup NewFlatFeeSetupDecorator
	s.Require().NoError(err, "flatfees SetParams")
	mfd := antewrapper.NewFlatFeeSetupDecorator()
	antehandler := sdk.ChainAnteDecorators(mfd)
	return antehandler
}
