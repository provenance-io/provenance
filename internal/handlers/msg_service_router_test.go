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
	sdkmath "cosmossdk.io/math"

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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	piosimapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/msgfees/types"
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

func TestRewardsProgramStartError(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin("atom", 1000),
		sdk.NewInt64Coin(NHash, 1_190_500_000),
	)
	app := piosimapp.SetupWithGenesisAccounts(t, "",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	blockTime := ctx.BlockTime()
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(NHash, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)

	txBytes, err := SignTxAndGetBytes(ctx,
		NewTestRewardsGasLimit(),
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000)),
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		&rewardProgram,
	)
	require.NoError(t, err, "SignTxAndGetBytes")
	blockRes, err := app.FinalizeBlock(
		&abci.RequestFinalizeBlock{
			Height: ctx.BlockHeight() + 1,
			Txs:    [][]byte{txBytes},
		},
	)
	assert.Equal(t, uint32(0x1), blockRes.TxResults[0].Code, "claims period error")
	assert.NoError(t, err, "FinalizeBlock")
}

func TestRewardsProgramStart(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	pioconfig.SetProvenanceConfig("nhash", 0)
	priv, _, addr := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin("hotdog", 1000),
		sdk.NewInt64Coin("atom", 1000),
		sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000),
	)
	app := piosimapp.SetupWithGenesisAccounts(t, "",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)

	txReward, err := SignTxAndGetBytes(
		ctx,
		NewTestRewardsGasLimit(),
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000)),
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		&rewardProgram,
	)
	require.NoError(t, err, "SignTxAndGetBytes")

	blockRes, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: ctx.BlockHeight() + 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txReward},
	},
	)
	assert.NoError(t, err, "FinalizeBlock expected no error")

	expEvents := []abci.Event{
		NewEvent(rewardtypes.EventTypeRewardProgramCreated,
			NewAttribute(rewardtypes.AttributeKeyRewardProgramID, "1")),
		NewEvent(sdk.EventTypeMessage,
			NewAttribute(sdk.AttributeKeyAction, sdk.MsgTypeURL(&rewardtypes.MsgCreateRewardProgramRequest{}))),
	}
	assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
}

