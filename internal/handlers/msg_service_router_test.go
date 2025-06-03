package handlers_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	piosimapp "github.com/provenance-io/provenance/app"
	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
)

// TestGasLimit is a test fee gas limit.
const TestGasLimit uint64 = 150000

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

func TestFailedTx(tt *testing.T) {
	pioconfig.SetProvConfig(sdk.DefaultBondDenom)
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit)+1))
	app := piosimapp.SetupWithGenesisAccounts(tt, "flatfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	encCfg := app.GetEncodingConfig()
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "flatfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeModuleAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	gasFeesString := fmt.Sprintf("%v%s", (TestGasLimit), sdk.DefaultBondDenom)

	flatFeesParams := flatfeestypes.Params{
		DefaultCost: sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit)),
		ConversionFactor: flatfeestypes.ConversionFactor{
			BaseAmount:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
			ConvertedAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
		},
	}
	require.NoError(tt, app.FlatFeesKeeper.SetParams(ctx, flatFeesParams), "FlatFeesKeeper.SetParams(%s)", flatFeesParams)

	// Check both account balances before we begin.
	addr1beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
	addr2beforeBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
	assert.Equal(tt, "150001stake", addr1beforeBalance, "addr1beforeBalance")
	assert.Equal(tt, "", addr2beforeBalance, "addr2beforeBalance")
	stopIfFailed(tt)

	tt.Run("no msg-based fee", func(t *testing.T) {
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2)))
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit)))
		txBytes, err := SignTxAndGetBytes(ctx, TestGasLimit, fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		assert.NoError(t, err, "FinalizeBlock expected no error")
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.Equal(t, uint32(0x5), blockRes.TxResults[0].Code, "code 5 insufficient funds error: %s", blockRes.TxResults[0].Log)

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

	// Give acct1 150000stake back (should now have 150001stake again).
	require.NoError(tt, banktestutil.FundAccount(ctx, app.BankKeeper, acct1.GetAddress(),
		sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit)))),
		fmt.Sprintf("funding acct1 with %s", gasFeesString))

	tt.Run("1stake fee over default", func(t *testing.T) {
		// The default cost (150000stake) should be charged first,
		// then the MsgSend will fail because there's only 1stake left in the account.
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2)))
		cost := flatFeesParams.DefaultCost.AddAmount(sdkmath.NewInt(1))
		msgbasedFee := flatfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), cost)
		require.NoError(t, app.FlatFeesKeeper.SetMsgFee(ctx, *msgbasedFee), "setting fee %s", msgbasedFee)
		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		fees := msgbasedFee.Cost
		txBytes, err := SignTxAndGetBytes(ctx, TestGasLimit, fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		assert.NoError(t, err, "FinalizeBlock")
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		t.Logf("Tx Log: %s", blockRes.TxResults[0].Log)
		assert.Equal(t, 5, int(blockRes.TxResults[0].Code), "code 5 insufficient funds error: %s", blockRes.TxResults[0].Log)

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
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeModuleAccount.GetAddress().String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit))))...)

		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})
}

