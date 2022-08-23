package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdkgas "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/antewrapper"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
	msgfeetype "github.com/provenance-io/provenance/x/msgfees/types"
)

func (suite *HandlerTestSuite) TestMsgFeeHandlerFeeChargedNoRemainingBaseFee() {
	encodingConfig, err := setUpApp(suite, "atom", 100)
	testTx, acct1 := createTestTx(suite, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	suite.Require().NoError(err, "txEncoder")

	suite.ctx = suite.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	suite.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(100000))})
	}, "panicked on adding fees")
	suite.ctx = suite.ctx.WithGasMeter(feeGasMeter)
	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  suite.app.AccountKeeper,
		BankKeeper:     suite.app.BankKeeper,
		FeegrantKeeper: suite.app.FeeGrantKeeper,
		MsgFeesKeeper:  suite.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	suite.Require().NoError(err)
	coins, _, err := feeChargeFn(suite.ctx, false)

	suite.Require().ErrorContains(err, "0nhash is smaller than 1000000nhash: insufficient funds: insufficient funds", "feeChargeFn 1")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	suite.Require().True(coins.IsZero(), "coins.IsZero() 1")

	suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(900000)))), "fund account")
	coins, _, err = feeChargeFn(suite.ctx, false)
	suite.Require().ErrorContains(err, "900000nhash is smaller than 1000000nhash: insufficient funds: insufficient funds", "feeChargeFn 2")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	suite.Require().True(coins.IsZero(), "coins.IsZero() 2")

	suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(100000)))), "fund account again")
	coins, _, err = feeChargeFn(suite.ctx, false)
	suite.Require().NoError(err, "feeChargeFn 3")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	suite.Require().True(coins.IsAllGTE(sdk.Coins{sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000))}), "coins all gt 1000000nhash")
}

func (suite *HandlerTestSuite) TestMsgFeeHandlerFeeChargedWithRemainingBaseFee() {
	encodingConfig, err := setUpApp(suite, "atom", 100)
	testTx, acct1 := createTestTx(suite, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 120000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	if err != nil {
		panic(err)
	}
	suite.ctx = suite.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	suite.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(100000))}) // fee consumed at ante handler
	}, "panicked on adding fees")
	suite.ctx = suite.ctx.WithGasMeter(feeGasMeter)
	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  suite.app.AccountKeeper,
		BankKeeper:     suite.app.BankKeeper,
		FeegrantKeeper: suite.app.FeeGrantKeeper,
		MsgFeesKeeper:  suite.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	suite.Require().NoError(err, "NewAdditionalMsgFeeHandler")

	suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)))), "funding account")
	coins, _, err := feeChargeFn(suite.ctx, false)
	suite.Require().ErrorContains(err, "0atom is smaller than 20000atom: insufficient funds: insufficient funds", "feeChargeFn 1")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	suite.Require().True(coins.IsZero(), "coins.IsZero() 1")

	suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("atom", 20000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000))), "funding account again")
	coins, _, err = feeChargeFn(suite.ctx, false)
	suite.Require().NoError(err, "feeChargeFn 2")
	expected := sdk.Coins{sdk.NewInt64Coin("atom", 20000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)}
	suite.Require().Equal(expected, coins, "final coins")
}

func (suite *HandlerTestSuite) TestMsgFeeHandlerFeeChargedFeeGranter() {
	encodingConfig, err := setUpApp(suite, "atom", 100)
	testTxWithFeeGrant, _ := createTestTxWithFeeGrant(suite, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTxWithFeeGrant)
	if err != nil {
		panic(err)
	}
	suite.ctx = suite.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	suite.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(100000))})
	}, "panicked on adding fees")
	suite.ctx = suite.ctx.WithGasMeter(feeGasMeter)
	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  suite.app.AccountKeeper,
		BankKeeper:     suite.app.BankKeeper,
		FeegrantKeeper: suite.app.FeeGrantKeeper,
		MsgFeesKeeper:  suite.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})

	coins, _, err := feeChargeFn(suite.ctx, false)
	suite.Require().Nil(err, "Got error when should not have.")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	suite.Require().True(coins.IsAllGTE(sdk.Coins{sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000))}))
}

