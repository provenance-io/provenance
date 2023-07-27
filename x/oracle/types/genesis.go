package types

import fmt "fmt"

func NewGenesisState(queryID uint64) *GenesisState {
	return &GenesisState{
		QueryId: queryID,
	}
}

// DefaultGenesis returns the default trigger genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(1)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.QueryId == 0 {
		return fmt.Errorf("invalid query id")
	}

	return nil
}
