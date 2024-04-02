package exchange

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/testutil/assertions"
)

// assertEverythingSet asserts that the provided proto.Message can be converted to
// an untyped event with the expected type string. Then, if assertAllSet = true,
// this asserts that none of the event attributes are empty.
// Returns true on success, false if one or more things aren't right.
func assertEventContent(t *testing.T, tev proto.Message, typeString string, assertAllSet bool) bool {
	t.Helper()
	event, err := sdk.TypedEventToEvent(tev)
	if !assert.NoError(t, err, "TypedEventToEvent(%T)", tev) {
		return false
	}

	expType := "provenance.exchange.v1." + typeString
	rv := assert.Equal(t, expType, event.Type, "%T event.Type", tev)
	if !assertAllSet {
		return rv
	}

	for i, attr := range event.Attributes {
		rv = assert.NotEmpty(t, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `""`, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `0`, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `"0"`, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEmpty(t, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `""`, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `0`, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `"0"`, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
	}
	return rv
}

// assertEverythingSet asserts that the provided proto.Message can be converted to
// an untyped event with the expected type string, and a value for all fields.
// Returns true on success, false if one or more things aren't right.
func assertEverythingSet(t *testing.T, tev proto.Message, typeString string) bool {
	return assertEventContent(t, tev, typeString, true)
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
		cancelledBy string
		expected    *EventOrderCancelled
	}{
		{
			name:        "ask order",
			order:       NewOrder(11).WithAsk(&AskOrder{MarketId: 71, ExternalId: "an external identifier"}),
			cancelledBy: "CancelledBy_________",
			expected: &EventOrderCancelled{
				OrderId:     11,
				CancelledBy: "CancelledBy_________",
				MarketId:    71,
				ExternalId:  "an external identifier",
			},
		},
		{
			name:        "bid order",
			order:       NewOrder(55).WithAsk(&AskOrder{MarketId: 88, ExternalId: "another external identifier"}),
			cancelledBy: "cancelled_by________",
			expected: &EventOrderCancelled{
				OrderId:     55,
				CancelledBy: "cancelled_by________",
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

func TestNewEventFundsCommitted(t *testing.T) {
	account := sdk.AccAddress("account_____________").String()
	marketID := uint32(4444)
	amount := sdk.NewCoins(sdk.NewInt64Coin("apple", 57), sdk.NewInt64Coin("banana", 99))
	tag := "help-help-i-have-been-committed"

	var event *EventFundsCommitted
	testFunc := func() {
		event = NewEventFundsCommitted(account, marketID, amount, tag)
	}
	require.NotPanics(t, testFunc, "NewEventFundsCommitted(%q, %d, %q, %q)", account, marketID, amount, tag)
	assert.Equal(t, account, event.Account, "Account")
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, amount.String(), event.Amount, "Amount")
	assert.Equal(t, tag, event.Tag, "Tag")
	assertEverythingSet(t, event, "EventFundsCommitted")
}

func TestNewEventCommitmentReleased(t *testing.T) {
	account := sdk.AccAddress("account_____________").String()
	marketID := uint32(4444)
	amount := sdk.NewCoins(sdk.NewInt64Coin("apple", 57), sdk.NewInt64Coin("banana", 99))
	tag := "i-have-been-released"

	var event *EventCommitmentReleased
	testFunc := func() {
		event = NewEventCommitmentReleased(account, marketID, amount, tag)
	}
	require.NotPanics(t, testFunc, "NewEventCommitmentReleased(%q, %d, %q, %q)", account, marketID, amount, tag)
	assert.Equal(t, account, event.Account, "Account")
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, amount.String(), event.Amount, "Amount")
	assert.Equal(t, tag, event.Tag, "Tag")
	assertEverythingSet(t, event, "EventCommitmentReleased")
}

func TestNewEventMarketWithdraw(t *testing.T) {
	marketID := uint32(55)
	amountWithdrawn := sdk.NewCoins(sdk.NewInt64Coin("mine", 188382), sdk.NewInt64Coin("yours", 3))
	destination := sdk.AccAddress("destination_________")
	withdrawnBy := sdk.AccAddress("withdrawnBy_________").String()

	var event *EventMarketWithdraw
	testFunc := func() {
		event = NewEventMarketWithdraw(marketID, amountWithdrawn, destination, withdrawnBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketWithdraw(%d, %q, %q, %q)",
		marketID, amountWithdrawn, string(destination), withdrawnBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, amountWithdrawn.String(), event.Amount, "Amount")
	assert.Equal(t, destination.String(), event.Destination, "Destination")
	assert.Equal(t, withdrawnBy, event.WithdrawnBy, "WithdrawnBy")
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

func TestNewEventMarketAcceptingOrdersUpdated(t *testing.T) {
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
			expected:  NewEventMarketOrdersEnabled(33, someAddr),
		},
		{
			name:      "disabled",
			marketID:  556,
			updatedBy: someAddr,
			isActive:  false,
			expected:  NewEventMarketOrdersDisabled(556, someAddr),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event proto.Message
			testFunc := func() {
				event = NewEventMarketAcceptingOrdersUpdated(tc.marketID, tc.updatedBy, tc.isActive)
			}
			require.NotPanics(t, testFunc, "NewEventMarketAcceptingOrdersUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy, tc.isActive)
			assert.Equal(t, tc.expected, event, "NewEventMarketAcceptingOrdersUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy, tc.isActive)
		})
	}
}

func TestNewEventMarketOrdersEnabled(t *testing.T) {
	marketID := uint32(919)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketOrdersEnabled
	testFunc := func() {
		event = NewEventMarketOrdersEnabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketOrdersEnabled(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketOrdersEnabled")
}

func TestNewEventMarketOrdersDisabled(t *testing.T) {
	marketID := uint32(5555)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketOrdersDisabled
	testFunc := func() {
		event = NewEventMarketOrdersDisabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketOrdersDisabled(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketOrdersDisabled")
}

func TestNewEventMarketUserSettleUpdated(t *testing.T) {
	someAddr := sdk.AccAddress("some_address________").String()

	tests := []struct {
		name      string
		marketID  uint32
		updatedBy string
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
				tc.marketID, tc.updatedBy, tc.isAllowed)
			assert.Equal(t, tc.expected, event, "NewEventMarketUserSettleUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy, tc.isAllowed)
		})
	}
}

func TestNewEventMarketUserSettleEnabled(t *testing.T) {
	marketID := uint32(123)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketUserSettleEnabled
	testFunc := func() {
		event = NewEventMarketUserSettleEnabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketUserSettleEnabled(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketUserSettleEnabled")
}

func TestNewEventMarketUserSettleDisabled(t *testing.T) {
	marketID := uint32(123)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketUserSettleDisabled
	testFunc := func() {
		event = NewEventMarketUserSettleDisabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketUserSettleDisabled(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketUserSettleDisabled")
}

func TestNewEventMarketAcceptingCommitmentsUpdated(t *testing.T) {
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	tests := []struct {
		name      string
		marketID  uint32
		updatedBy string
		isAllowed bool
		expected  proto.Message
	}{
		{
			name:      "enabled",
			marketID:  575,
			updatedBy: updatedBy,
			isAllowed: true,
			expected:  NewEventMarketCommitmentsEnabled(575, updatedBy),
		},
		{
			name:      "disabled",
			marketID:  406,
			updatedBy: updatedBy,
			isAllowed: false,
			expected:  NewEventMarketCommitmentsDisabled(406, updatedBy),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event proto.Message
			testFunc := func() {
				event = NewEventMarketAcceptingCommitmentsUpdated(tc.marketID, tc.updatedBy, tc.isAllowed)
			}
			require.NotPanics(t, testFunc, "NewEventMarketAcceptingCommitmentsUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy, tc.isAllowed)
			assert.Equal(t, tc.expected, event, "NewEventMarketAcceptingCommitmentsUpdated(%d, %q, %t) result",
				tc.marketID, tc.updatedBy, tc.isAllowed)
		})
	}
}

func TestNewEventMarketCommitmentsEnabled(t *testing.T) {
	marketID := uint32(4541)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketCommitmentsEnabled
	testFunc := func() {
		event = NewEventMarketCommitmentsEnabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketCommitmentsEnabled(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketCommitmentsEnabled")
}

func TestNewEventMarketCommitmentsDisabled(t *testing.T) {
	marketID := uint32(4541)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketCommitmentsDisabled
	testFunc := func() {
		event = NewEventMarketCommitmentsDisabled(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketCommitmentsDisabled(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketCommitmentsDisabled")
}

func TestNewEventMarketIntermediaryDenomUpdated(t *testing.T) {
	marketID := uint32(4541)
	updatedBy := sdk.AccAddress("updatedBy___________").String()

	var event *EventMarketIntermediaryDenomUpdated
	testFunc := func() {
		event = NewEventMarketIntermediaryDenomUpdated(marketID, updatedBy)
	}
	require.NotPanics(t, testFunc, "NewEventMarketIntermediaryDenomUpdated(%d, %q)", marketID, updatedBy)
	assert.Equal(t, marketID, event.MarketId, "MarketId")
	assert.Equal(t, updatedBy, event.UpdatedBy, "UpdatedBy")
	assertEverythingSet(t, event, "EventMarketIntermediaryDenomUpdated")
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

// newTestPayment creates a new Payment using the provided values.
func newTestPayment(t *testing.T, source, sourceAmount, target, targetAmount, externalID string) *Payment {
	t.Helper()
	rv := &Payment{
		Source:     source,
		Target:     target,
		ExternalId: externalID,
	}
	var err error
	rv.SourceAmount, err = sdk.ParseCoinsNormalized(sourceAmount)
	require.NoError(t, err, "source amount: ParseCoinsNormalized(%q)", sourceAmount)
	rv.TargetAmount, err = sdk.ParseCoinsNormalized(targetAmount)
	require.NoError(t, err, "target amount: ParseCoinsNormalized(%q)", targetAmount)
	return rv
}

func TestEventPaymentCreated(t *testing.T) {
	tests := []struct {
		name      string
		payment   *Payment
		expected  *EventPaymentCreated
		expAllSet bool
	}{
		{
			name:    "all payment fields have content",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentCreated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				Target:       "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:    "zero source amount",
			payment: newTestPayment(t, "source_addr", "", "target_addr", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentCreated{
				Source:       "source_addr",
				SourceAmount: "",
				Target:       "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:    "no target",
			payment: newTestPayment(t, "source_addr", "312strawberry", "", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentCreated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				Target:       "",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:    "zero target amount",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "", "just_some_identifier"),
			expected: &EventPaymentCreated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				Target:       "target_addr",
				TargetAmount: "",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:    "empty external id",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", ""),
			expected: &EventPaymentCreated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				Target:       "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventPaymentCreated
			testFunc := func() {
				event = NewEventPaymentCreated(tc.payment)
			}
			require.NotPanics(t, testFunc, "NewEventPaymentCreated")
			assert.Equal(t, tc.expected, event, "NewEventPaymentCreated result")
			assertEventContent(t, event, "EventPaymentCreated", tc.expAllSet)
		})
	}
}

func TestEventPaymentUpdated(t *testing.T) {
	tests := []struct {
		name      string
		payment   *Payment
		oldTarget string
		expected  *EventPaymentUpdated
		expAllSet bool
	}{
		{
			name:      "all payment fields have content",
			payment:   newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", "just_some_identifier"),
			oldTarget: "old_target_addr",
			expected: &EventPaymentUpdated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				OldTarget:    "old_target_addr",
				NewTarget:    "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:      "zero source amount",
			payment:   newTestPayment(t, "source_addr", "", "target_addr", "7tangerine", "just_some_identifier"),
			oldTarget: "old_target_addr",
			expected: &EventPaymentUpdated{
				Source:       "source_addr",
				SourceAmount: "",
				OldTarget:    "old_target_addr",
				NewTarget:    "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:      "no old target",
			payment:   newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", "just_some_identifier"),
			oldTarget: "",
			expected: &EventPaymentUpdated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				OldTarget:    "",
				NewTarget:    "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:      "no new target",
			payment:   newTestPayment(t, "source_addr", "312strawberry", "", "7tangerine", "just_some_identifier"),
			oldTarget: "old_target_addr",
			expected: &EventPaymentUpdated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				OldTarget:    "old_target_addr",
				NewTarget:    "",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:      "zero target amount",
			payment:   newTestPayment(t, "source_addr", "312strawberry", "target_addr", "", "just_some_identifier"),
			oldTarget: "old_target_addr",
			expected: &EventPaymentUpdated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				OldTarget:    "old_target_addr",
				NewTarget:    "target_addr",
				TargetAmount: "",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:      "empty external id",
			payment:   newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", ""),
			oldTarget: "old_target_addr",
			expected: &EventPaymentUpdated{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				OldTarget:    "old_target_addr",
				NewTarget:    "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventPaymentUpdated
			testFunc := func() {
				event = NewEventPaymentUpdated(tc.payment, tc.oldTarget)
			}
			require.NotPanics(t, testFunc, "NewEventPaymentUpdated")
			assert.Equal(t, tc.expected, event, "NewEventPaymentUpdated result")
			assertEventContent(t, event, "EventPaymentUpdated", tc.expAllSet)
		})
	}
}

func TestEventPaymentAccepted(t *testing.T) {
	tests := []struct {
		name      string
		payment   *Payment
		expected  *EventPaymentAccepted
		expAllSet bool
	}{
		{
			name:    "all payment fields have content",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentAccepted{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				Target:       "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:    "zero source amount",
			payment: newTestPayment(t, "source_addr", "", "target_addr", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentAccepted{
				Source:       "source_addr",
				SourceAmount: "",
				Target:       "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:    "no target",
			payment: newTestPayment(t, "source_addr", "312strawberry", "", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentAccepted{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				Target:       "",
				TargetAmount: "7tangerine",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:    "zero target amount",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "", "just_some_identifier"),
			expected: &EventPaymentAccepted{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				Target:       "target_addr",
				TargetAmount: "",
				ExternalId:   "just_some_identifier",
			},
		},
		{
			name:    "empty external id",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", ""),
			expected: &EventPaymentAccepted{
				Source:       "source_addr",
				SourceAmount: "312strawberry",
				Target:       "target_addr",
				TargetAmount: "7tangerine",
				ExternalId:   "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventPaymentAccepted
			testFunc := func() {
				event = NewEventPaymentAccepted(tc.payment)
			}
			require.NotPanics(t, testFunc, "NewEventPaymentAccepted")
			assert.Equal(t, tc.expected, event, "NewEventPaymentAccepted result")
			assertEventContent(t, event, "EventPaymentAccepted", tc.expAllSet)
		})
	}
}

func TestEventPaymentRejected(t *testing.T) {
	tests := []struct {
		name      string
		payment   *Payment
		expected  *EventPaymentRejected
		expAllSet bool
	}{
		{
			name:    "all payment fields have content",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentRejected{
				Source:     "source_addr",
				Target:     "target_addr",
				ExternalId: "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:    "zero source amount",
			payment: newTestPayment(t, "source_addr", "", "target_addr", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentRejected{
				Source:     "source_addr",
				Target:     "target_addr",
				ExternalId: "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:    "no target",
			payment: newTestPayment(t, "source_addr", "312strawberry", "", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentRejected{
				Source:     "source_addr",
				Target:     "",
				ExternalId: "just_some_identifier",
			},
		},
		{
			name:    "zero target amount",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "", "just_some_identifier"),
			expected: &EventPaymentRejected{
				Source:     "source_addr",
				Target:     "target_addr",
				ExternalId: "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:    "empty external id",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", ""),
			expected: &EventPaymentRejected{
				Source:     "source_addr",
				Target:     "target_addr",
				ExternalId: "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventPaymentRejected
			testFunc := func() {
				event = NewEventPaymentRejected(tc.payment)
			}
			require.NotPanics(t, testFunc, "NewEventPaymentRejected")
			assert.Equal(t, tc.expected, event, "NewEventPaymentRejected result")
			assertEventContent(t, event, "EventPaymentRejected", tc.expAllSet)
		})
	}
}

func TestNewEventsPaymentsRejected(t *testing.T) {
	eStringer := func(event *EventPaymentRejected) string {
		return fmt.Sprintf("%s+%q<->%s", event.Source, event.ExternalId, event.Target)
	}

	tests := []struct {
		name     string
		payments []*Payment
		expected []*EventPaymentRejected
	}{
		{
			name:     "nil payments",
			payments: nil,
			expected: []*EventPaymentRejected{},
		},
		{
			name:     "empty payments",
			payments: []*Payment{},
			expected: []*EventPaymentRejected{},
		},
		{
			name: "five payments",
			payments: []*Payment{
				newTestPayment(t, "source_addr_0", "300strawberry", "target_addr_0", "70tangerine", "just_some_identifier_0"),
				newTestPayment(t, "source_addr_1", "", "target_addr_1", "71tangerine", "just_some_identifier_1"),
				newTestPayment(t, "source_addr_2", "302strawberry", "", "72tangerine", "just_some_identifier_2"),
				newTestPayment(t, "source_addr_3", "303strawberry", "target_addr_3", "", "just_some_identifier_3"),
				newTestPayment(t, "source_addr_4", "304strawberry", "target_addr_4", "74tangerine", ""),
			},
			expected: []*EventPaymentRejected{
				{Source: "source_addr_0", Target: "target_addr_0", ExternalId: "just_some_identifier_0"},
				{Source: "source_addr_1", Target: "target_addr_1", ExternalId: "just_some_identifier_1"},
				{Source: "source_addr_2", Target: "", ExternalId: "just_some_identifier_2"},
				{Source: "source_addr_3", Target: "target_addr_3", ExternalId: "just_some_identifier_3"},
				{Source: "source_addr_4", Target: "target_addr_4", ExternalId: ""},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []*EventPaymentRejected
			testFunc := func() {
				actual = NewEventsPaymentsRejected(tc.payments)
			}
			require.NotPanics(t, testFunc, "NewEventsPaymentsRejected")
			assertEqualSlice(t, tc.expected, actual, eStringer, "NewEventsPaymentsRejected")
		})
	}
}

func TestEventPaymentCancelled(t *testing.T) {
	tests := []struct {
		name      string
		payment   *Payment
		expected  *EventPaymentCancelled
		expAllSet bool
	}{
		{
			name:    "all payment fields have content",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentCancelled{
				Source:     "source_addr",
				Target:     "target_addr",
				ExternalId: "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:    "zero source amount",
			payment: newTestPayment(t, "source_addr", "", "target_addr", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentCancelled{
				Source:     "source_addr",
				Target:     "target_addr",
				ExternalId: "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:    "no target",
			payment: newTestPayment(t, "source_addr", "312strawberry", "", "7tangerine", "just_some_identifier"),
			expected: &EventPaymentCancelled{
				Source:     "source_addr",
				Target:     "",
				ExternalId: "just_some_identifier",
			},
		},
		{
			name:    "zero target amount",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "", "just_some_identifier"),
			expected: &EventPaymentCancelled{
				Source:     "source_addr",
				Target:     "target_addr",
				ExternalId: "just_some_identifier",
			},
			expAllSet: true,
		},
		{
			name:    "empty external id",
			payment: newTestPayment(t, "source_addr", "312strawberry", "target_addr", "7tangerine", ""),
			expected: &EventPaymentCancelled{
				Source:     "source_addr",
				Target:     "target_addr",
				ExternalId: "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event *EventPaymentCancelled
			testFunc := func() {
				event = NewEventPaymentCancelled(tc.payment)
			}
			require.NotPanics(t, testFunc, "NewEventPaymentCancelled")
			assert.Equal(t, tc.expected, event, "NewEventPaymentCancelled result")
			assertEventContent(t, event, "EventPaymentCancelled", tc.expAllSet)
		})
	}
}

func TestNewEventsPaymentsCancelled(t *testing.T) {
	eStringer := func(event *EventPaymentCancelled) string {
		return fmt.Sprintf("%s+%q<->%s", event.Source, event.ExternalId, event.Target)
	}

	tests := []struct {
		name     string
		payments []*Payment
		expected []*EventPaymentCancelled
	}{
		{
			name:     "nil payments",
			payments: nil,
			expected: []*EventPaymentCancelled{},
		},
		{
			name:     "empty payments",
			payments: []*Payment{},
			expected: []*EventPaymentCancelled{},
		},
		{
			name: "five payments",
			payments: []*Payment{
				newTestPayment(t, "source_addr_0", "300strawberry", "target_addr_0", "70tangerine", "just_some_identifier_0"),
				newTestPayment(t, "source_addr_1", "", "target_addr_1", "71tangerine", "just_some_identifier_1"),
				newTestPayment(t, "source_addr_2", "302strawberry", "", "72tangerine", "just_some_identifier_2"),
				newTestPayment(t, "source_addr_3", "303strawberry", "target_addr_3", "", "just_some_identifier_3"),
				newTestPayment(t, "source_addr_4", "304strawberry", "target_addr_4", "74tangerine", ""),
			},
			expected: []*EventPaymentCancelled{
				{Source: "source_addr_0", Target: "target_addr_0", ExternalId: "just_some_identifier_0"},
				{Source: "source_addr_1", Target: "target_addr_1", ExternalId: "just_some_identifier_1"},
				{Source: "source_addr_2", Target: "", ExternalId: "just_some_identifier_2"},
				{Source: "source_addr_3", Target: "target_addr_3", ExternalId: "just_some_identifier_3"},
				{Source: "source_addr_4", Target: "target_addr_4", ExternalId: ""},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []*EventPaymentCancelled
			testFunc := func() {
				actual = NewEventsPaymentsCancelled(tc.payments)
			}
			require.NotPanics(t, testFunc, "NewEventsPaymentsCancelled")
			assertEqualSlice(t, tc.expected, actual, eStringer, "NewEventsPaymentsCancelled")
		})
	}
}

func TestTypedEventToEvent(t *testing.T) {
	quoteStr := func(str string) string {
		return fmt.Sprintf("%q", str)
	}
	account := "account_____________"
	accountQ := quoteStr(account)
	cancelledBy := "cancelledBy_________"
	cancelledByQ := quoteStr(cancelledBy)
	destination := sdk.AccAddress("destination_________")
	destinationQ := quoteStr(destination.String())
	withdrawnBy := sdk.AccAddress("withdrawnBy_________")
	withdrawnByQ := quoteStr(withdrawnBy.String())
	updatedBy := "updatedBy___________"
	updatedByQ := quoteStr(updatedBy)
	coins1 := sdk.NewCoins(sdk.NewInt64Coin("onecoin", 1), sdk.NewInt64Coin("twocoin", 2))
	coins1Q := quoteStr(coins1.String())
	coins2 := sdk.NewCoins(sdk.NewInt64Coin("threecoin", 3), sdk.NewInt64Coin("fourcoin", 4))
	coins2Q := quoteStr(coins2.String())
	acoin := sdk.NewInt64Coin("acoin", 55)
	acoinQ := quoteStr(acoin.String())
	pcoin := sdk.NewInt64Coin("pcoin", 66)
	pcoinQ := quoteStr(pcoin.String())
	fcoin := sdk.NewInt64Coin("fcoin", 33)
	fcoinQ := quoteStr(fcoin.String())
	payment := &Payment{
		Source:       "source______________",
		SourceAmount: coins1,
		Target:       "target______________",
		TargetAmount: coins2,
		ExternalId:   "something external",
	}
	sourceQ := quoteStr(payment.Source)
	targetQ := quoteStr(payment.Target)
	externalIDQ := quoteStr(payment.ExternalId)
	oldTarget := "old_target__________"
	oldTargetQ := quoteStr(oldTarget)

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
					{Key: "external_id", Value: quoteStr("stuff")},
					{Key: "market_id", Value: "88"},
					{Key: "order_id", Value: quoteStr("1")},
					{Key: "order_type", Value: quoteStr("ask")},
				},
			},
		},
		{
			name: "EventOrderCreated bid",
			tev:  NewEventOrderCreated(NewOrder(2).WithBid(&BidOrder{MarketId: 77, ExternalId: "something else"})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCreated",
				Attributes: []abci.EventAttribute{
					{Key: "external_id", Value: quoteStr("something else")},
					{Key: "market_id", Value: "77"},
					{Key: "order_id", Value: quoteStr("2")},
					{Key: "order_type", Value: quoteStr("bid")},
				},
			},
		},
		{
			name: "EventOrderCancelled ask",
			tev:  NewEventOrderCancelled(NewOrder(3).WithAsk(&AskOrder{MarketId: 66, ExternalId: "outside 8"}), cancelledBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCancelled",
				Attributes: []abci.EventAttribute{
					{Key: "cancelled_by", Value: cancelledByQ},
					{Key: "external_id", Value: quoteStr("outside 8")},
					{Key: "market_id", Value: "66"},
					{Key: "order_id", Value: quoteStr("3")},
				},
			},
		},
		{
			name: "EventOrderCancelled bid",
			tev:  NewEventOrderCancelled(NewOrder(3).WithBid(&BidOrder{MarketId: 55, ExternalId: "outside 8"}), cancelledBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderCancelled",
				Attributes: []abci.EventAttribute{
					{Key: "cancelled_by", Value: cancelledByQ},
					{Key: "external_id", Value: quoteStr("outside 8")},
					{Key: "market_id", Value: "55"},
					{Key: "order_id", Value: quoteStr("3")},
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
					{Key: "assets", Value: acoinQ},
					{Key: "external_id", Value: quoteStr("eeeeiiiiiddddd")},
					{Key: "fees", Value: fcoinQ},
					{Key: "market_id", Value: "33"},
					{Key: "order_id", Value: quoteStr("4")},
					{Key: "price", Value: pcoinQ},
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
					{Key: "assets", Value: acoinQ},
					{Key: "external_id", Value: quoteStr("that one thing")},
					{Key: "fees", Value: fcoinQ},
					{Key: "market_id", Value: "44"},
					{Key: "order_id", Value: quoteStr("104")},
					{Key: "price", Value: pcoinQ},
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
					{Key: "assets", Value: acoinQ},
					{Key: "external_id", Value: quoteStr("12345")},
					{Key: "fees", Value: fcoinQ},
					{Key: "market_id", Value: "22"},
					{Key: "order_id", Value: quoteStr("5")},
					{Key: "price", Value: pcoinQ},
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
					{Key: "assets", Value: acoinQ},
					{Key: "external_id", Value: quoteStr("67890")},
					{Key: "fees", Value: fcoinQ},
					{Key: "market_id", Value: "11"},
					{Key: "order_id", Value: quoteStr("5")},
					{Key: "price", Value: pcoinQ},
				},
			},
		},
		{
			name: "EventOrderExternalIDUpdated ask",
			tev:  NewEventOrderExternalIDUpdated(NewOrder(8).WithAsk(&AskOrder{MarketId: 99, ExternalId: "yellow"})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderExternalIDUpdated",
				Attributes: []abci.EventAttribute{
					{Key: "external_id", Value: quoteStr("yellow")},
					{Key: "market_id", Value: "99"},
					{Key: "order_id", Value: quoteStr("8")},
				},
			},
		},
		{
			name: "EventOrderExternalIDUpdated bid",
			tev:  NewEventOrderExternalIDUpdated(NewOrder(8).WithBid(&BidOrder{MarketId: 111, ExternalId: "yellow"})),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventOrderExternalIDUpdated",
				Attributes: []abci.EventAttribute{
					{Key: "external_id", Value: quoteStr("yellow")},
					{Key: "market_id", Value: "111"},
					{Key: "order_id", Value: quoteStr("8")},
				},
			},
		},
		{
			name: "EventFundsCommitted",
			tev:  NewEventFundsCommitted(account, 44, coins1, "tagTagTAG"),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventFundsCommitted",
				Attributes: []abci.EventAttribute{
					{Key: "account", Value: accountQ},
					{Key: "amount", Value: coins1Q},
					{Key: "market_id", Value: "44"},
					{Key: "tag", Value: quoteStr("tagTagTAG")},
				},
			},
		},
		{
			name: "EventCommitmentReleased",
			tev:  NewEventCommitmentReleased(account, 15, coins1, "something"),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventCommitmentReleased",
				Attributes: []abci.EventAttribute{
					{Key: "account", Value: accountQ},
					{Key: "amount", Value: coins1Q},
					{Key: "market_id", Value: "15"},
					{Key: "tag", Value: quoteStr("something")},
				},
			},
		},
		{
			name: "EventMarketWithdraw",
			tev:  NewEventMarketWithdraw(6, coins1, destination, withdrawnBy.String()),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketWithdraw",
				Attributes: []abci.EventAttribute{
					{Key: "amount", Value: coins1Q},
					{Key: "destination", Value: destinationQ},
					{Key: "market_id", Value: "6"},
					{Key: "withdrawn_by", Value: withdrawnByQ},
				},
			},
		},
		{
			name: "EventMarketDetailsUpdated",
			tev:  NewEventMarketDetailsUpdated(7, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketDetailsUpdated",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "7"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketOrdersEnabled",
			tev:  NewEventMarketOrdersEnabled(8, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketOrdersEnabled",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "8"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketOrdersDisabled",
			tev:  NewEventMarketOrdersDisabled(9, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketOrdersDisabled",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "9"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketUserSettleEnabled",
			tev:  NewEventMarketUserSettleEnabled(10, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketUserSettleEnabled",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "10"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketUserSettleDisabled",
			tev:  NewEventMarketUserSettleDisabled(11, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketUserSettleDisabled",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "11"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketCommitmentsEnabled",
			tev:  NewEventMarketCommitmentsEnabled(52, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketCommitmentsEnabled",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "52"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketCommitmentsDisabled",
			tev:  NewEventMarketCommitmentsDisabled(25, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketCommitmentsDisabled",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "25"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketIntermediaryDenomUpdated",
			tev:  NewEventMarketIntermediaryDenomUpdated(18, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketIntermediaryDenomUpdated",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "18"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketPermissionsUpdated",
			tev:  NewEventMarketPermissionsUpdated(12, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketPermissionsUpdated",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "12"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketReqAttrUpdated",
			tev:  NewEventMarketReqAttrUpdated(13, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketReqAttrUpdated",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "13"},
					{Key: "updated_by", Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketCreated",
			tev:  NewEventMarketCreated(14),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketCreated",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "14"},
				},
			},
		},
		{
			name: "EventMarketFeesUpdated",
			tev:  NewEventMarketFeesUpdated(15),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketFeesUpdated",
				Attributes: []abci.EventAttribute{
					{Key: "market_id", Value: "15"},
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
		{
			name: "EventPaymentCreated",
			tev:  NewEventPaymentCreated(payment),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentCreated",
				Attributes: []abci.EventAttribute{
					{Key: "external_id", Value: externalIDQ},
					{Key: "source", Value: sourceQ},
					{Key: "source_amount", Value: coins1Q},
					{Key: "target", Value: targetQ},
					{Key: "target_amount", Value: coins2Q},
				},
			},
		},
		{
			name: "EventPaymentUpdated",
			tev:  NewEventPaymentUpdated(payment, oldTarget),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentUpdated",
				Attributes: []abci.EventAttribute{
					{Key: "external_id", Value: externalIDQ},
					{Key: "new_target", Value: targetQ},
					{Key: "old_target", Value: oldTargetQ},
					{Key: "source", Value: sourceQ},
					{Key: "source_amount", Value: coins1Q},
					{Key: "target_amount", Value: coins2Q},
				},
			},
		},
		{
			name: "EventPaymentAccepted",
			tev:  NewEventPaymentAccepted(payment),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentAccepted",
				Attributes: []abci.EventAttribute{
					{Key: "external_id", Value: externalIDQ},
					{Key: "source", Value: sourceQ},
					{Key: "source_amount", Value: coins1Q},
					{Key: "target", Value: targetQ},
					{Key: "target_amount", Value: coins2Q},
				},
			},
		},
		{
			name: "EventPaymentRejected",
			tev:  NewEventPaymentRejected(payment),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentRejected",
				Attributes: []abci.EventAttribute{
					{Key: "external_id", Value: externalIDQ},
					{Key: "source", Value: sourceQ},
					{Key: "target", Value: targetQ},
				},
			},
		},
		{
			name: "EventPaymentCancelled",
			tev:  NewEventPaymentCancelled(payment),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentCancelled",
				Attributes: []abci.EventAttribute{
					{Key: "external_id", Value: externalIDQ},
					{Key: "source", Value: sourceQ},
					{Key: "target", Value: targetQ},
				},
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