func TestRewardsProgramStartPerformQualifyingActions(t *testing.T) {
	pioconfig.SetProvenanceConfig("nhash", 0)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin("hotdog", 10000000000),
		sdk.NewInt64Coin("atom", 10000000),
		sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000),
	)
	app := piosimapp.SetupWithGenesisAccounts(t, "",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)
	txReward, err := SignTxAndGetBytes(
		ctx,
		NewTestRewardsGasLimit(),
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000)),
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		&rewardProgram,
	)
	require.NoError(t, err, "SignTxAndGetBytes")

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: ctx.BlockHeight() + 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txReward},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	app.Commit()

	time.Sleep(110 * time.Millisecond)
	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	for height := int64(2); height <= int64(99); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Len(t, claimPeriodDistributions, 1, "claimPeriodDistributions")
		assert.Equal(t, 10, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.NotEqual(t, "0nhash", claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer)
	assert.Equal(t, 98, int(actionCounter), "ActionCounter transfer")
	assert.Equal(t, 10, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress unspecified")
	canCheck := assert.NotNil(t, byAddress, "byAddress unspecified")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState unspecified")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 1, "RewardAccountState unspecified")
		assert.Equal(t, "100nhash", byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim unspecified")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus unspecified")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress unclaimable")
	canCheck = assert.NotNil(t, byAddress, "byAddress unclaimable")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState unclaimable")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 1, "RewardAccountState unclaimable")
		assert.Equal(t, "100nhash", byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim unclaimable")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus unclaimable")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress claimable")
	assert.Empty(t, byAddress.RewardAccountState, "RewardAccountState claimable")
}

func TestRewardsProgramStartPerformQualifyingActionsRecordedRewardsUnclaimable(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 1)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, nil,
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	// get past the reward start time ( test that reward program starts up after 50ms)
	time.Sleep(55 * time.Millisecond)

	for height := int64(1); height < int64(21); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)

		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Len(t, claimPeriodDistributions, 1, "claimPeriodDistributions")
		assert.Equal(t, 10, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 100_000_000_000).String(),
			claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer)
	assert.Equal(t, 20, int(actionCounter), "ActionCounter transfer")
	assert.Equal(t, 10, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress unspecified")
	canCheck := assert.NotNil(t, byAddress, "byAddress unspecified")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState unspecified")
	if canCheck {
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim unspecified")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus unspecified")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress unclaimable")
	canCheck = assert.NotNil(t, byAddress, "byAddress unclaimable")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState unclaimable")
	if canCheck {
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim unclaimable")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus unclaimable")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress claimable")
	if assert.NotNil(t, byAddress, "byAddress claimable") {
		assert.Empty(t, byAddress.RewardAccountState, "RewardAccountState claimable")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress claimed")
	if assert.NotNil(t, byAddress, "byAddress claimed") {
		assert.Empty(t, byAddress.RewardAccountState, "RewardAccountState claimed")
	}
}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountFunded(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 1)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	// fund th =e deterministic rewards account, since genesis won't do that work
	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, nil,
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
		banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(150 * time.Millisecond)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(1); height < int64(6); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)

		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		app.Commit()
		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Equal(t, 1, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000), claimPeriodDistributions[0].RewardsPool, "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer)
	assert.Equal(t, 1, int(actionCounter), "ActionCounter transfer")
	assert.Equal(t, 1, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress unspecified")
	canCheck := assert.NotNil(t, byAddress, "byAddress unspecified")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState unspecified")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 5, "RewardAccountState unspecified")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim unspecified")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus unspecified")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress claimable")
	canCheck = assert.NotNil(t, byAddress, "byAddress claimable")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState claimable")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 4, "RewardAccountState claimable")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim claimable")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus claimable")
	}

	// get the accoutn balances of acct1
	balance := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// claim rewards for the address
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq), "SetSequence(%d)", seq)
	txClaim, errClaim := SignTxAndGetBytes(
		ctx,
		NewTestRewardsGasLimit(),
		fees,
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		msgClaim,
	)
	require.NoError(t, errClaim, "SignTxAndGetBytes")
	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 6,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txClaim},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")

	// unmarshal the TxMsgData
	var protoResult sdk.TxMsgData

	require.NoError(t, app.AppCodec().Unmarshal(res.TxResults[0].Data, &protoResult), "unmarshalling protoResult")
	require.Len(t, protoResult.MsgResponses, 1, "protoResult.MsgResponses")
	require.Equal(t, protoResult.MsgResponses[0].GetTypeUrl(), "/provenance.reward.v1.MsgClaimAllRewardsResponse",
		"protoResult.MsgResponses[0].GetTypeUrl()")
	claimResponse := rewardtypes.MsgClaimAllRewardsResponse{}
	require.NoError(t, claimResponse.Unmarshal(protoResult.MsgResponses[0].Value), "unmarshalling claimResponse")
	assert.Equal(t, sdk.NewInt64Coin("nhash", 50_000_000_000).String(), claimResponse.TotalRewardClaim[0].String(),
		"TotalRewardClaim")
	if assert.Len(t, claimResponse.ClaimDetails, 1, "ClaimDetails") {
		assert.Equal(t, 1, int(claimResponse.ClaimDetails[0].RewardProgramId), "RewardProgramId")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 50_000_000_000).String(),
			claimResponse.ClaimDetails[0].TotalRewardClaim.String(), "ClaimDetails TotalRewardClaim")
		if assert.Len(t, claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails, 5, "ClaimedRewardPeriodDetails") {
			assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
				claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails[0].ClaimPeriodReward.String(), "ClaimPeriodReward")
		}
	}
	balanceLater := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// make sure account balance has the tokens
	balanceChange := balanceLater.AmountOf(pioconfig.GetProvenanceConfig().FeeDenom).Sub(balance.AmountOf(pioconfig.GetProvenanceConfig().FeeDenom))
	assert.Equal(t, sdkmath.NewInt(50_000_000_000).String(), balanceChange.String(), "balance change")
}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountFundedDifferentDenom(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin("hotdog", 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("hotdog", 1000_000_000_000),
		sdk.NewInt64Coin("hotdog", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	// fund th =e deterministic rewards account, since genesis won't do that work
	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, nil,
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
		banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(150 * time.Millisecond)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(1); height < int64(6); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Equal(t, 1, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 10_000_000_000).String(),
			claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCount := int(rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer))
	assert.Equal(t, 1, actionCount, "ActionCounter transfer")
	assert.Equal(t, 1, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress unspecified")
	canCheck := assert.NotNil(t, byAddress, "byAddress unspecified")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState unspecified")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 5, "RewardAccountState unspecified")
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim unspecified")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus unspecified")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress claimable")
	canCheck = assert.NotNil(t, byAddress, "byAddress claimable")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState claimable")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 4, "claimable rewards should be 4 for this address.")
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim claimable")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus claimable")
	}

	// get the accoutn balances of acct1
	balance := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// claim rewards for the address
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq), "SetSequence(%d)", seq)
	txClaim, errClaim := SignTxAndGetBytes(
		ctx,
		NewTestRewardsGasLimit(),
		fees,
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		msgClaim,
	)
	require.NoError(t, errClaim, "SignTxAndGetBytes")
	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 6,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txClaim},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")

	// unmarshal the TxMsgData
	var protoResult sdk.TxMsgData
	require.NoError(t, app.AppCodec().Unmarshal(res.TxResults[0].Data, &protoResult), "unmarshalling protoResult")
	require.Len(t, protoResult.MsgResponses, 1, "protoResult.MsgResponses")
	require.Equal(t, protoResult.MsgResponses[0].GetTypeUrl(), "/provenance.reward.v1.MsgClaimAllRewardsResponse",
		"protoResult.MsgResponses[0].GetTypeUrl()")
	claimResponse := rewardtypes.MsgClaimAllRewardsResponse{}
	require.NoError(t, claimResponse.Unmarshal(protoResult.MsgResponses[0].Value), "unmarshalling claimResponse")
	if assert.NotEmpty(t, claimResponse.TotalRewardClaim, "TotalRewardClaim") {
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 50_000_000_000).String(),
			claimResponse.TotalRewardClaim[0].String(), "TotalRewardClaim")
	}
	if assert.NotEmpty(t, claimResponse.ClaimDetails, "ClaimDetails") {
		assert.Len(t, claimResponse.ClaimDetails, 1, "ClaimDetails")
		assert.Equal(t, 1, int(claimResponse.ClaimDetails[0].RewardProgramId), "RewardProgramId")
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 50_000_000_000).String(),
			claimResponse.ClaimDetails[0].TotalRewardClaim.String(), "ClaimDetails TotalRewardClaim")
		if assert.NotEmpty(t, claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails, "ClaimedRewardPeriodDetails") {
			assert.Len(t, claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails, 5, "ClaimedRewardPeriodDetails")
			assert.Equal(t, sdk.NewInt64Coin("hotdog", 10_000_000_000).String(),
				claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails[0].ClaimPeriodReward.String(), "ClaimPeriodReward")
		}
	}

	app.Commit()
	balanceLater := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	balanceChange := balanceLater.AmountOf("hotdog").Sub(balance.AmountOf("hotdog"))
	// make sure account balance has the tokens
	assert.Equal(t, sdkmath.NewInt(50_000_000_000).String(), balanceChange.String(), "balance change")
}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountFundedDifferentDenomClaimedTogether(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin("atom", 10000000),
		sdk.NewInt64Coin("hotdog", 1000000_000_000_000),
		sdk.NewInt64Coin("nhash", 1000000_000_000_000),
	)

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("hotdog", 1000_000_000_000),
		sdk.NewInt64Coin("hotdog", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
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
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	// fund th =e deterministic rewards account, since genesis won't do that work
	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(3), []rewardtypes.RewardProgram{rewardProgram, secondRewardProgram},
		[]authtypes.GenesisAccount{acct1}, nil,
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
		banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(150 * time.Millisecond)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(1); height < int64(6); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)

		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		app.Commit()

		// wait for claim period to end (claim period is 1s)
		time.Sleep(1500 * time.Millisecond)
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Equal(t, 1, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 10_000_000_000).String(),
			claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer)
	assert.Equal(t, 1, int(actionCounter), "ActionCounter transfer")
	assert.Equal(t, 1, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress unspecified")
	canCheck := assert.NotNil(t, byAddress, "byAddress unspecified")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState unspecified")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 10, "RewardAccountState unspecified")
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim unspecified")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus unspecified")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress claimable")
	canCheck = assert.NotNil(t, byAddress, "byAddress claimable")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState claimable")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 8, "RewardAccountState claimable")
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim claimable")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus claimable")
	}

	// get the accoutn balances of acct1
	balance := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// claim rewards for the address

	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq), "SetSequence(%d)", seq)
	// needs extra gas
	txClaim, errClaim := SignTxAndGetBytes(ctx, 300000, fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msgClaim)
	require.NoError(t, errClaim, "SignTxAndGetBytes")

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 6,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txClaim},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	app.Commit()

	// unmarshal the TxMsgData
	var protoResult sdk.TxMsgData
	require.NoError(t, app.AppCodec().Unmarshal(res.TxResults[0].Data, &protoResult), "unmarshalling protoResult")
	require.Len(t, protoResult.MsgResponses, 1, "protoResult.MsgResponses")
	require.Equal(t, protoResult.MsgResponses[0].GetTypeUrl(), "/provenance.reward.v1.MsgClaimAllRewardsResponse",
		"protoResult.MsgResponses[0].GetTypeUrl()")
	claimResponse := rewardtypes.MsgClaimAllRewardsResponse{}
	require.NoError(t, claimResponse.Unmarshal(protoResult.MsgResponses[0].Value), "unmarshalling claimResponse")
	assert.Equal(t, sdk.NewInt64Coin("hotdog", 50_000_000_000).String(),
		claimResponse.TotalRewardClaim[0].String(), "TotalRewardClaim")
	if assert.NotEmpty(t, claimResponse.ClaimDetails, "ClaimDetails") {
		assert.Len(t, claimResponse.ClaimDetails, 2)

		assert.Equal(t, 1, int(claimResponse.ClaimDetails[0].RewardProgramId), "[0].RewardProgramId")
		assert.Equal(t, sdk.NewInt64Coin("hotdog", 50_000_000_000).String(),
			claimResponse.ClaimDetails[0].TotalRewardClaim.String(), "[0].TotalRewardClaim")
		if assert.NotEmpty(t, claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails, "[0].ClaimedRewardPeriodDetails") {
			assert.Len(t, claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails, 5, "[0].ClaimedRewardPeriodDetails")
			assert.Equal(t, sdk.NewInt64Coin("hotdog", 10_000_000_000).String(),
				claimResponse.ClaimDetails[0].ClaimedRewardPeriodDetails[0].ClaimPeriodReward.String(), "[0].[0].ClaimPeriodReward")
		}

		assert.Equal(t, 2, int(claimResponse.ClaimDetails[1].RewardProgramId), "[1].RewardProgramId")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 50_000_000_000).String(),
			claimResponse.ClaimDetails[1].TotalRewardClaim.String(), "[1].TotalRewardClaim")
		if assert.NotEmpty(t, claimResponse.ClaimDetails[1].ClaimedRewardPeriodDetails, "[1].ClaimedRewardPeriodDetails") {
			assert.Len(t, claimResponse.ClaimDetails[1].ClaimedRewardPeriodDetails, 5, "[1].ClaimedRewardPeriodDetails")
			assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
				claimResponse.ClaimDetails[1].ClaimedRewardPeriodDetails[0].ClaimPeriodReward.String(), "[1].[0].ClaimPeriodReward")
		}
	}

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 7,
		Time:   time.Now().UTC(),
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	app.Commit()

	balanceLater := app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// make sure account balance has the tokens
	balanceChangeHotDog := balanceLater.AmountOf("hotdog").Sub(balance.AmountOf("hotdog"))
	assert.Equal(t, sdkmath.NewInt(50_000_000_000).String(), balanceChangeHotDog.String(), "balance change hotdog")

	balanceLater = app.BankKeeper.GetAllBalances(ctx, acct1.GetAddress())
	// make sure account balance has the tokens
	balanceChangeNHash := balanceLater.AmountOf("nhash").Sub(balance.AmountOf("nhash"))
	assert.Equal(t, sdkmath.NewInt(50_000_000_000).String(), balanceChangeNHash.String(), "balance change nhash")
}

