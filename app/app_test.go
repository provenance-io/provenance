package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil/assertions"
	markermodule "github.com/provenance-io/provenance/x/marker"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	opts := SetupOptions{
		Logger:  log.NewTestLogger(t),
		DB:      dbm.NewMemDB(),
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	}
	app := NewAppWithCustomOptions(t, false, opts)

	for acc := range maccPerms {
		require.True(
			t,
			app.BankKeeper.BlockedAddr(app.AccountKeeper.GetModuleAddress(acc)),
			"ensure that blocked addresses are properly set in bank keeper",
		)
	}

	// finalize block so we have CheckTx state set
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)

	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := New(log.NewTestLogger(t), opts.DB, nil, true, opts.AppOpts)
	require.NotPanics(t, func() {
		_, err = app2.ExportAppStateAndValidators(false, nil, nil)
	}, "exporting app state at current height")
	require.NoError(t, err, "ExportAppStateAndValidators at current height")

	require.NotPanics(t, func() {
		_, err = app2.ExportAppStateAndValidators(true, nil, nil)
	}, "exporting app state at zero height")
	require.NoError(t, err, "ExportAppStateAndValidators at zero height")
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}

func TestExportAppStateAndValidators(t *testing.T) {
	opts := SetupOptions{
		Logger:  log.NewTestLogger(t),
		DB:      dbm.NewMemDB(),
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	}
	app := NewAppWithCustomOptions(t, false, opts)
	ctx := app.BaseApp.NewContext(false)

	initAccts := app.AccountKeeper.GetAllAccounts(ctx)
	initAddrs := make([]sdk.AccAddress, len(initAccts))
	for i, acct := range initAccts {
		initAddrs[i] = acct.GetAddress()
	}

	// Create a few accounts
	addrs1 := AddTestAddrs(app, ctx, 10, sdkmath.NewInt(5_000))
	require.Len(t, addrs1, 10, "addrs1")

	managerAddr := addrs1[0]
	managerAllAccess := []markertypes.AccessGrant{{
		Address: managerAddr.String(),
		Permissions: markertypes.AccessList{
			markertypes.Access_Mint, markertypes.Access_Burn, markertypes.Access_Deposit,
			markertypes.Access_Withdraw, markertypes.Access_Delete, markertypes.Access_Admin,
			markertypes.Access_Transfer,
		},
	}}

	markerToAddr := map[string]sdk.AccAddress{}
	// Create some markers.
	for _, denom := range []string{"marker1coin", "marker2coin", "marker3coin"} {
		markerAddr := markertypes.MustGetMarkerAddress(denom)
		require.NoErrorf(t, app.MarkerKeeper.AddMarkerAccount(ctx, &markertypes.MarkerAccount{
			BaseAccount:            authtypes.NewBaseAccount(markerAddr, nil, 0, 0),
			Manager:                managerAddr.String(),
			AccessControl:          managerAllAccess,
			Status:                 markertypes.StatusProposed,
			Denom:                  denom,
			Supply:                 sdkmath.NewInt(1_000_000),
			MarkerType:             markertypes.MarkerType_RestrictedCoin,
			SupplyFixed:            true,
			AllowGovernanceControl: true,
			AllowForcedTransfer:    false,
			RequiredAttributes:     []string{},
		}), "adding %q account", denom)
		markerToAddr[denom] = markerAddr
	}
	require.Len(t, markerToAddr, 3, "markerToAddr")

	// Create some more accounts:
	addrs2 := AddTestAddrs(app, ctx, 10, sdkmath.NewInt(5_000))
	require.Len(t, addrs2, 10, "addrs2")

	// Delete one of the markers.
	require.NoError(t, app.MarkerKeeper.CancelMarker(ctx, managerAddr, "marker2coin"), "canceling marker2coin")
	require.NoError(t, app.MarkerKeeper.DeleteMarker(ctx, managerAddr, "marker2coin"), "deleting marker2coin")
	markermodule.BeginBlocker(ctx, app.MarkerKeeper, app.BankKeeper)
	deletedMarkerAddr := markerToAddr["marker2coin"]

	markerAddrs := []sdk.AccAddress{markerToAddr["marker1coin"], markerToAddr["marker3coin"]}

	logAddrs(t, initAddrs, "initAddrs")
	logAddrs(t, addrs1, "addrs1")
	logAddrs(t, addrs2, "addrs2")
	logAddrs(t, markerAddrs, "markerAddrs")
	t.Logf("deleted marker: %s", deletedMarkerAddr)

	allAccounts := app.AccountKeeper.GetAllAccounts(ctx)
	logAccounts(t, allAccounts, "allAccounts")

	// finalize block so we have CheckTx state set
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)

	app.Commit()

	// Get an export
	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err, "ExportAppStateAndValidators")

	var genState map[string]json.RawMessage
	err = json.Unmarshal(exported.AppState, &genState)
	require.NoError(t, err, "unmarshalling exported app state")

	var authGenState authtypes.GenesisState
	require.NoError(t, app.appCodec.UnmarshalJSON(genState[authtypes.ModuleName], &authGenState), "unmarshalling auth gen state")
	genAccounts := make([]sdk.AccountI, len(authGenState.Accounts))
	for i, acctAny := range authGenState.Accounts {
		switch acctAny.GetTypeUrl() {
		case "/cosmos.auth.v1beta1.ModuleAccount":
			acct, ok := acctAny.GetCachedValue().(*authtypes.ModuleAccount)
			if assert.Truef(t, ok, "casting %T to ModuleAccount", acctAny) {
				genAccounts[i] = acct
			}
		case "/cosmos.auth.v1beta1.BaseAccount":
			acct, ok := acctAny.GetCachedValue().(*authtypes.BaseAccount)
			if assert.Truef(t, ok, "casting %T to BaseAccount", acctAny) {
				genAccounts[i] = acct
			}
		default:
			acct, ok := acctAny.GetCachedValue().(sdk.AccountI)
			if assert.Truef(t, ok, "casting entry %d to AccountI", i) {
				genAccounts[i] = acct
			}
		}
	}
	logAccounts(t, genAccounts, "genAccounts")
	require.False(t, t.Failed(), "failed to convert one ore more genesis accounts from any to a known account type")

	var markerGenState markertypes.GenesisState
	require.NoError(t, app.appCodec.UnmarshalJSON(genState[markertypes.ModuleName], &markerGenState), "unmarshalling marker gen state")

	for i, marker := range markerGenState.Markers {
		t.Logf("genesis marker [%d]: \"%d\" - %s - %s - %s", i, marker.GetAccountNumber(), marker.Denom, marker.Status, marker.Address)
	}

	t.Run("same accounts from keeper and genesis", func(t *testing.T) {
		assert.Equal(t, len(allAccounts), len(genAccounts), "number of accounts: from keeper vs from genesis")
		// The marker accounts are put into genesis as base accounts, so they'd fail if we directly compared elements.
		// So just get all the base accounts from them and compare those lists.
		allBaseAccounts := toBaseAccounts(t, allAccounts, "allAccounts")
		genBaseAccounts := toBaseAccounts(t, genAccounts, "allAccounts")
		assert.ElementsMatch(t, allBaseAccounts, genBaseAccounts, "accounts: from keeper vs from genesis")
	})

	t.Run("initial addresses are present", func(t *testing.T) {
		for _, addr := range initAddrs {
			assertAddrInAccounts(t, addr, "initAddrs", allAccounts, "allAccounts")
			assertAddrInAccounts(t, addr, "initAddrs", genAccounts, "genAccounts")
		}
	})
	t.Run("first set of added addresses are present", func(t *testing.T) {
		for _, addr := range addrs1 {
			assertAddrInAccounts(t, addr, "addrs1 ", allAccounts, "allAccounts")
			assertAddrInAccounts(t, addr, "addrs1", genAccounts, "genAccounts")
		}
	})
	t.Run("markers addresses are present", func(t *testing.T) {
		for _, addr := range markerAddrs {
			assertAddrInAccounts(t, addr, "markerAddrs", allAccounts, "allAccounts")
			assertAddrInAccounts(t, addr, "markerAddrs", genAccounts, "genAccounts")
		}
		assertAddrNotInAccounts(t, deletedMarkerAddr, "deletedMarkerAddr", allAccounts, "allAccounts")
		assertAddrNotInAccounts(t, deletedMarkerAddr, "deletedMarkerAddr", genAccounts, "genAccounts")
	})
	t.Run("seconds set of added addresses are present", func(t *testing.T) {
		for _, addr := range addrs2 {
			assertAddrInAccounts(t, addr, "addrs2", allAccounts, "allAccounts")
			assertAddrInAccounts(t, addr, "addrs2", genAccounts, "genAccounts")
		}
	})
	t.Run("no duplicate account numbers", func(t *testing.T) {
		assertNoDupeAccountNumbers(t, ctx, app, allAccounts, "allAccounts")
		assertNoDupeAccountNumbers(t, ctx, app, genAccounts, "genAccounts")
	})
}

