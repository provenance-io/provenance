package types

import (
	time "time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	proto "github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
)

type TriggerID = uint64

const BLOCK_HEIGHT_PREFIX = "block-height"
const BLOCK_TIME_PREFIX = "block-time"

type TriggerEventI interface {
	proto.Message
	GetEventPrefix() string
}

func NewTrigger(id TriggerID, owner string, event *types.Any, action []*types.Any) Trigger {
	return Trigger{
		id,
		owner,
		event,
		action,
	}
}

func NewQueuedTrigger(trigger Trigger, blockTime time.Time, blockHeight uint64) QueuedTrigger {
	return QueuedTrigger{
		Time:        blockTime,
		BlockHeight: blockHeight,
		Trigger:     trigger,
	}
}

func (e TransactionEvent) Equals(other abci.Event) bool {
	if e.Name != other.GetType() {
		return false
	}

	for _, attr := range e.Attributes {
		hasAttribute := false

		for _, otherAttr := range other.Attributes {
			if attr.Equals(otherAttr) {
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

func (a Attribute) Equals(other abci.EventAttribute) bool {
	if a.GetName() != string(other.GetKey()) {
		return false
	}

	if a.GetValue() != "" && a.GetValue() != string(other.GetValue()) {
		return false
	}

	return true
}

func (e TransactionEvent) GetEventPrefix() string {
	return e.Name
}

func (e BlockHeightEvent) GetEventPrefix() string {
	return BLOCK_HEIGHT_PREFIX
}

func (e BlockTimeEvent) GetEventPrefix() string {
	return BLOCK_TIME_PREFIX
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

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m QueuedTrigger) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, m.Trigger.Actions)
}
