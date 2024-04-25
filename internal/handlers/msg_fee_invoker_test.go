package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	simapp "github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/antewrapper"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	msgfeetype "github.com/provenance-io/provenance/x/msgfees/types"
)

const (
	NHash = "nhash"
)

// HandlerTestSuite is a test suite for these handler tests.
type HandlerTestSuite struct {
	suite.Suite

	app       *simapp.App
	ctx       sdk.Context
	clientCtx client.Context
	txBuilder client.TxBuilder
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s *HandlerTestSuite) TestMsgFeeHandlerFeeChargedNoRemainingBaseFee() {
	encodingConfig, err := setUpApp(s, "atom", 100)
	testTx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin(NHash, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	s.Require().NoError(err, "txEncoder")

	s.ctx = s.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.NewTestLogger(s.T()), storetypes.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	s.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin(NHash, 1_000_000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewInt64Coin("atom", 100_000)})
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

	s.Require().ErrorContains(err, "spendable balance 0nhash is smaller than 1000000nhash: insufficient funds", "feeChargeFn 1")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsZero(), "coins.IsZero() 1")

	s.Require().NoError(banktestutil.FundAccount(s.ctx, s.app.BankKeeper, acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(NHash, 900_000))), "fund account")
	coins, _, err = feeChargeFn(s.ctx, false)
	s.Require().ErrorContains(err, "900000nhash is smaller than 1000000nhash: insufficient funds: insufficient funds", "feeChargeFn 2")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsZero(), "coins.IsZero() 2")

	s.Require().NoError(banktestutil.FundAccount(s.ctx, s.app.BankKeeper, acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(NHash, 100_000))), "fund account again")
	coins, _, err = feeChargeFn(s.ctx, false)
	s.Require().NoError(err, "feeChargeFn 3")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsAllGTE(sdk.Coins{sdk.NewInt64Coin(NHash, 1_000_000)}), "coins all gt 1000000nhash")
}

func (s *HandlerTestSuite) TestMsgFeeHandlerFeeChargedWithRemainingBaseFee() {
	encodingConfig, err := setUpApp(s, "atom", 100)
	testTx, acct1 := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 120000), sdk.NewInt64Coin(NHash, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	if err != nil {
		panic(err)
	}
	s.ctx = s.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.NewTestLogger(s.T()), storetypes.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	s.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin(NHash, 1_000_000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewInt64Coin("atom", 100_000)}) // fee consumed at ante handler
	}, "panicked on adding fees")
	s.ctx = s.ctx.WithGasMeter(feeGasMeter)

	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	s.Require().NoError(err, "NewAdditionalMsgFeeHandler")

	s.Require().NoError(banktestutil.FundAccount(s.ctx, s.app.BankKeeper, acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(NHash, 1_000_000))), "funding account")
	coins, _, err := feeChargeFn(s.ctx, false)
	s.Require().ErrorContains(err, "spendable balance 0atom is smaller than 20000atom: insufficient funds", "feeChargeFn 1")
	// fee gas meter has nothing to charge, so nothing should have been charged.
	s.Require().True(coins.IsZero(), "coins.IsZero() 1")

	s.Require().NoError(banktestutil.FundAccount(s.ctx, s.app.BankKeeper, acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("atom", 20000), sdk.NewInt64Coin(NHash, 1000000))), "funding account again")
	coins, _, err = feeChargeFn(s.ctx, false)
	s.Require().Nil(err, "Got error when should have successfully paid all msg fees and swept remaining base fees")
	s.Require().True(coins.Equal(sdk.Coins{sdk.NewInt64Coin(NHash, 1000000), sdk.NewInt64Coin("atom", 20000)}))
	s.Require().NoError(err, "feeChargeFn 2")
	expected := sdk.Coins{sdk.NewInt64Coin("atom", 20000), sdk.NewInt64Coin(NHash, 1000000)}
	s.Require().Equal(expected, coins, "final coins")

}

func (s *HandlerTestSuite) TestMsgFeeHandlerFeeChargedFeeGranter() {
	encodingConfig, err := setUpApp(s, "atom", 100)
	testTxWithFeeGrant, _ := createTestTxWithFeeGrant(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin(NHash, 1000000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTxWithFeeGrant)
	if err != nil {
		panic(err)
	}
	s.ctx = s.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.NewTestLogger(s.T()), storetypes.NewGasMeter(100000), false).(*antewrapper.FeeGasMeter)
	s.Require().NotPanics(func() {
		msgType := sdk.MsgTypeURL(&testdata.TestMsg{})
		feeGasMeter.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin(NHash, 1_000_000)), msgType, "")
		feeGasMeter.ConsumeBaseFee(sdk.Coins{sdk.NewInt64Coin("atom", 100_000)})
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
	s.Require().True(coins.IsAllGTE(sdk.Coins{sdk.NewInt64Coin(NHash, 1000000)}))

}

func (s *HandlerTestSuite) TestMsgFeeHandlerBadDecoder() {
	encodingConfig, err := setUpApp(s, "atom", 100)
	testTx, _ := createTestTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))

	// See comment for Check().
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(testTx)
	if err != nil {
		panic(err)
	}
	s.ctx = s.ctx.WithTxBytes(bz)
	feeGasMeter := antewrapper.NewFeeGasMeterWrapper(log.NewTestLogger(s.T()), storetypes.NewGasMeter(100), false).(*antewrapper.FeeGasMeter)
	s.ctx = s.ctx.WithGasMeter(feeGasMeter)

	emptyTxCfg := tx.NewTxConfig(codec.NewProtoCodec(codectestutil.CodecOptions{}.NewInterfaceRegistry()), tx.DefaultSignModes)

	feeChargeFn, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        emptyTxCfg.TxDecoder(),
	})
	s.Require().NoError(err)
	s.Require().Panics(func() { feeChargeFn(s.ctx, false) }, "Bad decoder while setting up app.")
}

func setUpApp(s *HandlerTestSuite, additionalFeeCoinDenom string, additionalFeeCoinAmt int64) (simappparams.EncodingConfig, error) {
	pioconfig.SetProvenanceConfig("", 0)
	encodingConfig := s.SetupTest(s.T()) // setup
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
func createTestApp(t *testing.T) (*simapp.App, sdk.Context) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams())

	return app, ctx
}

func createTestTx(s *HandlerTestSuite, err error, feeAmount sdk.Coins) (xauthsigning.Tx, sdk.AccountI) {
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

func createTestTxWithFeeGrant(s *HandlerTestSuite, err error, feeAmount sdk.Coins) (xauthsigning.Tx, sdk.AccountI) {
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
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin(NHash, 1000000)),
	})
	s.txBuilder.SetFeeGranter(acct2.GetAddress())

	s.Require().NoError(banktestutil.FundAccount(s.ctx, s.app.BankKeeper, acct2.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(NHash, 1_000_000))), "funding account")

	testTx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err, "CreateTestTx")
	return testTx, acct1
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *HandlerTestSuite) SetupTest(t *testing.T) simappparams.EncodingConfig {
	s.app, s.ctx = createTestApp(t)
	s.ctx = s.ctx.WithBlockHeight(1)

	// Set up TxConfig.
	encodingConfig := s.app.GetEncodingConfig()
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
				// SignMode:  s.clientCtx.TxConfig.SignModeHandler().DefaultMode(), // TODO[1760]: signing: same type name diff packages.
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
	// TODO[1760]: signing: SignWithPrivKey: Uncomment these lines.
	/*
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
	*/
	err = s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return s.txBuilder.GetTx(), nil
}
