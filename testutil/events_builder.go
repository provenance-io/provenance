package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"
)

// EventsBuilder helps put together a list of events for testing purposes.
// Create one using NewEventsBuilder, then add events using the methods like AddEvent or AddSendCoins.
// Once everything has been added, use Build() to get the final list of events.
type EventsBuilder struct {
	Events sdk.Events
	T      *testing.T
}

// NewEventsBuilder creates a new EventsBuilder that will use the provided T for error checking.
// If a T is provided, errors will cause test failures, otherwise errors will cause panics.
func NewEventsBuilder(t *testing.T) *EventsBuilder {
	return &EventsBuilder{T: t}
}

// Build returns the list of Events built so far.
func (b *EventsBuilder) Build() sdk.Events {
	return b.Events
}

// BuildABCI returns the list of Events built so far as a list of abci.Event types.
func (b *EventsBuilder) BuildABCI() []abci.Event {
	rv := make([]abci.Event, len(b.Events))
	for i, event := range b.Events {
		rv[i] = abci.Event(event)
	}
	return rv
}

// AddEvent adds one or more sdk.Event entries to this builder.
func (b *EventsBuilder) AddEvent(events ...sdk.Event) *EventsBuilder {
	b.Events = append(b.Events, events...)
	return b
}

// AddEvents adds all of the provided sdk.Events to this builder.
func (b *EventsBuilder) AddEvents(events sdk.Events) *EventsBuilder {
	b.Events = append(b.Events, events...)
	return b
}

// AddTypedEvent converts each of the provided events into an sdk.Event then adds it to this builder.
func (b *EventsBuilder) AddTypedEvent(tevs ...proto.Message) *EventsBuilder {
	if b.T != nil {
		b.T.Helper()
	}
	var err error
	events := make(sdk.Events, len(tevs))
	for i, tev := range tevs {
		events[i], err = sdk.TypedEventToEvent(tev)
		switch {
		case b.T != nil:
			require.NoError(b.T, err, "TypedEventToEvent(%#v)", tev)
		case err != nil:
			panic(err)
		}
	}
	return b.AddEvents(events)
}

// AddSendCoins adds the events emitted during SendCoins.
func (b *EventsBuilder) AddSendCoins(from, to sdk.AccAddress, amount sdk.Coins) *EventsBuilder {
	return b.AddEvents(SendCoinsEvents(from, to, amount))
}

// AddSendCoinsStrs adds the events emitted during SendCoins, but takes in the values as strings.
func (b *EventsBuilder) AddSendCoinsStrs(from, to, amount string) *EventsBuilder {
	return b.AddEvents(SendCoinsEventsStrs(from, to, amount))
}

// SendCoinsEvents creates the events emitted during SendCoins.
func SendCoinsEvents(from, to sdk.AccAddress, amount sdk.Coins) sdk.Events {
	return SendCoinsEventsStrs(from.String(), to.String(), amount.String())
}

// SendCoinsEventsStrs creates the events emitted during SendCoins, but takes in the values as strings.
func SendCoinsEventsStrs(from, to, amount string) sdk.Events {
	return sdk.Events{
		{Type: "coin_spent", Attributes: []abci.EventAttribute{
			{Key: "spender", Value: from},
			{Key: "amount", Value: amount},
		}},
		{Type: "coin_received", Attributes: []abci.EventAttribute{
			{Key: "receiver", Value: to},
			{Key: "amount", Value: amount},
		}},
		{Type: "transfer", Attributes: []abci.EventAttribute{
			{Key: "recipient", Value: to},
			{Key: "sender", Value: from},
			{Key: "amount", Value: amount},
		}},
		{Type: "message", Attributes: []abci.EventAttribute{
			{Key: "sender", Value: from},
		}},
	}
}

// AddFailedSendCoins adds the events emitted during SendCoins when there's an error from the send restrictions.
func (b *EventsBuilder) AddFailedSendCoins(from sdk.AccAddress, amount sdk.Coins) *EventsBuilder {
	return b.AddEvents(FailedSendCoinsEvents(from, amount))
}