func TestRewardsProgramStartPerformQualifyingActionsSomePeriodsClaimableModuleAccountNotFunded(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	// do not fund the deterministic rewards account
	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, nil,
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(150 * time.Millisecond)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(1); height < int64(6); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)

		time.Sleep(1500 * time.Millisecond)
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Equal(t, 1, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer)
	assert.Equal(t, 1, int(actionCounter), "ActionCounter transfer")
	assert.Equal(t, 1, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress unspecified")
	canCheck := assert.NotNil(t, byAddress, "byAddress unspecified")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState unspecified")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 5, "RewardAccountState unspecified")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim unspecified")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus unspecified")
	}

	byAddress, err = app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress claimable")
	canCheck = assert.NotNil(t, byAddress, "byAddress claimable")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState claimable")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 4, "RewardAccountState claimable")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim claimable")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus claimable")
	}

	// claim rewards for the address
	msgClaim := rewardtypes.NewMsgClaimAllRewardsRequest(acct1.Address)
	require.NoError(t, acct1.SetSequence(seq), "SetSequence(%d)", seq)
	txClaim, errClaim := SignTxAndGetBytes(ctx,
		NewTestRewardsGasLimit(),
		fees,
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		msgClaim,
	)
	require.NoError(t, errClaim, "SignTxAndGetBytes")

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 6,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txClaim},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	app.Commit()
}

