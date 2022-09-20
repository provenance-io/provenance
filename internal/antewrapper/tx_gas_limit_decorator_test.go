package antewrapper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
)

// These tests are kicked off by TestAnteTestSuite in testutil_test.go

func (suite *AnteTestSuite) TestNoErrorWhenMaxGasIsUnlimited() {
	err, antehandler := setUpTxGasLimitDecorator(suite, true)
	tx, _ := createTx(suite, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	suite.Require().NoError(err)

	// Set gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	suite.ctx = suite.ctx.WithMinGasPrices(highGasPrice)

	suite.ctx = suite.ctx.WithIsCheckTx(true)

	//Force the !isTestContext(ctx) to be true
	suite.ctx = suite.ctx.WithChainID("mainnet")

	params := &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxGas: int64(-1),
		},
	}

	suite.ctx = suite.ctx.WithConsensusParams(params)

	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NoError(err)
}

func (suite *AnteTestSuite) TestErrorOutWhenMaxGasIsLimited() {
	err, antehandler := setUpTxGasLimitDecorator(suite, true)
	tx, _ := createTx(suite, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	suite.Require().NoError(err)

	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	suite.ctx = suite.ctx.WithMinGasPrices(highGasPrice)

	suite.ctx = suite.ctx.WithIsCheckTx(true)
	suite.ctx = suite.ctx.WithChainID("mainnet")

	params := &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxGas: int64(60_000_000),
		},
	}

	suite.ctx = suite.ctx.WithConsensusParams(params)

	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().ErrorContains(err, "transaction gas exceeds maximum allowed")
}

func createTx(suite *AnteTestSuite, err error, feeAmount sdk.Coins) (signing.Tx, types.AccountI) {
	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acct1)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(5_000_000)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	return tx, acct1
}

func setUpTxGasLimitDecorator(suite *AnteTestSuite, checkTx bool) (error, sdk.AnteHandler) {
	suite.SetupTest(checkTx) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// setup NewTxGasLimitDecorator
	mfd := antewrapper.NewTxGasLimitDecorator()
	antehandler := sdk.ChainAnteDecorators(mfd)
	return nil, antehandler
}
