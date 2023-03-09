package handlers_test

import (
	"fmt"
	"strings"
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
	"github.com/provenance-io/provenance/x/marker/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func TestBankSend(tt *testing.T) {
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 1) // set denom as stake and floor gas price as 1 stake.
	// encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000)))
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	acct2Balance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000)))

	app := piosimapp.SetupWithGenesisAccounts(tt, "bank-restriction-testing",
		[]authtypes.GenesisAccount{acct1, acct2},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
		banktypes.Balance{Address: addr2.String(), Coins: acct2Balance},
	)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "bank-restriction-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	nrMarkerDenom := "nonrestrictedmarker"
	nrMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(nrMarkerDenom), nil, 200, 0)
	require.NoError(tt, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, markertypes.NewMarkerAccount(nrMarkerAcct, sdk.NewInt64Coin(nrMarkerDenom, 10_000), addr1, []markertypes.AccessGrant{{Address: acct1.Address,
		Permissions: []markertypes.Access{markertypes.Access_Withdraw}}}, markertypes.StatusProposed, markertypes.MarkerType_Coin, true, []string{})), "")

	rMarkerDenom := "restrictedmarker"
	rMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerDenom), nil, 300, 0)
	require.NoError(tt, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, markertypes.NewMarkerAccount(rMarkerAcct, sdk.NewInt64Coin(rMarkerDenom, 10_000), addr1, []markertypes.AccessGrant{{Address: acct1.Address,
		Permissions: []markertypes.Access{markertypes.Access_Withdraw, markertypes.Access_Transfer}}}, markertypes.StatusProposed, markertypes.MarkerType_RestrictedCoin, true, []string{})), "")

	raMarkerDenom := "restrictedmarkerattr"
	raMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(raMarkerDenom), nil, 400, 0)
	require.NoError(tt, app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, markertypes.NewMarkerAccount(raMarkerAcct, sdk.NewInt64Coin(raMarkerDenom, 10_000), addr1, []markertypes.AccessGrant{{Address: acct1.Address,
		Permissions: []markertypes.Access{markertypes.Access_Withdraw, markertypes.Access_Transfer}}}, markertypes.StatusProposed, markertypes.MarkerType_RestrictedCoin, true, []string{"some.kyc.provenance.io"})), "")

	// Check both account balances before we begin.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000000000stake", addr1beforeBalance, "addr1beforeBalance")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "1000000000stake", addr2beforeBalance, "addr2beforeBalance")

	// send withdraw for coins
	withdrawMsg := markertypes.NewMsgWithdrawRequest(addr1, addr1, nrMarkerDenom,
		sdk.NewCoins(sdk.NewInt64Coin(nrMarkerDenom, 1000)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv, withdrawMsg, abci.CodeTypeOK, "")
	// Check both account balances before we begin.
	addr1afterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000nonrestrictedmarker,999900000stake", addr1afterBalance, "addr1afterBalance")

	// send withdraw for coins
	withdrawMsg = markertypes.NewMsgWithdrawRequest(addr1, addr1, rMarkerDenom,
		sdk.NewCoins(sdk.NewInt64Coin(rMarkerDenom, 1000)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv, withdrawMsg, abci.CodeTypeOK, "")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000nonrestrictedmarker,1000restrictedmarker,999800000stake", addr1afterBalance, "addr1afterBalance")

	// send restricted marker from account with transfer rights and no required attributes, expect success
	sendRMarker := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin(rMarkerDenom, 100)))
	ConstructAndSendTx(tt, *app, ctx, acct1, priv, sendRMarker, abci.CodeTypeOK, "")
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000nonrestrictedmarker,900restrictedmarker,999700000stake", addr1afterBalance, "addr1afterBalance")
	addr2afterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "100restrictedmarker,1000000000stake", addr2afterBalance, "addr2beforeBalance")

	// send restricted marker from account without transfer rights and no required attributes, expect failure
	expectedError := fmt.Sprintf("%s does not have transfer permissions", addr2.String())
	sendRMarker = banktypes.NewMsgSend(addr2, addr1, sdk.NewCoins(sdk.NewInt64Coin(rMarkerDenom, 100)))
	ConstructAndSendTx(tt, *app, ctx, acct2, priv2, sendRMarker, 1, expectedError)
	addr1afterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000nonrestrictedmarker,900restrictedmarker,999700000stake", addr1afterBalance, "addr1afterBalance")
	addr2afterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "100restrictedmarker,999900000stake", addr2afterBalance, "addr2beforeBalance")

	stopIfFailed(tt)
}

func ConstructAndSendTx(tt *testing.T, app piosimapp.App, ctx sdk.Context, acct *authtypes.BaseAccount, priv cryptotypes.PrivKey, msg sdk.Msg, expectedCode uint32, expectedError string) {
	encCfg := sdksim.MakeTestEncodingConfig()
	fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000))
	acct = app.AccountKeeper.GetAccount(ctx, acct.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct, ctx.ChainID(), msg)
	require.NoError(tt, err, "SignTxAndGetBytes")
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(tt, expectedCode, res.Code, "res=%+v", res)
	if len(expectedError) > 0 {
		require.True(tt, strings.Contains(res.Log, expectedError), fmt.Sprintf("error msg does not contain %s", expectedError))
	}
}
