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

func (s *AnteTestSuite) TestNoErrorWhenMaxGasIsUnlimited() {
	err, antehandler := setUpTxGasLimitDecorator(s, true)
	tx, _ := createTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	s.Require().NoError(err)

	// Set gas price (1 atom)
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	s.ctx = s.ctx.WithIsCheckTx(true)

	//Force the !isTestContext(ctx) to be true
	s.ctx = s.ctx.WithChainID("mainnet")

	params := &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxGas: int64(-1),
		},
	}

	s.ctx = s.ctx.WithConsensusParams(params)

	_, err = antehandler(s.ctx, tx, false)
	s.Require().NoError(err)
}

func (s *AnteTestSuite) TestErrorOutWhenMaxGasIsLimited() {
	err, antehandler := setUpTxGasLimitDecorator(s, true)
	tx, _ := createTx(s, err, sdk.NewCoins(sdk.NewInt64Coin("atom", 100000)))
	s.Require().NoError(err)

	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(1))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	s.ctx = s.ctx.WithIsCheckTx(true)
	s.ctx = s.ctx.WithChainID("mainnet")

	params := &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxGas: int64(60_000_000),
		},
	}

	s.ctx = s.ctx.WithConsensusParams(params)

	_, err = antehandler(s.ctx, tx, false)
	s.Require().ErrorContains(err, "transaction gas exceeds maximum allowed")
}

func createTx(s *AnteTestSuite, err error, feeAmount sdk.Coins) (signing.Tx, authtypes.AccountI) {
	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.app.AccountKeeper.SetAccount(s.ctx, acct1)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(5_000_000)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	return tx, acct1
}

func setUpTxGasLimitDecorator(s *AnteTestSuite, checkTx bool) (error, sdk.AnteHandler) {
	s.SetupTest(checkTx) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// setup NewTxGasLimitDecorator
	mfd := antewrapper.NewTxGasLimitDecorator()
	antehandler := sdk.ChainAnteDecorators(mfd)
	return nil, antehandler
}