// Checks to see if delegation are met for a Qualifying action in this case Transfer
// this tests has transfers from an account which DOES NOT have the minimum delegation
// amount needed to get a share
func TestRewardsProgramStartPerformQualifyingActionsCriteriaNotMet(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
							Amount: sdkmath.NewInt(10_000_000_000)},
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, nil,
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(110 * time.Millisecond)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(1); height < int64(6); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		time.Sleep(1100 * time.Millisecond)

		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")

		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Equal(t, 0, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer)
	assert.Equal(t, 0, int(actionCounter), "ActionCounter transfer")
	assert.Equal(t, 0, int(accountState.SharesEarned), "SharesEarned")
}

// Checks to see if delegation are met for a Qualifying action in this case, Transfer, create an address with delegations
// transfers which map to QualifyingAction map to the delegated address
// delegation threshold is met
func TestRewardsProgramStartPerformQualifyingActionsTransferAndDelegationsPresent(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
							Amount: sdkmath.NewInt(100)},
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2),
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	// wait for program to start
	time.Sleep(150 * time.Millisecond)

	//go through 5 blocks, but take a time to cut blocks > claim period time interval.
	for height := int64(1); height < int64(6); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		time.Sleep(1100 * time.Millisecond)
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Equal(t, 1, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer)
	assert.Equal(t, 1, int(actionCounter), "ActionCounter transfer")
	assert.Equal(t, 1, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress")
	canCheck := assert.NotNil(t, byAddress, "byAddress unspecified")
	canCheck = canCheck && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState")
	if canCheck {
		assert.Len(t, byAddress.RewardAccountState, 5, "RewardAccountState")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus")
	}
}

