package handlers_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	piosimapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/msgfees/types"
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
	var missingEvents []abci.Event
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
			if sattr.Key == eattr.Key && sattr.Value == eattr.Value {
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
		Key:   key,
		Value: value,
	}
}

func msgFeesMsgSendEventJSON(count int, amount int, denom string, recipient string) string {
	return msgFeesEventJSON("/cosmos.bank.v1beta1.MsgSend", count, amount, denom, recipient)
}

func msgFeesEventJSON(msg_type string, count int, amount int, denom string, recipient string) string {
	return fmt.Sprintf(`{"msg_type":"%s","count":"%d","total":"%d%s","recipient":"%s"}`,
		msg_type, count, amount, denom, recipient)
}

func jsonArrayJoin(entries ...string) string {
	return "[" + strings.Join(entries, ",") + "]"
}

func getLastProposal(t *testing.T, ctx sdk.Context, app *piosimapp.App) *govtypesv1.Proposal {
	var rv *govtypesv1.Proposal
	var highestProposalID uint64 = 0

	err := app.GovKeeper.Proposals.Walk(ctx, nil, func(key uint64, value govtypesv1.Proposal) (stop bool, err error) {
		if value.Id > highestProposalID {
			highestProposalID = value.Id
			rv = &value
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("Error walking through proposals: %v", err)
	}

	require.NotNil(t, rv, "no gov proposals found")
	return rv
}

func TestRegisterMsgService(t *testing.T) {
	db := dbm.NewMemDB()

	// Create an encoding config that doesn't register testdata Msg services.
	encCfg := moduletestutil.MakeTestEncodingConfig()
	log.NewTestLogger(t)
	app := baseapp.NewBaseApp("test", log.NewTestLogger(t), db, encCfg.TxConfig.TxDecoder())
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
	encCfg := moduletestutil.MakeTestEncodingConfig()
	app := baseapp.NewBaseApp("test", log.NewTestLogger(t), db, encCfg.TxConfig.TxDecoder())
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

func TestFailedTx(tt *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 1) // will create a gas fee of 1stake * gas
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit())+1))
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "msgfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeModuleAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	gasFeesString := fmt.Sprintf("%v%s", (NewTestGasLimit()), sdk.DefaultBondDenom)

	// Check both account balances before we begin.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "150001stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)

	tt.Run("no msg-based fee", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2)))
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit())))
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		assert.NoError(t, err, "FinalizeBlock expected no error")
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.Equal(t, uint32(0x5), blockRes.TxResults[0].Code, "code 5 insufficient funds error")

		// Check both account balances after transaction
		// the 150000stake should have been deducted from account 1, and the send should have failed.
		// So account 2 should still be empty, and account 1 should only have 1 left.
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "1stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "", addr2AfterBalance, "addr2AfterBalance")

		// Make sure a couple events are in the list.
		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, gasFeesString),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyMinFeeCharged, gasFeesString),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
		}
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150000)))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})

	// Give acct1 150000stake back.
	require.NoError(tt, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(),
		sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit())))),
		fmt.Sprintf("funding acct1 with %s", gasFeesString))

	tt.Run("10stake fee associated with msg type", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2)))
		msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin(sdk.DefaultBondDenom, 10), "", 0)
		require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 10stake")
		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit())+10))
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		assert.NoError(t, err, "FinalizeBlock expected no error")
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.Equal(t, uint32(0x5), blockRes.TxResults[0].Code, "code 5 insufficient funds error")

		// Check both account balances after transaction
		// the 150000 should have been deducted from account 1, and the send should have failed.
		// So account 2 should still be empty, and account 1 should only have 1 left.
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "1stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "", addr2AfterBalance, "addr2AfterBalance")

		// Make sure a couple events are in the list.
		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, gasFeesString),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyMinFeeCharged, gasFeesString),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
		}
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit()))))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})
}

