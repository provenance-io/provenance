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
	rv := assert.Equal(t, expType, event.Type, "%T event.Type", tev)
	for i, attr := range event.Attributes {
		rv = assert.NotEmpty(t, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `""`, string(attr.Key), "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `0`, string(attr.Key), "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `"0"`, string(attr.Key), "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEmpty(t, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `""`, string(attr.Value), "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `0`, string(attr.Value), "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `"0"`, string(attr.Value), "%T event.attributes[%d].Value", tev, i) && rv
	}
	return rv
}

func TestNewEventOrderCreated(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected *EventOrderCreated
		expPanic string
	}{
		{
			name:     "nil order",
			order:    NewOrder(3),
			expPanic: "order 3 has unknown sub-order type <nil>: does not implement SubOrderI",
		},
		{
			name:  "order with ask",
			order: NewOrder(1).WithAsk(&AskOrder{MarketId: 97, ExternalId: "oneoneone"}),
			expected: &EventOrderCreated{
				OrderId:    1,
				OrderType:  "ask",
				MarketId:   97,
				ExternalId: "oneoneone",
			},
		},
		{
			name:  "order with bid",
			order: NewOrder(2).WithBid(&BidOrder{MarketId: 33, ExternalId: "twotwotwo"}),
			expected: &EventOrderCreated{
				OrderId:    2,
				OrderType:  "bid",
				MarketId:   33,
				ExternalId: "twotwotwo",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *EventOrderCreated
			testFunc := func() {
				actual = NewEventOrderCreated(tc.order)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "NewEventOrderCreated")
			if len(tc.expPanic) > 0 {
				return
			}
			assert.Equal(t, tc.expected, actual, "NewEventOrderCreated result")
			assertEverythingSet(t, actual, "EventOrderCreated")
		})
	}
}

func TestNewEventOrderCancelled(t *testing.T) {
	tests := []struct {
		name        string
		order       OrderI
		cancelledBy sdk.AccAddress
		expected    *EventOrderCancelled
	}{
		{
			name:        "ask order",
			order:       NewOrder(11).WithAsk(&AskOrder{MarketId: 71, ExternalId: "an external identifier"}),
			cancelledBy: sdk.AccAddress("CancelledBy_________"),
			expected: &EventOrderCancelled{
				OrderId:     11,
				CancelledBy: sdk.AccAddress("CancelledBy_________").String(),
				MarketId:    71,
				ExternalId:  "an external identifier",
			},
		},
		{
			name:        "bid order",
			order:       NewOrder(55).WithAsk(&AskOrder{MarketId: 88, ExternalId: "another external identifier"}),
			cancelledBy: sdk.AccAddress("cancelled_by________"),
			expected: &EventOrderCancelled{
				OrderId:     55,
				CancelledBy: sdk.AccAddress("cancelled_by________").String(),
				MarketId:    88,
				ExternalId:  "another external identifier",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventOrderCancelled
			testFunc := func() {
				event = NewEventOrderCancelled(tc.order, tc.cancelledBy)
			}
			require.NotPanics(t, testFunc, "NewEventOrderCancelled")
			assert.Equal(t, tc.expected, event, "NewEventOrderCancelled result")
			assertEverythingSet(t, event, "EventOrderCancelled")
		})
	}
}