func logAccounts(t *testing.T, accts []sdk.AccountI, name string) {
	t.Helper()
	for i, acctI := range accts {
		switch acct := acctI.(type) {
		case *authtypes.ModuleAccount:
			t.Logf("%s[%d]: \"%d\" - Module:%s - %s", name, i, acct.GetAccountNumber(), acct.Name, acct.Address)
		case *markertypes.MarkerAccount:
			t.Logf("%s[%d]: \"%d\" - Marker:%s - %s", name, i, acct.GetAccountNumber(), acct.Denom, acct.Address)
		case *authtypes.BaseAccount:
			t.Logf("%s[%d]: \"%d\" - Base - %s", name, i, acct.GetAccountNumber(), acct.Address)
		default:
			t.Logf("%s[%d]: \"%d\" - Unknown - %s", name, i, acctI.GetAccountNumber(), acctI.GetAddress().String())
		}
	}
}

func toBaseAccounts(t *testing.T, acctsI []sdk.AccountI, name string) []*authtypes.BaseAccount {
	t.Helper()
	rv := make([]*authtypes.BaseAccount, len(acctsI))
	for i, acctI := range acctsI {
		rv[i] = toBaseAccount(t, i, acctI, name)
	}
	return rv
}

func toBaseAccount(t *testing.T, i int, acctI sdk.AccountI, name string) *authtypes.BaseAccount {
	t.Helper()
	switch acct := acctI.(type) {
	case *authtypes.ModuleAccount:
		return acct.BaseAccount
	case *markertypes.MarkerAccount:
		return acct.BaseAccount
	case *authtypes.BaseAccount:
		return acct
	default:
		t.Logf("unknown account type: %s[%d]: %d %s %T", name, i, acctI.GetAccountNumber(), acctI.GetAddress(), acctI)
		return &authtypes.BaseAccount{
			Address:       acctI.GetAddress().String(),
			PubKey:        sdktypes.UnsafePackAny(acctI.GetPubKey()),
			AccountNumber: acctI.GetAccountNumber(),
			Sequence:      acctI.GetSequence(),
		}
	}
}

