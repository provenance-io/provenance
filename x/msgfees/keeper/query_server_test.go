package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/msgfees/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"
)

type QueryServerTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	cfg         network.Config
	queryClient types.QueryClient

	privkey1  cryptotypes.PrivKey
	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress
	acct1     authtypes.AccountI

	privkey2  cryptotypes.PrivKey
	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
	acct2     authtypes.AccountI
}

func (s *QueryServerTestSuite) SetupTest() {
	s.app = simapp.Setup(true)
	s.ctx = s.app.BaseApp.NewContext(true, tmproto.Header{})
	s.app.AccountKeeper.SetParams(s.ctx, authtypes.DefaultParams())
	s.app.BankKeeper.SetParams(s.ctx, banktypes.DefaultParams())
	s.cfg = testutil.DefaultTestNetworkConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MsgBasedFeeKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.privkey2 = secp256k1.GenPrivKey()
	s.pubkey2 = s.privkey2.PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.acct1 = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx,s.acct1)
	s.acct2 = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user2Addr)
	s.app.AccountKeeper.SetAccount(s.ctx,s.acct2)

	simapp.FundAccount(s.app, s.ctx, s.acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))))

}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (s *QueryServerTestSuite) TestCalculateTxFees() {
	queryClient := s.queryClient
	simulate1 := types.CalculateTxFeesRequest{
		TxBytes: nil,
	}
	response, err := queryClient.CalculateTxFees(s.ctx.Context(), &simulate1)
	s.Assert().Error(err)
	s.Assert().Nil(response)
	bankSend := banktypes.NewMsgSend(s.user1Addr, s.user1Addr, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))))
	theTx := s.cfg.TxConfig.NewTxBuilder()
	theTx.SetMsgs(bankSend)

	account1Sig := signing.SignatureV2{
		PubKey: s.pubkey1,
		Data: &signing.SingleSignatureData{
			SignMode: s.cfg.TxConfig.SignModeHandler().DefaultMode(),
		},
		Sequence: s.acct1.GetSequence(),
	}
	signerData := authsign.SignerData{
		ChainID:       s.cfg.ChainID,
		AccountNumber: s.acct1.GetAccountNumber(),
		Sequence:      s.acct1.GetSequence(),
	}
	signBytes, err := s.cfg.TxConfig.SignModeHandler().GetSignBytes(s.cfg.TxConfig.SignModeHandler().DefaultMode(), signerData, theTx.GetTx())
	if err != nil {
		panic(err)
	}
	sig, err := s.privkey1.Sign(signBytes)
	if err != nil {
		panic(err)
	}
	account1Sig.Data.(*signing.SingleSignatureData).Signature = sig
	err = theTx.SetSignatures(account1Sig)
	if err != nil {
		panic(err)
	}

	//theTx.SetMemo("memo")
	theTx.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1))))
	theTx.SetGasLimit(uint64(1000000000))

	txBytes, err := s.cfg.TxConfig.TxEncoder()(theTx.(sdk.Tx))
	s.Require().NoError(err)
	simulate1 = types.CalculateTxFeesRequest{
		TxBytes: txBytes,
	}
	println(s.app.AccountKeeper.GetParams(s.ctx).MaxMemoCharacters)
	response, err = queryClient.CalculateTxFees(s.ctx.Context(), &simulate1)
	s.Assert().NoError(err)
	s.Assert().Nil(response)
}
