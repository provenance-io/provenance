package sdk

import (
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// NoOpEventManager is an event manager that satisfies the sdk.EventManagerI interface, but does nothing.
type NoOpEventManager struct{}

var _ sdk.EventManagerI = (*NoOpEventManager)(nil)

// NewNoOpEventManager returns a new event manager that does nothing.
func NewNoOpEventManager() *NoOpEventManager {
	return &NoOpEventManager{}
}

// Events returns sdk.EmptyEvents().
func (x NoOpEventManager) Events() sdk.Events {
	// Returning sdk.EmptyEvents() here (instead of nil) to match sdk.EventManager behavior.
	return sdk.EmptyEvents()
}

// ABCIEvents returns sdk.EmptyABCIEvents().
func (x NoOpEventManager) ABCIEvents() []abci.Event {
	// Returning sdk.EmptyABCIEvents() here (instead of nil) to match sdk.EventManager behavior.
	return sdk.EmptyABCIEvents()
}

// EmitTypedEvent ignores the provided argument, does nothing, and always returns nil.
func (x NoOpEventManager) EmitTypedEvent(_ proto.Message) error {
	return nil
}

// EmitTypedEvents ignores the provided arguments, does nothing, and always returns nil.
func (x NoOpEventManager) EmitTypedEvents(_ ...proto.Message) error {
	return nil
}

// EmitEvent ignores the provided event and does nothing.
func (x NoOpEventManager) EmitEvent(_ sdk.Event) {}

// EmitEvents ignores the provided events and does nothing.
func (x NoOpEventManager) EmitEvents(_ sdk.Events) {}
