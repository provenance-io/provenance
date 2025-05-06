package antewrapper_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	simapp "github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
)

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

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	acc  sdk.AccountI
	priv cryptotypes.PrivKey
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
			FlatFeesKeeper:      s.app.FlatFeesKeeper,
			CircuitKeeper:       &s.app.CircuitKeeper,
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
	var sigsV2 []sdksigning.SignatureV2
	for i, priv := range privs {
		sigV2 := sdksigning.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &sdksigning.SingleSignatureData{
				SignMode:  sdksigning.SignMode_SIGN_MODE_DIRECT,
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
	sigsV2 = []sdksigning.SignatureV2{}
	for i, priv := range privs {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(s.ctx,
			sdksigning.SignMode_SIGN_MODE_DIRECT, signerData,
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
		msgFeeToCreate := flatfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), fee)
		err := s.app.FlatFeesKeeper.SetMsgFee(s.ctx, *msgFeeToCreate)
		if err != nil {
			return fmt.Errorf("could not create msg fee %q for %q: %w", fee.String(), sdk.MsgTypeURL(msg), err)
		}
	}
	return nil
}

// NewTestGasLimit is a test fee gas limit.
// they keep changing this value and our tests break, hence moving it to test.
func (s *AnteTestSuite) NewTestGasLimit() uint64 {
	return 100000
}

// ctxWithFlatFeeGasMeter will return a ctx (based off s.ctx) with a FlatFeeGasMeter in it.
func (s *AnteTestSuite) ctxWithFlatFeeGasMeter() sdk.Context {
	if _, err := antewrapper.GetFlatFeeGasMeter(s.ctx); err == nil {
		return s.ctx
	}
	return s.ctx.WithGasMeter(antewrapper.NewFlatFeeGasMeter(s.ctx.GasMeter(), s.ctx.Logger(), s.app.FlatFeesKeeper))
}

func genTxWithFeeGranter(ctx context.Context,
	gen client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string,
	accNums, accSeqs []uint64, feeGranter sdk.AccAddress, priv ...cryptotypes.PrivKey,
) (sdk.Tx, error) {
	sigs := make([]sdksigning.SignatureV2, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memo := simtypes.RandStringOfLength(r, r.Intn(101))

	signMode := gen.SignModeHandler().DefaultMode()

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = sdksigning.SignatureV2{
			PubKey: p.PubKey(),
			Data: &sdksigning.SingleSignatureData{
				SignMode: sdksigning.SignMode(signMode),
			},
			Sequence: accSeqs[i],
		}
	}

	txb := gen.NewTxBuilder()
	err := txb.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = txb.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	txb.SetMemo(memo)
	txb.SetFeeAmount(feeAmt)
	txb.SetGasLimit(gas)
	txb.SetFeeGranter(feeGranter)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := signing.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}

		theTx := txb.GetTx()
		adaptableTx, ok := theTx.(authsigning.V2AdaptableTx)
		if !ok {
			return nil, fmt.Errorf("%T does not implement the authsigning.V2AdaptableTx interface", theTx)
		}
		txData := adaptableTx.GetSigningTxData()
		signBytes, err := gen.SignModeHandler().GetSignBytes(ctx, signMode, signerData, txData)
		if err != nil {
			return nil, err
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			return nil, err
		}
		sigs[i].Data.(*sdksigning.SingleSignatureData).Signature = sig
		err = txb.SetSignatures(sigs...)
		if err != nil {
			return nil, err
		}
	}

	return txb.GetTx(), nil
}