func TestNewEventOrderFilled(t *testing.T) {
	coinP := func(denom string, amount int64) *sdk.Coin {
		rv := sdk.NewInt64Coin(denom, amount)
		return &rv
	}

	tests := []struct {
		name     string
		order    OrderI
		expected *EventOrderFilled
	}{
		{
			name: "ask",
			order: NewOrder(4).WithAsk(&AskOrder{
				MarketId:                57,
				Assets:                  sdk.NewInt64Coin("apple", 22),
				Price:                   sdk.NewInt64Coin("plum", 18),
				SellerSettlementFlatFee: coinP("fig", 57),
				ExternalId:              "one",
			}),
			expected: &EventOrderFilled{
				OrderId:    4,
				Assets:     "22apple",
				Price:      "18plum",
				Fees:       "57fig",
				MarketId:   57,
				ExternalId: "one",
			},
		},
		{
			name: "filled ask",
			order: NewFilledOrder(NewOrder(4).WithAsk(&AskOrder{
				MarketId:                1234,
				Assets:                  sdk.NewInt64Coin("apple", 22),
				Price:                   sdk.NewInt64Coin("plum", 18),
				SellerSettlementFlatFee: coinP("fig", 57),
				ExternalId:              "two",
			}), sdk.NewInt64Coin("plum", 88), sdk.NewCoins(sdk.NewInt64Coin("fig", 61), sdk.NewInt64Coin("grape", 12))),
			expected: &EventOrderFilled{
				OrderId:    4,
				Assets:     "22apple",
				Price:      "88plum",
				Fees:       "61fig,12grape",
				MarketId:   1234,
				ExternalId: "two",
			},
		},
		{
			name: "bid",
			order: NewOrder(104).WithBid(&BidOrder{
				MarketId:            87878,
				Assets:              sdk.NewInt64Coin("apple", 23),
				Price:               sdk.NewInt64Coin("plum", 19),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 58)),
				ExternalId:          "three",
			}),
			expected: &EventOrderFilled{
				OrderId:    104,
				Assets:     "23apple",
				Price:      "19plum",
				Fees:       "58fig",
				MarketId:   87878,
				ExternalId: "three",
			},
		},
		{
			name: "filled bid",
			order: NewFilledOrder(NewOrder(105).WithBid(&BidOrder{
				MarketId:            9119,
				Assets:              sdk.NewInt64Coin("apple", 24),
				Price:               sdk.NewInt64Coin("plum", 20),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 59)),
				ExternalId:          "four",
			}), sdk.NewInt64Coin("plum", 89), sdk.NewCoins(sdk.NewInt64Coin("fig", 62), sdk.NewInt64Coin("grape", 13))),
			expected: &EventOrderFilled{
				OrderId:    105,
				Assets:     "24apple",
				Price:      "89plum",
				Fees:       "62fig,13grape",
				MarketId:   9119,
				ExternalId: "four",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventOrderFilled
			testFunc := func() {
				event = NewEventOrderFilled(tc.order)
			}
			require.NotPanics(t, testFunc, "NewEventOrderFilled")
			assert.Equal(t, tc.expected, event, "NewEventOrderFilled result")
			assertEverythingSet(t, event, "EventOrderFilled")
		})
	}
}

func TestNewEventOrderPartiallyFilled(t *testing.T) {
	coinP := func(denom string, amount int64) *sdk.Coin {
		rv := sdk.NewInt64Coin(denom, amount)
		return &rv
	}

	tests := []struct {
		name     string
		order    OrderI
		expected *EventOrderPartiallyFilled
	}{
		{
			name: "ask",
			order: NewOrder(4).WithAsk(&AskOrder{
				MarketId:                432,
				Assets:                  sdk.NewInt64Coin("apple", 22),
				Price:                   sdk.NewInt64Coin("plum", 18),
				SellerSettlementFlatFee: coinP("fig", 57),
				ExternalId:              "five",
			}),
			expected: &EventOrderPartiallyFilled{
				OrderId:    4,
				Assets:     "22apple",
				Price:      "18plum",
				Fees:       "57fig",
				MarketId:   432,
				ExternalId: "five",
			},
		},
		{
			name: "filled ask",
			order: NewFilledOrder(NewOrder(4).WithAsk(&AskOrder{
				MarketId:                456,
				Assets:                  sdk.NewInt64Coin("apple", 22),
				Price:                   sdk.NewInt64Coin("plum", 18),
				SellerSettlementFlatFee: coinP("fig", 57),
				ExternalId:              "six",
			}), sdk.NewInt64Coin("plum", 88), sdk.NewCoins(sdk.NewInt64Coin("fig", 61), sdk.NewInt64Coin("grape", 12))),
			expected: &EventOrderPartiallyFilled{
				OrderId:    4,
				Assets:     "22apple",
				Price:      "88plum",
				Fees:       "61fig,12grape",
				MarketId:   456,
				ExternalId: "six",
			},
		},
		{
			name: "bid",
			order: NewOrder(104).WithBid(&BidOrder{
				MarketId:            765,
				Assets:              sdk.NewInt64Coin("apple", 23),
				Price:               sdk.NewInt64Coin("plum", 19),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 58)),
				ExternalId:          "seven",
			}),
			expected: &EventOrderPartiallyFilled{
				OrderId:    104,
				Assets:     "23apple",
				Price:      "19plum",
				Fees:       "58fig",
				MarketId:   765,
				ExternalId: "seven",
			},
		},
		{
			name: "filled bid",
			order: NewFilledOrder(NewOrder(104).WithBid(&BidOrder{
				MarketId:            818,
				Assets:              sdk.NewInt64Coin("apple", 23),
				Price:               sdk.NewInt64Coin("plum", 19),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 58)),
				ExternalId:          "eight",
			}), sdk.NewInt64Coin("plum", 89), sdk.NewCoins(sdk.NewInt64Coin("fig", 62), sdk.NewInt64Coin("grape", 13))),
			expected: &EventOrderPartiallyFilled{
				OrderId:    104,
				Assets:     "23apple",
				Price:      "89plum",
				Fees:       "62fig,13grape",
				MarketId:   818,
				ExternalId: "eight",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventOrderPartiallyFilled
			testFunc := func() {
				event = NewEventOrderPartiallyFilled(tc.order)
			}
			require.NotPanics(t, testFunc, "NewEventOrderPartiallyFilled")
			assert.Equal(t, tc.expected, event, "NewEventOrderPartiallyFilled result")
			assertEverythingSet(t, event, "EventOrderPartiallyFilled")
		})
	}
}

