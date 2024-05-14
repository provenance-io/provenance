package antewrapper_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	simapp "github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	msgfeetype "github.com/provenance-io/provenance/x/msgfees/types"
)

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	acc  sdk.AccountI
	priv cryptotypes.PrivKey
}

// AnteTestSuite is a test s to be used with ante handler tests.
type AnteTestSuite struct {
	suite.Suite

	app         *simapp.App
	anteHandler sdk.AnteHandler
	ctx         sdk.Context
	clientCtx   client.Context
	txBuilder   client.TxBuilder

	encodingConfig simappparams.EncodingConfig
}

// returns context and app with params set on account keeper
func createTestApp(t *testing.T, isCheckTx bool) (*simapp.App, sdk.Context) {
	var app *simapp.App
	if isCheckTx {
		app = simapp.SetupQuerier(t)
	} else {
		app = simapp.Setup(t)
	}
	ctx := app.BaseApp.NewContext(isCheckTx)
	err := app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams())
	if err != nil {
		panic(err)
	}

	return app, ctx
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *AnteTestSuite) SetupTest(isCheckTx bool) {
	s.app, s.ctx = createTestApp(s.T(), isCheckTx)
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 1)
	s.ctx = s.ctx.WithBlockHeight(1)

	// Set up TxConfig.
	s.encodingConfig = s.app.GetEncodingConfig()
	// We're using TestMsg encoding in some tests, so register it here.
	s.encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(s.encodingConfig.InterfaceRegistry)

	s.clientCtx = client.Context{}.
		WithTxConfig(s.encodingConfig.TxConfig)

	anteHandler, err := antewrapper.NewAnteHandler(
		antewrapper.HandlerOptions{
			AccountKeeper:       s.app.AccountKeeper,
			BankKeeper:          s.app.BankKeeper,
			FeegrantKeeper:      s.app.FeeGrantKeeper,
			TxSigningHandlerMap: s.encodingConfig.TxConfig.SignModeHandler(),
			SigGasConsumer:      ante.DefaultSigVerificationGasConsumer,
			MsgFeesKeeper:       s.app.MsgFeesKeeper,
		},
	)

	s.Require().NoError(err)
	s.anteHandler = anteHandler
}

// CreateTestAccounts creates `numAccs` accounts, and return all relevant
// information about them including their private keys.
func (s *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
		err := acc.SetAccountNumber(uint64(i))
		s.Require().NoError(err)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		someCoins := sdk.Coins{
			sdk.NewInt64Coin("atom", 10000000),
		}
		err = s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, someCoins)
		s.Require().NoError(err)

		err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr, someCoins)
		s.Require().NoError(err)

		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (authsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
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
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(s.ctx,
			signing.SignMode_SIGN_MODE_DIRECT, signerData,
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

// TestCase represents a test case used in test tables.
type TestCase struct {
	desc     string
	malleate func()
	simulate bool
	expPass  bool
	expErr   error
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *AnteTestSuite) RunTestCase(privs []cryptotypes.PrivKey, msgs []sdk.Msg, feeAmount sdk.Coins, gasLimit uint64, accNums, accSeqs []uint64, chainID string, tc TestCase) {
	s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
		s.Require().NoError(s.txBuilder.SetMsgs(msgs...))
		s.txBuilder.SetFeeAmount(feeAmount)
		s.txBuilder.SetGasLimit(gasLimit)

		// Theoretically speaking, ante handler unit tests should only test
		// ante handlers, but here we sometimes also test the tx creation
		// process.
		ttx, txErr := s.CreateTestTx(privs, accNums, accSeqs, chainID)
		newCtx, anteErr := s.anteHandler(s.ctx, ttx, tc.simulate)

		if tc.expPass {
			s.Require().NoError(txErr)
			s.Require().NoError(anteErr)
			s.Require().NotNil(newCtx)

			s.ctx = newCtx
		} else {
			switch {
			case txErr != nil:
				s.Require().Error(txErr)
				s.Require().True(errors.Is(txErr, tc.expErr))

			case anteErr != nil:
				s.Require().Error(anteErr)
				s.Require().True(errors.Is(anteErr, tc.expErr))

			default:
				s.Fail("expected one of txErr,anteErr to be an error")
			}
		}
	})
}

func (s *AnteTestSuite) CreateMsgFee(fee sdk.Coin, msgs ...sdk.Msg) error {
	for _, msg := range msgs {
		msgFeeToCreate := msgfeetype.NewMsgFee(sdk.MsgTypeURL(msg), fee, "", msgfeetype.DefaultMsgFeeBips)
		err := s.app.MsgFeesKeeper.SetMsgFee(s.ctx, msgFeeToCreate)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewTestGasLimit is a test fee gas limit.
// they keep changing this value and our tests break, hence moving it to test.
func (s *AnteTestSuite) NewTestGasLimit() uint64 {
	return 100000
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}