func TestMsgService(tt *testing.T) {
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 1) // set denom as stake and floor gas price as 1 stake.
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin(sdk.DefaultBondDenom, 600_500))
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "msgfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeModuleAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	gasFeesString := fmt.Sprintf("%v%s", (NewTestGasLimit()), sdk.DefaultBondDenom)

	// Check both account balances before we begin.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	assert.Equal(tt, "1000hotdog,600500stake", addr1beforeBalance, "addr1beforeBalance")
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)

	tt.Run("no fee associated with msg type", func(t *testing.T) {
		// Sending 100hotdog with fees of 150000stake.
		// account 1 will lose 100hotdog,150000stake
		// account 2 will gain 100hotdog
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 100)))
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150000))
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.NoError(t, err, "FinalizeBlock() error")

		// Check both account balances after transaction
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "900hotdog,450500stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "100hotdog", addr2AfterBalance, "addr2AfterBalance")

		// Make sure a couple events are in the list.
		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, gasFeesString),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyMinFeeCharged, gasFeesString),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
		}
		// fee charge in antehandler
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit()))))...)
		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})

	tt.Run("800hotdog fee associated with msg type", func(t *testing.T) {
		// Sending 50hotdog with fees of 150100stake,800hotdog.
		// The send message will have a fee of 800hotdog.
		// account 1 will lose 100100stake,800hotdog.
		// account 2 will gain 50hotdog.
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 50)))
		msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin("hotdog", 800), "", 0)
		require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 800hotdog")
		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit())+100), sdk.NewInt64Coin("hotdog", 800))
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.NoError(t, err, "FinalizeBlock() error")

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "50hotdog,300400stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "150hotdog", addr2AfterBalance, "addr2AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "800hotdog,150100stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyMinFeeCharged, gasFeesString),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "800hotdog"),
				NewAttribute(antewrapper.AttributeKeyBaseFee, "150100stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", jsonArrayJoin(msgFeesMsgSendEventJSON(1, 800, "hotdog", "")))),
		}
		// fee charge in antehandler
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit()))))...)
		// fee charged for msg based fee
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("hotdog", 800)))...)
		// swept fee amount
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})

	tt.Run("10stake fee associated with msg type", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 50)))
		msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin(sdk.DefaultBondDenom, 10), "", 0)
		require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 10stake")

		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit())+111))
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.NoError(t, err, "FinalizeBlock() error")

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "150289stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "200hotdog", addr2AfterBalance, "addr2AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "150111stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyMinFeeCharged, gasFeesString),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "10stake"),
				NewAttribute(antewrapper.AttributeKeyBaseFee, "150101stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", jsonArrayJoin(msgFeesMsgSendEventJSON(1, 10, "stake", "")))),
		}
		// fee charge in antehandler
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit()))))...)
		// fee charged for msg based fee
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)))...)
		// swept fee amount
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 101)))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})
}

func TestMsgServiceMsgFeeWithRecipient(t *testing.T) {
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 1)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	gasAmt := NewTestGasLimit() + 20_000
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1_000), sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(gasAmt)))
	app := piosimapp.SetupWithGenesisAccounts(t, "msgfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "msgfee-testing"})
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeModuleAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	// Check both account balances before transaction
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(t, "1000hotdog,170000stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(t, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(t)

	// Sending 100hotdog coin from 1 to 2.
	// Will have a msg fee of 800hotdog, 600 will go to 2, 200 to module.
	msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 100)))
	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin("hotdog", 800), addr2.String(), 7_500)
	require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 800hotdog addr2 75%")

	fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(gasAmt)), sdk.NewInt64Coin("hotdog", 800))
	txBytes, err := SignTxAndGetBytes(ctx, gasAmt, fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err, "SignTxAndGetBytes")
	blockRes, err := app.FinalizeBlock(
		&abci.RequestFinalizeBlock{
			Height: ctx.BlockHeight() + 1,
			Txs:    [][]byte{txBytes},
		},
	)
	assert.NoError(t, err, "FinalizeBlock() error")

	// Check both account balances after transaction
	addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(t, "100hotdog", addr1AfterBalance, "addr1AfterBalance")
	assert.Equal(t, "700hotdog", addr2AfterBalance, "addr2AfterBalance")

	expEvents := []abci.Event{
		NewEvent(sdk.EventTypeTx,
			NewAttribute(sdk.AttributeKeyFee, "800hotdog,170000stake"),
			NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
		NewEvent(sdk.EventTypeTx,
			NewAttribute(antewrapper.AttributeKeyAdditionalFee, "800hotdog"),
			NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
		NewEvent(sdk.EventTypeTx,
			NewAttribute(antewrapper.AttributeKeyBaseFee, "170000stake"),
			NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
		NewEvent("provenance.msgfees.v1.EventMsgFees",
			NewAttribute("msg_fees",
				jsonArrayJoin(msgFeesMsgSendEventJSON(1, 200, "hotdog", ""), msgFeesMsgSendEventJSON(1, 600, "hotdog", addr2.String())))),
	}
	// fee charge in antehandler
	expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(gasAmt))))...)
	// fee charged for msg based fee
	expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("hotdog", 200)))...)
	// fee charged for msg based fee
	expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), addr2.String(), sdk.NewCoins(sdk.NewInt64Coin("hotdog", 600)))...)

	assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
}