func TestMsgService(tt *testing.T) {
	pioconfig.SetProvConfig(sdk.DefaultBondDenom) // Set denom as stake.
	priv, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 1000), sdk.NewInt64Coin(sdk.DefaultBondDenom, 600_500))
	app := piosimapp.SetupWithGenesisAccounts(tt, "flatfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	encCfg := app.GetEncodingConfig()
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "flatfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeCollectorAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	feeCollectorAddr := feeCollectorAccount.GetAddress().String()
	gasFeesString := fmt.Sprintf("%v%s", (TestGasLimit), sdk.DefaultBondDenom)

	flatFeesParams := flatfeestypes.Params{
		DefaultCost: sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit)),
		ConversionFactor: flatfeestypes.ConversionFactor{
			BaseAmount:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
			ConvertedAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
		},
	}
	require.NoError(tt, app.FlatFeesKeeper.SetParams(ctx, flatFeesParams), "FlatFeesKeeper.SetParams(%s)", flatFeesParams)

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
		txBytes, err := SignTxAndGetBytes(ctx, TestGasLimit, fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
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
		expEvents = append(expEvents, CreateSendCoinEvents(addr1.String(), feeCollectorAddr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit))))...)
		assertEventsContains(t, blockRes.TxResults[0].Events, expEvents)
	})

	tt.Run("800hotdog fee associated with msg type", func(t *testing.T) {
		// Sending 50hotdog with fees of 800hotdog.
		// account 1 will lose 850hotdog.
		// account 2 will gain 50hotdog.
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 50)))
		msgbasedFee := flatfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin("hotdog", 800))
		require.NoError(t, app.FlatFeesKeeper.SetMsgFee(ctx, *msgbasedFee), "setting fee 800hotdog")
		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		fees := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 800))
		txBytes, err := SignTxAndGetBytes(ctx, TestGasLimit, fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
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
		assert.Equal(t, "50hotdog,450500stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "150hotdog", addr2AfterBalance, "addr2AfterBalance")

		eb := testutil.NewEventsBuilder(t).AddEvent(
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFee, fees.String()),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(antewrapper.AttributeKeyMinFeeCharged, ""),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr1.String()),
				sdk.NewAttribute(antewrapper.AttributeKeyBaseFee, ""),
				sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, "800hotdog")),
		)
		// No send coins from the antehandler because the default fee denom is different from the msg fee denom.
		// Fee charged for msg based fee (from post handler).
		eb.AddSendCoinsStrs(addr1.String(), feeCollectorAddr, "800hotdog")

		assertEventsContains(t, blockRes.TxResults[0].Events, eb.BuildABCI())
	})

	tt.Run("10stake,5hotdog fee associated with msg type", func(t *testing.T) {
		// Sending 12hotdog with fees of 11stake,5hotdog (1stake over what's needed).
		// account 1 will lose 11stake,17hotdog.
		// account 2 will gain 12hotdog.
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 12)))
		msgbasedFee := flatfeestypes.NewMsgFee(sdk.MsgTypeURL(msg), sdk.NewInt64Coin(sdk.DefaultBondDenom, 10), sdk.NewInt64Coin("hotdog", 5))
		require.NoError(t, app.FlatFeesKeeper.SetMsgFee(ctx, *msgbasedFee), "setting fee 10stake")

		acct1 = app.AccountKeeper.GetAccount(ctx, acct1.GetAddress()).(*authtypes.BaseAccount)
		fees := msgbasedFee.Cost.Add(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
		txBytes, err := SignTxAndGetBytes(ctx, TestGasLimit, fees, encCfg, priv.PubKey(), priv, *acct1, ctx.ChainID(), msg)
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

		assert.Equal(t, "33hotdog,450489stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "162hotdog", addr2AfterBalance, "addr2AfterBalance")

		eb := testutil.NewEventsBuilder(t).AddEvent(
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFee, fees.String()),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(antewrapper.AttributeKeyMinFeeCharged, "10stake"),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr1.String())),
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr1.String()),
				sdk.NewAttribute(antewrapper.AttributeKeyBaseFee, "10stake"),
				sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, "5hotdog"),
				sdk.NewAttribute(antewrapper.AttributeKeyFeeOverage, "1stake")),
		)
		// Up-front fee charged.
		eb.AddSendCoinsStrs(addr1.String(), feeCollectorAddr, "10stake")
		// On success fee charged.
		eb.AddSendCoinsStrs(addr1.String(), feeCollectorAddr, "5hotdog,1stake")

		assertEventsContains(t, blockRes.TxResults[0].Events, eb.BuildABCI())
	})
}

