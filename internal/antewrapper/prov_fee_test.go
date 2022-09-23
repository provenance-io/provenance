package antewrapper_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	pioante "github.com/provenance-io/provenance/internal/antewrapper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// These tests are kicked off by TestAnteTestSuite in testutil_test.go

func (s *AnteTestSuite) TestEnsureMempoolFees() {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	s.SetupTest(true) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(s.app.AccountKeeper, s.app.BankKeeper, s.app.FeeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	testaccs := s.CreateTestAccounts(1)
	priv1 := testaccs[0].priv
	addr1 := testaccs[0].acc.GetAddress()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails
	stakePrice := sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	highGasPrice := []sdk.DecCoin{stakePrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().ErrorContains(err, "insufficient fees", "Decorator should have errored on too low fee for local gasPrice")

	// Set IsCheckTx to false
	s.ctx = s.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "MempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	s.ctx = s.ctx.WithIsCheckTx(true)

	stakePrice = sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{stakePrice}
	s.ctx = s.ctx.WithMinGasPrices(lowGasPrice)

	_, err = antehandler(s.ctx, tx, false)
	s.Require().Nil(err, "Decorator should not have errored on fee higher than local gasPrice")
}

func (s *AnteTestSuite) TestDeductFees() {
	s.SetupTest(false) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200000))
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)
	// Set account with insufficient funds
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)))
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, coins)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)), s.app.BankKeeper.GetAllBalances(s.ctx, addr1), "should have the new balance after funding account")

	decorators := []sdk.AnteDecorator{pioante.NewFeeMeterContextDecorator(), pioante.NewProvenanceDeductFeeDecorator(s.app.AccountKeeper, s.app.BankKeeper, nil, s.app.MsgFeesKeeper)}
	antehandler := sdk.ChainAnteDecorators(decorators...)

	_, err = antehandler(s.ctx, tx, false)

	s.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200_000))))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200_010)), s.app.BankKeeper.GetAllBalances(s.ctx, addr1), "Balance before tx")
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NoError(err, "Tx errored after account has been set with sufficient funds")
	s.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)), s.app.BankKeeper.GetAllBalances(s.ctx, addr1), "Balance after tx")
}

func (s *AnteTestSuite) TestEnsureAdditionalFeesPaid() {
	// given
	s.SetupTest(true)
	newCoin := sdk.NewInt64Coin("steak", 100)
	s.CreateMsgFee(newCoin, &testdata.TestMsg{})

	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// when
	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	s.ctx.ChainID()
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// then
	// Set the account with insufficient funds (base fee coin insufficient)
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)))
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, coins)
	s.Require().NoError(err)

	decorators := []sdk.AnteDecorator{pioante.NewFeeMeterContextDecorator(), pioante.NewProvenanceDeductFeeDecorator(s.app.AccountKeeper, s.app.BankKeeper, nil, s.app.MsgFeesKeeper)}

	antehandler := sdk.ChainAnteDecorators(decorators...)

	_, err = antehandler(s.ctx, tx, false)

	s.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds for base fees and but not additional fees
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200_000))))
	s.Require().NoError(err)

	_, err = antehandler(s.ctx, tx, false)

	s.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// valid case
	// set gas fee and msg fees (steak)
	// Set account with sufficient funds
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, sdk.NewCoins(sdk.NewCoin("steak", sdk.NewInt(100))))
	s.Require().NoError(err)

	s.txBuilder.SetFeeAmount(NewTestFeeAmountMultiple())
	s.txBuilder.SetGasLimit(gasLimit)

	tx, err = s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	_, err = antehandler(s.ctx, tx, false)

	s.Require().Nil(err, "Tx did not error when fee payer had insufficient funds")
}

// NewTestFeeAmount is a test fee amount with multiple coins.
func NewTestFeeAmountMultiple() sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150), sdk.NewInt64Coin("steak", 100))
}
