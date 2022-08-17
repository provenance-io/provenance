package handlers_test

import (
	"fmt"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/protobuf/proto"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"

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
	require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee))

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
	require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee))

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
	require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee))

	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
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
	require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee))

	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
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
	require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee))
	require.NoError(t, app.AuthzKeeper.SaveGrant(ctx, addr2, addr, banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("hotdog", 500))), time.Now().Add(time.Hour)))

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
	require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee))
	require.NoError(t, app.AuthzKeeper.SaveGrant(ctx, addr2, addr, banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("atom", 500))), time.Now().Add(time.Hour)))

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

func TestRewardsProgramStartError(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	//_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin("atom", 1000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000))
	app := piosimapp.SetupWithGenesisAccounts([]authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	blockTime := ctx.BlockTime()
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	rewardProgram := *rewardtypes.NewMsgCreateRewardProgramRequest(
		"title",
		"description",
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		blockTime,
		0,
		3,
		3,
		uint64(blockTime.Day()),
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.ZeroInt()),
					},
				},
			},
		},
	)

	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &rewardProgram)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, true, res.IsErr(), "should return error", res)
}

func TestRewardsProgramStart(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin("atom", 1000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000))
	app := piosimapp.SetupWithGenesisAccounts([]authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	rewardProgram := *rewardtypes.NewMsgCreateRewardProgramRequest(
		"title",
		"description",
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time.Now().Add(time.Duration(1)*time.Second),
		9,
		3,
		3,
		10,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.ZeroInt()),
					},
				},
			},
		},
	)
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2, Time: time.Now().UTC()}})
	txReward, err := SignTx(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &rewardProgram)
	require.NoError(t, err)
	_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), txReward)
	require.NoError(t, errFromDeliverTx)
	assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	assert.Equal(t, 13, len(res.Events))
	lookForRewardCreatedEvent, found := contains(res.Events, "reward_program_created")
	assert.Equal(t, true, found, "event reward_program_created should exist")
	assert.Equal(t, 1, len(*lookForRewardCreatedEvent), "event reward_program_created should exist")
	lookForMessageEvent, foundMessage := contains(res.Events, "message")
	assert.Equal(t, true, foundMessage, "event message should exist")
	_, foundAttribute := containsAttribute(*lookForMessageEvent, "create_reward_program")
	assert.Equal(t, true, foundAttribute, "event attribute create_reward_program should exist")
}

func TestRewardsProgramStartPerformQualifyingActions(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000))
	app := piosimapp.SetupWithGenesisAccounts([]authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	rewardProgram := *rewardtypes.NewMsgCreateRewardProgramRequest(
		"title",
		"description",
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time.Now().Add(time.Duration(100)*time.Millisecond),
		9,
		3,
		3,
		10,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.ZeroInt()),
					},
				},
			},
		},
	)
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2, Time: time.Now().UTC()}})
	txReward, err := SignTx(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &rewardProgram)
	require.NoError(t, err)
	_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), txReward)
	require.NoError(t, errFromDeliverTx)
	assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	time.Sleep(110 * time.Millisecond)
	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(50))))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	for height := int64(3); height <= int64(100); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) == 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	assert.Equal(t, 98, int(accountState.ActionCounter["ActionTransfer"]), "account state incorrect")
	assert.Equal(t, 10, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(100)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 1, len(byAddress.RewardAccountState), "only one reward account for one claim period.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(100)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 1, len(byAddress.RewardAccountState), "only one reward account for one claim period.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, len(byAddress.RewardAccountState), "none of them should be claimable.")
}

