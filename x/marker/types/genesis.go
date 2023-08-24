package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, markers []MarkerAccount, denySendAddress []DenySendAddress) *GenesisState {
	return &GenesisState{
		Params:            params,
		Markers:           markers,
		DenySendAddresses: denySendAddress,
	}
}

// Validate ensures a genesis state is valid.
func (state GenesisState) Validate() error {
	for _, m := range state.Markers {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// DefaultGenesisState returns the initial module genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []MarkerAccount{}, []DenySendAddress{})
}

// GetGenesisStateFromAppState returns x/marker GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