func TestNewEventOrderExternalIDUpdated(t *testing.T) {
	tests := []struct {
		name     string
		order    OrderI
		expected *EventOrderExternalIDUpdated
	}{
		{
			name:  "ask",
			order: NewOrder(51).WithAsk(&AskOrder{MarketId: 9, ExternalId: "orange-red"}),
			expected: &EventOrderExternalIDUpdated{
				OrderId:    51,
				MarketId:   9,
				ExternalId: "orange-red",
			},
		},
		{
			name:  "bid",
			order: NewOrder(777).WithAsk(&AskOrder{MarketId: 53, ExternalId: "purple-purple"}),
			expected: &EventOrderExternalIDUpdated{
				OrderId:    777,
				MarketId:   53,
				ExternalId: "purple-purple",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventOrderExternalIDUpdated
			testFunc := func() {
				event = NewEventOrderExternalIDUpdated(tc.order)
			}
			require.NotPanics(t, testFunc, "NewEventOrderExternalIDUpdated")
			assert.Equal(t, tc.expected, event, "NewEventOrderExternalIDUpdated result")
			assertEverythingSet(t, event, "EventOrderExternalIDUpdated")
		})
	}
}

func TestNewEventMarketWithdraw(t *testing.T) {
	marketID := uint32(55)
	amountWithdrawn := sdk.NewCoins(sdk.NewInt64Coin("mine", 188382), sdk.NewInt64Coin("yours", 3))
	destination := sdk.AccAddress("destination_________")
	withdrawnBy := sdk.AccAddress("withdrawnBy_________")

	var event *EventMarketWithdraw
	testFunc := func() {
		event = NewEventMarketWithdraw(marketID, amountWithdrawn, destination, withdrawnBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketWithdraw(%d, %q, %q, %q)",
		marketID, amountWithdrawn, string(destination), string(withdrawnBy))
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, amountWithdrawn.String(), event.Amount, "Amount")
	assert.Equal(t, destination.String(), event.Destination, "Destination")
	assert.Equal(t, withdrawnBy.String(), event.WithdrawnBy, "WithdrawnBy")
	assertEverythingSet(t, event, "EventMarketWithdraw")
}