func TestRewardsProgramStartPerformQualifyingActionsRecordedRewardsUnclaimable(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(50*time.Millisecond),
		uint64(30),
		10,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.ZeroInt()),
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(50))))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	// get past the reward start time ( test that reward program starts up after 50ms)
	time.Sleep(55 * time.Millisecond)

	for height := int64(2); height < int64(22); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) == 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 20, int(accountState.ActionCounter["ActionTransfer"]), "account state incorrect")
	assert.Equal(t, 10, int(accountState.SharesEarned), "account state incorrect")
	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
	})

	assert.Equal(t, 0, len(byAddress.RewardAccountState), "none of the rewards should be in claimable state since claim period has not ended yet.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMED,
	})

	assert.Equal(t, 0, len(byAddress.RewardAccountState), "none of the rewards should be in claimed state.")
}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountFunded(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(1),
		100,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.ZeroInt()),
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	// fund th =e deterministic rewards account, since genesis won't do that work
	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance}, banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(50))))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(150 * time.Millisecond)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(2); height < int64(7); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) > 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(1), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 1, int(accountState.ActionCounter["ActionTransfer"]), "account state incorrect")
	assert.Equal(t, 1, int(accountState.SharesEarned), "account state incorrect")
	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 5, len(byAddress.RewardAccountState), "claimable and un claimable sum rewards should be 5 for this address.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 4, len(byAddress.RewardAccountState), "claimable rewards should be 4 for this address.")

	// get the accoutn balances of acct1
	balance := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	println(balance.AmountOf("nhash").String())
	// claim rewards for the address
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 7, Time: time.Now().UTC()}})
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq))
	txClaim, errClaim := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msgClaim)
	require.NoError(t, errClaim)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txClaim})
	require.Equal(t, true, res.IsOK(), "res=%+v", res)
	// unmarshal the TxMsgData
	var protoResult sdk.TxMsgData
	err3 := proto.Unmarshal(res.Data, &protoResult)
	require.NoError(t, err3)
	var claimResponse rewardtypes.MsgClaimAllRewardsResponse
	err4 := proto.Unmarshal(protoResult.Data[0].Data, &claimResponse)
	require.NoError(t, err4)
	require.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(50_000_000_000)), claimResponse.TotalRewardClaim)
	require.Equal(t, 1, len(claimResponse.ClaimDetails))
	require.Equal(t, uint64(1), claimResponse.ClaimDetails[0].RewardProgramId)
	require.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(50_000_000_000)), claimResponse.ClaimDetails[0].TotalRewardClaim)
	require.Equal(t, 5, len(claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails))
	require.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)), claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails[0].ClaimPeriodReward)
	app.EndBlock(abci.RequestEndBlock{Height: 7})
	app.Commit()
	balanceLater := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// make sure account balance has the tokens
	require.Equal(t, sdk.NewInt(50_000_000_000), balanceLater.AmountOf(pioconfig.DefaultBondDenom).Sub(balance.AmountOf(pioconfig.DefaultBondDenom)))

}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountNotFunded(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(1),
		100,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.ZeroInt()),
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	// do not fund the deterministic rewards account
	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(50))))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(150 * time.Millisecond)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(2); height < int64(7); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) > 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(1), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 1, int(accountState.ActionCounter["ActionTransfer"]), "account state incorrect")
	assert.Equal(t, 1, int(accountState.SharesEarned), "account state incorrect")
	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 5, len(byAddress.RewardAccountState), "claimable and un claimable sum rewards should be 5 for this address.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 4, len(byAddress.RewardAccountState), "claimable rewards should be 4 for this address.")

	// claim rewards for the address
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 7, Time: time.Now().UTC()}})
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq))
	txClaim, errClaim := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msgClaim)
	require.NoError(t, errClaim)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txClaim})
	require.Equal(t, true, res.IsErr(), "res=%+v", res)
	app.EndBlock(abci.RequestEndBlock{Height: 7})
	app.Commit()
}

// Checks to see if delegation are met for a Qualifying action in this case Transfer
// this tests has transfers from an account which DOES NOT have the minimum delegation
// amount needed to get a share
func TestRewardsProgramStartPerformQualifyingActionsCriteriaNotMet(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(1),
		100,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions: 0,
						MaximumActions: 10,
						MinimumDelegationAmount: sdk.Coin{
							Denom:  "nhash",
							Amount: sdk.NewInt(10_000_000_000)},
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(50))))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(110 * time.Millisecond)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(2); height < int64(7); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		time.Sleep(1100 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) > 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(0), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 0, int(accountState.ActionCounter["ActionTransfer"]), "account state incorrect")
	assert.Equal(t, 0, int(accountState.SharesEarned), "account state incorrect")
}

// Checks to see if delegation are met for a Qualifying action in this case, Transfer, create an address with delegations
// transfers which map to QualifyingAction map to the delegated address
// delegation threshold is met
func TestRewardsProgramStartPerformQualifyingActionsTransferAndDelegationsPresent(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(1),
		100,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions: 0,
						MaximumActions: 10,
						MinimumDelegationAmount: sdk.Coin{
							Denom:  "nhash",
							Amount: sdk.NewInt(100)},
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING
	err, bondedVal1, bondedVal2 := createTestValidators(pubKey, pubKey2, addr, addr2)

	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, []stakingtypes.Validator{bondedVal1, bondedVal2}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(50))))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	// wait for program to start
	time.Sleep(150 * time.Millisecond)

	//go through 5 blocks, but take a time to cut blocks > claim period time interval.
	for height := int64(2); height < int64(7); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		time.Sleep(1100 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) > 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(1), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 1, int(accountState.ActionCounter["ActionTransfer"]), "account state incorrect")
	assert.Equal(t, 1, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 5, len(byAddress.RewardAccountState), "claimable and un claimable sum rewards should be 5 for this address.")

}

