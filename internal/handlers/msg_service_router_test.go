package handlers_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	piosimapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/handlers"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestRegisterMsgService(t *testing.T) {
	db := dbm.NewMemDB()

	// Create an encoding config that doesn't register testdata Msg services.
	encCfg := simapp.MakeTestEncodingConfig()
	app := baseapp.NewBaseApp("test", log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, encCfg.TxConfig.TxDecoder())
	router := handlers.NewPioMsgServiceRouter(encCfg.TxConfig.TxDecoder())
	router.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	app.SetMsgServiceRouter(router)
	require.Panics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})

	// Register testdata Msg services, and rerun `RegisterService`.
	testdata.RegisterInterfaces(encCfg.InterfaceRegistry)
	require.NotPanics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})
}

func TestRegisterMsgServiceTwice(t *testing.T) {
	// Setup baseapp.
	db := dbm.NewMemDB()
	encCfg := simapp.MakeTestEncodingConfig()
	app := baseapp.NewBaseApp("test", log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, encCfg.TxConfig.TxDecoder())
	router := handlers.NewPioMsgServiceRouter(encCfg.TxConfig.TxDecoder())
	router.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	app.SetMsgServiceRouter(router)
	testdata.RegisterInterfaces(encCfg.InterfaceRegistry)

	// First time registering service shouldn't panic.
	require.NotPanics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})

	// Second time should panic.
	require.Panics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})
}

func TestMsgService(t *testing.T) {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(1000)), sdk.NewCoin("atom", sdk.NewInt(1000)))
	app := piosimapp.SetupWithGenesisAccounts([]authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	// tx without a fee associated with msg type
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), testdata.NewTestFeeAmount(), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 13, len(res.Events))
	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "150atom", string(res.Events[4].Attributes[0].Value))

	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewCoin("hotdog", sdk.NewInt(800)))
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)

	// tx with a fee associated with msg type and account has funds
	msg = banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(50))))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin("hotdog", 800))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 16, len(res.Events))
	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "150atom,800hotdog", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[5].Type)
	assert.Equal(t, "acc_seq", string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[6].Type)
	assert.Equal(t, "signature", string(res.Events[6].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[13].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[13].Attributes[0].Key))
	assert.Equal(t, "800hotdog", string(res.Events[13].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "150atom", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[15].Type)
	assert.Equal(t, "msg_fees", string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"800hotdog\",\"recipient\":\"\"}]", string(res.Events[15].Attributes[0].Value))

	msgbasedFee = msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin("atom", 10))
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)

	// tx with a fee associated with msg type, additional cost is in same base as fee
	msg = banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(50))))
	fees = sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 16, len(res.Events))
	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "150atom", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[5].Type)
	assert.Equal(t, "acc_seq", string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[6].Type)
	assert.Equal(t, "signature", string(res.Events[6].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[13].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[13].Attributes[0].Key))
	assert.Equal(t, "10atom", string(res.Events[13].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "140atom", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[15].Type)
	assert.Equal(t, "msg_fees", string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"10atom\",\"recipient\":\"\"}]", string(res.Events[15].Attributes[0].Value))

	msgbasedFee = msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 10))
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)

	check(simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	// tx with a fee associated with msg type, additional cost is in same base as fee
	msg = banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(50))))
	fees = sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500010))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 16, len(res.Events))
	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "190500010nhash", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[5].Type)
	assert.Equal(t, "acc_seq", string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[6].Type)
	assert.Equal(t, "signature", string(res.Events[6].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[13].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[13].Attributes[0].Key))
	assert.Equal(t, "10nhash", string(res.Events[13].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "190500000nhash", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[15].Type)
	assert.Equal(t, "msg_fees", string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"10nhash\",\"recipient\":\"\"}]", string(res.Events[15].Attributes[0].Value))

	msgbasedFee = msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin("atom", 100))
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)

	check(simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	// tx with a fee associated with msg type, additional cost is in diff denom(atom) but using default denom, nhash for base fee.
	msg = banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(50))))
	fees = sdk.NewCoins(sdk.NewInt64Coin(msgfeestypes.NhashDenom, 190500010), sdk.NewInt64Coin("atom", 100))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 16, len(res.Events))
	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "100atom,190500010nhash", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[5].Type)
	assert.Equal(t, "acc_seq", string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[6].Type)
	assert.Equal(t, "signature", string(res.Events[6].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[13].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[13].Attributes[0].Key))
	assert.Equal(t, "100atom", string(res.Events[13].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "190500010nhash", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[15].Type)
	assert.Equal(t, "msg_fees", string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"100atom\",\"recipient\":\"\"}]", string(res.Events[15].Attributes[0].Value))
}

