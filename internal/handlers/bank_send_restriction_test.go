package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	piosimapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/marker/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func TestBankSend(tt *testing.T) {
	txFailureCode := uint32(1)
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 1) // set denom as stake and floor gas price as 1 stake.
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	priv3, _, addr3 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv1.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000)))
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	acct2Balance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000)))
	acct3 := authtypes.NewBaseAccount(addr3, priv3.PubKey(), 2, 0)

	app := piosimapp.SetupWithGenesisAccounts(tt, "bank-restriction-testing",
		[]authtypes.GenesisAccount{acct1, acct2, acct3},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
		banktypes.Balance{Address: addr2.String(), Coins: acct2Balance},
	)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "bank-restriction-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	require.NoError(tt, app.NameKeeper.SetNameRecord(ctx, "some.kyc.provenance.io", addr1, false))
	require.NoError(tt, app.AttributeKeeper.SetAttribute(ctx, attrtypes.NewAttribute("some.kyc.provenance.io", acct3.Address, attrtypes.AttributeType_Bytes, []byte{}, nil), addr1))

	nrMarkerDenom := "nonrestrictedmarker"
	nrMarkerBaseAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(nrMarkerDenom), nil, 200, 0)
	nrMarkerAcct := markertypes.NewMarkerAccount(nrMarkerBaseAcct, sdk.NewInt64Coin(nrMarkerDenom, 10_000), addr1, []markertypes.AccessGrant{{Address: acct1.Address,
		Permissions: []markertypes.Access{markertypes.Access_Withdraw}}}, markertypes.StatusProposed, markertypes.MarkerType_Coin, true, true, false, []string{})
	require.NoError(tt, app.MarkerKeeper.SetNetAssetValue(ctx, nrMarkerAcct, markertypes.NetAssetValue{
		Value:  sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, 1),
		Volume: 1,
	}, "test"), "SetNetAssetValue failed to create nav for marker")
	require.NoError(tt, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, nrMarkerAcct),
		"AddFinalizeAndActivateMarker failed to create marker")

	restrictedMarkerDenom := "restrictedmarker"
	rMarkerBaseAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(restrictedMarkerDenom), nil, 300, 0)
	rMarkerAcct := markertypes.NewMarkerAccount(rMarkerBaseAcct, sdk.NewInt64Coin(restrictedMarkerDenom, 10_000), addr1, []markertypes.AccessGrant{{Address: acct1.Address,
		Permissions: []markertypes.Access{markertypes.Access_Withdraw, markertypes.Access_Transfer}}}, markertypes.StatusProposed, markertypes.MarkerType_RestrictedCoin, true, true, false, []string{})
	require.NoError(tt, app.MarkerKeeper.SetNetAssetValue(ctx, rMarkerAcct, markertypes.NetAssetValue{
		Value:  sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, 1),
		Volume: 1,
	}, "test"), "SetNetAssetValue failed to create nav for marker")
	require.NoError(tt, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, rMarkerAcct), "AddFinalizeAndActivateMarker failed to create marker")

	restrictedAttrMarkerDenom := "restrictedmarkerattr"
	raMarkerBaseAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(restrictedAttrMarkerDenom), nil, 400, 0)
	raMarkerAcct := markertypes.NewMarkerAccount(raMarkerBaseAcct, sdk.NewInt64Coin(restrictedAttrMarkerDenom, 10_000), addr1, []markertypes.AccessGrant{{Address: acct1.Address,
		Permissions: []markertypes.Access{markertypes.Access_Withdraw, markertypes.Access_Transfer}}}, markertypes.StatusProposed, markertypes.MarkerType_RestrictedCoin, true, true, false, []string{"some.kyc.provenance.io"})
	require.NoError(tt, app.MarkerKeeper.SetNetAssetValue(ctx, raMarkerAcct, markertypes.NetAssetValue{
		Value:  sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, 1),
		Volume: 1,
	}, "test"), "SetNetAssetValue failed to create nav for marker")
	require.NoError(tt, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, raMarkerAcct), "AddFinalizeAndActivateMarker failed to create marker")

	// Check both account balances before we begin.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000000000stake", addr1beforeBalance, "addr1beforeBalance")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "1000000000stake", addr2beforeBalance, "addr2beforeBalance")

	// send withdraw for coins
	withdrawMsg := markertypes.NewMsgWithdrawRequest(addr1, addr1, nrMarkerDenom,
		sdk.NewCoins(sdk.NewInt64Coin(nrMarkerDenom, 1000)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv1, withdrawMsg, abci.CodeTypeOK, "")
	// Check both account balances before we begin.
	addr1afterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000nonrestrictedmarker,999850000stake", addr1afterBalance, "addr1afterBalance")

	// send withdraw for coins
	withdrawMsg = markertypes.NewMsgWithdrawRequest(addr1, addr1, restrictedMarkerDenom,
		sdk.NewCoins(sdk.NewInt64Coin(restrictedMarkerDenom, 1000)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv1, withdrawMsg, abci.CodeTypeOK, "")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000nonrestrictedmarker,1000restrictedmarker,999700000stake", addr1afterBalance, "addr1afterBalance")

	// send withdraw for coins
	withdrawMsg = markertypes.NewMsgWithdrawRequest(addr1, addr1, restrictedAttrMarkerDenom,
		sdk.NewCoins(sdk.NewInt64Coin(restrictedAttrMarkerDenom, 1000)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv1, withdrawMsg, abci.CodeTypeOK, "")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000nonrestrictedmarker,1000restrictedmarker,1000restrictedmarkerattr,999550000stake", addr1afterBalance, "addr1afterBalance")

	// send restricted marker from account with transfer rights and no required attributes, expect success
	sendRMarker := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin(restrictedMarkerDenom, 100)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv1, sendRMarker, abci.CodeTypeOK, "")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000nonrestrictedmarker,900restrictedmarker,1000restrictedmarkerattr,999400000stake", addr1afterBalance, "addr1afterBalance")
	addr2afterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "100restrictedmarker,1000000000stake", addr2afterBalance, "addr2beforeBalance")

	// send non restricted marker, expect success
	sendRMarker = banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin(nrMarkerDenom, 100)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv1, sendRMarker, abci.CodeTypeOK, "")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "900nonrestrictedmarker,900restrictedmarker,1000restrictedmarkerattr,999250000stake", addr1afterBalance, "addr1afterBalance")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "100nonrestrictedmarker,100restrictedmarker,1000000000stake", addr2afterBalance, "addr2beforeBalance")

	sendNRMarker := banktypes.NewMsgSend(addr2, addr1, sdk.NewCoins(sdk.NewInt64Coin(nrMarkerDenom, 50)))
	ConstructAndSendTx(tt, *app, ctx, acct2, priv2, sendNRMarker, abci.CodeTypeOK, "")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "950nonrestrictedmarker,900restrictedmarker,1000restrictedmarkerattr,999250000stake", addr1afterBalance, "addr1afterBalance")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "50nonrestrictedmarker,100restrictedmarker,999850000stake", addr2afterBalance, "addr2beforeBalance")

	// On a restricted coin without required attributes where the sender doesn't have TRANSFER permission.
	sendRMarker = banktypes.NewMsgSend(addr2, addr1, sdk.NewCoins(sdk.NewInt64Coin(restrictedMarkerDenom, 100)))
	ConstructAndSendTx(tt, *app, ctx, acct2, priv2, sendRMarker, txFailureCode, addr2.String()+" does not have transfer permissions")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "950nonrestrictedmarker,900restrictedmarker,1000restrictedmarkerattr,999250000stake", addr1afterBalance, "addr1afterBalance")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "50nonrestrictedmarker,100restrictedmarker,999700000stake", addr2afterBalance, "addr2beforeBalance")

	// On a restricted coin with required attributes from a sender that has TRANSFER permission, but the receiver doesn't have the required attributes.
	sendRMarker = banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin(restrictedAttrMarkerDenom, 100)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv1, sendRMarker, abci.CodeTypeOK, "")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "950nonrestrictedmarker,900restrictedmarker,900restrictedmarkerattr,999100000stake", addr1afterBalance, "addr1afterBalance")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "50nonrestrictedmarker,100restrictedmarker,100restrictedmarkerattr,999700000stake", addr2afterBalance, "addr2beforeBalance")

	// On a restricted coin with required attributes from a sender that does not have TRANSFER permission, but the receiver DOES have the required attributes.
	sendRMarker = banktypes.NewMsgSend(addr2, addr3, sdk.NewCoins(sdk.NewInt64Coin(restrictedAttrMarkerDenom, 25)))
	ConstructAndSendTx(tt, *app, ctx, acct2, priv2, sendRMarker, abci.CodeTypeOK, "")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "50nonrestrictedmarker,100restrictedmarker,75restrictedmarkerattr,999550000stake", addr2afterBalance, "addr1afterBalance")
	addr3afterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
	assert.Equal(tt, "25restrictedmarkerattr", addr3afterBalance, "addr3beforeBalance")

	// MsgTransfer Tests
	// On a restricted coin with required attributes using an admin that has TRANSFER permission, but the receiver doesn't have the required attributes.
	tranferRMarker := markertypes.NewMsgTransferRequest(addr1, addr1, addr2, sdk.NewInt64Coin(restrictedMarkerDenom, 25))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv1, tranferRMarker, abci.CodeTypeOK, "")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "950nonrestrictedmarker,875restrictedmarker,900restrictedmarkerattr,998950000stake", addr2afterBalance, "addr1afterBalance")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "50nonrestrictedmarker,125restrictedmarker,75restrictedmarkerattr,999550000stake", addr2afterBalance, "addr2beforeBalance")

	// On a restricted coin with required attributes using an admin that does not have TRANSFER permission, but the receiver DOES have the required attributes.
	tranferRAMarker := markertypes.NewMsgTransferRequest(addr2, addr2, addr3, sdk.NewInt64Coin(restrictedAttrMarkerDenom, 25))
	ConstructAndSendTx(tt, *app, ctx, acct2, priv2, tranferRAMarker, txFailureCode, addr2.String()+" is not allowed to broker transfers")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "50nonrestrictedmarker,125restrictedmarker,75restrictedmarkerattr,999400000stake", addr2afterBalance, "addr1afterBalance")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr3).String()
	assert.Equal(tt, "25restrictedmarkerattr", addr3afterBalance, "addr3beforeBalance")
}

func ConstructAndSendTx(tt *testing.T, app piosimapp.App, ctx sdk.Context, acct *authtypes.BaseAccount, priv cryptotypes.PrivKey, msg sdk.Msg, expectedCode uint32, expectedError string) {
	encCfg := sdksim.MakeTestEncodingConfig()
	fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit())))
	acct = app.AccountKeeper.GetAccount(ctx, acct.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct, ctx.ChainID(), msg)
	require.NoError(tt, err, "SignTxAndGetBytes")
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(tt, expectedCode, res.Code, "res=%+v", res)
	if len(expectedError) > 0 {
		require.Contains(tt, res.Log, expectedError, "DeliverTx result.Log")
	}
}