func TestMsgServiceAuthz(tt *testing.T) {
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 1)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	_, _, addr3 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	acct3 := authtypes.NewBaseAccount(addr3, priv2.PubKey(), 2, 0)
	initBalance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 10_000), sdk.NewInt64Coin(sdk.DefaultBondDenom, 721_000))
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing",
		[]authtypes.GenesisAccount{acct1, acct2, acct3},
		banktypes.Balance{Address: addr1.String(), Coins: initBalance},
		banktypes.Balance{Address: addr2.String(), Coins: initBalance},
	)
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "msgfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeModuleAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	// Create an authz grant from addr1 to addr2 for 500hotdog.
	now := ctx.BlockHeader().Time
	exp1Hour := now.Add(time.Hour)
	sendAuth := banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("hotdog", 500)), nil)
	require.NoError(tt, app.AuthzKeeper.SaveGrant(ctx, addr2, addr1, sendAuth, &exp1Hour), "Save Grant addr2 addr1 500hotdog")
	// Set a MsgSend msg-based fee of 800hotdog.
	msgbasedFee := msgfeestypes.NewMsgFee(sdk.MsgTypeURL(&banktypes.MsgSend{}), sdk.NewInt64Coin("hotdog", 800), "", 0)
	require.NoError(tt, app.MsgFeesKeeper.SetMsgFee(ctx, msgbasedFee), "setting fee 800hotdog")

	// Check all account balances before we start.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	addr3beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
	assert.Equal(tt, "10000hotdog,721000stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "10000hotdog,721000stake", addr2beforeBalance, "addr2beforeBalance")
	assert.Equal(tt, "", addr3beforeBalance, "addr3beforeBalance")
	stopIfFailed(tt)

	tt.Run("exec one send", func(t *testing.T) {
		// tx authz send message with correct amount of fees associated
		gasAmt := NewTestGasLimit() + 20_000
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)

		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 100)))
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(gasAmt)), sdk.NewInt64Coin("hotdog", 800))
		txBytes, err := SignTxAndGetBytes(ctx, gasAmt, fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		assert.NoError(t, err, "FinalizeBlock() error")

		// acct1 sent 100hotdog to acct3 with acct2 paying fees 100000stake in gas, 800hotdog msgfees
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "9900hotdog,721000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "9200hotdog,551000stake", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "100hotdog", addr3AfterBalance, "addr3AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "800hotdog,170000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "170000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "800hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", jsonArrayJoin(msgFeesMsgSendEventJSON(1, 800, "hotdog", "")))),
		}
		// fee charge in antehandler
		expEvents = append(expEvents, CreateSendCoinEvents(addr2.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(gasAmt))))...)
		// fee charged for msg based fee
		expEvents = append(expEvents, CreateSendCoinEvents(addr2.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("hotdog", 800)))...)
		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})

	tt.Run("exec two sends", func(t *testing.T) {
		// send 2 successful authz messages
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 80)))
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg, msg})
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 300000), sdk.NewInt64Coin("hotdog", 1600))
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit()*2, fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		assert.NoError(t, err, "FinalizeBlock() error")

		// acct1 2x sent 100hotdog to acct3 with acct2 paying fees 200000stake in gas, 1600hotdog msgfees
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "9740hotdog,721000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "7600hotdog,251000stake", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "260hotdog", addr3AfterBalance, "addr3AfterBalance")

		expEvents := []abci.Event{
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "1600hotdog,300000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyBaseFee, "300000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "1600hotdog"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", jsonArrayJoin(msgFeesMsgSendEventJSON(2, 1600, "hotdog", "")))),
		}
		// fee charge in antehandler
		expEvents = append(expEvents, CreateSendCoinEvents(addr2.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 300000)))...)
		// fee charged for msg based fee
		expEvents = append(expEvents, CreateSendCoinEvents(addr2.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1600)))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})

	tt.Run("not enough fees", func(t *testing.T) {
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 100)))
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin("hotdog", 799))
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.Equal(t, uint32(0xd), blockRes.TxResults[0].Code, "code 13 insufficient fee")

		// addr2 pays the base fee, but nothing else is changes.
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "9740hotdog,721000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "7600hotdog,101000stake", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "260hotdog", addr3AfterBalance, "addr3AfterBalance")
	})
}

