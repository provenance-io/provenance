package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// assertEventContent asserts that tev converts to an untyped SDK event whose
// Type equals "provenance.flatfees.v1.<typeString>". When assertAllSet is true
// it also verifies that every attribute key and value is non-empty.
// Returns true on success.
func assertEventContent(t *testing.T, tev proto.Message, typeString string, assertAllSet bool) bool {
	t.Helper()
	event, err := sdk.TypedEventToEvent(tev)
	if !assert.NoError(t, err, "TypedEventToEvent(%T)", tev) {
		return false
	}

	expType := "provenance.flatfees.v1." + typeString
	rv := assert.Equal(t, expType, event.Type, "%T event.Type", tev)
	if !assertAllSet {
		return rv
	}

	for i, attr := range event.Attributes {
		rv = assert.NotEmpty(t, attr.Key, "%T event.Attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `""`, attr.Key, "%T event.Attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `0`, attr.Key, "%T event.Attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `"0"`, attr.Key, "%T event.Attributes[%d].Key", tev, i) && rv
		rv = assert.NotEmpty(t, attr.Value, "%T event.Attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `""`, attr.Value, "%T event.Attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `0`, attr.Value, "%T event.Attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `"0"`, attr.Value, "%T event.Attributes[%d].Value", tev, i) && rv
	}
	return rv
}

// assertEverythingSet is a convenience wrapper over assertEventContent that
// also asserts every attribute is non-empty.
func assertEverythingSet(t *testing.T, tev proto.Message, typeString string) bool {
	t.Helper()
	return assertEventContent(t, tev, typeString, true)
}

func TestNewEventParamsUpdated(t *testing.T) {
	exp := &EventParamsUpdated{}

	var act *EventParamsUpdated
	testFunc := func() {
		act = NewEventParamsUpdated()
	}
	require.NotPanics(t, testFunc, "NewEventParamsUpdated")
	assert.Equal(t, exp, act, "NewEventParamsUpdated result")
	// EventParamsUpdated has no proto fields, so we only check the event type.
	assertEventContent(t, act, "EventParamsUpdated", false)
}

func TestNewEventConversionFactorUpdated(t *testing.T) {
	tests := []struct {
		name string
		cf   ConversionFactor
		exp  *EventConversionFactorUpdated
	}{
		{
			name: "non-zero amounts",
			cf: ConversionFactor{
				DefinitionAmount: sdk.NewInt64Coin("musd", 1),
				ConvertedAmount:  sdk.NewInt64Coin("nhash", 2_000),
			},
			exp: &EventConversionFactorUpdated{
				DefinitionAmount: sdk.NewInt64Coin("musd", 1).String(),
				ConvertedAmount:  sdk.NewInt64Coin("nhash", 2_000).String(),
			},
		},
		{
			name: "large amounts",
			cf: ConversionFactor{
				DefinitionAmount: sdk.NewInt64Coin("musd", 1_000_000),
				ConvertedAmount:  sdk.NewInt64Coin("nhash", 1_000_000_000),
			},
			exp: &EventConversionFactorUpdated{
				DefinitionAmount: sdk.NewInt64Coin("musd", 1_000_000).String(),
				ConvertedAmount:  sdk.NewInt64Coin("nhash", 1_000_000_000).String(),
			},
		},
		{
			name: "same denom same amount (identity conversion factor)",
			cf: ConversionFactor{
				DefinitionAmount: sdk.NewInt64Coin("nhash", 1),
				ConvertedAmount:  sdk.NewInt64Coin("nhash", 1),
			},
			exp: &EventConversionFactorUpdated{
				DefinitionAmount: sdk.NewInt64Coin("nhash", 1).String(),
				ConvertedAmount:  sdk.NewInt64Coin("nhash", 1).String(),
			},
		},
		{
			name: "zero converted amount (free msgs)",
			cf: ConversionFactor{
				DefinitionAmount: sdk.NewInt64Coin("musd", 1),
				ConvertedAmount:  sdk.NewInt64Coin("nhash", 0),
			},
			exp: &EventConversionFactorUpdated{
				DefinitionAmount: sdk.NewInt64Coin("musd", 1).String(),
				ConvertedAmount:  sdk.NewInt64Coin("nhash", 0).String(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventConversionFactorUpdated
			testFunc := func() {
				act = NewEventConversionFactorUpdated(tc.cf)
			}
			require.NotPanics(t, testFunc, "NewEventConversionFactorUpdated")
			assert.Equal(t, tc.exp, act, "NewEventConversionFactorUpdated result")
			assertEverythingSet(t, act, "EventConversionFactorUpdated")
		})
	}
}

func TestNewEventMsgFeeSet(t *testing.T) {
	tests := []struct {
		name       string
		msgTypeURL string
		exp        *EventMsgFeeSet
	}{
		{
			name:       "bank send",
			msgTypeURL: "/cosmos.bank.v1beta1.MsgSend",
			exp:        &EventMsgFeeSet{MsgTypeUrl: "/cosmos.bank.v1beta1.MsgSend"},
		},
		{
			name:       "provenance marker add",
			msgTypeURL: "/provenance.marker.v1.MsgAddMarkerRequest",
			exp:        &EventMsgFeeSet{MsgTypeUrl: "/provenance.marker.v1.MsgAddMarkerRequest"},
		},
		{
			name:       "generic type url",
			msgTypeURL: "some.module.v1.SomeMsg",
			exp:        &EventMsgFeeSet{MsgTypeUrl: "some.module.v1.SomeMsg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventMsgFeeSet
			testFunc := func() {
				act = NewEventMsgFeeSet(tc.msgTypeURL)
			}
			require.NotPanics(t, testFunc, "NewEventMsgFeeSet")
			assert.Equal(t, tc.exp, act, "NewEventMsgFeeSet result")
			assertEverythingSet(t, act, "EventMsgFeeSet")
		})
	}
}

func TestNewEventMsgFeeUnset(t *testing.T) {
	tests := []struct {
		name       string
		msgTypeURL string
		exp        *EventMsgFeeUnset
	}{
		{
			name:       "bank send",
			msgTypeURL: "/cosmos.bank.v1beta1.MsgSend",
			exp:        &EventMsgFeeUnset{MsgTypeUrl: "/cosmos.bank.v1beta1.MsgSend"},
		},
		{
			name:       "provenance name bind",
			msgTypeURL: "/provenance.name.v1.MsgBindNameRequest",
			exp:        &EventMsgFeeUnset{MsgTypeUrl: "/provenance.name.v1.MsgBindNameRequest"},
		},
		{
			name:       "generic type url",
			msgTypeURL: "some.module.v1.SomeMsg",
			exp:        &EventMsgFeeUnset{MsgTypeUrl: "some.module.v1.SomeMsg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventMsgFeeUnset
			testFunc := func() {
				act = NewEventMsgFeeUnset(tc.msgTypeURL)
			}
			require.NotPanics(t, testFunc, "NewEventMsgFeeUnset")
			assert.Equal(t, tc.exp, act, "NewEventMsgFeeUnset result")
			assertEverythingSet(t, act, "EventMsgFeeUnset")
		})
	}
}

func TestNewEventOracleAddressAdded(t *testing.T) {
	addr1 := sdk.AccAddress("oracle1_____________").String()
	addr2 := sdk.AccAddress("oracle2_____________").String()

	tests := []struct {
		name    string
		address string
		exp     *EventOracleAddressAdded
	}{
		{
			name:    "oracle1",
			address: addr1,
			exp:     &EventOracleAddressAdded{OracleAddress: addr1},
		},
		{
			name:    "oracle2",
			address: addr2,
			exp:     &EventOracleAddressAdded{OracleAddress: addr2},
		},
		{
			name:    "governance address",
			address: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			exp:     &EventOracleAddressAdded{OracleAddress: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventOracleAddressAdded
			testFunc := func() {
				act = NewEventOracleAddressAdded(tc.address)
			}
			require.NotPanics(t, testFunc, "NewEventOracleAddressAdded")
			assert.Equal(t, tc.exp, act, "NewEventOracleAddressAdded result")
			assertEverythingSet(t, act, "EventOracleAddressAdded")
		})
	}
}

func TestNewEventOracleAddressRemoved(t *testing.T) {
	addr1 := sdk.AccAddress("oracle1_____________").String()
	addr2 := sdk.AccAddress("oracle2_____________").String()

	tests := []struct {
		name    string
		address string
		exp     *EventOracleAddressRemoved
	}{
		{
			name:    "oracle1",
			address: addr1,
			exp:     &EventOracleAddressRemoved{OracleAddress: addr1},
		},
		{
			name:    "oracle2",
			address: addr2,
			exp:     &EventOracleAddressRemoved{OracleAddress: addr2},
		},
		{
			name:    "governance address",
			address: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			exp:     &EventOracleAddressRemoved{OracleAddress: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventOracleAddressRemoved
			testFunc := func() {
				act = NewEventOracleAddressRemoved(tc.address)
			}
			require.NotPanics(t, testFunc, "NewEventOracleAddressRemoved")
			assert.Equal(t, tc.exp, act, "NewEventOracleAddressRemoved result")
			assertEverythingSet(t, act, "EventOracleAddressRemoved")
		})
	}
}
