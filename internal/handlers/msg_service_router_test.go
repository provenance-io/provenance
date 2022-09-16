package handlers_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	piosimapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
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
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1) // will create a gas fee of 1stake * gas
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(1000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(400500)))
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing", []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr1.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	// Check both account balances before we begin.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000hotdog,400500stake", addr1beforeBalance, "addr1beforeBalance")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)

	tt.Run("no fee associated with msg type", func(t *testing.T) {
		// Sending 100hotdog with fees of 100000stake.
		// account 1 will lose 100hotdog,100000stake
		// account 2 will gain 100hotdog
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000))

		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		// Check both account balances after transaction
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "900hotdog,300500stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "100hotdog", addr2AfterBalance, "addr2AfterBalance")

		// Make sure a couple events are in the list.
		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "100000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
		}
		assertEventsContains(t, res.Events, expEvents)
	})

	tt.Run("800hotdog fee associated with msg type", func(t *testing.T) {
		// Sending 50hotdog with fees of 100100stake,800hotdog.
		// The send message will have a fee of 800hotdog.
		// account 1 will lose 100100stake,800hotdog.
		// account 2 will gain 50hotdog.
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(50))))
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100100), sdk.NewInt64Coin("hotdog", 800))
		msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewCoin("hotdog", sdk.NewInt(800)), "", 0)
		require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 800hotdog")
		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "50hotdog,200400stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "150hotdog", addr2AfterBalance, "addr2AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "800hotdog,100100stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "800hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", `[{"msg_type":"/cosmos.bank.v1beta1.MsgSend","count":"1","total":"800hotdog","recipient":""}]`)),
		}
		assertEventsContains(t, res.Events, expEvents)
	})

	tt.Run("10stake fee associated with msg type", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(50))))
		msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin(sdk.DefaultBondDenom, 10), "", 0)
		require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 10stake")

		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100111))
		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "100289stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "200hotdog", addr2AfterBalance, "addr2AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "100111stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "10stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", `[{"msg_type":"/cosmos.bank.v1beta1.MsgSend","count":"1","total":"10stake","recipient":""}]`)),
		}
		assertEventsContains(t, res.Events, expEvents)
	})
}

func TestMsgServiceMsgFeeWithRecipient(t *testing.T) {
	// TODO: Required for v1.13.x: Remove this t.Skip() line and fix things so these tests pass. https://github.com/provenance-io/provenance/issues/1006
	t.Skip("This test is disabled, but must be re-enabled before v1.13 can be ready.")

	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1) // will create a gas fee of 1stake * gas
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(1_000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100_000)))
	app := piosimapp.SetupWithGenesisAccounts(t, []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr1.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))
	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewCoin("hotdog", sdk.NewInt(800)), addr2.String(), msgfeestypes.DefaultMsgFeeBips)
	require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 800hotdog")

	// Check both account balances before transaction
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "1000hotdog,100000stake", addr1beforeBalance, "should have the new balance after funding account")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "", addr2beforeBalance, "should have the new balance after funding account")

	fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000), sdk.NewInt64Coin("hotdog", 800))
	txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})

	// Check both account balances after transaction
	addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	require.Equal(t, "100hotdog", addr1AfterBalance, "should have the new balance after running transaction")
	addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	require.Equal(t, "500hotdog", addr2AfterBalance, "should have the new balance after running transaction")
	require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)
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
	expectedTypedEventJson := fmt.Sprintf("[{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"400hotdog\",\"recipient\":\"\"},{\"msg_type\":\"/cosmos.bank.v1beta1.MsgSend\",\"count\":\"1\",\"total\":\"400hotdog\",\"recipient\":\"%s\"}]", addr2.String())
	assert.Equal(t, expectedTypedEventJson, string(res.Events[16].Attributes[0].Value), "typed event should reflect msg fees with recipient and calculated bip amount")
}