func TestMsgServiceAssessMsgFee(tt *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	pioconfig.ChangeMsgFeeFloorDenom(1, sdk.DefaultBondDenom)

	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin("hotdog", 1000),
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 151000),
		sdk.NewInt64Coin(NHash, 1_190_500_001),
	)
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "msgfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeModuleAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	// Check both account balances before we start.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "1000hotdog,1190500001nhash,151000stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)
	tt.Run("assess custom msg fee", func(t *testing.T) {

		msgFeeCoin := sdk.NewInt64Coin(msgfeestypes.UsdDenom, 7)
		msg := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("test", msgFeeCoin, addr2.String(), addr1.String(), "")
		fees := sdk.NewCoins(
			sdk.NewInt64Coin(sdk.DefaultBondDenom, 150000),
			sdk.NewInt64Coin(NHash, 1_190_500_001),
		)
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.NoError(t, err, "FinalizeBlock() error")

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "1000hotdog,1000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "175000000nhash", addr2AfterBalance, "addr2AfterBalance") // addr2 gets all the fee as recipient

		expEvents := []abci.Event{
			NewEvent(
				types.EventTypeAssessCustomMsgFee,
				NewAttribute(types.KeyAttributeName, "test"),
				NewAttribute(types.KeyAttributeAmount, "7usd"),
				NewAttribute(types.KeyAttributeRecipient, addr2.String()),
				NewAttribute(types.KeyAttributeBips, "10000"),
			),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "1190500001nhash,150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyMinFeeCharged, "150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(msgfeestypes.EventTypeAssessCustomMsgFee,
				NewAttribute(msgfeestypes.KeyAttributeName, "test"),
				NewAttribute(msgfeestypes.KeyAttributeAmount, "7usd"),
				NewAttribute(msgfeestypes.KeyAttributeRecipient, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "175000000nhash"),
				NewAttribute(antewrapper.AttributeKeyBaseFee, "1015500001nhash,150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", jsonArrayJoin(
					msgFeesEventJSON("/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest", 1, 175000000, "nhash", addr2.String())))),
		}
		// fee charge in antehandler
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 1015500001)))...)
		// fee charged for msg based fee to recipient from assess msg split
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), addr2.String(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 175000000)))...)
		// swept amount
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 1015500001)))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})
}

