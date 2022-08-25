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
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1) // will create a gas fee of 1stake * gas
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(1000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(400500)))
	app := piosimapp.SetupWithGenesisAccounts("msgfee-testing", []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr1.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000))

	// tx without a fee associated with msg type
	msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))

	// Check both account balances before transaction
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(t, "1000hotdog,400500stake", addr1beforeBalance, "should have the new balance after funding account")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "", addr2beforeBalance, "should have the new balance after funding account")

	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})

	// Check both account balances after transaction
	addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "900hotdog,300500stake", addr1AfterBalance, "should have the new balance running transaction")
	addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "100hotdog", addr2AfterBalance, "should have the new balance running transaction")
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 15, len(res.Events))
	assert.Equal(t, sdk.EventTypeTx, res.Events[4].Type)
	assert.Equal(t, sdk.AttributeKeyFee, string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "100000stake", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[5].Type)
	assert.Equal(t, antewrapper.AttributeKeyMinFeeCharged, string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "100000stake", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "100000stake", string(res.Events[14].Attributes[0].Value))

	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewCoin("hotdog", sdk.NewInt(800)), "", msgfeestypes.DefaultMsgFeeSplit)
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)

	// tx with a fee associated with msg type and account has funds
	msg = banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(50))))
	fees = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100100), sdk.NewInt64Coin("hotdog", 800))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})

	addr1AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "50hotdog,200400stake", addr1AfterBalance, "should have the new balance running transaction")
	addr2AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "150hotdog", addr2AfterBalance, "should have the new balance running transaction")
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 17, len(res.Events))
	assert.Equal(t, sdk.EventTypeTx, res.Events[4].Type)
	assert.Equal(t, sdk.AttributeKeyFee, string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "800hotdog,100100stake", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[5].Type)
	assert.Equal(t, antewrapper.AttributeKeyMinFeeCharged, string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "100000stake", string(res.Events[5].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "800hotdog", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[15].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "100100stake", string(res.Events[15].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[16].Type)
	assert.Equal(t, "msg_fees", string(res.Events[16].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"800hotdog\",\"recipient\":\"\"}]", string(res.Events[16].Attributes[0].Value))

	msgbasedFee = msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin(sdk.DefaultBondDenom, 10), "", msgfeestypes.DefaultMsgFeeSplit)
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)

	// tx with a fee associated with msg type, additional cost is in same base as fee
	msg = banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(50))))
	fees = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100111))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})

	addr1AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "100289stake", addr1AfterBalance, "should have the new balance running transaction")
	addr2AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "200hotdog", addr2AfterBalance, "should have the new balance running transaction")
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
	assert.Equal(t, 17, len(res.Events))
	assert.Equal(t, sdk.EventTypeTx, res.Events[4].Type)
	assert.Equal(t, sdk.AttributeKeyFee, string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "100111stake", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[5].Type)
	assert.Equal(t, antewrapper.AttributeKeyMinFeeCharged, string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "100000stake", string(res.Events[5].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "10stake", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[15].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "100101stake", string(res.Events[15].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[16].Type)
	assert.Equal(t, "msg_fees", string(res.Events[16].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"10stake\",\"recipient\":\"\"}]", string(res.Events[16].Attributes[0].Value))

}

func TestMsgServiceAuthz(t *testing.T) {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	_, _, addr3 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	acct3 := authtypes.NewBaseAccount(addr3, priv2.PubKey(), 2, 0)
	initBalance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(401000)))
	app := piosimapp.SetupWithGenesisAccounts("msgfee-testing", []authtypes.GenesisAccount{acct1, acct2, acct3}, banktypes.Balance{Address: addr1.String(), Coins: initBalance}, banktypes.Balance{Address: addr2.String(), Coins: initBalance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	// Check both account balances before transaction
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(t, "10000hotdog,401000stake", addr1beforeBalance, "should have the new balance after funding account")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "10000hotdog,401000stake", addr2beforeBalance, "should have the new balance after funding account")
	addr3beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
	require.Equal(t, "", addr3beforeBalance, "should have the new balance after funding account")

	msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))
	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewCoin("hotdog", sdk.NewInt(800)), "", msgfeestypes.DefaultMsgFeeSplit)
	app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee)
	app.AuthzKeeper.SaveGrant(ctx, addr2, addr1, banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("hotdog", 500))), time.Now().Add(time.Hour))

	// tx authz send message with correct amount of fees associated
	msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
	fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin("hotdog", 800))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

	// acct1 sent 100hotdog to acct3 with acct2 paying fees 100000stake in gas, 800hotdog msgfees
	addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "9900hotdog,401000stake", addr1AfterBalance, "should have the new balance running transaction")
	addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "9200hotdog,301000stake", addr2AfterBalance, "should have the new balance running transaction")
	addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
	require.Equal(t, "100hotdog", addr3AfterBalance, "should have the new balance running transaction")

	assert.Equal(t, 17, len(res.Events))
	assert.Equal(t, sdk.EventTypeTx, res.Events[4].Type)
	assert.Equal(t, sdk.AttributeKeyFee, string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "800hotdog,100000stake", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[5].Type)
	assert.Equal(t, antewrapper.AttributeKeyMinFeeCharged, string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "100000stake", string(res.Events[5].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[14].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[14].Attributes[0].Key))
	assert.Equal(t, "800hotdog", string(res.Events[14].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[15].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[15].Attributes[0].Key))
	assert.Equal(t, "100000stake", string(res.Events[15].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[16].Type)
	assert.Equal(t, "msg_fees", string(res.Events[16].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"800hotdog\",\"recipient\":\"\"}]", string(res.Events[16].Attributes[0].Value))

	// send 2 successful authz messages
	msgExec = authztypes.NewMsgExec(addr2, []sdk.Msg{msg, msg})
	fees = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200000), sdk.NewInt64Coin("hotdog", 1600))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit()*2, fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

	// acct1 2x sent 100hotdog to acct3 with acct2 paying fees 200000stake in gas, 1600hotdog msgfees
	addr1AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "9700hotdog,401000stake", addr1AfterBalance, "should have the new balance running transaction")
	addr2AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "7600hotdog,101000stake", addr2AfterBalance, "should have the new balance running transaction")
	addr3AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr3).String()
	require.Equal(t, "300hotdog", addr3AfterBalance, "should have the new balance running transaction")

	assert.Equal(t, 22, len(res.Events))
	assert.Equal(t, sdk.EventTypeTx, res.Events[4].Type)
	assert.Equal(t, sdk.AttributeKeyFee, string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "1600hotdog,200000stake", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[5].Type)
	assert.Equal(t, antewrapper.AttributeKeyMinFeeCharged, string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "200000stake", string(res.Events[5].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[19].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[19].Attributes[0].Key))
	assert.Equal(t, "1600hotdog", string(res.Events[19].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[20].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[20].Attributes[0].Key))
	assert.Equal(t, "200000stake", string(res.Events[20].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[21].Type)
	assert.Equal(t, "msg_fees", string(res.Events[21].Attributes[0].Key))
	assert.Equal(t, "[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"2\",\"total\":\"1600hotdog\",\"recipient\":\"\"}]", string(res.Events[21].Attributes[0].Value))

	// tx authz single send message without enough fees associated
	fees = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin("hotdog", 1))
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	txBytes, err = SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
	require.NoError(t, err)
	res = app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, uint32(0xd), res.Code, "res=%+v", res)
	addr1AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "9700hotdog,401000stake", addr1AfterBalance, "should have the new balance running transaction")
	addr2AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "7600hotdog,1000stake", addr2AfterBalance, "should have the new balance running transaction")
	addr3AfterBalance = app.BankKeeper.GetAllBalances(ctx, addr3).String()
	require.Equal(t, "300hotdog", addr3AfterBalance, "should have the new balance running transaction")
}

