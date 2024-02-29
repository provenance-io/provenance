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
		marketID, amountWithdrawn, string(destination), string(withdrawnBy))
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
	quoteBz := func(str string) []byte {
		return []byte(fmt.Sprintf("%q", str))
	}
	account := "account_____________"
	accountQ := quoteBz(account)
	cancelledBy := "cancelledBy_________"
	cancelledByQ := quoteBz(cancelledBy)
	destination := sdk.AccAddress("destination_________")
	destinationQ := quoteBz(destination.String())
	withdrawnBy := sdk.AccAddress("withdrawnBy_________")
	withdrawnByQ := quoteBz(withdrawnBy.String())
	updatedBy := "updatedBy___________"
	updatedByQ := quoteBz(updatedBy)
	coins1 := sdk.NewCoins(sdk.NewInt64Coin("onecoin", 1), sdk.NewInt64Coin("twocoin", 2))
	coins1Q := quoteBz(coins1.String())
	coins2 := sdk.NewCoins(sdk.NewInt64Coin("threecoin", 3), sdk.NewInt64Coin("fourcoin", 4))
	coins2Q := quoteBz(coins2.String())
	acoin := sdk.NewInt64Coin("acoin", 55)
	acoinQ := quoteBz(acoin.String())
	pcoin := sdk.NewInt64Coin("pcoin", 66)
	pcoinQ := quoteBz(pcoin.String())
	fcoin := sdk.NewInt64Coin("fcoin", 33)
	fcoinQ := quoteBz(fcoin.String())
	payment := &Payment{
		Source:       "source______________",
		SourceAmount: coins1,
		Target:       "target______________",
		TargetAmount: coins2,
		ExternalId:   "something external",
	}
	sourceQ := quoteBz(payment.Source)
	targetQ := quoteBz(payment.Target)
	externalIDQ := quoteBz(payment.ExternalId)
	oldTarget := "old_target__________"
	oldTargetQ := quoteBz(oldTarget)

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
			name: "EventFundsCommitted",
			tev:  NewEventFundsCommitted(account, 44, coins1, "tagTagTAG"),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventFundsCommitted",
				Attributes: []abci.EventAttribute{
					{Key: []byte("account"), Value: accountQ},
					{Key: []byte("amount"), Value: coins1Q},
					{Key: []byte("market_id"), Value: []byte("44")},
					{Key: []byte("tag"), Value: quoteBz("tagTagTAG")},
				},
			},
		},
		{
			name: "EventCommitmentReleased",
			tev:  NewEventCommitmentReleased(account, 15, coins1, "something"),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventCommitmentReleased",
				Attributes: []abci.EventAttribute{
					{Key: []byte("account"), Value: accountQ},
					{Key: []byte("amount"), Value: coins1Q},
					{Key: []byte("market_id"), Value: []byte("15")},
					{Key: []byte("tag"), Value: quoteBz("something")},
				},
			},
		},
		{
			name: "EventMarketWithdraw",
			tev:  NewEventMarketWithdraw(6, coins1, destination, withdrawnBy.String()),
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
			name: "EventMarketOrdersEnabled",
			tev:  NewEventMarketOrdersEnabled(8, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketOrdersEnabled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("8")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketOrdersDisabled",
			tev:  NewEventMarketOrdersDisabled(9, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketOrdersDisabled",
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
			name: "EventMarketCommitmentsEnabled",
			tev:  NewEventMarketCommitmentsEnabled(52, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketCommitmentsEnabled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("52")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketCommitmentsDisabled",
			tev:  NewEventMarketCommitmentsDisabled(25, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketCommitmentsDisabled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("25")},
					{Key: []byte("updated_by"), Value: updatedByQ},
				},
			},
		},
		{
			name: "EventMarketIntermediaryDenomUpdated",
			tev:  NewEventMarketIntermediaryDenomUpdated(18, updatedBy),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventMarketIntermediaryDenomUpdated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("market_id"), Value: []byte("18")},
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
		{
			name: "EventPaymentCreated",
			tev:  NewEventPaymentCreated(payment),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentCreated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: externalIDQ},
					{Key: []byte("source"), Value: sourceQ},
					{Key: []byte("source_amount"), Value: coins1Q},
					{Key: []byte("target"), Value: targetQ},
					{Key: []byte("target_amount"), Value: coins2Q},
				},
			},
		},
		{
			name: "EventPaymentUpdated",
			tev:  NewEventPaymentUpdated(payment, oldTarget),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentUpdated",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: externalIDQ},
					{Key: []byte("new_target"), Value: targetQ},
					{Key: []byte("old_target"), Value: oldTargetQ},
					{Key: []byte("source"), Value: sourceQ},
					{Key: []byte("source_amount"), Value: coins1Q},
					{Key: []byte("target_amount"), Value: coins2Q},
				},
			},
		},
		{
			name: "EventPaymentAccepted",
			tev:  NewEventPaymentAccepted(payment),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentAccepted",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: externalIDQ},
					{Key: []byte("source"), Value: sourceQ},
					{Key: []byte("source_amount"), Value: coins1Q},
					{Key: []byte("target"), Value: targetQ},
					{Key: []byte("target_amount"), Value: coins2Q},
				},
			},
		},
		{
			name: "EventPaymentRejected",
			tev:  NewEventPaymentRejected(payment),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentRejected",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: externalIDQ},
					{Key: []byte("source"), Value: sourceQ},
					{Key: []byte("target"), Value: targetQ},
				},
			},
		},
		{
			name: "EventPaymentCancelled",
			tev:  NewEventPaymentCancelled(payment),
			expEvent: sdk.Event{
				Type: "provenance.exchange.v1.EventPaymentCancelled",
				Attributes: []abci.EventAttribute{
					{Key: []byte("external_id"), Value: externalIDQ},
					{Key: []byte("source"), Value: sourceQ},
					{Key: []byte("target"), Value: targetQ},
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