func (suite *HandlerTestSuite) TestMsgFeeHandlerBadDecoder() {
	encodingConfig, err := setUpApp(suite, "atom", 100)
	testTx, _ := createTestTx(suite, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	if err != nil {
		panic(err)
	}
	suite.ctx = suite.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100), false).(*antewrapper.FeeGasMeter)
	suite.ctx = suite.ctx.WithGasMeter(feeGasMeter)
	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  suite.app.AccountKeeper,
		BankKeeper:     suite.app.BankKeeper,
		FeegrantKeeper: suite.app.FeeGrantKeeper,
		MsgFeesKeeper:  suite.app.MsgFeesKeeper,
		Decoder:        sdksim.MakeTestEncodingConfig().TxConfig.TxDecoder(),
	})
	suite.Require().NoError(err)
	suite.Require().Panics(func() { feeChargeFn(suite.ctx, false) }, "Bad decoder while setting up app.")
}

func setUpApp(suite *HandlerTestSuite, additionalFeeCoinDenom string, additionalFeeCoinAmt int64) (params.EncodingConfig, error) {
	encodingConfig := suite.SetupTest(suite.T()) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	// create fee in stake
	newCoin := sdk.NewInt64Coin(additionalFeeCoinDenom, additionalFeeCoinAmt)
	err := suite.CreateMsgFee(newCoin, &testdata.TestMsg{})
	if err != nil {
		panic(err)
	}

	return encodingConfig, err
}

// returns context and app with params set on account keeper
func createTestApp(t *testing.T) (*simapp.App, sdk.Context) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}

func createTestTx(suite *HandlerTestSuite, err error, feeAmount sdk.Coins) (xauthsigning.Tx, types.AccountI) {
	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acct1)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	testTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)
	return testTx, acct1
}

func createTestTxWithFeeGrant(suite *HandlerTestSuite, err error, feeAmount sdk.Coins) (xauthsigning.Tx, types.AccountI) {
	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acct1)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	gasLimit := testdata.NewTestGasLimit()
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	// fee granter account
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct2 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr2)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acct2)

	// grant fee allowance from `addr2` to `addr1` (plenty to pay)
	err = suite.app.FeeGrantKeeper.GrantAllowance(suite.ctx, addr2, addr1, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)),
	})
	suite.txBuilder.SetFeeGranter(acct2.GetAddress())

	suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, acct2.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)))), "funding account")

	testTx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err, "CreateTestTx")
	return testTx, acct1
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (suite *HandlerTestSuite) SetupTest(t *testing.T) params.EncodingConfig {
	suite.app, suite.ctx = createTestApp(t)
	suite.ctx = suite.ctx.WithBlockHeight(1)

	// Set up TxConfig.
	encodingConfig := sdksim.MakeTestEncodingConfig()
	// We're using TestMsg encoding in some tests, so register it here.
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	suite.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig)
	return encodingConfig
}

func (suite *HandlerTestSuite) CreateMsgFee(fee sdk.Coin, msgs ...sdk.Msg) error {
	for _, msg := range msgs {
		msgFeeToCreate := msgfeetype.NewMsgFee(sdk.MsgTypeURL(msg), fee)
		err := suite.app.MsgFeesKeeper.SetMsgFee(suite.ctx, msgFeeToCreate)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *HandlerTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			suite.txBuilder, priv, suite.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return suite.txBuilder.GetTx(), nil
}

// AnteTestSuite is a test suite to be used with ante handler tests.
type HandlerTestSuite struct {
	suite.Suite

	app       *simapp.App
	ctx       sdk.Context
	clientCtx client.Context
	txBuilder client.TxBuilder
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