func TestMsgServiceAssessMsgFeeWithBips(tt *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	pioconfig.ChangeMsgFeeFloorDenom(1, sdk.DefaultBondDenom)

	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin("hotdog", 1000),
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 151000),
		sdk.NewInt64Coin(NHash, 1_190_500_001),
	)
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "msgfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeModuleAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	// Check both account balances before we start.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "1000hotdog,1190500001nhash,151000stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)
	tt.Run("assess custom msg fee", func(t *testing.T) {
		msgFeeCoin := sdk.NewInt64Coin(msgfeestypes.UsdDenom, 7)
		msg := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("test", msgFeeCoin, addr2.String(), addr1.String(), "2500")
		fees := sdk.NewCoins(
			sdk.NewInt64Coin(sdk.DefaultBondDenom, 150000),
			sdk.NewInt64Coin(NHash, 1_190_500_001),
		)
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.NoError(t, err, "FinalizeBlock() error")

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		assert.Equal(t, "1000hotdog,1000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "43750000nhash", addr2AfterBalance, "addr2AfterBalance") // addr2 gets all the fee as recipient

		expEvents := []abci.Event{
			NewEvent(
				types.EventTypeAssessCustomMsgFee,
				NewAttribute(types.KeyAttributeName, "test"),
				NewAttribute(types.KeyAttributeAmount, "7usd"),
				NewAttribute(types.KeyAttributeRecipient, addr2.String()),
				NewAttribute(types.KeyAttributeBips, "2500"),
			),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "1190500001nhash,150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyMinFeeCharged, "150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(msgfeestypes.EventTypeAssessCustomMsgFee,
				NewAttribute(msgfeestypes.KeyAttributeName, "test"),
				NewAttribute(msgfeestypes.KeyAttributeAmount, "7usd"),
				NewAttribute(msgfeestypes.KeyAttributeRecipient, addr2.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "175000000nhash"),
				NewAttribute(antewrapper.AttributeKeyBaseFee, "1015500001nhash,150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", jsonArrayJoin(
					msgFeesEventJSON("/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest", 1, 131250000, "nhash", ""),
					msgFeesEventJSON("/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest", 1, 43750000, "nhash", addr2.String())))),
		}
		// fee charge in antehandler
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 1015500001)))...)
		// fee charged for msg based fee to recipient from assess msg split
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), addr2.String(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 43750000)))...)
		// swept amount
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 1015500001)))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})
}

// TestMsgServiceAssessMsgFeeNoRecipient tests that if no recipient the full fee goes to the module account
func TestMsgServiceAssessMsgFeeNoRecipient(tt *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	pioconfig.ChangeMsgFeeFloorDenom(1, sdk.DefaultBondDenom)

	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin("hotdog", 1000),
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 151000),
		sdk.NewInt64Coin(NHash, 1_190_500_001),
	)
	app := piosimapp.SetupWithGenesisAccounts(tt, "msgfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "msgfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeModuleAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	// Check both account balances before we start.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "1000hotdog,1190500001nhash,151000stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)
	tt.Run("assess custom msg fee", func(t *testing.T) {

		msgFeeCoin := sdk.NewInt64Coin(msgfeestypes.UsdDenom, 7)
		msg := msgfeestypes.NewMsgAssessCustomMsgFeeRequest("test", msgFeeCoin, "", addr1.String(), "")
		fees := sdk.NewCoins(
			sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(NewTestGasLimit())),
			sdk.NewInt64Coin(NHash, 1_190_500_001),
		)
		txBytes, err := SignTxAndGetBytes(ctx, NewTestGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), &msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.NoError(t, err, "FinalizeBlock() error")

		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		assert.Equal(t, "1000hotdog,1000stake", addr1AfterBalance, "addr1AfterBalance")

		expEvents := []abci.Event{
			NewEvent(
				types.EventTypeAssessCustomMsgFee,
				NewAttribute(types.KeyAttributeName, "test"),
				NewAttribute(types.KeyAttributeAmount, "7usd"),
				NewAttribute(types.KeyAttributeRecipient, ""),
				NewAttribute(types.KeyAttributeBips, ""),
			),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(sdk.AttributeKeyFee, "1190500001nhash,150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyMinFeeCharged, "150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent(msgfeestypes.EventTypeAssessCustomMsgFee,
				NewAttribute(msgfeestypes.KeyAttributeName, "test"),
				NewAttribute(msgfeestypes.KeyAttributeAmount, "7usd"),
				NewAttribute(msgfeestypes.KeyAttributeRecipient, "")),
			NewEvent(sdk.EventTypeTx,
				NewAttribute(antewrapper.AttributeKeyAdditionalFee, "175000000nhash"),
				NewAttribute(antewrapper.AttributeKeyBaseFee, "1015500001nhash,150000stake"),
				NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			NewEvent("provenance.msgfees.v1.EventMsgFees",
				NewAttribute("msg_fees", jsonArrayJoin(
					msgFeesEventJSON("/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest", 1, 175000000, "nhash", "")))),
		}
		// fee charge in antehandler
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 1015500001)))...)
		// swept amount
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 1015500001)))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})
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

