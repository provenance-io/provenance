package types

import (
	time "time"

	abci "github.com/tendermint/tendermint/abci/types"

	types "github.com/cosmos/cosmos-sdk/codec/types"
)

type TriggerID = uint64

func NewTrigger(id TriggerID, owner string, event Event, action *types.Any) Trigger {
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

func (e Event) Equals(other abci.Event) bool {
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