func TestMsgServiceAuthz(tt *testing.T) {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	_, _, addr3 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	acct3 := authtypes.NewBaseAccount(addr3, priv2.PubKey(), 2, 0)
	initBalance := sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(10000)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(401000)))
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
	assert.Equal(tt, "10000hotdog,401000stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "10000hotdog,401000stake", addr2beforeBalance, "addr2beforeBalance")
	assert.Equal(tt, "", addr3beforeBalance, "addr3beforeBalance")
	stopIfFailed(tt)

	tt.Run("exec one send", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewCoin("hotdog", sdk.NewInt(100))))

		// tx authz send message with correct amount of fees associated
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin("hotdog", 800))
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		// acct1 sent 100hotdog to acct3 with acct2 paying fees 100000stake in gas, 800hotdog msgfees
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "9900hotdog,401000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "9200hotdog,301000stake", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "100hotdog", addr3AfterBalance, "addr3AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "800hotdog,100000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000stake"),
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
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200000), sdk.NewInt64Coin("hotdog", 1600))
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit()*2, fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		// acct1 2x sent 100hotdog to acct3 with acct2 paying fees 200000stake in gas, 1600hotdog msgfees
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "9740hotdog,401000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "7600hotdog,101000stake", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "260hotdog", addr3AfterBalance, "addr3AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "1600hotdog,200000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "200000stake"),
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
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin("hotdog", 799))
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err)
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, uint32(13), res.Code, "res=%+v", res)

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "9740hotdog,401000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "7600hotdog,101000stake", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "260hotdog", addr3AfterBalance, "addr3AfterBalance")
	})
}

func TestMsgServiceAssessMsgFee(tt *testing.T) {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin(sdk.DefaultBondDenom, 101000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_001))
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing", []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr1.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "msgfee-testing"})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	// Check both account balances before we start.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "1000hotdog,1190500001nhash,101000stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)

	tt.Run("assess custom msg fee", func(t *testing.T) {
		msg := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("test", sdk.NewInt64Coin(msgfeestypes.UsdDenom, 7), addr2.String(), addr1.String())
		txBytes, err := SignTxAndGetBytes(NewTestGasLimit(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_001)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.Equal(t, abci.CodeTypeOK, res.Code, "res=%+v", res)

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "1000hotdog,1000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "87500000nhash", addr2AfterBalance, "addr2AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "1190500001nhash,100000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "100000stake"),
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

func TestRewardsProgramStartError(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	//_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin("atom", 1000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000))
	app := piosimapp.SetupWithGenesisAccounts(t, []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	blockTime := ctx.BlockTime()
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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

	txBytes, err := SignTxAndGetBytes(NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &rewardProgram)
	require.NoError(t, err)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	require.Equal(t, true, res.IsErr(), "should return error", res)
}

func TestRewardsProgramStart(t *testing.T) {
	// TODO: Required for v1.13.x: Remove this t.Skip() line and fix things so these tests pass. https://github.com/provenance-io/provenance/issues/1006
	t.Skip("This test is disabled, but must be re-enabled before v1.13 can be ready.")

	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin("atom", 1000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000))
	app := piosimapp.SetupWithGenesisAccounts(t, []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
	txReward, err := SignTx(NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &rewardProgram)
	require.NoError(t, err)
	_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), txReward)
	require.NoError(t, errFromDeliverTx)
	assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	assert.Len(t, res.Events, 15)
	containsEvent(t, res.Events, "reward_program_created", 1)
	containsEvent(t, res.Events, "message", 3)
	containsEventWithAttribute(t, res.Events, "message", "create_reward_program", 1)
}

func TestRewardsProgramStartPerformQualifyingActions(t *testing.T) {
	// TODO: Required for v1.13.x: Remove this t.Skip() line and fix things so these tests pass. https://github.com/provenance-io/provenance/issues/1006
	t.Skip("This test is disabled, but must be re-enabled before v1.13 can be ready.")

	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000))
	app := piosimapp.SetupWithGenesisAccounts(t, []authtypes.GenesisAccount{acct1}, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
	txReward, err := SignTx(NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &rewardProgram)
	require.NoError(t, err)
	_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), txReward)
	require.NoError(t, errFromDeliverTx)
	assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	assert.Equal(t, uint64(98), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
	assert.Equal(t, 10, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(100)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Len(t, byAddress.RewardAccountState, 1, "only one reward account for one claim period.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(100)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Len(t, byAddress.RewardAccountState, 1, "only one reward account for one claim period.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
	})
	require.NoError(t, err)
	assert.Empty(t, byAddress.RewardAccountState, "none of them should be claimable.")
}