func createValSet(t *testing.T, pubKeys ...cryptotypes.PubKey) *cmttypes.ValidatorSet {
	validators := make([]*cmttypes.Validator, len(pubKeys))
	for i, key := range pubKeys {
		pk, err := cryptocodec.ToTmPubKeyInterface(key)
		require.NoError(t, err, "ToTmPubKeyInterface")
		validators[i] = cmttypes.NewValidator(pk, 1)
	}
	return cmttypes.NewValidatorSet(validators)
}

func signAndGenTx(
	ctx sdk.Context,
	gaslimit uint64,
	fees sdk.Coins,
	encCfg moduletestutil.TestEncodingConfig,
	pubKey cryptotypes.PubKey,
	privKey cryptotypes.PrivKey,
	acct authtypes.BaseAccount,
	chainId string,
	msg []sdk.Msg,
) (client.TxBuilder, error) {
	txBuilder := encCfg.TxConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(fees)
	txBuilder.SetGasLimit(gaslimit)
	err := txBuilder.SetMsgs(msg...)
	if err != nil {
		return nil, err
	}

	signingMode := signing.SignMode(encCfg.TxConfig.SignModeHandler().DefaultMode())
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signingMode,
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
		Address:       sdk.AccAddress(pubKey.Bytes()).String(),
		ChainID:       chainId,
		AccountNumber: acct.AccountNumber,
		Sequence:      acct.Sequence,
		PubKey:        pubKey,
	}
	sigV2, err = tx.SignWithPrivKey(
		ctx, signingMode, signerData,
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

func SignTxAndGetBytes(
	ctx sdk.Context,
	gaslimit uint64,
	fees sdk.Coins,
	encCfg moduletestutil.TestEncodingConfig,
	pubKey cryptotypes.PubKey,
	privKey cryptotypes.PrivKey,
	acct authtypes.BaseAccount,
	chainId string,
	msg ...sdk.Msg,
) ([]byte, error) {
	txBuilder, err := signAndGenTx(ctx, gaslimit, fees, encCfg, pubKey, privKey, acct, chainId, msg)
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

func SignTx(
	ctx sdk.Context,
	gaslimit uint64,
	fees sdk.Coins,
	encCfg moduletestutil.TestEncodingConfig,
	pubKey cryptotypes.PubKey,
	privKey cryptotypes.PrivKey,
	acct authtypes.BaseAccount,
	chainId string,
	msg ...sdk.Msg,
) (sdk.Tx, error) {
	txBuilder, err := signAndGenTx(ctx, gaslimit, fees, encCfg, pubKey, privKey, acct, chainId, msg)
	if err != nil {
		return nil, err
	}
	// Send the tx to the app
	return txBuilder.GetTx(), nil
}

// NewTestGasLimit is a test fee gas limit.
// they keep changing this value and our tests break, hence moving it to test.
func NewTestGasLimit() uint64 {
	return 150000
}

// CreateSendCoinEvents creates the sequence of events that are created on bankkeeper.SendCoins
func CreateSendCoinEvents(fromAddress, toAddress string, amt sdk.Coins) []abci.Event {
	events := sdk.NewEventManager().Events()
	// subUnlockedCoins event `coin_spent`
	events = events.AppendEvent(sdk.NewEvent(
		banktypes.EventTypeCoinSpent,
		sdk.NewAttribute(banktypes.AttributeKeySpender, fromAddress),
		sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
	))
	// addCoins event
	events = events.AppendEvent(sdk.NewEvent(
		banktypes.EventTypeCoinReceived,
		sdk.NewAttribute(banktypes.AttributeKeyReceiver, toAddress),
		sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
	))

	// SendCoins function
	events = events.AppendEvent(sdk.NewEvent(
		banktypes.EventTypeTransfer,
		sdk.NewAttribute(banktypes.AttributeKeyRecipient, toAddress),
		sdk.NewAttribute(banktypes.AttributeKeySender, fromAddress),
		sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
	))
	events = events.AppendEvent(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(banktypes.AttributeKeySender, fromAddress),
	))

	return events.ToABCIEvents()
}
