package types

import (
	fmt "fmt"

	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
)

func NewGenesisState(port string, params Params, sequence uint64) *GenesisState {
	return &GenesisState{
		PortId:   port,
		Params:   params,
		Sequence: sequence,
	}
}

// DefaultGenesis returns the default trigger genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(PortID, DefaultParams(), 1)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.Sequence == 0 {
		return fmt.Errorf("invalid query id")
	}

	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}

	return gs.Params.Validate()
}