// Checks to see if delegation are met for a Qualifying action in this case Transfer, create an address with delegations
// transfers which map to QualifyingAction map to the delegated address
// delegation threshold is *not* met
func TestRewardsProgramStartPerformQualifyingActionsThreshHoldNotMet(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
							Amount: sdkmath.NewInt(100_000_000)},
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2),
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")

	// tx with a fee associated with msg type and account has funds
	msg := banktypes.NewMsgSend(addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 50)))
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(1 * time.Second)

	//go through 5 blocks, but take a long time to cut blocks.
	for height := int64(1); height < int64(6); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		time.Sleep(1100 * time.Millisecond)
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Equal(t, 0, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, true, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeTransfer)
	assert.Equal(t, 0, int(actionCounter), "ActionCounter transfer")
	assert.Equal(t, 0, int(accountState.SharesEarned), "SharesEarned")
}

func TestRewardsProgramStartPerformQualifyingActions_Vote(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, _, addr := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000),
		sdk.NewInt64Coin("atom", 10000000),
		sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000),
	)

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 0),
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, nil,
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
		coinsPos,
		addr,
	)
	require.NoError(t, err, "NewMsgSubmitProposal expected to create msg")
	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence

	txGov, err := SignTxAndGetBytes(
		ctx,
		NewTestRewardsGasLimit(),
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000)),
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		msg,
	)
	require.NoError(t, err, "SignTxAndGetBytes")
	time.Sleep(200 * time.Millisecond)

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txGov},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	require.Len(t, res.TxResults, 1, "TxResults expected length not met")
	require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)
	app.Commit()

	seq = seq + 1
	proposal := getLastProposal(t, ctx, app)

	// tx with a fee associated with msg type and account has funds
	vote1 := govtypesv1beta1.NewMsgVote(addr, proposal.Id, govtypesv1beta1.OptionYes)

	for height := int64(2); height < int64(22); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		require.Len(t, res.TxResults, 1, "TxResults expected length not met")
		require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Len(t, claimPeriodDistributions, 1, "claimPeriodDistributions")
		assert.Equal(t, 10, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.NotEqual(t, "0nhash", claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeVote)
	assert.Equal(t, 20, int(actionCounter), "ActionCounter vote")
	assert.Equal(t, 10, int(accountState.SharesEarned), "SharesEarned")
}

func TestRewardsProgramStartPerformQualifyingActions_Vote_InvalidDelegations(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv1, pubKey1, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv1.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acctBalance := sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000),
		sdk.NewInt64Coin("atom", 10000000),
		sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000),
	)

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 1000),
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
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct2.GetAddress(), fundCoins),
		"funding acct2 with 290500010nhash")
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
		coinsPos,
		addr1,
	)
	require.NoError(t, err, "NewMsgSubmitProposal expected to create msg")

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	time.Sleep(200 * time.Millisecond)

	txGov, err := SignTxAndGetBytes(
		ctx,
		NewTestRewardsGasLimit(),
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000)),
		encCfg, priv1.PubKey(), priv1, *acct1, ctx.ChainID(),
		msg,
	)
	require.NoError(t, err, "SignTxAndGetBytes")

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txGov},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	require.Len(t, res.TxResults, 1, "TxResults expected length not met")
	require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)

	app.Commit()

	proposal := getLastProposal(t, ctx, app)

	// tx with a fee associated with msg type and account has funds
	vote2 := govtypesv1beta1.NewMsgVote(addr2, proposal.Id, govtypesv1beta1.OptionYes)
	acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
	seq := acct2.Sequence

	for height := int64(2); height < int64(4); height++ {
		require.NoError(t, acct2.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), vote2)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		require.Len(t, res.TxResults, 1, "TxResults expected length not met")
		require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)

		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Len(t, claimPeriodDistributions, 1, "claimPeriodDistributions")
		assert.Equal(t, 0, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.Equal(t, "100000000000nhash", claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct2.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeVote)
	assert.Equal(t, 0, int(actionCounter), "ActionCounter vote")
	assert.Equal(t, 0, int(accountState.SharesEarned), "SharesEarned")

	byAddress1, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress acct1")
	assert.Empty(t, byAddress1.RewardAccountState, "RewardAccountState acct1")

	byAddress2, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct2.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress acct2")
	assert.Empty(t, byAddress2.RewardAccountState, "RewardAccountState acct2")
}