func logAddrs(t *testing.T, addrs []sdk.AccAddress, name string) {
	t.Helper()
	for i, addr := range addrs {
		t.Logf("%s[%d]: %s", name, i, addr)
	}
}

func assertNoDupeAccountNumbers(t *testing.T, _ sdk.Context, _ *App, accts []sdk.AccountI, name string) bool {
	t.Helper()
	byAcctNum := map[uint64][]sdk.AccountI{}
	acctNums := []uint64{}
	for _, acct := range accts {
		acctNum := acct.GetAccountNumber()
		byAcctNum[acctNum] = append(byAcctNum[acctNum], acct)
		acctNums = append(acctNums, acctNum)
	}
	sort.Slice(acctNums, func(i, j int) bool {
		return acctNums[i] < acctNums[j]
	})
	rv := true
	for i, acctNum := range acctNums {
		if i > 0 && acctNums[i-1] == acctNum {
			continue
		}
		if !assert.Equalf(t, len(byAcctNum[acctNum]), 1, "%s entries with Account Number %d", name, acctNum) {
			rv = false
			logAccounts(t, byAcctNum[acctNum], fmt.Sprintf("byAcctNum[%d]", acctNum))
		}
	}
	return rv
}

func assertAddrInAccounts(t *testing.T, addr sdk.AccAddress, addrName string, accts []sdk.AccountI, acctsName string) bool {
	t.Helper()
	for _, acct := range accts {
		if addr.Equals(acct.GetAddress()) {
			return true
		}
	}
	return assert.Fail(t, fmt.Sprintf("%s address not found in %s", addrName, acctsName), addr.String())
}

func assertAddrNotInAccounts(t *testing.T, addr sdk.AccAddress, addrName string, accts []sdk.AccountI, acctsName string) bool {
	t.Helper()
	for _, acct := range accts {
		if addr.Equals(acct.GetAddress()) {
			return assert.Fail(t, fmt.Sprintf("%s address found in %s", addrName, acctsName), addr.String())
		}
	}
	return true
}

func createEvent(eventType string, attributes []abci.EventAttribute) abci.Event {
	return abci.Event{
		Type:       eventType,
		Attributes: attributes,
	}
}