func TestNewEventMarketDetailsUpdated(t *testing.T) {
	marketID := uint32(84)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketDetailsUpdated
	testFunc := func() {
		event = NewEventMarketDetailsUpdated(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketDetailsUpdated(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketDetailsUpdated")
}

func TestNewEventMarketActiveUpdated(t *testing.T) {
	someAddr := sdk.AccAddress("some_address________").String()

	tests := []struct {
		name      string
		marketID  uint32
		updatedBy string
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
			var event proto.Message
			testFunc := func() {
				event = NewEventMarketActiveUpdated(tc.marketID, tc.updatedBy, tc.isActive)
			}
			require.NotPanics(t, testFunc, "NewEventMarketActiveUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy, tc.isActive)
			assert.Equal(t, tc.expected, event, "NewEventMarketActiveUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy, tc.isActive)
		})
	}
}

func TestNewEventMarketEnabled(t *testing.T) {
	marketID := uint32(919)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketEnabled
	testFunc := func() {
		event = NewEventMarketEnabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketEnabled(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketEnabled")
}

func TestNewEventMarketDisabled(t *testing.T) {
	marketID := uint32(5555)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketDisabled
	testFunc := func() {
		event = NewEventMarketDisabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketDisabled(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
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
			var event proto.Message
			testFunc := func() {
				event = NewEventMarketUserSettleUpdated(tc.marketID, tc.updatedBy, tc.isAllowed)
			}
			require.NotPanics(t, testFunc, "NewEventMarketUserSettleUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy.String(), tc.isAllowed)
			assert.Equal(t, tc.expected, event, "NewEventMarketUserSettleUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy.String(), tc.isAllowed)
		})
	}
}

func TestNewEventMarketUserSettleEnabled(t *testing.T) {
	marketID := uint32(123)
	updatedBy := sdk.AccAddress("updatedBy___________")

	var event *EventMarketUserSettleEnabled
	testFunc := func() {
		event = NewEventMarketUserSettleEnabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketUserSettleEnabled(%d, %q)", marketID, string(updatedBy))
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketUserSettleEnabled")
}

func TestNewEventMarketUserSettleDisabled(t *testing.T) {
	marketID := uint32(123)
	updatedBy := sdk.AccAddress("updatedBy___________")

	var event *EventMarketUserSettleDisabled
	testFunc := func() {
		event = NewEventMarketUserSettleDisabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketUserSettleDisabled(%d, %q)", marketID, string(updatedBy))
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketUserSettleDisabled")
}

func TestNewEventMarketPermissionsUpdated(t *testing.T) {
	marketID := uint32(5432)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketPermissionsUpdated
	testFunc := func() {
		event = NewEventMarketPermissionsUpdated(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketPermissionsUpdated(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketPermissionsUpdated")
}

func TestNewEventMarketReqAttrUpdated(t *testing.T) {
	marketID := uint32(3334)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketReqAttrUpdated
	testFunc := func() {
		event = NewEventMarketReqAttrUpdated(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketReqAttrUpdated(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketReqAttrUpdated")
}

func TestNewEventMarketCreated(t *testing.T) {
	marketID := uint32(10111213)

	var event *EventMarketCreated
	testFunc := func() {
		event = NewEventMarketCreated(marketID)
	}
	require.NotPanics(t, testFunc, "NewEventMarketCreated(%d)", marketID)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assertEverythingSet(t, event, "EventMarketCreated")
}

func TestNewEventMarketFeesUpdated(t *testing.T) {
	marketID := uint32(1415)

	var event *EventMarketFeesUpdated
	testFunc := func() {
		event = NewEventMarketFeesUpdated(marketID)
	}
	require.NotPanics(t, testFunc, "NewEventMarketFeesUpdated(%d)", marketID)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assertEverythingSet(t, event, "EventMarketFeesUpdated")
}

func TestNewEventParamsUpdated(t *testing.T) {
	var event *EventParamsUpdated
	testFunc := func() {
		event = NewEventParamsUpdated()
	}
	require.NotPanics(t, testFunc, "NewEventParamsUpdated()")
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
	acoin := sdk.NewInt64Coin("acoin", 55)
	acoinQ := quoteBz(acoin.String())
	pcoin := sdk.NewInt64Coin("pcoin", 66)
	pcoinQ := quoteBz(pcoin.String())
	fcoin := sdk.NewInt64Coin("fcoin", 33)
	fcoinQ := quoteBz(fcoin.String())

	tests := []struct {
		name     string
		tev      proto.Message
		expEvent sdk.Event
	}{
		{
			name: "EventOrderCreated ask",
			tev:  NewEventOrderCreated(NewOrder(1).WithAsk(&AskOrder{MarketId: 88, ExternalId: "stuff"})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCreated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: quoteBz("stuff")},
					{Key: []byte("market_id"), Value: []byte("88")},
					{Key: []byte("order_id"), Value: quoteBz("1")},
					{Key: []byte("order_type"), Value: quoteBz("ask")},
				},
			},
		},
		{
			name: "EventOrderCreated bid",
			tev:  NewEventOrderCreated(NewOrder(2).WithBid(&BidOrder{MarketId: 77, ExternalId: "something else"})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCreated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: quoteBz("something else")},
					{Key: []byte("market_id"), Value: []byte("77")},
					{Key: []byte("order_id"), Value: quoteBz("2")},
					{Key: []byte("order_type"), Value: quoteBz("bid")},
				},
			},
		},
		{
			name: "EventOrderCancelled ask",
			tev:  NewEventOrderCancelled(NewOrder(3).WithAsk(&AskOrder{MarketId: 66, ExternalId: "outside 8"}), cancelledBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCancelled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("cancelled_by"), Value: cancelledByQ},
					{Key: []byte("external_id"), Value: quoteBz("outside 8")},
					{Key: []byte("market_id"), Value: []byte("66")},
					{Key: []byte("order_id"), Value: quoteBz("3")},
				},
			},
		},
		{
			name: "EventOrderCancelled bid",
			tev:  NewEventOrderCancelled(NewOrder(3).WithBid(&BidOrder{MarketId: 55, ExternalId: "outside 8"}), cancelledBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCancelled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("cancelled_by"), Value: cancelledByQ},
					{Key: []byte("external_id"), Value: quoteBz("outside 8")},
					{Key: []byte("market_id"), Value: []byte("55")},
					{Key: []byte("order_id"), Value: quoteBz("3")},
				},
			},
		},
		{
			name: "EventOrderFilled ask",
			tev: NewEventOrderFilled(NewOrder(4).WithAsk(&AskOrder{
				MarketId:                33,
				Assets:                  acoin,
				Price:                   pcoin,
				SellerSettlementFlatFee: &fcoin,
				ExternalId:              "eeeeiiiiiddddd",
			})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderFilled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("assets"), Value: acoinQ},
					{Key: []byte("external_id"), Value: quoteBz("eeeeiiiiiddddd")},
					{Key: []byte("fees"), Value: fcoinQ},
					{Key: []byte("market_id"), Value: []byte("33")},
					{Key: []byte("order_id"), Value: quoteBz("4")},
					{Key: []byte("price"), Value: pcoinQ},
				},
			},
		},
		{
			name: "EventOrderFilled bid",
			tev: NewEventOrderFilled(NewOrder(104).WithBid(&BidOrder{
				MarketId:            44,
				Assets:              acoin,
				Price:               pcoin,
				BuyerSettlementFees: sdk.Coins{fcoin},
				ExternalId:          "that one thing",
			})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderFilled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("assets"), Value: acoinQ},
					{Key: []byte("external_id"), Value: quoteBz("that one thing")},
					{Key: []byte("fees"), Value: fcoinQ},
					{Key: []byte("market_id"), Value: []byte("44")},
					{Key: []byte("order_id"), Value: quoteBz("104")},
					{Key: []byte("price"), Value: pcoinQ},
				},
			},
		},
		{
			name: "EventOrderPartiallyFilled ask",
			tev: NewEventOrderPartiallyFilled(NewOrder(5).WithAsk(&AskOrder{
				MarketId:                22,
				Assets:                  acoin,
				Price:                   pcoin,
				SellerSettlementFlatFee: &fcoin,
				ExternalId:              "12345",
			})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderPartiallyFilled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("assets"), Value: acoinQ},
					{Key: []byte("external_id"), Value: quoteBz("12345")},
					{Key: []byte("fees"), Value: fcoinQ},
					{Key: []byte("market_id"), Value: []byte("22")},
					{Key: []byte("order_id"), Value: quoteBz("5")},
					{Key: []byte("price"), Value: pcoinQ},
				},
			},
		},
		{
			name: "EventOrderPartiallyFilled bid",
			tev: NewEventOrderPartiallyFilled(NewOrder(5).WithBid(&BidOrder{
				MarketId:            11,
				Assets:              acoin,
				Price:               pcoin,
				BuyerSettlementFees: sdk.Coins{fcoin},
				ExternalId:          "67890",
			})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderPartiallyFilled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("assets"), Value: acoinQ},
					{Key: []byte("external_id"), Value: quoteBz("67890")},
					{Key: []byte("fees"), Value: fcoinQ},
					{Key: []byte("market_id"), Value: []byte("11")},
					{Key: []byte("order_id"), Value: quoteBz("5")},
					{Key: []byte("price"), Value: pcoinQ},
				},
			},
		},
		{
			name: "EventOrderExternalIDUpdated ask",
			tev:  NewEventOrderExternalIDUpdated(NewOrder(8).WithAsk(&AskOrder{MarketId: 99, ExternalId: "yellow"})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderExternalIDUpdated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: quoteBz("yellow")},
					{Key: []byte("market_id"), Value: []byte("99")},
					{Key: []byte("order_id"), Value: quoteBz("8")},
				},
			},
		},
		{
			name: "EventOrderExternalIDUpdated bid",
			tev:  NewEventOrderExternalIDUpdated(NewOrder(8).WithBid(&BidOrder{MarketId: 111, ExternalId: "yellow"})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderExternalIDUpdated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: quoteBz("yellow")},
					{Key: []byte("market_id"), Value: []byte("111")},
					{Key: []byte("order_id"), Value: quoteBz("8")},
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
			tev:  NewEventMarketDetailsUpdated(7, updatedBy.String()),
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
			tev:  NewEventMarketEnabled(8, updatedBy.String()),
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
			tev:  NewEventMarketDisabled(9, updatedBy.String()),
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
			tev:  NewEventMarketPermissionsUpdated(12, updatedBy.String()),
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
			tev:  NewEventMarketReqAttrUpdated(13, updatedBy.String()),
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