func TestRewardsProgramStartPerformQualifyingActions_Vote_ValidDelegations(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, _ := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000),
		sdk.NewInt64Coin("atom", 10000000),
		sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000),
	)

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 1000),
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2),
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)
	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
		coinsPos,
		addr,
	)
	require.NoError(t, err, "NewMsgSubmitProposal expected to create msg")

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(200 * time.Millisecond)

	txGov, err := SignTxAndGetBytes(
		ctx,
		NewTestRewardsGasLimit(),
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000)),
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		msg,
	)
	require.NoError(t, err, "SignTxAndGetBytes")

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txGov},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	require.Len(t, res.TxResults, 1, "TxResults expected length not met")
	require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)

	app.Commit()

	seq = seq + 1
	proposal := getLastProposal(t, ctx, app)

	// tx with a fee associated with msg type and account has funds
	vote1 := govtypesv1beta1.NewMsgVote(addr, proposal.Id, govtypesv1beta1.OptionYes)

	// threshold will be met after 10 actions
	for height := int64(2); height < int64(22); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		require.Len(t, res.TxResults, 1, "TxResults expected length not met")
		require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)

		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Len(t, claimPeriodDistributions, 1, "claimPeriodDistributions")
		assert.Equal(t, 10, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.NotEqual(t, "0nhash", claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeVote)
	assert.Equal(t, 20, int(actionCounter), "ActionCounter vote")
	assert.Equal(t, 10, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress")
	if assert.NotNil(t, byAddress, "byAddress") && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState") {
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus")
	}
}

// checks to see that votes are coming from an address that has delegated enough coins and is a validator, and gets the multiplier applied
func TestRewardsProgramStartPerformQualifyingActions_Vote_ValidDelegations_Multiplier_Present(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, _ := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin("nhash", 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 1000),
						ValidatorMultiplier:     10,
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING
	//err, bondedVal1, bondedVal2 := createTestValidators(pubKey, pubKey2, addr, addr2)
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2), banktypes.Balance{Address: addr.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false)
	ctx.WithBlockTime(time.Now())
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 290_500_010))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
		coinsPos,
		addr,
	)
	require.NoError(t, err, "NewMsgSubmitProposal expected to create msg")

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(200 * time.Millisecond)

	txGov, err := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin("nhash", 1_190_500_000)), encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
	require.NoError(t, err, "SignTxAndGetBytes")

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txGov},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	require.Len(t, res.TxResults, 1, "TxResults expected length not met")
	require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)

	app.Commit()

	seq = seq + 1
	proposal := getLastProposal(t, ctx, app)

	// tx with a fee associated with msg type and account has funds
	vote1 := govtypesv1beta1.NewMsgVote(addr, proposal.Id, govtypesv1beta1.OptionYes)

	assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")

	// threshold will be met after 10 actions
	for height := int64(2); height < int64(22); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		require.Len(t, res.TxResults, 1, "TxResults expected length not met")
		require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(100), claimPeriodDistributions[0].TotalShares, "shares should have accumulated to value of 100 ( 10 action * multiplier (10) shares")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended.")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.Equal(sdk.Coin{
		Denom:  "nhash",
		Amount: sdkmath.ZeroInt(),
	}), "claim period has not ended so rewards still haven't been calculated(hence 0 coins)")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(20), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionVote"), "account state incorrect")
	assert.Equal(t, 100, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct1.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
}

