package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simappCosmos "github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdkgas "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/antewrapper"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
	msgfeetype "github.com/provenance-io/provenance/x/msgfees/types"
)

func (s *HandlerTestSuite) TestMsgFeeHandlerFeeChargedNoRemainingBaseFee() {
	encodingConfig, err := setUpApp(s, false, "atom", 100)
	testTx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	if err != nil {
		panic(err)
	}
	s.ctx = s.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	s.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(100000))})
	}, "panicked on adding fees")
	s.ctx = s.ctx.WithGasMeter(feeGasMeter)
	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	s.Require().NoError(err)
	coins, _, err := feeChargeFn(s.ctx, false)

	s.Require().Contains(err.Error(), "0nhash is smaller than 1000000nhash: insufficient funds: insufficient funds", "got wrong message")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsZero())

	check(simapp.FundAccount(s.app, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(900000)))))
	coins, _, err = feeChargeFn(s.ctx, false)
	s.Require().Contains(err.Error(), "900000nhash is smaller than 1000000nhash: insufficient funds: insufficient funds", "got wrong message")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsZero())

	check(simapp.FundAccount(s.app, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(100000)))))
	coins, _, err = feeChargeFn(s.ctx, false)
	s.Require().Nil(err, "Got error when should not have.")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsAllGTE(sdk.Coins{sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000))}))
}

func (s *HandlerTestSuite) TestMsgFeeHandlerFeeChargedWithRemainingBaseFee() {
	encodingConfig, err := setUpApp(s, false, "atom", 100)
	testTx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 120000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	if err != nil {
		panic(err)
	}
	s.ctx = s.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	s.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(100000))}) // fee consumed at ante handler
	}, "panicked on adding fees")
	s.ctx = s.ctx.WithGasMeter(feeGasMeter)
	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	s.Require().NoError(err)

	check(simapp.FundAccount(s.app, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)))))
	coins, _, err := feeChargeFn(s.ctx, false)
	s.Require().Contains(err.Error(), "0atom is smaller than 20000atom: insufficient funds: insufficient funds", "should have enough to pay msg fees, but not enough to sweep remaining base fees of 20000atom")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsZero())

	check(simapp.FundAccount(s.app, s.ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("atom", 20000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000))))
	coins, _, err = feeChargeFn(s.ctx, false)
	s.Require().Nil(err, "Got error when should have successfully paid all msg fees and swept remaining base fees")
	s.Require().True(coins.IsEqual(sdk.Coins{sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000), sdk.NewInt64Coin("atom", 20000)}))
}

func (s *HandlerTestSuite) TestMsgFeeHandlerFeeChargedFeeGranter() {
	encodingConfig, err := setUpApp(s, false, "atom", 100)
	testTxWithFeeGrant, _ := createTestTxWithFeeGrant(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTxWithFeeGrant)
	if err != nil {
		panic(err)
	}
	s.ctx = s.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	s.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewCoin("atom", sdk.NewInt(100000))})
	}, "panicked on adding fees")
	s.ctx = s.ctx.WithGasMeter(feeGasMeter)
	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})

	coins, _, err := feeChargeFn(s.ctx, false)
	s.Require().Nil(err, "Got error when should not have.")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsAllGTE(sdk.Coins{sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000))}))
}

func (s *HandlerTestSuite) TestMsgFeeHandlerBadDecoder() {
	encodingConfig, err := setUpApp(s, false, "atom", 100)
	testTx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	if err != nil {
		panic(err)
	}
	s.ctx = s.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100), false).(*antewrapper.FeeGasMeter)
	s.ctx = s.ctx.WithGasMeter(feeGasMeter)
	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        simappCosmos.MakeTestEncodingConfig().TxConfig.TxDecoder(),
	})
	s.Require().NoError(err)
	s.Require().Panics(func() { feeChargeFn(s.ctx, false) }, "Bad decoder while setting up app.")
}

func setUpApp(s *HandlerTestSuite, checkTx bool, additionalFeeCoinDenom string, additionalFeeCoinAmt int64) (params.EncodingConfig, error) {
	encodingConfig := s.SetupTest(checkTx) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
	// create fee in stake
	newCoin := sdk.NewInt64Coin(additionalFeeCoinDenom, additionalFeeCoinAmt)
	err := s.CreateMsgFee(newCoin, &testdata.TestMsg{})
	if err != nil {
		panic(err)
	}

	return encodingConfig, err
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.App, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}

func createTestTx(s *HandlerTestSuite, err error, feeAmount sdk.Coins) (xauthsigning.Tx, types.AccountI) {
	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acct1)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	testTx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)
	return testTx, acct1
}

func createTestTxWithFeeGrant(s *HandlerTestSuite, err error, feeAmount sdk.Coins) (xauthsigning.Tx, types.AccountI) {
	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acct1)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	// fee granter account
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct2 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr2)
	s.app.AccountKeeper.SetAccount(s.ctx, acct2)

	// grant fee allowance from `addr2` to `addr1` (plenty to pay)
	err = s.app.FeeGrantKeeper.GrantAllowance(s.ctx, addr2, addr1, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin(msgfeetype.NhashDenom, 1000000)),
	})
	s.txBuilder.SetFeeGranter(acct2.GetAddress())

	check(simapp.FundAccount(s.app, s.ctx, acct2.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeetype.NhashDenom, sdk.NewInt(1000000)))))

	testTx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)
	return testTx, acct1
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *HandlerTestSuite) SetupTest(isCheckTx bool) params.EncodingConfig {
	s.app, s.ctx = createTestApp(isCheckTx)
	s.ctx = s.ctx.WithBlockHeight(1)

	// Set up TxConfig.
	encodingConfig := simappCosmos.MakeTestEncodingConfig()
	// We're using TestMsg encoding in some tests, so register it here.
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	s.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig)
	return encodingConfig
}

func (s *HandlerTestSuite) CreateMsgFee(fee sdk.Coin, msgs ...sdk.Msg) error {
	for _, msg := range msgs {
		msgFeeToCreate := msgfeetype.NewMsgFee(sdk.MsgTypeURL(msg), fee, "", msgfeetype.DefaultMsgFeeBips)
		err := s.app.MsgFeesKeeper.SetMsgFee(s.ctx, msgFeeToCreate)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *HandlerTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  s.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := s.txBuilder.SetSignatures(sigsV2...)
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
			s.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			s.txBuilder, priv, s.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return s.txBuilder.GetTx(), nil
}

// AnteTestSuite is a test s to be used with ante handler tests.
type HandlerTestSuite struct {
	suite.Suite

	app       *simapp.App
	ctx       sdk.Context
	clientCtx client.Context
	txBuilder client.TxBuilder
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
