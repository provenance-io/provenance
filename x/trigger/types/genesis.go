package types

import (
	"errors"
	fmt "fmt"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	internalsdk "github.com/provenance-io/provenance/internal/sdk"
)

var _ types.UnpackInterfacesMessage = (*GenesisState)(nil)

func NewGenesisState(triggerID, queueStart uint64, triggers []Trigger, queuedTriggers []QueuedTrigger) *GenesisState {
	return &GenesisState{
		TriggerId:      triggerID,
		QueueStart:     queueStart,
		Triggers:       triggers,
		QueuedTriggers: queuedTriggers,
	}
}

// DefaultGenesis returns the default trigger genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(1, 1, []Trigger{}, []QueuedTrigger{})
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.TriggerId == 0 {
		return fmt.Errorf("invalid trigger id")
	}
	if gs.QueueStart == 0 {
		return fmt.Errorf("invalid queue start")
	}

	triggers := append([]Trigger{}, gs.Triggers...)
	for _, queuedTrigger := range gs.QueuedTriggers {
		triggers = append(triggers, queuedTrigger.GetTrigger())
	}

	if len(gs.GasLimits) != 0 {
		return errors.New("gas limits are deprecated and must be empty")
	}

	triggerMap := make(map[uint64]bool)
	for _, trigger := range triggers {
		msgs, err := sdktx.GetMsgs(trigger.Actions, "Genesis - Validate")
		if err != nil {
			return fmt.Errorf("could not get msgs for trigger with id %d: %w", trigger.GetId(), err)
		}

		for idx, msg := range msgs {
			if err = internalsdk.ValidateBasic(msg); err != nil {
				return fmt.Errorf("trigger id: %d, msg: %d, err: %w", trigger.GetId(), idx, err)
			}
		}

		if trigger.GetId() > gs.TriggerId {
			return fmt.Errorf("trigger id %d is invalid and cannot exceed %d", trigger.GetId(), gs.TriggerId)
		}

		event, err := trigger.GetTriggerEventI()
		if err != nil {
			return fmt.Errorf("could not get event for trigger with id %d: %w", trigger.GetId(), err)
		}
		if err = event.Validate(); err != nil {
			return fmt.Errorf("could not validate event for trigger with id %d: %w", trigger.GetId(), err)
		}

		if _, found := triggerMap[trigger.GetId()]; found {
			return fmt.Errorf("trigger id %d is not unique within the set all triggers and queued triggers", trigger.GetId())
		}
		triggerMap[trigger.GetId()] = true
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (gs GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for i, t := range gs.Triggers {
		err := t.UnpackInterfaces(unpacker)
		if err != nil {
			return fmt.Errorf("failed to unpack Triggers[%d]: %w", i, err)
		}
	}
	for i, q := range gs.QueuedTriggers {
		err := q.UnpackInterfaces(unpacker)
		if err != nil {
			return fmt.Errorf("failed to unpack QueuedTriggers[%d]: %w", i, err)
		}
	}
	return nil
}