func TestMsgServiceAssessMsgFee(t *testing.T) {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)
	encCfg := simapp.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin(sdk.DefaultBondDenom, 101000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_001))
	app := piosimapp.SetupWithGenesisAccounts("msgfee-testing", []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr1.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	// Check both account balances before transaction
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(t, "1000hotdog,1190500001nhash,101000stake", addr1beforeBalance, "should have the new balance after funding account")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "", addr2beforeBalance, "should have the new balance after funding account")

	msg := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("test", sdk.NewInt64Coin(msgfeestypes.UsdDenom, 7), addr2.String(), addr1.String())
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_001)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &msg)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

	addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "1000hotdog,1000stake", addr1AfterBalance, "should have the new balance running transaction")
	addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "87500000nhash", addr2AfterBalance, "should have the new balance running transaction")

	assert.Equal(t, 13, len(res.Events))

	assert.Equal(t, sdk.EventTypeTx, res.Events[4].Type)
	assert.Equal(t, sdk.AttributeKeyFee, string(res.Events[4].Attributes[0].Key))
	assert.Equal(t, "1190500001nhash,100000stake", string(res.Events[4].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[5].Type)
	assert.Equal(t, antewrapper.AttributeKeyMinFeeCharged, string(res.Events[5].Attributes[0].Key))
	assert.Equal(t, "100000stake", string(res.Events[5].Attributes[0].Value))
	assert.Equal(t, msgfeestypes.EventTypeAssessCustomMsgFee, res.Events[9].Type)
	assert.Equal(t, msgfeestypes.KeyAttributeName, string(res.Events[9].Attributes[0].Key))
	assert.Equal(t, "test", string(res.Events[9].Attributes[0].Value))
	assert.Equal(t, msgfeestypes.KeyAttributeAmount, string(res.Events[9].Attributes[1].Key))
	assert.Equal(t, "7usd", string(res.Events[9].Attributes[1].Value))
	assert.Equal(t, msgfeestypes.KeyAttributeRecipient, string(res.Events[9].Attributes[2].Key))
	assert.Equal(t, addr2.String(), string(res.Events[9].Attributes[2].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[10].Type)
	assert.Equal(t, antewrapper.AttributeKeyAdditionalFee, string(res.Events[10].Attributes[0].Key))
	assert.Equal(t, "175000000nhash", string(res.Events[10].Attributes[0].Value))
	assert.Equal(t, sdk.EventTypeTx, res.Events[11].Type)
	assert.Equal(t, antewrapper.AttributeKeyBaseFee, string(res.Events[11].Attributes[0].Key))
	assert.Equal(t, "1015500001nhash,100000stake", string(res.Events[11].Attributes[0].Value))
	assert.Equal(t, "provenance.msgfees.v1.EventMsgFees", res.Events[12].Type)
	assert.Equal(t, "msg_fees", string(res.Events[12].Attributes[0].Key))
	assert.Equal(t, fmt.Sprintf("[{\"msg_type\":\"/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest\",\"count\":\"1\",\"total\":\"87500000nhash\",\"recipient\":\"\"},{\"msg_type\":\"/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest\",\"count\":\"1\",\"total\":\"87500000nhash\",\"recipient\":\"%s\"}]", addr2.String()), string(res.Events[12].Attributes[0].Value))

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