// checks to see that votes are coming from an address that has delegated enough coins but is not a validator, hence does not get the multiplier applied
func TestRewardsProgramStartPerformQualifyingActions_Vote_ValidDelegations_Multiplier_Present_But_NotValidatorVotes(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, _ := testdata.KeyTestPubAddr()
	// an address which is not a validator
	priv3, _, addr3 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	acct3 := authtypes.NewBaseAccount(addr3, priv3.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000), sdk.NewInt64Coin("atom", 10000000), sdk.NewInt64Coin("nhash", 1000000_000_000_000))

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 1000),
						ValidatorMultiplier:     10,
					},
				},
			},
		},
	)

	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING
	app := piosimapp.SetupWithGenesisRewardsProgram(t, uint64(2), []rewardtypes.RewardProgram{rewardProgram}, []authtypes.GenesisAccount{acct3}, createValSet(t, pubKey, pubKey2), banktypes.Balance{Address: addr.String(), Coins: acct1Balance}, banktypes.Balance{Address: addr3.String(), Coins: acct1Balance})

	ctx := app.BaseApp.NewContext(false)
	ctx.WithBlockTime(time.Now())
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 290_500_010))))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct3.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("nhash", 290_500_010))))
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
		coinsPos,
		addr3,
	)
	require.NoError(t, err, "NewMsgSubmitProposal expected to create msg")

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	ctx.WithBlockTime(time.Now())
	acct3 = app.AccountKeeper.GetAccount(ctx, acct3.GetAddress()).(*authtypes.BaseAccount)
	seq := acct3.Sequence
	ctx.WithBlockTime(time.Now())
	time.Sleep(200 * time.Millisecond)

	txGov, err := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin("nhash", 1_190_500_000)), encCfg, priv3.PubKey(), priv3, *acct3, ctx.ChainID(), msg)
	require.NoError(t, err, "SignTxAndGetBytes")
	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txGov},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	require.Len(t, res.TxResults, 1, "TxResults expected length not met")
	require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)

	app.Commit()

	seq = seq + 1
	proposal := getLastProposal(t, ctx, app)

	// vote with an account which is not a validator
	vote := govtypesv1beta1.NewMsgVote(addr3, proposal.Id, govtypesv1beta1.OptionYes)

	assert.NotEmpty(t, res.GetEvents(), "should have emitted an event.")

	// threshold will be met after 10 actions
	for height := int64(2); height < int64(22); height++ {
		require.NoError(t, acct3.SetSequence(seq))
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv3.PubKey(), priv3, *acct3, ctx.ChainID(), vote)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		require.Len(t, res.TxResults, 1, "TxResults expected length not met")
		require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)
		app.Commit()
		seq = seq + 1
	}
	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err)
	assert.Len(t, claimPeriodDistributions, 1, "claim period reward distributions should exist")
	assert.Equal(t, int64(10), claimPeriodDistributions[0].TotalShares, "shares should have accumulated to value of 10, ( 10 action leading to 1 share each) (no multiplier is applied) ")
	assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "claim period has not ended.")
	assert.Equal(t, false, claimPeriodDistributions[0].RewardsPool.Equal(sdk.Coin{
		Denom:  "nhash",
		Amount: sdkmath.ZeroInt(),
	}), "claim period has not ended so rewards still haven't been calculated(hence 0 coins)")

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct3.Address)
	require.NoError(t, err)
	assert.Equal(t, uint64(20), rewardtypes.GetActionCount(accountState.ActionCounter, "ActionVote"), "account state incorrect")
	assert.Equal(t, 10, int(accountState.SharesEarned), "account state incorrect")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx), &rewardtypes.QueryRewardDistributionsByAddressRequest{
		Address:     acct3.Address,
		ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(), byAddress.RewardAccountState[0].TotalRewardClaim.String(), "RewardDistributionsByAddress incorrect")
	assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, byAddress.RewardAccountState[0].ClaimStatus, "claim status incorrect")
}

