package types

import (
	fmt "fmt"

	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

func NewGenesisState(triggerID, queueStart uint64, triggers []Trigger, gasLimits []GasLimit, queuedTriggers []QueuedTrigger) *GenesisState {
	return &GenesisState{
		TriggerId:      triggerID,
		QueueStart:     queueStart,
		Triggers:       triggers,
		GasLimits:      gasLimits,
		QueuedTriggers: queuedTriggers,
	}
}

// DefaultGenesis returns the default trigger genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(1, 1, []Trigger{}, []GasLimit{}, []QueuedTrigger{})
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
	if len(gs.Triggers)+len(gs.QueuedTriggers) != len(gs.GasLimits) {
		return fmt.Errorf("gas limit list length must match sum of triggers and queued triggers length")
	}

	triggers := gs.Triggers
	for _, queuedTrigger := range gs.QueuedTriggers {
		triggers = append(triggers, queuedTrigger.GetTrigger())
	}

	gasLimitMap := make(map[uint64]bool)
	for _, gasLimit := range gs.GasLimits {
		if _, found := gasLimitMap[gasLimit.TriggerId]; found {
			return fmt.Errorf("cannot have duplicate trigger id in gas limits")
		}
		gasLimitMap[gasLimit.TriggerId] = true
	}

	triggerMap := make(map[uint64]bool)
	for _, trigger := range triggers {
		msgs, err := sdktx.GetMsgs(trigger.Actions, "Genesis - Validate")
		if err != nil {
			return err
		}

		for idx, msg := range msgs {
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("msg: %d, err: %w", idx, err)
			}
		}

		if trigger.GetId() > gs.TriggerId {
			return fmt.Errorf("trigger id is invalid and cannot exceed %d", gs.TriggerId)
		}

		event, err := trigger.GetTriggerEventI()
		if err != nil {
			return err
		}
		if err = event.Validate(); err != nil {
			return err
		}

		if _, found := gasLimitMap[trigger.GetId()]; !found {
			return fmt.Errorf("trigger or queued trigger does not have a gas limit that matches it with id %d", trigger.GetId())
		}

		if _, found := triggerMap[trigger.GetId()]; found {
			return fmt.Errorf("all trigger ids shared between triggers and queued triggers must be unique")
		}
		triggerMap[trigger.GetId()] = true
	}

	return nil
}
