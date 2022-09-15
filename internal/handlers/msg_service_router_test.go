package handlers_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	piosimapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/handlers"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

func stopIfFailed(t *testing.T) {
	t.Helper()
	if t.Failed() {
		t.FailNow()
	}
}

func assertEventsContains(t *testing.T, events, contains []abci.Event, msgAndArgs ...interface{}) bool {
	t.Helper()
	missingEvents := []abci.Event{}
LookingForLoop:
	for _, lookingFor := range contains {
		for _, event := range events {
			if eventContains(event, lookingFor) {
				continue LookingForLoop
			}
		}
		missingEvents = append(missingEvents, lookingFor)
	}
	if len(missingEvents) > 0 {
		err := fmt.Sprintf("Events missing %d/%d expected entries.\n"+
			"Events:\n%s\n"+
			"Missing:\n%s",
			len(missingEvents), len(contains),
			eventsString(events, true), eventsString(missingEvents, false),
		)
		msg := messageFromMsgAndArgs(msgAndArgs...)
		if len(msg) > 0 {
			err = err + "\nMessage: " + msg
		}
		t.Error(err)
		return false
	}
	return true
}

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	switch len(msgAndArgs) {
	case 0:
		return ""
	case 1:
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	default:
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}

// eventContains returns true if the provided event has the same type as contains and has all
// attributes (key/value) that contains has. The event can have extra attributes too, but it has to have
// at least all the attributes in contains.
func eventContains(event, contains abci.Event) bool {
	if event.Type != contains.Type {
		return false
	}
	return eventAttributesHasSubset(event.Attributes, contains.Attributes)
}

func eventAttributesHasSubset(eventAttributes, subset []abci.EventAttribute) bool {
SubsetLoop:
	for _, sattr := range subset {
		for _, eattr := range eventAttributes {
			if bytes.Equal(sattr.Key, eattr.Key) && bytes.Equal(sattr.Value, eattr.Value) {
				continue SubsetLoop
			}
		}
		return false
	}
	return true
}

func eventsString(events []abci.Event, includeIndex bool) string {
	strs := make([]string, len(events))
	for i, e := range events {
		if includeIndex {
			strs[i] = fmt.Sprintf("    [%d]: %s", i, eventString(e))
		} else {
			strs[i] = fmt.Sprintf("    %s", eventString(e))
		}
	}
	return strings.Join(strs, "\n")
}

func eventString(event abci.Event) string {
	attrs := make([]string, len(event.Attributes))
	for i, attr := range event.Attributes {
		attrs[i] = fmt.Sprintf("%q = %q", attr.Key, attr.Value)
	}
	return fmt.Sprintf("%s: %s", event.Type, strings.Join(attrs, ", "))
}

func NewEvent(ty string, attrs ...abci.EventAttribute) abci.Event {
	return abci.Event{
		Type:       ty,
		Attributes: attrs,
	}
}

func NewAttribute(key, value string) abci.EventAttribute {
	return abci.EventAttribute{
		Key:   []byte(key),
		Value: []byte(value),
	}
}

func TestRegisterMsgService(t *testing.T) {
	db := dbm.NewMemDB()

	// Create an encoding config that doesn't register testdata Msg services.
	encCfg := sdksim.MakeTestEncodingConfig()
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
	encCfg := sdksim.MakeTestEncodingConfig()
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

func TestMsgService(tt *testing.T) {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 1) // will create a gas fee of 1atom * gas
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(1000)), sdk.NewCoin("atom", sdk.NewInt(400500)))
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing", []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr1.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	// Check both account balances before we begin.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "400500atom,1000hotdog", addr1beforeBalance, "addr1beforeBalance")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)

	tt.Run("no fee associated with msg type", func(t *testing.T) {
		// Sending 100hotdog with fees of 100000atom.
		// account 1 will lose 100hotdog,100000atom
		// account 2 will gain 100hotdog
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))
		fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 100000))

		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		// Check both account balances after transaction
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "300500atom,900hotdog", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "100hotdog", addr2AfterBalance, "addr2AfterBalance")

		// Make sure a couple events are in the list.
		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "100000atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
		}
		assertEventsContains(t, res.Events, expEvents)
	})

	tt.Run("800hotdog fee associated with msg type", func(t *testing.T) {
		// Sending 50hotdog with fees of 100100atom,800hotdog.
		// The send message will have a fee of 800hotdog.
		// account 1 will lose 100100atom,800hotdog.
		// account 2 will gain 50hotdog.
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(50))))
		fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 100100), sdk.NewInt64Coin("hotdog", 800))
		msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewCoin("hotdog", sdk.NewInt(800)))
		require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 800hotdog")

		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "200400atom,50hotdog", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "150hotdog", addr2AfterBalance, "addr2AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "100100atom,800hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "800hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", `[{"msg_type":"/cosmos.bank.v1beta1.MsgSend","count":"1","total":"800hotdog","recipient":""}]`)),
		}
		assertEventsContains(t, res.Events, expEvents)
	})

	tt.Run("10atom fee associated with msg type", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(50))))
		msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin("atom", 10))
		require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 10atom")

		fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 100111))
		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "100289atom", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "200hotdog", addr2AfterBalance, "addr2AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "100111atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "10atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", `[{"msg_type":"/cosmos.bank.v1beta1.MsgSend","count":"1","total":"10atom","recipient":""}]`)),
		}
		assertEventsContains(t, res.Events, expEvents)
	})
}