func TestRewardsProgramStartPerformQualifyingActionsRecordedRewardsUnclaimable(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
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

	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(20), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
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

	assert.Empty(t, byAddress.RewardAccountState, "none of the rewards should be in claimed state.")
}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountFunded(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
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
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance}, banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(claimPeriodDistributions), 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(1), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
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
	assert.Len(t, byAddress.RewardAccountState, 4, "claimable rewards should be 4 for this address.")

	// get the accoutn balances of acct1
	balance := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// claim rewards for the address
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 7, Time: time.Now().UTC()}})
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq))
	txClaim, errClaim := SignTxAndGetBytes(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msgClaim)
	require.NoError(t, errClaim)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txClaim})
	require.Equal(t, true, res.IsOK(), "res=%+v", res)
	// unmarshal the TxMsgData
	var protoResult sdk.TxMsgData
	require.NoError(t, proto.Unmarshal(res.Data, &protoResult), "unmarshalling protoResult")
	require.Len(t, protoResult.MsgResponses, 1, "protoResult.MsgResponses")
	require.Equal(t, protoResult.MsgResponses[0].GetTypeUrl(), "/provenance.reward.v1.MsgClaimAllRewardsResponse", "protoResult.MsgResponses[0].GetTypeUrl()")
	claimResponse := rewardtypes.MsgClaimAllRewardsResponse{}
	require.NoError(t, claimResponse.Unmarshal(protoResult.MsgResponses[0].Value), "unmarshalling claimResponse")
	require.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(50_000_000_000)), claimResponse.TotalRewardClaim[0])
	require.Len(t, claimResponse.ClaimDetails, 1)
	require.Equal(t, uint64(1), claimResponse.ClaimDetails[0].RewardProgramId)
	require.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(50_000_000_000)), claimResponse.ClaimDetails[0].TotalRewardClaim)
	require.Len(t, claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails, 5)
	require.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)), claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails[0].ClaimPeriodReward)
	app.EndBlock(abci.RequestEndBlock{Height: 7})
	app.Commit()
	balanceLater := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// make sure account balance has the tokens
	require.Equal(t, sdk.NewInt(50_000_000_000), balanceLater.AmountOf(pioconfig.DefaultBondDenom).Sub(balance.AmountOf(pioconfig.DefaultBondDenom)))

}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountFundedDifferentDenom(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin("hotdog", 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("hotdog", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("hotdog", sdk.NewInt(10_000_000_000)),
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
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance}, banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(claimPeriodDistributions), 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(1), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "hotdog",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
	assert.Equal(t, 1, int(accountState.SharesEarned), "account state incorrect")
	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 5, len(byAddress.RewardAccountState), "claimable and un claimable sum rewards should be 5 for this address.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Len(t, byAddress.RewardAccountState, 4, "claimable rewards should be 4 for this address.")

	// get the accoutn balances of acct1
	balance := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// claim rewards for the address
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 7, Time: time.Now().UTC()}})
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq))
	txClaim, errClaim := SignTxAndGetBytes(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msgClaim)
	require.NoError(t, errClaim)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txClaim})
	require.Equal(t, true, res.IsOK(), "res=%+v", res)
	// unmarshal the TxMsgData
	var protoResult sdk.TxMsgData
	require.NoError(t, proto.Unmarshal(res.Data, &protoResult), "unmarshalling protoResult")
	require.Len(t, protoResult.MsgResponses, 1, "protoResult.MsgResponses")
	require.Equal(t, protoResult.MsgResponses[0].GetTypeUrl(), "/provenance.reward.v1.MsgClaimAllRewardsResponse", "protoResult.MsgResponses[0].GetTypeUrl()")
	claimResponse := rewardtypes.MsgClaimAllRewardsResponse{}
	require.NoError(t, claimResponse.Unmarshal(protoResult.MsgResponses[0].Value), "unmarshalling claimResponse")
	require.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(50_000_000_000)), claimResponse.TotalRewardClaim[0])
	require.Len(t, claimResponse.ClaimDetails, 1)
	require.Equal(t, uint64(1), claimResponse.ClaimDetails[0].RewardProgramId)
	require.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(50_000_000_000)), claimResponse.ClaimDetails[0].TotalRewardClaim)
	require.Len(t, claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails, 5)
	require.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(10_000_000_000)), claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails[0].ClaimPeriodReward)
	app.EndBlock(abci.RequestEndBlock{Height: 7})
	app.Commit()
	balanceLater := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// make sure account balance has the tokens
	require.Equal(t, sdk.NewInt(50_000_000_000), balanceLater.AmountOf("hotdog").Sub(balance.AmountOf("hotdog")))
}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountFundedDifferentDenomClaimedTogether(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin("hotdog", 1000000_000_000_000), sdk.NewInt64Coin("nhash", 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewCoin("hotdog", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("hotdog", sdk.NewInt(10_000_000_000)),
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

	secondRewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		2,
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
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(3), []rewardtypes.RewardProgram{rewardProgram, secondRewardProgram}, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance}, banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(claimPeriodDistributions), 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(1), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "hotdog",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
	assert.Equal(t, 1, int(accountState.SharesEarned), "account state incorrect")
	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Equal(t, 10, len(byAddress.RewardAccountState), "claimable and un claimable sum rewards should be 10 for this address.")

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(10_000_000_000)).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
	assert.Len(t, byAddress.RewardAccountState, 8, "claimable rewards should be 8 for this address.")

	// get the accoutn balances of acct1
	balance := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// claim rewards for the address
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 7, Time: time.Now().UTC()}})
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq))
	// needs extra gas
	txClaim, errClaim := SignTxAndGetBytes(300000, fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msgClaim)
	require.NoError(t, errClaim)
	res := app.DeliverTx(abci.RequestDeliverTx{Tx: txClaim})
	require.Equal(t, true, res.IsOK(), "res=%+v", res)
	// unmarshal the TxMsgData
	var protoResult sdk.TxMsgData
	require.NoError(t, proto.Unmarshal(res.Data, &protoResult), "unmarshalling protoResult")
	require.Len(t, protoResult.MsgResponses, 1, "protoResult.MsgResponses")
	require.Equal(t, protoResult.MsgResponses[0].GetTypeUrl(), "/provenance.reward.v1.MsgClaimAllRewardsResponse", "protoResult.MsgResponses[0].GetTypeUrl()")
	claimResponse := rewardtypes.MsgClaimAllRewardsResponse{}
	require.NoError(t, claimResponse.Unmarshal(protoResult.MsgResponses[0].Value), "unmarshalling claimResponse")
	require.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(50_000_000_000)), claimResponse.TotalRewardClaim[0])
	require.Len(t, claimResponse.ClaimDetails, 2)
	require.Equal(t, uint64(1), claimResponse.ClaimDetails[0].RewardProgramId)
	require.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(50_000_000_000)), claimResponse.ClaimDetails[0].TotalRewardClaim)
	require.Len(t, claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails, 5)
	require.Equal(t, sdk.NewCoin("hotdog", sdk.NewInt(10_000_000_000)), claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails[0].ClaimPeriodReward)

	require.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(50_000_000_000)), claimResponse.ClaimDetails[1].TotalRewardClaim)
	require.Len(t, claimResponse.ClaimDetails[1].ClaimedRewardPeriodDetails, 5)
	require.Equal(t, sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)), claimResponse.ClaimDetails[1].ClaimedRewardPeriodDetails[0].ClaimPeriodReward)

	app.EndBlock(abci.RequestEndBlock{Height: 7})
	app.Commit()
	balanceLater := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// make sure account balance has the tokens
	require.Equal(t, sdk.NewInt(50_000_000_000), balanceLater.AmountOf("hotdog").Sub(balance.AmountOf("hotdog")))

	balanceLater = app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// make sure account balance has the tokens
	require.Equal(t, sdk.NewInt(50_000_000_000), balanceLater.AmountOf("nhash").Sub(balance.AmountOf("nhash")))

}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountNotFunded(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
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
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(claimPeriodDistributions), 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(1), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
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
	assert.Len(t, byAddress.RewardAccountState, 4, "claimable rewards should be 4 for this address.")

	// claim rewards for the address
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 7, Time: time.Now().UTC()}})
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq))
	txClaim, errClaim := SignTxAndGetBytes(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msgClaim)
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
	encCfg := sdksim.MakeTestEncodingConfig()
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

	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
		time.Sleep(1100 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(claimPeriodDistributions), 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(0), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
	assert.Equal(t, 0, int(accountState.SharesEarned), "account state incorrect")
}

