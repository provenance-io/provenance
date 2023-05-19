package types

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
	return nil
}
