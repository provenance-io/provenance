package types

import (
	fmt "fmt"

	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

func NewGenesisState(triggerID, queueStart uint64, triggers []Trigger, gasLimits []uint64, queuedTriggers []QueuedTrigger) *GenesisState {
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
	return NewGenesisState(1, 1, []Trigger{}, []uint64{}, []QueuedTrigger{})
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
	if len(gs.Triggers) != len(gs.GasLimits) {
		return fmt.Errorf("trigger list length must match gas limit list length")
	}

	triggers := gs.Triggers
	for _, queuedTrigger := range gs.QueuedTriggers {
		triggers = append(triggers, queuedTrigger.GetTrigger())
	}
	for _, trigger := range triggers {
		msgs, err := sdktx.GetMsgs(trigger.Actions, "Genesis - Validate")
		if err != nil {
			return err
		}

		for idx, msg := range msgs {
			if err := msg.ValidateBasic(); err != nil {
				return fmt.Errorf("msg: %d, err: %w", idx, err)
			}
		}

		if trigger.GetId() > gs.TriggerId {
			return fmt.Errorf("trigger id is invalid and cannot exceed %d", gs.TriggerId)
		}
	}

	return nil
}