// Checks to see if delegation are met for a Qualifying action in this case, Transfer, create an address with delegations
// transfers which map to QualifyingAction map to the delegated address
// delegation threshold is met
func TestRewardsProgramStartPerformQualifyingActionsTransferAndDelegationsPresent(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
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
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2), banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		time.Sleep(1100 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(claimPeriodDistributions), 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(1), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
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
	encCfg := sdksim.MakeTestEncodingConfig()
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
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2), banktypes.Balance{Address: addr.String(), Coins: acct1Balance})
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))

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
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.Equal(t, true, len(res.GetEvents()) >= 1, "should have emitted an event.")
		time.Sleep(1100 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(claimPeriodDistributions), 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(0), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.NewInt(10_000_000_000),
	}, claimPeriodDistributions[0].RewardsPool)

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionTransfer"), "account state incorrect")
	assert.Equal(t, 0, int(accountState.SharesEarned), "account state incorrect")

}

func TestRewardsProgramStartPerformQualifyingActions_Vote(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
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

	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, nil, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
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
	txGov, err := SignTx(NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.NotEmpty(t, proposal, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	vote1 := govtypesv1beta1.NewMsgVote(addr, proposal[0].Id, govtypesv1beta1.OptionYes)

	assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")

	for height := int64(3); height < int64(23); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.GreaterOrEqual(t, len(res.GetEvents()), 1, "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(20), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionVote"), "account state incorrect")
	assert.Equal(t, 10, int(accountState.SharesEarned), "account state incorrect")
}

func TestRewardsProgramStartPerformQualifyingActions_Vote_InvalidDelegations(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
	priv1, pubKey1, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv1.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	acctBalance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1000000_000_000_000))

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

	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1, acct2}, createValSet(t, pubKey1),
		banktypes.Balance{Address: addr1.String(), Coins: acctBalance},
		banktypes.Balance{Address: addr2.String(), Coins: acctBalance},
	)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct2.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
		coinsPos,
		addr1,
	)

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	time.Sleep(200 * time.Millisecond)

	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: 2, Time: time.Now().UTC()}})
	txGov, err := SignTx(NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv1.PubKey(), priv1, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)
	assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	proposal := app.GovKeeper.GetProposals(ctx)
	require.NotEmpty(t, proposal, "proposal has to exist")

	// tx with a fee associated with msg type and account has funds
	vote2 := govtypesv1beta1.NewMsgVote(addr2, proposal[0].Id, govtypesv1beta1.OptionYes)
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	seq := acct2.Sequence

	for height := int64(3); height < int64(5); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct2.SetSequence(seq))
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), vote2)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(0), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, "100000000000nhash", claimPeriodDistributions[0].RewardsPool.String(), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct2.Address)
	require.NoError(t, err, "GetRewardAccountState")
	assert.Equal(t, uint64(0), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionVote"), "account state incorrect")
	assert.Equal(t, 0, int(accountState.SharesEarned), "account state incorrect")

	byAddress1, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err, "RewardDistributionsByAddress acct1")
	assert.Empty(t, byAddress1.RewardAccountState, "RewardDistributionsByAddress incorrect acct1")

	byAddress2, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct2.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err, "RewardDistributionsByAddress acct2")
	assert.Empty(t, byAddress2.RewardAccountState, "RewardDistributionsByAddress incorrect acct2")
}

