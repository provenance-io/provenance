package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
)

func NewGenesisState(port string, oracle string) *GenesisState {
	return &GenesisState{
		PortId: port,
		Oracle: oracle,
	}
}

// DefaultGenesis returns the default oracle genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(PortID, "")
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}

	_, err := sdk.AccAddressFromBech32(gs.Oracle)
	if len(gs.Oracle) > 0 && err != nil {
		return err
	}

	return nil
}