func TestMsgServiceAuthz(tt *testing.T) {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 1)
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	_, _, addr3 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	acct3 := authtypes.NewBaseAccount(addr3, priv2.PubKey(), 2, 0)
	initBalance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10000)), sdk.NewCoin("atom", sdk.NewInt(401000)))
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing", []authtypes.GenesisAccount{acct1, acct2, acct3}, banktypes.Balance{Address: addr1.String(), Coins: initBalance}, banktypes.Balance{Address: addr2.String(), Coins: initBalance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	// Create an authz grant from addr1 to addr2 for 500hotdog.
	now := ctx.BlockHeader().Time
	require.NotNil(tt, now, "now")
	exp1Hour := now.Add(time.Hour)
	require.NoError(tt, app.AuthzKeeper.SaveGrant(ctx, addr2, addr1, banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("hotdog", 500))), &exp1Hour), "Save Grant addr2 addr1 500hotdog")
	// Set a MsgSend msg-based fee of 800hotdog.
	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(&banktypes.MsgSend{}), sdk.NewCoin("hotdog", sdk.NewInt(800)))
	require.NoError(tt, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 800hotdog")

	// Check all account balances before we start.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	addr3beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
	assert.Equal(tt, "401000atom,10000hotdog", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "401000atom,10000hotdog", addr2beforeBalance, "addr2beforeBalance")
	assert.Equal(tt, "", addr3beforeBalance, "addr3beforeBalance")
	stopIfFailed(tt)

	tt.Run("exec one send", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))

		// tx authz send message with correct amount of fees associated
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
		fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin("hotdog", 800))
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		// acct1 sent 100hotdog to acct3 with acct2 paying fees 100000atom in gas, 800hotdog msgfees
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "401000atom,9900hotdog", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "301000atom,9200hotdog", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "100hotdog", addr3AfterBalance, "addr3AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "100000atom,800hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "800hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", `[{"msg_type":"/cosmos.bank.v1beta1.MsgSend","count":"1","total":"800hotdog","recipient":""}]`)),
		}
		assertEventsContains(t, res.Events, expEvents)
	})

	tt.Run("exec two sends", func(t *testing.T) {
		// send 2 successful authz messages
		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(80))))
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg, msg})
		fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 200000), sdk.NewInt64Coin("hotdog", 1600))
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit()*2, fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		// acct1 2x sent 100hotdog to acct3 with acct2 paying fees 200000atom in gas, 1600hotdog msgfees
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "401000atom,9740hotdog", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "101000atom,7600hotdog", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "260hotdog", addr3AfterBalance, "addr3AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "200000atom,1600hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "200000atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "1600hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", `[{"msg_type":"/cosmos.bank.v1beta1.MsgSend","count":"2","total":"1600hotdog","recipient":""}]`)),
		}
		assertEventsContains(t, res.Events, expEvents)
	})

	tt.Run("not enough fees", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
		fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin("hotdog", 799))
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err)
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, uint32(13), res.Code, "res=%+v", res)

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "401000atom,9740hotdog", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "1000atom,7600hotdog", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "260hotdog", addr3AfterBalance, "addr3AfterBalance")
	})
}

func TestMsgServiceAssessMsgFee(tt *testing.T) {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 1)
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin("atom", 101000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_001))
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing", []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr1.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	// Check both account balances before we start.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "101000atom,1000hotdog,1190500001nhash", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)

	tt.Run("assess custom msg fee", func(t *testing.T) {
		msg := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("test", sdk.NewInt64Coin(msgfeestypes.UsdDenom, 7), addr2.String(), addr1.String())
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 100000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_001)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "1000atom,1000hotdog", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "87500000nhash", addr2AfterBalance, "addr2AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "100000atom,1190500001nhash"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000atom"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(msgfeestypes.EventTypeAssessCustomMsgFee,
				NewAttribute(msgfeestypes.KeyAttributeName, "test"),
				NewAttribute(msgfeestypes.KeyAttributeAmount, "7usd"),
				NewAttribute(msgfeestypes.KeyAttributeRecipient, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "175000000nhash"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", fmt.Sprintf(`[{"msg_type":"/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest","count":"1","total":"87500000nhash","recipient":""},{"msg_type":"/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest","count":"1","total":"87500000nhash","recipient":"%s"}]`, addr2.String()))),
		}
		assertEventsContains(t, res.Events, expEvents)
	})
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
