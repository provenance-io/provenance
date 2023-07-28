package types

import (
	fmt "fmt"

	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
)

func NewGenesisState(queryID uint64) *GenesisState {
	return &GenesisState{
		QueryId: queryID,
	}
}

// DefaultGenesis returns the default trigger genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		PortId:  PortID,
		Params:  DefaultParams(),
		QueryId: 1,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.QueryId == 0 {
		return fmt.Errorf("invalid query id")
	}

	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}

	return gs.Params.Validate()
}
