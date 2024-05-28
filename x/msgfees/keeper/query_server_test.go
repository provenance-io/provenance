package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
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
	acct1     sdk.AccountI

	privkey2  cryptotypes.PrivKey
	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
	acct2     sdk.AccountI

	minGasPrice       sdk.Coin
	usdConversionRate uint64
}

func (s *QueryServerTestSuite) SetupTest() {
	s.app = simapp.SetupQuerier(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
	s.app.AccountKeeper.Params.Set(s.ctx, authtypes.DefaultParams())
	s.app.BankKeeper.SetParams(s.ctx, banktypes.DefaultParams())
	s.cfg = testutil.DefaultTestNetworkConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MsgFeesKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.minGasPrice = sdk.Coin{
		Denom:  s.cfg.BondDenom,
		Amount: sdkmath.NewInt(10),
	}
	s.usdConversionRate = 7
	s.app.MsgFeesKeeper.SetParams(s.ctx, types.NewParams(s.minGasPrice, s.usdConversionRate, pioconfig.GetProvenanceConfig().FeeDenom))

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.privkey2 = secp256k1.GenPrivKey()
	s.pubkey2 = s.privkey2.PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.acct1 = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, s.acct1)
	s.acct2 = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user2Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, s.acct2)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, markertypes.NewEmptyMarkerAccount("navcoin", s.acct1.GetAddress().String(), []markertypes.AccessGrant{})))
	s.Require().NoError(banktestutil.FundAccount(s.ctx, s.app.BankKeeper, s.acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 100_000))))
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (s *QueryServerTestSuite) TestCalculateTxFees() {
	bankSend := banktypes.NewMsgSend(s.user1Addr, s.user2Addr, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 100)))
	simulateReq := s.createTxFeesRequest(s.pubkey1, s.privkey1, s.acct1, bankSend)

	// do send without additional fee
	response, err := s.queryClient.CalculateTxFees(s.ctx.Context(), &simulateReq)
	s.Assert().NoError(err)
	s.Assert().NotNil(response)
	s.Assert().True(response.AdditionalFees.Empty())
	expectedTotalFees := sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, s.minGasPrice.Amount.MulRaw(int64(response.EstimatedGas))))
	s.Assert().Equal(expectedTotalFees.String(), response.TotalFees.String())

	// do send with an additional fee
	sendAddFee := sdk.NewInt64Coin(s.cfg.BondDenom, 1)
	s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, types.NewMsgFee("/cosmos.bank.v1beta1.MsgSend", sendAddFee, "", types.DefaultMsgFeeBips)))
	response, err = s.queryClient.CalculateTxFees(s.ctx.Context(), &simulateReq)
	s.Assert().NoError(err)
	s.Assert().NotNil(response)
	expectedTotalFees = response.AdditionalFees.Add(sdk.NewCoin(s.cfg.BondDenom, s.minGasPrice.Amount.MulRaw(int64(response.EstimatedGas))))
	s.Assert().Equal(expectedTotalFees, response.TotalFees)
	s.Assert().Equal(sdk.NewCoins(sendAddFee), response.AdditionalFees)

	// do multiple sends in tx with fee
	bankSend1 := banktypes.NewMsgSend(s.user1Addr, s.user2Addr, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 2)))
	bankSend2 := banktypes.NewMsgSend(s.user1Addr, s.user2Addr, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 3)))
	simulateReq = s.createTxFeesRequest(s.pubkey1, s.privkey1, s.acct1, bankSend1, bankSend2)

	response, err = s.queryClient.CalculateTxFees(s.ctx.Context(), &simulateReq)
	s.Assert().NoError(err)
	s.Assert().NotNil(response)
	expectedTotalFees = response.AdditionalFees.Add(sdk.NewCoin(s.cfg.BondDenom, s.minGasPrice.Amount.MulRaw(int64(response.EstimatedGas))))
	s.Assert().Equal(expectedTotalFees, response.TotalFees)
	s.Assert().Equal(sdk.NewCoins(sdk.NewCoin(sendAddFee.Denom, sendAddFee.Amount.MulRaw(2))), response.AdditionalFees)
}

func (s *QueryServerTestSuite) TestCalculateTxFeesAuthz() {
	server := markerkeeper.NewMsgServerImpl(s.app.MarkerKeeper)

	hotdogDenom := "hotdog"
	_, err := server.AddMarker(s.ctx, markertypes.NewMsgAddMarkerRequest(hotdogDenom, sdkmath.NewInt(100), s.user1Addr, s.user1Addr, markertypes.MarkerType_RestrictedCoin, true, true, false, []string{}, 0, 0))
	s.Require().NoError(err)
	access := markertypes.AccessGrant{
		Address:     s.user1,
		Permissions: markertypes.AccessListByNames("DELETE,MINT,WITHDRAW,TRANSFER,ADMIN,BURN"),
	}
	_, err = server.AddAccess(s.ctx, markertypes.NewMsgAddAccessRequest(hotdogDenom, s.user1Addr, access))
	s.Require().NoError(err)
	_, err = server.Finalize(s.ctx, markertypes.NewMsgFinalizeRequest(hotdogDenom, s.user1Addr))
	s.Require().NoError(err)
	_, err = server.Activate(s.ctx, markertypes.NewMsgActivateRequest(hotdogDenom, s.user1Addr))
	s.Require().NoError(err)
	_, err = server.Mint(s.ctx, markertypes.NewMsgMintRequest(s.user1Addr, sdk.NewInt64Coin(hotdogDenom, 1000)))
	s.Require().NoError(err)
	_, err = server.Withdraw(s.ctx, markertypes.NewMsgWithdrawRequest(s.user1Addr, s.user1Addr, hotdogDenom, sdk.NewCoins(sdk.NewInt64Coin(hotdogDenom, 10))))
	s.Require().NoError(err)
	msgGrant := &authz.MsgGrant{
		Granter: s.user1,
		Grantee: s.user2,
		Grant:   authz.Grant{},
	}
	err = msgGrant.SetAuthorization(markertypes.NewMarkerTransferAuthorization(sdk.NewCoins(sdk.NewInt64Coin(hotdogDenom, 10)), []sdk.AccAddress{s.user1Addr}))
	s.Require().NoError(err)
	_, err = s.app.AuthzKeeper.Grant(s.ctx, msgGrant)
	s.Require().NoError(err)

	transferRequest := markertypes.NewMsgTransferRequest(s.user1Addr, s.user1Addr, s.user2Addr, sdk.NewInt64Coin(hotdogDenom, 9))
	simulateReq := s.createTxFeesRequest(s.pubkey2, s.privkey2, s.acct2, transferRequest)
	response, err := s.queryClient.CalculateTxFees(s.ctx.Context(), &simulateReq)
	s.Assert().NoError(err)
	s.Assert().NotNil(response)
	s.Assert().True(response.AdditionalFees.Empty())
}