// Checks to see if delegation are met for a Qualifying action in this case Transfer, create an address with delegations
// transfers which map to QualifyingAction map to the delegated address
// delegation threshold is *not* met
func TestRewardsProgramStartPerformQualifyingActionsThreshHoldNotMet(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(1),
		100,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Transfer{
					Transfer: &rewardtypes.ActionTransfer{
						MinimumActions: 0,
						MaximumActions: 10,
						MinimumDelegationAmount: sdk.Coin{
							Denom:  "nhash",
							Amount: sdk.NewInt(100000000)},
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING
	err, bondedVal1, bondedVal2 := createTestValidators(pubKey, pubKey2, addr, addr2)

	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, []stakingtypes.Validator{bondedVal1, bondedVal2}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(50))))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(1 * time.Second)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(2); height < int64(7); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		time.Sleep(1100 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) > 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(0), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 0, int(accountState.ActionCounter["ActionTransfer"]), "account state incorrect")
	assert.Equal(t, 0, int(accountState.SharesEarned), "account state incorrect")

}

func TestRewardsProgramStartPerformQualifyingActions_Vote(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	//_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(30),
		10,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Vote{
					Vote: &rewardtypes.ActionVote{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.ZeroInt()),
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypes.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypes.ProposalTypeText),
		coinsPos,
		addr,
	)

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(200 * time.Millisecond)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2, Time: time.Now().UTC()}})
	txGov, err := SignTx(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.Equal(t, true, len(proposal) > 0, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	vote1 := govtypes.NewMsgVote(addr, proposal[0].ProposalId, govtypes.OptionYes)

	assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")

	for height := int64(3); height < int64(23); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) == 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 20, int(accountState.ActionCounter["ActionVote"]), "account state incorrect")
	assert.Equal(t, 10, int(accountState.SharesEarned), "account state incorrect")
}

func TestRewardsProgramStartPerformQualifyingActions_Vote_InvalidDelegations(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	//_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(30),
		10,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Vote{
					Vote: &rewardtypes.ActionVote{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.NewInt(1000)),
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypes.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypes.ProposalTypeText),
		coinsPos,
		addr,
	)

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(200 * time.Millisecond)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2, Time: time.Now().UTC()}})
	txGov, err := SignTx(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.Equal(t, true, len(proposal) > 0, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	vote1 := govtypes.NewMsgVote(addr, proposal[0].ProposalId, govtypes.OptionYes)

	assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")

	for height := int64(3); height < int64(5); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) == 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(0), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 0, int(accountState.ActionCounter["ActionVote"]), "account state incorrect")
	assert.Equal(t, 0, int(accountState.SharesEarned), "account state incorrect")
	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, true, len(byAddress.RewardAccountState) == 0, "RewardDistributionsByAddress incorrect")

}

func TestRewardsProgramStartPerformQualifyingActions_Vote_ValidDelegations(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(30),
		10,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Vote{
					Vote: &rewardtypes.ActionVote{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.NewInt(1000)),
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING
	err, bondedVal1, bondedVal2 := createTestValidators(pubKey, pubKey2, addr, addr2)
	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, []stakingtypes.Validator{bondedVal1, bondedVal2}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypes.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypes.ProposalTypeText),
		coinsPos,
		addr,
	)

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(200 * time.Millisecond)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2, Time: time.Now().UTC()}})
	txGov, err := SignTx(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.Equal(t, true, len(proposal) > 0, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	vote1 := govtypes.NewMsgVote(addr, proposal[0].ProposalId, govtypes.OptionYes)

	assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")

	// threshold will be met after 10 actions
	for height := int64(3); height < int64(23); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) == 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 20, int(accountState.ActionCounter["ActionVote"]), "account state incorrect")
	assert.Equal(t, 10, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
}

func TestRewardsProgramStartPerformQualifyingActions_Delegate_NoQualifyingActions(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(30),
		10,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Delegate{
					Delegate: &rewardtypes.ActionDelegate{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: nil,
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING
	err, bondedVal1, bondedVal2 := createTestValidators(pubKey, pubKey2, addr, addr2)
	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, []stakingtypes.Validator{bondedVal1, bondedVal2}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypes.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypes.ProposalTypeText),
		coinsPos,
		addr,
	)

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(110 * time.Millisecond)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2, Time: time.Now().UTC()}})
	txGov, err := SignTx(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.Equal(t, true, len(proposal) > 0, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	vote1 := govtypes.NewMsgVote(addr, proposal[0].ProposalId, govtypes.OptionYes)

	assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")

	for height := int64(3); height < int64(15); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) == 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(0), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 0, int(accountState.ActionCounter["ActionVote"]), "account state incorrect")
	assert.Equal(t, 0, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, true, len(byAddress.RewardAccountState) == 0, "RewardDistributionsByAddress incorrect")
}

