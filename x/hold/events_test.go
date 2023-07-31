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

func TestNewEventEscrowAdded(t *testing.T) {
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		amount sdk.Coins
		exp    *EventEscrowAdded
	}{
		{
			name:   "both nil",
			addr:   nil,
			amount: nil,
			exp:    &EventEscrowAdded{Address: "", Amount: ""},
		},
		{
			name:   "both empty",
			addr:   sdk.AccAddress{},
			amount: sdk.Coins{},
			exp:    &EventEscrowAdded{Address: "", Amount: ""},
		},
		{
			name:   "normal address and two denoms",
			addr:   sdk.AccAddress("normal_address______"),
			amount: sdk.NewCoins(sdk.NewInt64Coin("fingercoin", 10), sdk.NewInt64Coin("toecoin", 9)),
			exp: &EventEscrowAdded{
				Address: sdk.AccAddress("normal_address______").String(),
				Amount:  "10fingercoin,9toecoin",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := NewEventEscrowAdded(tc.addr, tc.amount)
			assert.Equal(t, tc.exp, event, "NewEventEscrowAdded")
		})
	}
}

func TestNewEventEscrowRemoved(t *testing.T) {
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		amount sdk.Coins
		exp    *EventEscrowRemoved
	}{
		{
			name:   "both nil",
			addr:   nil,
			amount: nil,
			exp:    &EventEscrowRemoved{Address: "", Amount: ""},
		},
		{
			name:   "both empty",
			addr:   sdk.AccAddress{},
			amount: sdk.Coins{},
			exp:    &EventEscrowRemoved{Address: "", Amount: ""},
		},
		{
			name:   "normal address and two denoms",
			addr:   sdk.AccAddress("normal_address______"),
			amount: sdk.NewCoins(sdk.NewInt64Coin("fingercoin", 10), sdk.NewInt64Coin("toecoin", 9)),
			exp: &EventEscrowRemoved{
				Address: sdk.AccAddress("normal_address______").String(),
				Amount:  "10fingercoin,9toecoin",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := NewEventEscrowRemoved(tc.addr, tc.amount)
			assert.Equal(t, tc.exp, event, "NewEventEscrowRemoved")
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
			name: "EventEscrowAdded",
			tev:  NewEventEscrowAdded(addr, coins),
			expEvent: sdk.Event{
				Type: "provenance.hold.v1.EventEscrowAdded",
				Attributes: []abci.EventAttribute{
					{Key: []byte("address"), Value: []byte(addrQ)},
					{Key: []byte("amount"), Value: []byte(coinsQ)},
				},
			},
		},
		{
			name: "EventEscrowRemoved",
			tev:  NewEventEscrowRemoved(addr, coins),
			expEvent: sdk.Event{
				Type: "provenance.hold.v1.EventEscrowRemoved",
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
