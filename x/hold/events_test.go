package hold

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

func TestNewEventHoldAdded(t *testing.T) {
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		amount sdk.Coins
		reason string
		exp    *EventHoldAdded
	}{
		{
			name:   "both nil",
			addr:   nil,
			amount: nil,
			exp:    &EventHoldAdded{Address: "", Amount: ""},
		},
		{
			name:   "both empty",
			addr:   sdk.AccAddress{},
			amount: sdk.Coins{},
			exp:    &EventHoldAdded{Address: "", Amount: ""},
		},
		{
			name:   "normal address and two denoms",
			addr:   sdk.AccAddress("normal_address______"),
			amount: sdk.NewCoins(sdk.NewInt64Coin("fingercoin", 10), sdk.NewInt64Coin("toecoin", 9)),
			exp: &EventHoldAdded{
				Address: sdk.AccAddress("normal_address______").String(),
				Amount:  "10fingercoin,9toecoin",
			},
		},
		{
			name:   "only a reason",
			reason: "this is a test reason",
			exp:    &EventHoldAdded{Reason: "this is a test reason"},
		},
		{
			name:   "control",
			addr:   sdk.AccAddress("control_address_____"),
			amount: sdk.NewCoins(sdk.NewInt64Coin("cherry", 4)),
			reason: "control reason",
			exp: &EventHoldAdded{
				Address: sdk.AccAddress("control_address_____").String(),
				Amount:  sdk.NewCoins(sdk.NewInt64Coin("cherry", 4)).String(),
				Reason:  "control reason",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := NewEventHoldAdded(tc.addr, tc.amount, tc.reason)
			assert.Equal(t, tc.exp, event, "NewEventHoldAdded")
		})
	}
}

func TestNewEventHoldReleased(t *testing.T) {
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		amount sdk.Coins
		exp    *EventHoldReleased
	}{
		{
			name:   "both nil",
			addr:   nil,
			amount: nil,
			exp:    &EventHoldReleased{Address: "", Amount: ""},
		},
		{
			name:   "both empty",
			addr:   sdk.AccAddress{},
			amount: sdk.Coins{},
			exp:    &EventHoldReleased{Address: "", Amount: ""},
		},
		{
			name:   "normal address and two denoms",
			addr:   sdk.AccAddress("normal_address______"),
			amount: sdk.NewCoins(sdk.NewInt64Coin("fingercoin", 10), sdk.NewInt64Coin("toecoin", 9)),
			exp: &EventHoldReleased{
				Address: sdk.AccAddress("normal_address______").String(),
				Amount:  "10fingercoin,9toecoin",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := NewEventHoldReleased(tc.addr, tc.amount)
			assert.Equal(t, tc.exp, event, "NewEventHoldReleased")
		})
	}
}

func TestNewEventVestingAccountUnlocked(t *testing.T) {
	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  *EventVestingAccountUnlocked
	}{
		{
			name: "nil addr",
			addr: nil,
			exp:  &EventVestingAccountUnlocked{Address: ""},
		},
		{
			name: "empty addr",
			addr: sdk.AccAddress{},
			exp:  &EventVestingAccountUnlocked{Address: ""},
		},
		{
			name: "non-empty addr",
			addr: sdk.AccAddress("some_addr___________"),
			exp:  &EventVestingAccountUnlocked{Address: sdk.AccAddress("some_addr___________").String()},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventVestingAccountUnlocked
			testFunc := func() {
				act = NewEventVestingAccountUnlocked(tc.addr)
			}
			require.NotPanics(t, testFunc, "NewEventVestingAccountUnlocked(%q)", string(tc.addr))
			assert.Equal(t, tc.exp, act, "NewEventVestingAccountUnlocked(%q) result", string(tc.addr))
		})
	}
}

func TestTypedEventToEvent(t *testing.T) {
	addr := sdk.AccAddress("address_in_the_event")
	coins := sdk.NewCoins(sdk.NewInt64Coin("elbowcoin", 4), sdk.NewInt64Coin("kneecoin", 2))
	addrQ := fmt.Sprintf("%q", addr.String())
	coinsQ := fmt.Sprintf("%q", coins.String())

	tests := []struct {
		name     string
		tev      proto.Message
		expEvent sdk.Event
	}{
		{
			name: "EventHoldAdded",
			tev:  NewEventHoldAdded(addr, coins, "test reason"),
			expEvent: sdk.Event{
				Type: "provenance.hold.v1.EventHoldAdded",
				Attributes: []abci.EventAttribute{
					{Key: "address", Value: addrQ},
					{Key: "amount", Value: coinsQ},
					{Key: "reason", Value: `"test reason"`},
				},
			},
		},
		{
			name: "EventHoldReleased",
			tev:  NewEventHoldReleased(addr, coins),
			expEvent: sdk.Event{
				Type: "provenance.hold.v1.EventHoldReleased",
				Attributes: []abci.EventAttribute{
					{Key: "address", Value: addrQ},
					{Key: "amount", Value: coinsQ},
				},
			},
		},
		{
			name: "NewEventVestingAccountUnlocked",
			tev:  NewEventVestingAccountUnlocked(addr),
			expEvent: sdk.Event{
				Type: "provenance.hold.v1.EventVestingAccountUnlocked",
				Attributes: []abci.EventAttribute{
					{Key: "address", Value: addrQ},
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