func TestMsgServiceAuthz(tt *testing.T) {
	pioconfig.SetProvConfig(sdk.DefaultBondDenom)
	priv, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	_, _, addr3 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct2 := authtypes.NewBaseAccount(addr2, priv2.PubKey(), 1, 0)
	acct3 := authtypes.NewBaseAccount(addr3, priv2.PubKey(), 2, 0)
	initBalance := sdk.NewCoins(sdk.NewInt64Coin("hotdog", 10_000), sdk.NewInt64Coin(sdk.DefaultBondDenom, 721_000))
	app := piosimapp.SetupWithGenesisAccounts(tt, "flatfee-testing",
		[]authtypes.GenesisAccount{acct1, acct2, acct3},
		banktypes.Balance{Address: addr1.String(), Coins: initBalance},
		banktypes.Balance{Address: addr2.String(), Coins: initBalance},
	)
	encCfg := app.GetEncodingConfig()
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "flatfee-testing"})
	require.NoError(tt, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")
	feeCollectorAccount := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	feeCollectorAddr := feeCollectorAccount.GetAddress().String()

	flatFeesParams := flatfeestypes.Params{
		DefaultCost: sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit)),
		ConversionFactor: flatfeestypes.ConversionFactor{
			BaseAmount:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
			ConvertedAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
		},
	}
	require.NoError(tt, app.FlatFeesKeeper.SetParams(ctx, flatFeesParams), "FlatFeesKeeper.SetParams(%s)", flatFeesParams)

	// Create an authz grant from addr1 to addr2 for 500hotdog.
	now := ctx.BlockHeader().Time
	exp1Hour := now.Add(time.Hour)
	sendAuth := banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("hotdog", 500)), nil)
	require.NoError(tt, app.AuthzKeeper.SaveGrant(ctx, addr2, addr1, sendAuth, &exp1Hour), "Save Grant addr2 addr1 500hotdog")
	// Set a MsgSend msg-based fee of 800hotdog.
	msgbasedFee := flatfeestypes.NewMsgFee(sdk.MsgTypeURL(&banktypes.MsgSend{}), sdk.NewInt64Coin("hotdog", 800))
	require.NoError(tt, app.FlatFeesKeeper.SetMsgFee(ctx, *msgbasedFee), "setting fee 800hotdog")

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
		gasAmt := TestGasLimit + 20_000
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

		eb := testutil.NewEventsBuilder(t).AddEvent(
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFee, fees.String()),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(antewrapper.AttributeKeyMinFeeCharged, "150000stake"),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr2.String()),
				sdk.NewAttribute(antewrapper.AttributeKeyBaseFee, "150000stake"),
				sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, "800hotdog"),
				sdk.NewAttribute(antewrapper.AttributeKeyFeeOverage, "20000stake")),
		)

		// Up-front fee charged.
		eb.AddSendCoinsStrs(addr2.String(), feeCollectorAddr, "150000stake")
		// On success fee charged.
		eb.AddSendCoinsStrs(addr2.String(), feeCollectorAddr, "800hotdog,20000stake")
		assertEventsContains(t, blockRes.TxResults[0].Events, eb.BuildABCI())
	})

	tt.Run("exec two sends", func(t *testing.T) {
		// send 2 successful authz messages
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 80)))
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg, msg})
		// 150000stake for the MsgExec, and 800hotdog * 2 (for the two MsgSend).
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150000), sdk.NewInt64Coin("hotdog", 1600))
		txBytes, err := SignTxAndGetBytes(ctx, TestGasLimit*2, fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
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
		assert.Equal(t, "7600hotdog,401000stake", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "260hotdog", addr3AfterBalance, "addr3AfterBalance")

		eb := testutil.NewEventsBuilder(t).AddEvent(
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFee, fees.String()),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(antewrapper.AttributeKeyMinFeeCharged, "150000stake"),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr2.String())),
			sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, addr2.String()),
				sdk.NewAttribute(antewrapper.AttributeKeyBaseFee, "150000stake"),
				sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, "1600hotdog")),
		)
		// Up-front fee charged.
		eb.AddSendCoinsStrs(addr2.String(), feeCollectorAddr, "150000stake")
		// On success fee charged.
		eb.AddSendCoinsStrs(addr2.String(), feeCollectorAddr, "1600hotdog")

		assertEventsContains(t, blockRes.TxResults[0].Events, eb.BuildABCI())
	})

	tt.Run("not enough fees", func(t *testing.T) {
		acct2 = app.AccountKeeper.GetAccount(ctx, acct2.GetAddress()).(*authtypes.BaseAccount)
		msg := banktypes.NewMsgSend(addr1, addr3, sdk.NewCoins(sdk.NewInt64Coin("hotdog", 100)))
		msgExec := authztypes.NewMsgExec(addr2, []sdk.Msg{msg})
		fees := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000), sdk.NewInt64Coin("hotdog", 799))
		txBytes, err := SignTxAndGetBytes(ctx, TestGasLimit, fees, encCfg, priv2.PubKey(), priv2, *acct2, ctx.ChainID(), &msgExec)
		require.NoError(t, err, "SignTxAndGetBytes")
		blockRes, err := app.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ctx.BlockHeight() + 1,
				Txs:    [][]byte{txBytes},
			},
		)
		t.Logf("Events:\n%s\n", eventsString(blockRes.TxResults[0].Events, true))
		assert.Equal(t, 13, int(blockRes.TxResults[0].Code), "code 13 insufficient fee")

		// No fee paid. Same balances as previous.
		addr1AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr1).String()
		addr2AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr2).String()
		addr3AfterBalance := app.BankKeeper.GetAllBalances(ctx, addr3).String()
		assert.Equal(t, "9740hotdog,721000stake", addr1AfterBalance, "addr1AfterBalance")
		assert.Equal(t, "7600hotdog,401000stake", addr2AfterBalance, "addr2AfterBalance")
		assert.Equal(t, "260hotdog", addr3AfterBalance, "addr3AfterBalance")
	})
}