func TestMsgServiceAuthz(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	initBalance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10000)), sdk.NewCoin("atom", sdk.NewInt(1000)))
	app := piosimapp.SetupWithGenesisAccounts([]authtypes.GenesisAccount{acct1, acct2}, banktypes.Balance{Address: addr.String(), Coins: initBalance}, banktypes.Balance{Address: addr2.String(), Coins: initBalance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))
	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewCoin("hotdog", sdk.NewInt(800)))
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)
	app.AuthzKeeper.SaveGrant(ctx, addr2, addr, banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("hotdog", 500))), time.Now().Add(time.Hour))

	// tx authz send message with correct amount of fees associated
	msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin("hotdog", 800))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 16, len(res.Events))
	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "150atom,800hotdog", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[5].Type)
	assert.Equal(t, "acc_seq", string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[6].Type)
	assert.Equal(t, "signature", string(res.Events[6].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[13].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[13].Attributes[0].Key))
	assert.Equal(t, "800hotdog", string(res.Events[13].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "150atom", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[15].Type)
	assert.Equal(t, "msg_fees", string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"800hotdog\",\"recipient\":\"\"}]", string(res.Events[15].Attributes[0].Value))

	// send 2 successful authz messages
	msgExec = authztypes.NewMsgExec(addr2, []sdk.Msg{msg, msg})
	fees = sdk.NewCoins(sdk.NewInt64Coin("atom", 300), sdk.NewInt64Coin("hotdog", 1600))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit()*2, fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 21, len(res.Events))
	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "300atom,1600hotdog", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[5].Type)
	assert.Equal(t, "acc_seq", string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[6].Type)
	assert.Equal(t, "signature", string(res.Events[6].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[18].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[18].Attributes[0].Key))
	assert.Equal(t, "1600hotdog", string(res.Events[18].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[19].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[19].Attributes[0].Key))
	assert.Equal(t, "300atom", string(res.Events[19].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[20].Type)
	assert.Equal(t, "msg_fees", string(res.Events[20].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"2\",\"total\":\"1600hotdog\",\"recipient\":\"\"}]", string(res.Events[20].Attributes[0].Value))

	// tx authz single send message without enough fees associated
	fees = sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin("hotdog", 1))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, uint32(0xd), res.Code, "res=%+v", res)

	// tx authz with 2 send msgs that will exhaust the fees on the second msg
	msgExec = authztypes.NewMsgExec(addr2, []sdk.Msg{msg, msg})
	fees = sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin("hotdog", 1000))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	// using the higher gas limit coming from testdata pkg, since actually tx costs more than 100000
	txBytes, err = SignTxAndGetBytes(testdata.NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, uint32(0xd), res.Code, "res=%+v", res)

	// tx contains 1 regular send and one send in authz, should fail since authz's send will exhaust supplied fees
	msgsend := banktypes.NewMsgSend(addr2, addr, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))
	msgExec = authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
	fees = sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin("hotdog", 1000))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	// using the higher gas limit coming from testdata pkg, since actually tx costs more than 100000
	txBytes, err = SignTxAndGetBytes(testdata.NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), msgsend, &msgExec)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, uint32(0xd), res.Code, "res=%+v", res)
}

