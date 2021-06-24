package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NameRecords within the GenesisState
type NameRecords []NameRecord

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, nameRecords NameRecords) *GenesisState {
	return &GenesisState{
		Params:   params,
		Bindings: nameRecords,
	}
}

// Contains returns true if the given name exists in a slice of NameRecord genesis objects.
func (nrs NameRecords) Contains(name string) bool {
	for _, nr := range nrs {
		if nr.Name == strings.ToLower(strings.TrimSpace(name)) {
			return true
		}
	}

	return false
}

// GetGenesisStateFromAppState returns x/name GenesisState given raw application genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate ensures a genesis state is valid.
func (state GenesisState) Validate() error {
	for _, record := range state.Bindings {
		if strings.TrimSpace(record.Name) == "" {
			return fmt.Errorf("name cannot be empty")
		}
		if strings.TrimSpace(record.Address) == "" {
			return fmt.Errorf("address cannot be empty")
		}
	}
	return nil
}

// DefaultGenesisState returns the initial set of name -> address bindings.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		Bindings: NameRecords{},
	}
}
