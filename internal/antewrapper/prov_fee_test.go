package antewrapper_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	pioante "github.com/provenance-io/provenance/internal/antewrapper"
)

// These tests are kicked off by TestAnteTestSuite in testutil_test.go

func (s *AnteTestSuite) TestProvenanceDeductFeeDecoratorChecksFunds() {
	s.SetupTest(false) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200000))
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg), "SetMsgs")
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
	s.Require().NoError(err, "funding account with %s", coins)
	s.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)), s.app.BankKeeper.GetAllBalances(s.ctx, addr1), "should have the new balance after funding account")

	decorators := []sdk.AnteDecorator{pioante.NewFeeMeterContextDecorator(), pioante.NewProvenanceDeductFeeDecorator(s.app.AccountKeeper, s.app.BankKeeper, nil, s.app.MsgFeesKeeper)}
	antehandler := sdk.ChainAnteDecorators(decorators...)

	_, err = antehandler(s.ctx, tx, false)
	// Example error: "10stake is smaller than 200000stake: insufficient funds: insufficient funds"
	s.Require().Error(err, "antehandler insufficient funds")
	s.Assert().ErrorContains(err, addr1.String())
	s.Assert().ErrorContains(err, "10stake is smaller than 200000stake")
	s.Assert().ErrorContains(err, "insufficient funds")
	if s.T().Failed() {
		s.T().FailNow()
	}

	// Set account with sufficient funds
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	plusCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200_000)))
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, plusCoins)
	s.Require().NoError(err, "funding account with %s", plusCoins)
	s.Require().Equal(coins.Add(plusCoins...), s.app.BankKeeper.GetAllBalances(s.ctx, addr1), "Balance before tx")
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NoError(err, "antehandler sufficient funds")
	s.Require().Equal(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)), s.app.BankKeeper.GetAllBalances(s.ctx, addr1), "Balance after tx")
}

func (s *AnteTestSuite) TestProvenanceDeductFeeDecoratorAdditionalFees() {
	// given
	s.SetupTest(true)
	newCoin := sdk.NewInt64Coin("steak", 100)
	s.Require().NoError(s.CreateMsgFee(newCoin, &testdata.TestMsg{}), "creating 100steak message fee")

	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// when
	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg), "SetMsgs")
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err, "CreateTestTx")

	// then
	// Set the account with insufficient funds (base fee coin insufficient)
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)))
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, coins)
	s.Require().NoError(err, "funding account with 10stake")

	decorators := []sdk.AnteDecorator{pioante.NewFeeMeterContextDecorator(), pioante.NewProvenanceDeductFeeDecorator(s.app.AccountKeeper, s.app.BankKeeper, nil, s.app.MsgFeesKeeper)}
	antehandler := sdk.ChainAnteDecorators(decorators...)

	s.Run("insufficient funds for both base and additional fees", func() {
		_, err = antehandler(s.ctx, tx, false)
		// Example error: account cosmos1flu4xj7c66tdmvdjjas3a62a6jynf93ezrgysj does not have enough balance to pay for "100steak", balance: "": insufficient funds
		s.Require().Error(err, "antehandler")
		s.Assert().ErrorContains(err, addr1.String())
		s.Assert().ErrorContains(err, "does not have enough balance to pay")
		s.Assert().ErrorContains(err, `"100steak"`)
		s.Assert().ErrorContains(err, `balance: ""`)
		s.Assert().ErrorContains(err, `insufficient funds`)
	})

	// Set account with sufficient funds for base fees and but not additional fees
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200_000))))
	s.Require().NoError(err, "funding account with 200000stake")

	s.Run("insufficient funds for just additional fees", func() {
		_, err = antehandler(s.ctx, tx, false)
		s.Require().Error(err, "antehandler")
		s.Assert().ErrorContains(err, addr1.String())
		s.Assert().ErrorContains(err, "does not have enough balance to pay")
		s.Assert().ErrorContains(err, `"100steak"`)
		s.Assert().ErrorContains(err, `balance: ""`)
		s.Assert().ErrorContains(err, `insufficient funds`)
	})

	// valid case
	// set gas fee and msg fees (steak)
	// Set account with sufficient funds
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
	err = testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, sdk.NewCoins(sdk.NewCoin("steak", sdk.NewInt(200))))
	s.Require().NoError(err, "funding account with 200steak")

	s.Run("sufficient funds", func() {
		s.txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150), sdk.NewInt64Coin("steak", 100)))
		tx, err = s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
		s.Require().NoError(err, "CreateTestTx")

		_, err = antehandler(s.ctx, tx, false)
		s.Require().NoError(err, "antehandler")
	})
}
