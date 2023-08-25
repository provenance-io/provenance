package exchange

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
			order:    nil,
			expPanic: "OrderType() missing case for <nil>",
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
	assert.Equal(t, amountWithdrawn.String(), event.AmountWithdrawn, "AmountWithdrawn")
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
	marketID := uint32(123)
	updatedBy := sdk.AccAddress("updatedBy___________")

	event := NewEventMarketUserSettleUpdated(marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy.String(), event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketUserSettleUpdated")
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

func TestNewEventCreateMarketSubmitted(t *testing.T) {
	marketID := uint32(5445)
	proposalID := uint64(88888)
	submittedBy := sdk.AccAddress("submittedBy_________")

	event := NewEventCreateMarketSubmitted(marketID, proposalID, submittedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, proposalID, event.ProposalId, "ProposalId")
	assert.Equal(t, submittedBy.String(), event.SubmittedBy, "SubmittedBy")
	assertEverythingSet(t, event, "EventCreateMarketSubmitted")
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
