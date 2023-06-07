package types

import (
	fmt "fmt"
	"strings"
	time "time"

	proto "github.com/gogo/protobuf/proto"

	abci "github.com/tendermint/tendermint/abci/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

type TriggerID = uint64

const (
	BlockHeightPrefix = "block-height"
	BlockTimePrefix   = "block-time"
)

type TriggerEventI interface {
	proto.Message
	GetEventPrefix() string
	Validate() error
	ValidateContext(ctx sdk.Context) error
}

var _ TriggerEventI = &TransactionEvent{}
var _ TriggerEventI = &BlockHeightEvent{}
var _ TriggerEventI = &BlockTimeEvent{}

// Matches checks if two TransactionEvents have the same type and matching attributes.
func (e TransactionEvent) Matches(other abci.Event) bool {
	if e.Name != other.GetType() {
		return false
	}

	for _, attr := range e.Attributes {
		hasAttribute := false

		for _, otherAttr := range other.Attributes {
			if attr.Matches(otherAttr) {
				hasAttribute = true
				break
			}
		}

		if !hasAttribute {
			return false
		}
	}

	return true
}

// Matches checks if two Attributes have the same name and an equal value if one is specified.
func (a Attribute) Matches(other abci.EventAttribute) bool {
	if a.GetName() != string(other.GetKey()) {
		return false
	}

	if a.GetValue() != "" && a.GetValue() != string(other.GetValue()) {
		return false
	}

	return true
}

// GetEventPrefix gets the prefix for a TransactionEvent.
func (e TransactionEvent) GetEventPrefix() string {
	return e.Name
}

// Validate checks if the event data is valid.
func (e TransactionEvent) Validate() error {
	if strings.TrimSpace(e.Name) == "" {
		return fmt.Errorf("empty event name")
	}
	for _, attribute := range e.Attributes {
		if strings.TrimSpace(attribute.Name) == "" {
			return fmt.Errorf("empty attribute name")
		}
	}
	return nil
}

// Validate checks if this event is valid with the current context.
func (e TransactionEvent) ValidateContext(_ sdk.Context) error {
	return nil
}

// GetEventPrefix gets the prefix for a BlockHeightEvent.
func (e BlockHeightEvent) GetEventPrefix() string {
	return BlockHeightPrefix
}

// Validate checks if the event data is valid.
func (e BlockHeightEvent) Validate() error {
	return nil
}

// Validate checks if this event is valid with the current context.
func (e BlockHeightEvent) ValidateContext(ctx sdk.Context) error {
	if e.BlockHeight <= uint64(ctx.BlockHeight()) {
		return ErrInvalidBlockHeight
	}
	return nil
}

// GetEventPrefix gets the prefix for a BlockTimeEvent.
func (e BlockTimeEvent) GetEventPrefix() string {
	return BlockTimePrefix
}

// Validate checks if the event data is valid.
func (e BlockTimeEvent) Validate() error {
	return nil
}

// Validate checks if this event is valid with the current context.
func (e BlockTimeEvent) ValidateContext(ctx sdk.Context) error {
	if e.Time.UTC().Equal(ctx.BlockTime().UTC()) || e.Time.Before(ctx.BlockTime().UTC()) {
		return ErrInvalidBlockTime
	}
	return nil
}

// NewTrigger creates a new trigger.
func NewTrigger(id TriggerID, owner string, event *codectypes.Any, action []*codectypes.Any) Trigger {
	return Trigger{
		id,
		owner,
		event,
		action,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m Trigger) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m.Event != nil {
		var event TriggerEventI
		err := unpacker.UnpackAny(m.Event, &event)
		if err != nil {
			return err
		}
	}
	return sdktx.UnpackInterfaces(unpacker, m.Actions)
}

// NewQueuedTrigger creates a new trigger for queueing.
func NewQueuedTrigger(trigger Trigger, blockTime time.Time, blockHeight uint64) QueuedTrigger {
	return QueuedTrigger{
		Time:        blockTime,
		BlockHeight: blockHeight,
		Trigger:     trigger,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m QueuedTrigger) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m.Trigger.Event != nil {
		var event TriggerEventI
		err := unpacker.UnpackAny(m.Trigger.Event, &event)
		if err != nil {
			return err
		}
	}
	return sdktx.UnpackInterfaces(unpacker, m.Trigger.Actions)
}

// GetTriggerEventI returns unpacked TriggerEvent
func (m Trigger) GetTriggerEventI() (TriggerEventI, error) {
	event, ok := m.GetEvent().GetCachedValue().(TriggerEventI)
	if !ok {
		return nil, ErrNoTriggerEvent.Wrap("failed to get event")
	}

	return event, nil
}