func TestRewardsProgramStartPerformQualifyingActions_Delegate_NoQualifyingActions(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, _ := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000),
		sdk.NewInt64Coin("atom", 10000000),
		sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000),
	)

	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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

	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, createValSet(t, pubKey, pubKey2),
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)

	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
		coinsPos,
		addr,
	)
	require.NoError(t, err, "NewMsgSubmitProposal expected to create msg")

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(110 * time.Millisecond)

	txGov, err := SignTxAndGetBytes(
		ctx,
		NewTestRewardsGasLimit(),
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000)),
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		msg,
	)
	require.NoError(t, err, "SignTxAndGetBytes")

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txGov},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	require.Len(t, res.TxResults, 1, "TxResults expected length not met")
	require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)

	app.Commit()

	seq = seq + 1
	proposal := getLastProposal(t, ctx, app)

	// tx with a fee associated with msg type and account has funds
	vote1 := govtypesv1beta1.NewMsgVote(addr, proposal.Id, govtypesv1beta1.OptionYes)

	for height := int64(2); height < int64(14); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), vote1)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		require.Len(t, res.TxResults, 1, "TxResults expected length not met")
		require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)
		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Len(t, claimPeriodDistributions, 1, "claimPeriodDistributions")
		assert.Equal(t, 0, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.NotEqual(t, "0nhash", claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeVote)
	assert.Equal(t, 0, int(actionCounter), "ActionCounter vote")
	assert.Equal(t, 0, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress")
	if assert.NotNil(t, byAddress, "byAddress") {
		assert.Empty(t, byAddress.RewardAccountState, "RewardAccountState")
	}
}

func TestRewardsProgramStartPerformQualifyingActions_Delegate_QualifyingActionsPresent(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	priv, pubKey, addr := testdata.KeyTestPubAddr()
	_, pubKey2, _ := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	pioconfig.SetProvenanceConfig("nhash", 0)
	acct1Balance := sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000000),
		sdk.NewInt64Coin("atom", 10000000),
		sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1000000_000_000_000),
	)
	minDelegation := sdk.NewInt64Coin("nhash", 4)
	maxDelegation := sdk.NewInt64Coin("nhash", 2001000)
	rewardProgram := rewardtypes.NewRewardProgram(
		"title",
		"description",
		1,
		acct1.Address,
		sdk.NewInt64Coin("nhash", 1000_000_000_000),
		sdk.NewInt64Coin("nhash", 10_000_000_000),
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
						MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(100, 0),
					},
				},
			},
		},
	)
	rewardProgram.State = rewardtypes.RewardProgram_STATE_PENDING

	valSet := createValSet(t, pubKey, pubKey2)
	app := piosimapp.SetupWithGenesisRewardsProgram(t,
		uint64(2), []rewardtypes.RewardProgram{rewardProgram},
		[]authtypes.GenesisAccount{acct1}, valSet,
		banktypes.Balance{Address: addr.String(), Coins: acct1Balance},
	)

	ctx := app.BaseApp.NewContext(false)
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	fundCoins := sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 290_500_010))
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(), fundCoins),
		"funding acct1 with 290500010nhash")
	coinsPos := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000))
	msg, err := govtypesv1beta1.NewMsgSubmitProposal(
		ContentFromProposalType("title", "description", govtypesv1beta1.ProposalTypeText),
		coinsPos,
		addr,
	)
	require.NoError(t, err, "NewMsgSubmitProposal expected to create msg")

	fees := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
	acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
	seq := acct1.Sequence
	time.Sleep(200 * time.Millisecond)

	txGov, err := SignTxAndGetBytes(ctx,
		NewTestRewardsGasLimit(),
		sdk.NewCoins(sdk.NewInt64Coin("atom", 150), sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_190_500_000)),
		encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(),
		msg,
	)
	require.NoError(t, err, "SignTxAndGetBytes")

	res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   time.Now().UTC(),
		Txs:    [][]byte{txGov},
	},
	)
	require.NoError(t, err, "FinalizeBlock expected no error")
	require.Len(t, res.TxResults, 1, "TxResults expected length not met")
	require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)

	app.Commit()

	seq = seq + 1
	_ = getLastProposal(t, ctx, app) // Just making sure it exists.

	// tx with a fee associated with msg type and account has funds
	delAddr, _ := valSet.GetByIndex(0)
	delegation := stakingtypes.NewMsgDelegate(addr.String(), sdk.ValAddress(delAddr).String(), sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))

	for height := int64(2); height < int64(22); height++ {
		require.NoError(t, acct1.SetSequence(seq), "[%d]: SetSequence(%d)", height, seq)
		tx1, err1 := SignTxAndGetBytes(ctx, NewTestRewardsGasLimit(), fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), delegation)
		require.NoError(t, err1, "[%d]: SignTxAndGetBytes", height)
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: height,
			Time:   time.Now().UTC(),
			Txs:    [][]byte{tx1},
		},
		)
		require.NoError(t, err, "FinalizeBlock expected no error")
		require.Len(t, res.TxResults, 1, "TxResults expected length not met")
		require.Equal(t, uint32(0), res.TxResults[0].Code, "TxResults tx unexpected result code "+res.TxResults[0].Log)
		time.Sleep(100 * time.Millisecond)

		app.Commit()
		seq = seq + 1
	}

	claimPeriodDistributions, err := app.RewardKeeper.GetAllClaimPeriodRewardDistributions(ctx)
	require.NoError(t, err, "GetAllClaimPeriodRewardDistributions")
	if assert.NotEmpty(t, claimPeriodDistributions, "claimPeriodDistributions") {
		assert.Len(t, claimPeriodDistributions, 1, "claimPeriodDistributions")
		assert.Equal(t, 10, int(claimPeriodDistributions[0].TotalShares), "TotalShares")
		assert.Equal(t, false, claimPeriodDistributions[0].ClaimPeriodEnded, "ClaimPeriodEnded")
		assert.NotEqual(t, "0nhash", claimPeriodDistributions[0].RewardsPool.String(), "RewardsPool")
	}

	accountState, err := app.RewardKeeper.GetRewardAccountState(ctx, uint64(1), uint64(1), acct1.Address)
	require.NoError(t, err, "GetRewardAccountState")
	actionCounter := rewardtypes.GetActionCount(accountState.ActionCounter, rewardtypes.ActionTypeDelegate)
	assert.Equal(t, 20, int(actionCounter), "ActionCounter delegate")
	assert.Equal(t, 10, int(accountState.SharesEarned), "SharesEarned")

	byAddress, err := app.RewardKeeper.RewardDistributionsByAddress(sdk.WrapSDKContext(ctx),
		&rewardtypes.QueryRewardDistributionsByAddressRequest{
			Address:     acct1.Address,
			ClaimStatus: rewardtypes.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		},
	)
	require.NoError(t, err, "RewardDistributionsByAddress")
	if assert.NotNil(t, byAddress, "byAddress") && assert.NotEmpty(t, byAddress.RewardAccountState, "RewardAccountState") {
		assert.Equal(t, sdk.NewInt64Coin("nhash", 10_000_000_000).String(),
			byAddress.RewardAccountState[0].TotalRewardClaim.String(), "TotalRewardClaim")
		assert.Equal(t, rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE,
			byAddress.RewardAccountState[0].ClaimStatus, "ClaimStatus")
	}
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

func NewTestRewardsGasLimit() uint64 {
	return 200000
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