func TestRewardsProgramStartPerformQualifyingActions_Vote_ValidDelegations(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, _ := testdata.KeyTestPubAddr()
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
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2), banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
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
	txGov, err := SignTx(NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.NotEmpty(t, proposal, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	vote1 := govtypesv1beta1.NewMsgVote(addr, proposal[0].Id, govtypesv1beta1.OptionYes)

	assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")

	// threshold will be met after 10 actions
	for height := int64(3); height < int64(23); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(20), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionVote"), "account state incorrect")
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
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, _ := testdata.KeyTestPubAddr()
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
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2), banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
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
	txGov, err := SignTx(NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.NotEmpty(t, proposal, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	vote1 := govtypesv1beta1.NewMsgVote(addr, proposal[0].Id, govtypesv1beta1.OptionYes)

	assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")

	for height := int64(3); height < int64(15); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(0), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionVote"), "account state incorrect")
	assert.Equal(t, 0, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Empty(t, byAddress.RewardAccountState, "RewardDistributionsByAddress incorrect")
}

func TestRewardsProgramStartPerformQualifyingActions_Delegate_QualifyingActionsPresent(t *testing.T) {
	encCfg := sdksim.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, _ := testdata.KeyTestPubAddr()
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
	valSet := createValSet(t, pubKey, pubKey2)
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, valSet, banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx.WithBlockTime(time.Now())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, acct1.GetAddress(), sdk.NewCoins(sdk.NewCoin(msgfeestypes.NhashDenom, sdk.NewInt(290500010)))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
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
	txGov, err := SignTx(NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(msgfeestypes.NhashDenom, 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err)

	_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), txGov)
	require.NoError(t, errFromDeliverTx)

	app.EndBlock(abci.RequestEndBlock{Height: 2})
	app.Commit()

	seq = seq + 1
	proposal := app.GovKeeper.GetProposals(ctx)

	require.NotEmpty(t, proposal, "proposal has to exist")
	// tx with a fee associated with msg type and account has funds
	delAddr, _ := valSet.GetByIndex(0)
	delegation := stakingtypes.NewMsgDelegate(addr, delAddr, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)))

	assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")

	for height := int64(3); height < int64(23); height++ {
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: height, Time: time.Now().UTC()}})
		require.NoError(t, acct1.SetSequence(seq))
		tx1, err1 := SignTx(NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), delegation)
		require.NoError(t, err1)
		_, res, errFromDeliverTx := app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx1)
		require.NoError(t, errFromDeliverTx)
		assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")
		time.Sleep(100 * time.Millisecond)
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended so shares have to be 0")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.IsEqual(sdk.Coin{
		Denom:  "nhash",
		Amount: sdk.ZeroInt(),
	}), "claim period has not ended so shares have to be 0")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(20), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionDelegate"), "account state incorrect")
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
func ContentFromProposalType(title, desc, ty string) govtypesv1beta1.Content {
	switch ty {
	case govtypesv1beta1.ProposalTypeText:
		return govtypesv1beta1.NewTextProposal(title, desc)

	default:
		return nil
	}
}