func TestMsgServiceAuthzAdditionalMsgFeeInDefaultDenom(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	initBalance := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(1000)))
	app := piosimapp.SetupWithGenesisAccounts([]authtypes.GenesisAccount{acct1, acct2}, banktypes.Balance{Address: addr.String(), Coins: initBalance}, banktypes.Balance{Address: addr2.String(), Coins: initBalance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(100))))
	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewCoin("atom", sdk.NewInt(10)))
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)
	app.AuthzKeeper.SaveGrant(ctx, addr2, addr, banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("atom", 500))), time.Now().Add(time.Hour))

	// tx authz send message with correct amount of fees associated
	msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 16, len(res.Events))
	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "150atom", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[5].Type)
	assert.Equal(t, "acc_seq", string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[6].Type)
	assert.Equal(t, "signature", string(res.Events[6].Attributes[0].Key))
	assert.Equal(t, "tx", res.Events[13].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[13].Attributes[0].Key))
	assert.Equal(t, "10atom", string(res.Events[13].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "140atom", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[15].Type)
	assert.Equal(t, "msg_fees", string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"10atom\",\"recipient\":\"\"}]", string(res.Events[15].Attributes[0].Value))
}

func TestMsgServiceAssessMsgFee(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin("atom", 1000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000))
	app := piosimapp.SetupWithGenesisAccounts([]authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	msg := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("test", sdk.NewInt64Coin(msgfeestypes.UsdDenom, 7), addr2.String(), addr.String())
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &msg)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 12, len(res.Events))

	assert.Equal(t, "tx", res.Events[4].Type)
	assert.Equal(t, "fee", string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "150atom,1190500000nhash", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, msgfeestypes.EventTypeAssessCustomMsgFee, res.Events[8].Type)
	assert.Equal(t, msgfeestypes.KeyAttributeName, string(res.Events[8].Attributes[0].Key))
	assert.Equal(t, "test", string(res.Events[8].Attributes[0].Value))
	assert.Equal(t, msgfeestypes.KeyAttributeAmount, string(res.Events[8].Attributes[1].Key))
	assert.Equal(t, "7usd", string(res.Events[8].Attributes[1].Value))
	assert.Equal(t, msgfeestypes.KeyAttributeRecipient, string(res.Events[8].Attributes[2].Key))
	assert.Equal(t, addr2.String(), string(res.Events[8].Attributes[2].Value))
	assert.Equal(t, "tx", res.Events[9].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[9].Attributes[0].Key))
	assert.Equal(t, "175000000nhash", string(res.Events[9].Attributes[0].Value))
	assert.Equal(t, "tx", res.Events[10].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[10].Attributes[0].Key))
	assert.Equal(t, "150atom,1015500000nhash", string(res.Events[10].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[11].Type)
	assert.Equal(t, "msg_fees", string(res.Events[11].Attributes[0].Key))
	assert.Equal(t, fmt.Sprintf("[{\"msg_type\":\"/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest\",\"count\":\"1\",\"total\":\"87500000nhash\",\"recipient\":\"\"},{\"msg_type\":\"/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest\",\"count\":\"1\",\"total\":\"87500000nhash\",\"recipient\":\"%s\"}]", addr2.String()), string(res.Events[11].Attributes[0].Value))

}

func SignTxAndGetBytes(gaslimit uint64, fees sdk.Coins, encCfg simappparams.EncodingConfig, pubKey types.PubKey, privKey types.PrivKey, acct authtypes.BaseAccount, chainId string, msg ...sdk.Msg) ([]byte, error) {
	txBuilder := encCfg.TxConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(fees)
	txBuilder.SetGasLimit(gaslimit)
	err := txBuilder.SetMsgs(msg...)
	if err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  encCfg.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: acct.Sequence,
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	signerData := authsigning.SignerData{
		ChainID:       chainId,
		AccountNumber: acct.AccountNumber,
		Sequence:      acct.Sequence,
	}
	sigV2, err = tx.SignWithPrivKey(
		encCfg.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, privKey, encCfg.TxConfig, acct.Sequence)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}
	// Send the tx to the app
	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

// NewTestGasLimit is a test fee gas limit.
// they keep changing this value and our tests break, hence moving it to test.
func NewTestGasLimit() uint64 {
	return 100000
}
