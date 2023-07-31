package hold

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestNewEventHoldAdded(t *testing.T) {
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		amount sdk.Coins
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := NewEventHoldAdded(tc.addr, tc.amount)
			assert.Equal(t, tc.exp, event, "NewEventHoldAdded")
		})
	}
}

func TestNewEventHoldRemoved(t *testing.T) {
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		amount sdk.Coins
		exp    *EventHoldRemoved
	}{
		{
			name:   "both nil",
			addr:   nil,
			amount: nil,
			exp:    &EventHoldRemoved{Address: "", Amount: ""},
		},
		{
			name:   "both empty",
			addr:   sdk.AccAddress{},
			amount: sdk.Coins{},
			exp:    &EventHoldRemoved{Address: "", Amount: ""},
		},
		{
			name:   "normal address and two denoms",
			addr:   sdk.AccAddress("normal_address______"),
			amount: sdk.NewCoins(sdk.NewInt64Coin("fingercoin", 10), sdk.NewInt64Coin("toecoin", 9)),
			exp: &EventHoldRemoved{
				Address: sdk.AccAddress("normal_address______").String(),
				Amount:  "10fingercoin,9toecoin",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := NewEventHoldRemoved(tc.addr, tc.amount)
			assert.Equal(t, tc.exp, event, "NewEventHoldRemoved")
		})
	}
}

func TestTypedEventToEvent(t *testing.T) {
	addr := sdk.AccAddress("address_in_the_event")
	coins := sdk.NewCoins(sdk.NewInt64Coin("elbowcoin", 4), sdk.NewInt64Coin("kneecoin", 2))
	addrQ := fmt.Sprintf("%q", addr.String())
	coinsQ := fmt.Sprintf("%q", coins.String())

	// attrsToStrings converts the provided attributes to strings.
	// These are used to compare actual/expected so that any differences are
	// easier to understand in the test failure output.
	attrsToStrings := func(attrs []abci.EventAttribute) []string {
		rv := make([]string, len(attrs))
		for i, attr := range attrs {
			rv[i] = fmt.Sprintf("[%d]: %q = %q", i, string(attr.Key), string(attr.Value))
			if attr.Index {
				rv[i] = rv[i] + " (indexed)"
			}
		}
		return rv
	}

	tests := []struct {
		name     string
		tev      proto.Message
		expEvent sdk.Event
	}{
		{
			name: "EventHoldAdded",
			tev:  NewEventHoldAdded(addr, coins),
			expEvent: sdk.Event{
				Type: "provenance.hold.v1.EventHoldAdded",
				Attributes: []abci.EventAttribute{
					{Key: []byte("address"), Value: []byte(addrQ)},
					{Key: []byte("amount"), Value: []byte(coinsQ)},
				},
			},
		},
		{
			name: "EventHoldRemoved",
			tev:  NewEventHoldRemoved(addr, coins),
			expEvent: sdk.Event{
				Type: "provenance.hold.v1.EventHoldRemoved",
				Attributes: []abci.EventAttribute{
					{Key: []byte("address"), Value: []byte(addrQ)},
					{Key: []byte("amount"), Value: []byte(coinsQ)},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event, err := sdk.TypedEventToEvent(tc.tev)
			require.NoError(t, err, "TypedEventToEvent error")
			if assert.NotNilf(t, event, "TypedEventToEvent result") {
				assert.Equal(t, tc.expEvent.Type, event.Type, "event type")
				expAttrs := attrsToStrings(tc.expEvent.Attributes)
				actAttrs := attrsToStrings(event.Attributes)
				assert.Equal(t, expAttrs, actAttrs, "event attributes")
			}
		})
	}
}
