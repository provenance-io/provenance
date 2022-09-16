package antewrapper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/provenance-io/provenance/internal/antewrapper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

func TestIsGovernanceMessage(t *testing.T) {
	tests := []struct {
		msg sdk.Msg
		exp bool
	}{
		{&govtypesv1beta1.MsgSubmitProposal{}, true},
		{&govtypesv1.MsgSubmitProposal{}, true},
		{&govtypesv1.MsgExecLegacyContent{}, true},
		{&govtypesv1beta1.MsgVote{}, true},
		{&govtypesv1.MsgVote{}, true},
		{&govtypesv1beta1.MsgDeposit{}, true},
		{&govtypesv1.MsgDeposit{}, true},
		{&govtypesv1beta1.MsgVoteWeighted{}, true},
		{&govtypesv1.MsgVoteWeighted{}, true},
		{nil, false},
		{&authztypes.MsgGrant{}, false},
		{&nametypes.MsgBindNameRequest{}, false},
	}

	for _, tc := range tests {
		name := "nil"
		if tc.msg != nil {
			name = sdk.MsgTypeURL(tc.msg)[1:]
		}
		t.Run(name, func(tt *testing.T) {
			act := antewrapper.IsGovMessage(tc.msg)
			assert.Equal(tt, tc.exp, act, "isGovMessage")
		})
	}
}

func TestIsOnlyGovMsgs(t *testing.T) {
	tests := []struct {
		name string
		msgs []sdk.Msg
		exp  bool
	}{
		{
			name: "empty",
			msgs: []sdk.Msg{},
			exp:  false,
		},
		{
			name: "only MsgSubmitProposal",
			msgs: []sdk.Msg{&govtypesv1beta1.MsgSubmitProposal{}},
			exp:  true,
		},
		{
			name: "all the gov messages",
			msgs: []sdk.Msg{
				&govtypesv1beta1.MsgSubmitProposal{},
				&govtypesv1.MsgSubmitProposal{},
				&govtypesv1.MsgExecLegacyContent{},
				&govtypesv1beta1.MsgVote{},
				&govtypesv1.MsgVote{},
				&govtypesv1beta1.MsgDeposit{},
				&govtypesv1.MsgDeposit{},
				&govtypesv1beta1.MsgVoteWeighted{},
				&govtypesv1.MsgVoteWeighted{},
			},
			exp: true,
		},
		{
			name: "MsgSubmitProposal and MsgStoreCode",
			msgs: []sdk.Msg{
				&govtypesv1beta1.MsgSubmitProposal{},
				&wasmtypes.MsgStoreCode{},
			},
			exp: false,
		},
		{
			name: "MsgStoreCode and MsgSubmitProposal",
			msgs: []sdk.Msg{
				&wasmtypes.MsgStoreCode{},
				&govtypesv1beta1.MsgSubmitProposal{},
			},
			exp: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			act := antewrapper.IsOnlyGovMsgs(tc.msgs)
			assert.Equal(tt, tc.exp, act, "isOnlyGovMsgs")
		})
	}
}

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

func createTx(suite *AnteTestSuite, err error, feeAmount sdk.Coins) (signing.Tx, authtypes.AccountI) {
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