func createValSet(t *testing.T, pubKeys ...cryptotypes.PubKey) *tmtypes.ValidatorSet {
	validators := make([]*tmtypes.Validator, len(pubKeys))
	for i, key := range pubKeys {
		pk, err := cryptocodec.ToTmPubKeyInterface(key)
		require.NoError(t, err, "ToTmPubKeyInterface")
		validators[i] = tmtypes.NewValidator(pk, 1)
	}
	return tmtypes.NewValidatorSet(validators)
}

func containsEvent(t *testing.T, haystack []abci.Event, needle string, amount int) {
	t.Helper()

	counter := 0
	var eTypes []string
	for _, n := range haystack {
		if needle == n.Type {
			counter += 1
		}
		eTypes = append(eTypes, n.Type)
	}

	if counter != amount {
		t.Errorf("Found %d %s. Need exactly %d within %v", counter, needle, amount, eTypes)
	}
}

func containsEventWithAttribute(t *testing.T, haystack []abci.Event, needle, attribute string, amount int) {
	t.Helper()

	type AbciEvent struct {
		name       string
		attributes []string
	}

	var events []AbciEvent
	counter := 0
	for _, n := range haystack {
		event := AbciEvent{}
		event.name = n.Type

		for _, a := range n.Attributes {
			event.attributes = append(event.attributes, string(a.Value))
			if string(a.Value) == attribute && event.name == needle {
				counter += 1
				break
			}
		}

		events = append(events, event)
	}

	if counter != amount {
		t.Errorf("Found %d %s with attribute %s. Need exactly %d within %v", counter, needle, attribute, amount, events)
	}
}

func signAndGenTx(gaslimit uint64, fees sdk.Coins, encCfg simappparams.EncodingConfig, pubKey cryptotypes.PubKey, privKey cryptotypes.PrivKey, acct authtypes.BaseAccount, chainId string, msg []sdk.Msg) (client.TxBuilder, error) {
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

func SignTxAndGetBytes(gaslimit uint64, fees sdk.Coins, encCfg simappparams.EncodingConfig, pubKey cryptotypes.PubKey, privKey cryptotypes.PrivKey, acct authtypes.BaseAccount, chainId string, msg ...sdk.Msg) ([]byte, error) {
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

func SignTx(gaslimit uint64, fees sdk.Coins, encCfg simappparams.EncodingConfig, pubKey cryptotypes.PubKey, privKey cryptotypes.PrivKey, acct authtypes.BaseAccount, chainId string, msg ...sdk.Msg) (sdk.Tx, error) {
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
	return 100000
}

func NewTestRewardsGasLimit() uint64 {
	return 200000
}