// MockFlatFeesKeeper is a mock flat-fees keeper for use in a flat-fee gas meter.
type MockFlatFeesKeeper struct {
	CalculateMsgCostArgs []string
}

func NewMockFlatFeesKeeper() *MockFlatFeesKeeper {
	return &MockFlatFeesKeeper{}
}

func (k *MockFlatFeesKeeper) CalculateMsgCost(ctx sdk.Context, msgs ...sdk.Msg) (upFront sdk.Coins, onSuccess sdk.Coins, err error) {
	for _, msg := range msgs {
		k.CalculateMsgCostArgs = append(k.CalculateMsgCostArgs, sdk.MsgTypeURL(msg))
	}
	return sdk.Coins{sdk.NewInt64Coin("acoin", 5)}, sdk.Coins{sdk.NewInt64Coin("bcoin", 7)}, nil
}

func (k *MockFlatFeesKeeper) ExpandMsgs(msgs []sdk.Msg) ([]sdk.Msg, error) {
	return msgs, nil
}

func TestHandlersConsumeMsgs(t *testing.T) {
	pioconfig.SetProvConfig(sdk.DefaultBondDenom) // Set denom as stake.
	priv, _, addr1 := testdata.KeyTestPubAddr()
	acct1 := authtypes.NewBaseAccount(addr1, priv.PubKey(), 0, 0)
	acct1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000))
	app := piosimapp.SetupWithGenesisAccounts(t, "flatfee-testing",
		[]authtypes.GenesisAccount{acct1},
		banktypes.Balance{Address: addr1.String(), Coins: acct1Balance},
	)
	encCfg := app.GetEncodingConfig()
	ctx := app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "flatfee-testing"})
	require.NoError(t, app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams()), "Setting default account params")

	flatFeesParams := flatfeestypes.Params{
		DefaultCost: sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(TestGasLimit)),
		ConversionFactor: flatfeestypes.ConversionFactor{
			BaseAmount:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
			ConvertedAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
		},
	}
	require.NoError(t, app.FlatFeesKeeper.SetParams(ctx, flatFeesParams), "FlatFeesKeeper.SetParams(%s)", flatFeesParams)

	// expectedExtraMsgsCost is the sum of the two coins returned from the mock CalculateMsgCost method.
	expExtraMsgsCostStr := "5acoin,7bcoin"

	// These are all the Msg types that are registered in the interface registry (as a Msg) but don't have a handler.
	expNoHandlers := []string{
		"/cosmwasm.wasm.v1.MsgIBCCloseChannel",
		"/cosmwasm.wasm.v1.MsgIBCSend",
		"/ibc.applications.interchain_accounts.controller.v1.MsgRegisterInterchainAccount",
		"/ibc.applications.interchain_accounts.controller.v1.MsgSendTx",
		"/ibc.applications.interchain_accounts.controller.v1.MsgUpdateParams",
		"/provenance.metadata.v1.MsgP8eMemorializeContractRequest",
		"/provenance.metadata.v1.MsgWriteP8eContractSpecRequest",
		"/provenance.msgfees.v1.MsgAddMsgFeeProposalRequest",
		"/provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest",
		"/provenance.msgfees.v1.MsgRemoveMsgFeeProposalRequest",
		"/provenance.msgfees.v1.MsgUpdateConversionFeeDenomProposalRequest",
		"/provenance.msgfees.v1.MsgUpdateMsgFeeProposalRequest",
		"/provenance.msgfees.v1.MsgUpdateNhashPerUsdMilProposalRequest",
		"/provenance.oracle.v1.QueryOracleRequest",
		"/provenance.oracle.v1.QueryOracleResponse",
	}
	slices.Sort(expNoHandlers)
	var actNoHandlers []string

	allMsgTypeURLs := encCfg.InterfaceRegistry.ListImplementations("cosmos.base.v1beta1.Msg")
	slices.Sort(allMsgTypeURLs)

	for _, msgTypeURL := range allMsgTypeURLs {
		t.Run(strings.TrimPrefix(msgTypeURL, "/"), func(t *testing.T) {
			handler := app.MsgServiceRouter().HandlerByTypeURL(msgTypeURL)
			// There's a few Msg types that don't have handlers (e.g. for removed endpoints).
			// There's nothing in here to do for those, so just note and skip them.
			if handler == nil {
				actNoHandlers = append(actNoHandlers, msgTypeURL)
				t.Skipf("No handler found for msgTypeURL %q", msgTypeURL)
			}

			// We're trying to test that there's a call to ConsumeMsg in the handler so that
			// we know that any msgs not directly in a tx (e.g. executed by a WASM contract) are recorded
			// That way, we can make sure they're paid for (in the post-handler).
			// When ConsumeMsg is called with a new msg, it adds an entry to the private extraMsgs field.
			// We can't get at that directly, but we can call .Finalize(), which will call the flat-fees
			// keeper's CalculateMsgCost method with each entry of extraMsgs. Further, after .Finalize(),
			// there should be an extra msgs cost in the gas meter too.

			ffk := NewMockFlatFeesKeeper()
			gm := antewrapper.NewFlatFeeGasMeter(storetypes.NewInfiniteGasMeter(), log.NewNopLogger(), ffk)
			ctx = app.BaseApp.NewContextLegacy(false, cmtproto.Header{ChainID: "flatfee-testing"}).WithGasMeter(gm)

			msg, err := encCfg.InterfaceRegistry.Resolve(msgTypeURL)
			require.NoError(t, err, "InterfaceRegistry.Resolve(%q)", msgTypeURL)

			// We don't actually care about the responses from the handler, just that it has been invoked.
			// The msg is a zero-value anyway, so this would probably have an error, and possibly panic.
			// But the call to ConsumeMsg should happen before anything has a chance to break in there.
			_, _ = safeRunHandler(handler, ctx, msg)

			// ConsumeMsg should also add an entry to the private msgTypeURLs field. The MsgCountsString result
			// is based on just the content of that field, so we can use that to check here too.
			msgCountsSt := gm.MsgCountsString()
			assert.Equal(t, msgTypeURL, msgCountsSt, "gm.MsgCountsString()")

			err = gm.Finalize(ctx)
			require.NoError(t, err, "gm.Finalize()")

			extraMsgCost := gm.GetExtraMsgsCost()
			assert.Equal(t, expExtraMsgsCostStr, extraMsgCost.String(), "gm.GetExtraMsgsCost()")
			assert.Equal(t, []string{msgTypeURL}, ffk.CalculateMsgCostArgs, "args provided to CalculateMsgCost")
		})
	}

	t.Run("msgTypeURLs without handlers", func(t *testing.T) {
		assert.Equal(t, expNoHandlers, actNoHandlers, "msgTypeURLs without a handler")
	})
}

// safeRunHandler runs the provided handler with the provided ctx and msg and converts panics to an error.
func safeRunHandler(handler baseapp.MsgServiceHandler, ctx sdk.Context, msg sdk.Msg) (res *sdk.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic (recovered): %v", r)
		}
	}()
	return handler(ctx, msg)
}

func signAndGenTx(
	ctx sdk.Context,
	gaslimit uint64,
	fees sdk.Coins,
	encCfg simappparams.EncodingConfig,
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
	encCfg simappparams.EncodingConfig,
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

	return events.ToABCIEvents()
}
