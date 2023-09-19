package exchange

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

// assertEverythingSet asserts that the provided proto.Message has a value for all fields.
// Returns true on success, false if one or more things are missing.
func assertEverythingSet(t *testing.T, tev proto.Message, typeString string) bool {
	t.Helper()
	event, err := sdk.TypedEventToEvent(tev)
	if !assert.NoError(t, err, "TypedEventToEvent(%T)", tev) {
		return false
	}

	expType := "provenance.exchange.v1." + typeString
	rv := assert.Equal(t, expType, event.Type, "%T event.Type")
	for i, attrs := range event.Attributes {
		rv = assert.NotEmpty(t, attrs.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `""`, attrs.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEmpty(t, attrs.Value, "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `""`, attrs.Value, "%T event.attributes[%d].Value", tev, i) && rv
	}
	return rv
}

func TestNewEventOrderCreated(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected *EventOrderCreated
		expPanic interface{}
	}{
		{
			name:     "nil order",
			order:    NewOrder(3),
			expPanic: "GetOrderType() missing case for <nil>",
		},
		{
			name:  "order with ask",
			order: NewOrder(1).WithAsk(&AskOrder{}),
			expected: &EventOrderCreated{
				OrderId:   1,
				OrderType: "ask",
			},
		},
		{
			name:  "order with bid",
			order: NewOrder(2).WithBid(&BidOrder{}),
			expected: &EventOrderCreated{
				OrderId:   2,
				OrderType: "bid",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *EventOrderCreated
			testFunc := func() {
				actual = NewEventOrderCreated(tc.order)
			}

			if tc.expPanic != nil {
				require.PanicsWithValue(t, tc.expPanic, testFunc, "NewEventOrderCreated")
			} else {
				require.NotPanics(t, testFunc, "NewEventOrderCreated")
				assert.Equal(t, tc.expected, actual, "NewEventOrderCreated result")
				assertEverythingSet(t, actual, "EventOrderCreated")
			}
		})
	}
}

func TestNewEventOrderCancelled(t *testing.T) {
	orderID := uint64(101)
	cancelledBy := sdk.AccAddress("cancelledBy_________")

	event := NewEventOrderCancelled(orderID, cancelledBy)
	assert.Equal(t, orderID, event.OrderId, "OrderId")
	assert.Equal(t, cancelledBy.String(), event.CancelledBy, "CancelledBy")
	assertEverythingSet(t, event, "EventOrderCancelled")
}

func TestNewEventOrderFilled(t *testing.T) {
	orderID := uint64(5)

	event := NewEventOrderFilled(orderID)
	assert.Equal(t, orderID, event.OrderId, "OrderId")
	assertEverythingSet(t, event, "EventOrderFilled")
}

func TestNewEventOrderPartiallyFilled(t *testing.T) {
	orderID := uint64(18)
	assetsFilled := sdk.NewCoins(sdk.NewInt64Coin("first", 111), sdk.NewInt64Coin("second", 22))
	feesFilled := sdk.NewCoins(sdk.NewInt64Coin("charge", 8), sdk.NewInt64Coin("fee", 15))

	event := NewEventOrderPartiallyFilled(orderID, assetsFilled, feesFilled)
	assert.Equal(t, orderID, event.OrderId, "OrderId")
	assert.Equal(t, assetsFilled.String(), event.AssetsFilled, "AssetsFilled")
	assert.Equal(t, feesFilled.String(), event.FeesFilled, "FeesFilled")
	assertEverythingSet(t, event, "EventOrderPartiallyFilled")
}

func TestNewEventMarketWithdraw(t *testing.T) {
	marketID := uint32(55)
	amountWithdrawn := sdk.NewCoins(sdk.NewInt64Coin("mine", 188382), sdk.NewInt64Coin("yours", 3))
	destination := sdk.AccAddress("destination_________")
	withdrawnBy := sdk.AccAddress("withdrawnBy_________")

	event := NewEventMarketWithdraw(marketID, amountWithdrawn, destination, withdrawnBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, amountWithdrawn.String(), event.Amount, "Amount")
	assert.Equal(t, destination.String(), event.Destination, "Destination")
	assert.Equal(t, withdrawnBy.String(), event.WithdrawnBy, "WithdrawnBy")
	assertEverythingSet(t, event, "EventMarketWithdraw")
}

func TestNewEventMarketDetailsUpdated(t *testing.T) {
	marketID := uint32(84)
	updatedBy := sdk.AccAddress("updatedBy___________")

	event := NewEventMarketDetailsUpdated(marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketDetailsUpdated")
}

func TestNewEventMarketActiveUpdated(t *testing.T) {
	someAddr := sdk.AccAddress("some_address________")

	tests := []struct {
		name      string
		marketID  uint32
		updatedBy sdk.AccAddress
		isActive  bool
		expected  proto.Message
	}{
		{
			name:      "enabled",
			marketID:  33,
			updatedBy: someAddr,
			isActive:  true,
			expected:  NewEventMarketEnabled(33, someAddr),
		},
		{
			name:      "disabled",
			marketID:  556,
			updatedBy: someAddr,
			isActive:  false,
			expected:  NewEventMarketDisabled(556, someAddr),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewEventMarketActiveUpdated(tc.marketID, tc.updatedBy, tc.isActive)
			assert.Equal(t, tc.expected, actual, "NewEventMarketActiveUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy.String(), tc.isActive)
		})
	}
}

func TestNewEventMarketEnabled(t *testing.T) {
	marketID := uint32(919)
	updatedBy := sdk.AccAddress("updatedBy___________")

	event := NewEventMarketEnabled(marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketEnabled")
}

func TestNewEventMarketDisabled(t *testing.T) {
	marketID := uint32(5555)
	updatedBy := sdk.AccAddress("updatedBy___________")

	event := NewEventMarketDisabled(marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketDisabled")
}

func TestNewEventMarketUserSettleUpdated(t *testing.T) {
	someAddr := sdk.AccAddress("some_address________")

	tests := []struct {
		name      string
		marketID  uint32
		updatedBy sdk.AccAddress
		isAllowed bool
		expected  proto.Message
	}{
		{
			name:      "enabled",
			marketID:  33,
			updatedBy: someAddr,
			isAllowed: true,
			expected:  NewEventMarketUserSettleEnabled(33, someAddr),
		},
		{
			name:      "disabled",
			marketID:  556,
			updatedBy: someAddr,
			isAllowed: false,
			expected:  NewEventMarketUserSettleDisabled(556, someAddr),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewEventMarketUserSettleUpdated(tc.marketID, tc.updatedBy, tc.isAllowed)
			assert.Equal(t, tc.expected, actual, "NewEventMarketUserSettleUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy.String(), tc.isAllowed)
		})
	}
}

func TestNewEventMarketUserSettleEnabled(t *testing.T) {
	marketID := uint32(123)
	updatedBy := sdk.AccAddress("updatedBy___________")

	event := NewEventMarketUserSettleEnabled(marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketUserSettleEnabled")
}

func TestNewEventMarketUserSettleDisabled(t *testing.T) {
	marketID := uint32(123)
	updatedBy := sdk.AccAddress("updatedBy___________")

	event := NewEventMarketUserSettleDisabled(marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketUserSettleDisabled")
}

func TestNewEventMarketPermissionsUpdated(t *testing.T) {
	marketID := uint32(5432)
	updatedBy := sdk.AccAddress("updatedBy___________")

	event := NewEventMarketPermissionsUpdated(marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketPermissionsUpdated")
}

func TestNewEventMarketReqAttrUpdated(t *testing.T) {
	marketID := uint32(3334)
	updatedBy := sdk.AccAddress("updatedBy___________")

	event := NewEventMarketReqAttrUpdated(marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketReqAttrUpdated")
}

func TestNewEventMarketCreated(t *testing.T) {
	marketID := uint32(10111213)

	event := NewEventMarketCreated(marketID)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assertEverythingSet(t, event, "EventMarketCreated")
}

func TestNewEventMarketFeesUpdated(t *testing.T) {
	marketID := uint32(1415)

	event := NewEventMarketFeesUpdated(marketID)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assertEverythingSet(t, event, "EventMarketFeesUpdated")
}

func TestNewEventParamsUpdated(t *testing.T) {
	event := NewEventParamsUpdated()
	assertEverythingSet(t, event, "EventParamsUpdated")
}

func TestTypedEventToEvent(t *testing.T) {
	quoteBz := func(str string) []byte {
		return []byte(fmt.Sprintf("%q", str))
	}
	cancelledBy := sdk.AccAddress("cancelledBy_________")
	cancelledByQ := quoteBz(cancelledBy.String())
	destination := sdk.AccAddress("destination_________")
	destinationQ := quoteBz(destination.String())
	withdrawnBy := sdk.AccAddress("withdrawnBy_________")
	withdrawnByQ := quoteBz(withdrawnBy.String())
	updatedBy := sdk.AccAddress("updatedBy___________")
	updatedByQ := quoteBz(updatedBy.String())
	coins1 := sdk.NewCoins(sdk.NewInt64Coin("onecoin", 1), sdk.NewInt64Coin("twocoin", 2))
	coins1Q := quoteBz(coins1.String())
	coins2 := sdk.NewCoins(sdk.NewInt64Coin("threecoin", 3), sdk.NewInt64Coin("fourcoin", 4))
	coins2Q := quoteBz(coins2.String())

	tests := []struct {
		name     string
		tev      proto.Message
		expEvent sdk.Event
	}{
		{
			name: "EventOrderCreated ask",
			tev:  NewEventOrderCreated(NewOrder(1).WithAsk(&AskOrder{})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCreated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("order_id"), Value: quoteBz("1")},
					{Key: []byte("order_type"), Value: quoteBz("ask")},
				},
			},
		},
		{
			name: "EventOrderCreated bid",
			tev:  NewEventOrderCreated(NewOrder(2).WithBid(&BidOrder{})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCreated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("order_id"), Value: quoteBz("2")},
					{Key: []byte("order_type"), Value: quoteBz("bid")},
				},
			},
		},
		{
			name: "EventOrderCancelled",
			tev:  NewEventOrderCancelled(3, cancelledBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCancelled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("cancelled_by"), Value: cancelledByQ},
					{Key: []byte("order_id"), Value: quoteBz("3")},
				},
			},
		},
		{
			name: "EventOrderFilled",
			tev:  NewEventOrderFilled(4),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderFilled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("order_id"), Value: quoteBz("4")},
				},
			},
		},
		{
			name: "EventOrderPartiallyFilled",
			tev:  NewEventOrderPartiallyFilled(5, coins1, coins2),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderPartiallyFilled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("assets_filled"), Value: coins1Q},
					{Key: []byte("fees_filled"), Value: coins2Q},
					{Key: []byte("order_id"), Value: quoteBz("5")},
				},
			},
		},
		{
			name: "EventMarketWithdraw",
			tev:  NewEventMarketWithdraw(6, coins1, destination, withdrawnBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketWithdraw",
				Attributes: []abci.EventAttribute{
					{Key: []byte("amount"), Value: coins1Q},
					{Key: []byte("destination"), Value: destinationQ},
					{Key: []byte("market_id"), Value: []byte("6")},
					{Key: []byte("withdrawn_by"), Value: withdrawnByQ},
				},
			},
		},
		{
			name: "EventMarketDetailsUpdated",
			tev:  NewEventMarketDetailsUpdated(7, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketDetailsUpdated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("7")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketEnabled",
			tev:  NewEventMarketEnabled(8, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketEnabled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("8")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketDisabled",
			tev:  NewEventMarketDisabled(9, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketDisabled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("9")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketUserSettleEnabled",
			tev:  NewEventMarketUserSettleEnabled(10, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketUserSettleEnabled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("10")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketUserSettleDisabled",
			tev:  NewEventMarketUserSettleDisabled(11, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketUserSettleDisabled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("11")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketPermissionsUpdated",
			tev:  NewEventMarketPermissionsUpdated(12, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketPermissionsUpdated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("12")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketReqAttrUpdated",
			tev:  NewEventMarketReqAttrUpdated(13, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketReqAttrUpdated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("13")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketCreated",
			tev:  NewEventMarketCreated(14),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketCreated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("14")},
				},
			},
		},
		{
			name: "EventMarketFeesUpdated",
			tev:  NewEventMarketFeesUpdated(15),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketFeesUpdated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("15")},
				},
			},
		},
		{
			name: "EventParamsUpdated",
			tev:  NewEventParamsUpdated(),
			expEvent: sdk.Event{
				Type:       "provenance.exchange.v1.EventParamsUpdated",
				Attributes: nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event, err := sdk.TypedEventToEvent(tc.tev)
			require.NoError(t, err, "TypedEventToEvent error")
			if assert.NotNil(t, event, "TypedEventToEvent result") {
				assert.Equal(t, tc.expEvent.Type, event.Type, "event type")
				expAttrs := assertions.AttrsToStrings(tc.expEvent.Attributes)
				actAttrs := assertions.AttrsToStrings(event.Attributes)
				assert.Equal(t, expAttrs, actAttrs, "event attributes")
			}
		})
	}
}