func (s *QueryServerTestSuite) TestCalculateTxFeesWithAssessCustomFees() {
	additionalAccessedFeesCoin := sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 100)
	assessCustomFeeMsg := types.NewMsgAssessCustomMsgFeeRequest("name", additionalAccessedFeesCoin, s.user2, s.user1, "")
	simulateReq := s.createTxFeesRequest(s.pubkey1, s.privkey1, s.acct1, &assessCustomFeeMsg)

	// do assessCustomFee
	response, err := s.queryClient.CalculateTxFees(s.ctx.Context(), &simulateReq)
	s.Assert().NoError(err)
	s.Assert().NotNil(response)
	s.Assert().Equal(sdk.NewCoins(additionalAccessedFeesCoin), response.AdditionalFees)
	expectedGasFees := sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, s.minGasPrice.Amount.MulRaw(int64(response.EstimatedGas))))
	s.Assert().Equal(fmt.Sprintf("%s,%s", additionalAccessedFeesCoin.String(), expectedGasFees.String()), response.TotalFees.String())

	// do assessCustomFee where custom fee has a message fee associated with it
	additionalAccessedFeesCoin = sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 100)
	s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, types.NewMsgFee(sdk.MsgTypeURL(&assessCustomFeeMsg), additionalAccessedFeesCoin, "", types.DefaultMsgFeeBips)))
	response, err = s.queryClient.CalculateTxFees(s.ctx.Context(), &simulateReq)
	s.Assert().NoError(err)
	s.Assert().NotNil(response)
	additionalAccessedFeesCoin = sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 200)
	s.Assert().Equal(sdk.NewCoins(additionalAccessedFeesCoin), response.AdditionalFees)
	expectedGasFees = sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, s.minGasPrice.Amount.MulRaw(int64(response.EstimatedGas))))
	s.Assert().Equal(fmt.Sprintf("%s,%s", additionalAccessedFeesCoin.String(), expectedGasFees.String()), response.TotalFees.String())
}

func (s *QueryServerTestSuite) createTxFeesRequest(pubKey cryptotypes.PubKey, privKey cryptotypes.PrivKey, acct sdk.AccountI, msgs ...sdk.Msg) types.CalculateTxFeesRequest {
	theTx := s.cfg.TxConfig.NewTxBuilder()
	s.Require().NoError(theTx.SetMsgs(msgs...))
	s.signTx(theTx, pubKey, privKey, acct)
	txBytes, err := s.cfg.TxConfig.TxEncoder()(theTx.(sdk.Tx))
	s.Require().NoError(err)
	return types.CalculateTxFeesRequest{
		TxBytes:          txBytes,
		DefaultBaseDenom: s.cfg.BondDenom,
	}
}

func (s *QueryServerTestSuite) signTx(txb client.TxBuilder, pubKey cryptotypes.PubKey, privKey cryptotypes.PrivKey, acct sdk.AccountI) {
	signerData := signing.SignerData{
		ChainID:       s.cfg.ChainID,
		AccountNumber: acct.GetAccountNumber(),
		Sequence:      acct.GetSequence(),
	}
	theTx := txb.GetTx()
	adaptableTx, ok := theTx.(authsigning.V2AdaptableTx)
	s.Require().True(ok, "%T does not implement the authsigning.V2AdaptableTx interface", theTx)

	txData := adaptableTx.GetSigningTxData()
	signBytes, err := s.cfg.TxConfig.SignModeHandler().GetSignBytes(s.ctx, s.cfg.TxConfig.SignModeHandler().DefaultMode(), signerData, txData)
	s.Require().NoError(err, "GetSignBytes")

	sig, err := privKey.Sign(signBytes)
	s.Require().NoError(err, "privKey.Sign(signBytes)")

	accountSig := sdksigning.SignatureV2{
		PubKey: pubKey,
		Data: &sdksigning.SingleSignatureData{
			SignMode:  sdksigning.SignMode(s.cfg.TxConfig.SignModeHandler().DefaultMode()),
			Signature: sig,
		},
		Sequence: acct.GetSequence(),
	}
	err = txb.SetSignatures(accountSig)
	s.Require().NoError(err, "SetSignatures")
}