func TestRewardsProgramStartPerformQualifyingActions_Delegate_QualifyingActionsPresent(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))
	minDelegation := sdk.NewInt64Coin("nhash", 4)
	maxDelegation := sdk.NewInt64Coin("nhash", 2001000)
	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(30),
		10,
		10,
		3,
		[]rewardtypes.QualifyingAction{
			{
				Type: &rewardtypes.QualifyingAction_Delegate{
					Delegate: &rewardtypes.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(100, 0),
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING
	err, bondedVal1, bondedVal2 := createTestValidators(pubKey, pubKey2, addr, addr2)
	app := piosimapp.SetupWithGenesisRewardsProgram(uint64(2), rewardProgram, []authtypes.GenesisAccount{acct1}, []stakingtypes.Validator{bondedVal1, bondedVal2}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypes.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypes.ProposalTypeText),
		coinsPos,
		addr,
	)

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(200 * time.Millisecond)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2, Time: time.Now().UTC()}})
	txGov, err := SignTx(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.Equal(t, true, len(proposal) > 0, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	delAddr, _ := sdk.ValAddressFromBech32(bondedVal1.OperatorAddress)
	delegation := stakingtypes.NewMsgDelegate(addr, delAddr, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)))

	assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")

	for height := int64(3); height < int64(23); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), delegation)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.Deliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		time.Sleep(100 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Equal(t, true, len(claimPeriodDistributions) == 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, 20, int(accountState.ActionCounter["ActionDelegate"]), "account state incorrect")
	assert.Equal(t, 10, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")

}

// ContentFromProposalType returns a Content object based on the proposal type.
func ContentFromProposalType(title, desc, ty string) govtypes.Content {
	switch ty {
	case govtypes.ProposalTypeText:
		return govtypes.NewTextProposal(title, desc)

	default:
		return nil
	}
}

func createTestValidators(pubKey types.PubKey, pubKey2 types.PubKey, addr sdk.AccAddress, addr2 sdk.AccAddress) (error, stakingtypes.Validator, stakingtypes.Validator) {
	pk0, err := codectypes.NewAnyWithValue(pubKey)
	pk1, err := codectypes.NewAnyWithValue(pubKey2)
	// initialize the validators
	bondedVal1 := stakingtypes.Validator{
		OperatorAddress: sdk.ValAddress(addr).String(),
		ConsensusPubkey: pk0,
		Description:     stakingtypes.NewDescription("hotdog", "", "", "", ""),
	}
	bondedVal2 := stakingtypes.Validator{
		OperatorAddress: sdk.ValAddress(addr2).String(),
		ConsensusPubkey: pk1,
		Description:     stakingtypes.NewDescription("corndog", "", "", "", ""),
	}
	return err, bondedVal1, bondedVal2
}

// Contains tells if Event type exists in abci.Event.
func contains(a []abci.Event, x string) (*[]abci.Event, bool) {
	var eventSlice []abci.Event
	var found bool
	for _, n := range a {
		if x == n.Type {
			found = true
			eventSlice = append(eventSlice, n)
		}
	}
	return &eventSlice, found
}

// containsAttribute function to return back an attribute based on attribute value, since in tendermit you can have events
// which can have the same type(classic example is message)
func containsAttribute(a []abci.Event, x string) (*abci.EventAttribute, bool) {
	for _, n := range a {
		temp, found := containsEventAttribute(n, x)
		if found {
			return temp, found
		}
	}
	return nil, false
}

func containsEventAttribute(a abci.Event, x string) (*abci.EventAttribute, bool) {
	for _, n := range a.Attributes {
		if x == string(n.Value) {
			return &n, true
		}
	}
	return nil, false
}

func signAndGenTx(gaslimit uint64, fees sdk.Coins, encCfg simappparams.EncodingConfig, pubKey types.PubKey, privKey types.PrivKey, acct authtypes.BaseAccount, chainId string, msg []sdk.Msg) (client.TxBuilder, error) {
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
	return txBuilder, nil
}

func SignTxAndGetBytes(gaslimit uint64, fees sdk.Coins, encCfg simappparams.EncodingConfig, pubKey types.PubKey, privKey types.PrivKey, acct authtypes.BaseAccount, chainId string, msg ...sdk.Msg) ([]byte, error) {
	txBuilder, err := signAndGenTx(gaslimit, fees, encCfg, pubKey, privKey, acct, chainId, msg)
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

func SignTx(gaslimit uint64, fees sdk.Coins, encCfg simappparams.EncodingConfig, pubKey types.PubKey, privKey types.PrivKey, acct authtypes.BaseAccount, chainId string, msg ...sdk.Msg) (sdk.Tx, error) {
	txBuilder, err := signAndGenTx(gaslimit, fees, encCfg, pubKey, privKey, acct, chainId, msg)
	if err != nil {
		return nil, err
	}
	// Send the tx to the app
	return txBuilder.GetTx(), nil
}

// NewTestGasLimit is a test fee gas limit.
// they keep changing this value and our tests break, hence moving it to test.
func NewTestGasLimit() uint64 {
	return 200000
}