// AddFailedSendCoinsStrs adds the events emitted during SendCoins when there's
// an error from the send restrictions, but takes in the values as strings.
func (b *EventsBuilder) AddFailedSendCoinsStrs(from, amount string) *EventsBuilder {
	return b.AddEvents(FailedSendCoinsEventsStrs(from, amount))
}

// FailedSendCoinsEvents creates the events emitted during SendCoins when there's an error from the send restrictions.
func FailedSendCoinsEvents(from sdk.AccAddress, amount sdk.Coins) sdk.Events {
	return FailedSendCoinsEventsStrs(from.String(), amount.String())
}

// FailedSendCoinsEventsStrs creates the events emitted during SendCoins when there's
// an error from the send restrictions, but takes in the values as strings.
func FailedSendCoinsEventsStrs(from, amount string) sdk.Events {
	return sdk.Events{{Type: "coin_spent", Attributes: []abci.EventAttribute{
		{Key: "spender", Value: from},
		{Key: "amount", Value: amount},
	}}}
}

// AddMintCoins adds the events emitted during MintCoins.
func (b *EventsBuilder) AddMintCoins(moduleName string, amount sdk.Coins) *EventsBuilder {
	return b.AddEvents(MintCoinsEvents(moduleName, amount))
}

// AddMintCoinsStrs adds the events emitted during MintCoins, but takes in the values as strings.
// Note that the first argument should be the bech32 address string of the module account (not the module name).
func (b *EventsBuilder) AddMintCoinsStrs(moduleAddr, amount string) *EventsBuilder {
	return b.AddEvents(MintCoinsEventsStrs(moduleAddr, amount))
}

// MintCoinsEvents creates the events emitted during MintCoins.
func MintCoinsEvents(moduleName string, amount sdk.Coins) sdk.Events {
	return MintCoinsEventsStrs(authtypes.NewModuleAddress(moduleName).String(), amount.String())
}

// MintCoinsEventsStrs creates the events emitted during MintCoins, but takes in the values as strings.
// Note that the first argument should be the bech32 address string of the module account (not the module name).
func MintCoinsEventsStrs(moduleAddr, amount string) sdk.Events {
	return sdk.Events{
		{Type: "coin_received", Attributes: []abci.EventAttribute{
			{Key: "receiver", Value: moduleAddr},
			{Key: "amount", Value: amount},
		}},
		{Type: "coinbase", Attributes: []abci.EventAttribute{
			{Key: "minter", Value: moduleAddr},
			{Key: "amount", Value: amount},
		}},
	}
}

// AddBurnCoins adds the events emitted during BurnCoins.
func (b *EventsBuilder) AddBurnCoins(moduleName string, amount sdk.Coins) *EventsBuilder {
	return b.AddEvents(BurnCoinsEvents(moduleName, amount))
}

// AddBurnCoinsStrs adds the events emitted during BurnCoins, but takes in the values as strings.
// Note that the first argument should be the bech32 address string of the module account (not the module name).
func (b *EventsBuilder) AddBurnCoinsStrs(moduleAddr, amount string) *EventsBuilder {
	return b.AddEvents(BurnCoinsEventsStrs(moduleAddr, amount))
}

// BurnCoinsEvents creates the events emitted during BurnCoins.
func BurnCoinsEvents(moduleName string, amount sdk.Coins) sdk.Events {
	return BurnCoinsEventsStrs(authtypes.NewModuleAddress(moduleName).String(), amount.String())
}

// BurnCoinsEventsStrs creates the events emitted during BurnCoins, but takes in the values as strings.
// Note that the first argument should be the bech32 address string of the module account (not the module name).
func BurnCoinsEventsStrs(moduleAddr, amount string) sdk.Events {
	return sdk.Events{
		{Type: "coin_spent", Attributes: []abci.EventAttribute{
			{Key: "spender", Value: moduleAddr},
			{Key: "amount", Value: amount},
		}},
		{Type: "burn", Attributes: []abci.EventAttribute{
			{Key: "burner", Value: moduleAddr},
			{Key: "amount", Value: amount},
		}},
	}
}