func TestShouldFilterEvent(t *testing.T) {
	tests := []struct {
		name   string
		event  abci.Event
		expect bool
	}{
		{"Empty commission event", createEvent(distrtypes.EventTypeCommission, []abci.EventAttribute{{Key: "amount", Value: ""}}), true},
		{"Non-empty commission event", createEvent(distrtypes.EventTypeCommission, []abci.EventAttribute{{Key: "amount", Value: "100"}}), false},

		{"Empty rewards event", createEvent(distrtypes.EventTypeRewards, []abci.EventAttribute{{Key: "amount", Value: ""}}), true},
		{"Non-empty rewards event", createEvent(distrtypes.EventTypeRewards, []abci.EventAttribute{{Key: "amount", Value: "100"}}), false},

		{"Empty proposer_reward event", createEvent(distrtypes.EventTypeProposerReward, []abci.EventAttribute{{Key: "amount", Value: ""}}), true},
		{"Non-empty proposer_reward event", createEvent(distrtypes.EventTypeProposerReward, []abci.EventAttribute{{Key: "amount", Value: "100"}}), false},

		{"Empty transfer event", createEvent(banktypes.EventTypeTransfer, []abci.EventAttribute{{Key: "amount", Value: ""}}), true},
		{"Non-empty transfer event", createEvent(banktypes.EventTypeTransfer, []abci.EventAttribute{{Key: "amount", Value: "100"}}), false},

		{"Empty coin_spent event", createEvent(banktypes.EventTypeCoinSpent, []abci.EventAttribute{{Key: "amount", Value: ""}}), true},
		{"Non-empty coin_spent event", createEvent(banktypes.EventTypeCoinSpent, []abci.EventAttribute{{Key: "amount", Value: "100"}}), false},

		{"Empty coin_received event", createEvent(banktypes.EventTypeCoinReceived, []abci.EventAttribute{{Key: "amount", Value: ""}}), true},
		{"Non-empty coin_received event", createEvent(banktypes.EventTypeCoinReceived, []abci.EventAttribute{{Key: "amount", Value: "100"}}), false},

		{"Unhandled event type with empty amount", createEvent("unhandled_type", []abci.EventAttribute{{Key: "amount", Value: ""}}), false},
		{"Unhandled event type with non-empty amount", createEvent("unhandled_type", []abci.EventAttribute{{Key: "amount", Value: "100"}}), false},
		{"Event with no attributes", createEvent(distrtypes.EventTypeCommission, []abci.EventAttribute{}), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := shouldFilterEvent(tc.event)
			assert.Equal(t, tc.expect, result, "Test %v failed: expected %v, got %v", tc.name, tc.expect, result)
		})
	}
}

func TestFilterBeginBlockerEvents(t *testing.T) {
	tests := []struct {
		name     string
		events   []abci.Event
		expected []abci.Event
	}{
		{
			name: "Filter out events with empty amounts",
			events: []abci.Event{
				createEvent(distrtypes.EventTypeCommission, []abci.EventAttribute{{Key: sdk.AttributeKeyAmount, Value: ""}}),
				createEvent(distrtypes.EventTypeRewards, []abci.EventAttribute{{Key: sdk.AttributeKeyAmount, Value: "100"}}),
			},
			expected: []abci.Event{
				createEvent(distrtypes.EventTypeRewards, []abci.EventAttribute{{Key: sdk.AttributeKeyAmount, Value: "100"}}),
			},
		},
		{
			name: "No filtering when all events are valid",
			events: []abci.Event{
				createEvent(banktypes.EventTypeTransfer, []abci.EventAttribute{{Key: sdk.AttributeKeyAmount, Value: "100"}}),
			},
			expected: []abci.Event{
				createEvent(banktypes.EventTypeTransfer, []abci.EventAttribute{{Key: sdk.AttributeKeyAmount, Value: "100"}}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			responseBeginBlock := sdk.BeginBlock{Events: tc.events}
			actualEvents := filterBeginBlockerEvents(responseBeginBlock)
			assert.Equal(t, len(tc.expected), len(actualEvents), "Number of events mismatch")

			for i, expectedEvent := range tc.expected {
				actualEvent := actualEvents[i]
				assert.Equal(t, expectedEvent.Type, actualEvent.Type, "Event types mismatch")

				assert.Equal(t, len(expectedEvent.Attributes), len(actualEvent.Attributes), "Number of attributes mismatch in event %v", expectedEvent.Type)

				for j, expectedAttribute := range expectedEvent.Attributes {
					actualAttribute := actualEvent.Attributes[j]
					assert.Equal(t, expectedAttribute.Key, actualAttribute.Key, "Attribute keys mismatch in event %v", expectedEvent.Type)
					assert.Equal(t, expectedAttribute.Value, actualAttribute.Value, "Attribute values mismatch in event %v", expectedEvent.Type)
				}
			}
		})
	}
}

func TestMsgServerProtoAnnotations(t *testing.T) {
	expErr := "service icq.v1.Msg does not have cosmos.msg.v1.service proto annotation"

	// Create an app so that we know everything's been registered.
	logger := log.NewNopLogger()
	db, err := dbm.NewDB("proto-test", dbm.MemDBBackend, "")
	require.NoError(t, err, "dbm.NewDB")
	appOpts := newSimAppOpts(t)
	baseAppOpts := []func(*baseapp.BaseApp){
		fauxMerkleModeOpt,
		baseapp.SetChainID(pioconfig.SimAppChainID),
	}
	_ = New(logger, db, nil, true, appOpts, baseAppOpts...)

	protoFiles, err := proto.MergedRegistry()
	require.NoError(t, err, "proto.MergedRegistry()")
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	assertions.AssertErrorValue(t, err, expErr, "ValidateProtoAnnotations")
}
