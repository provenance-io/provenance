package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState creates new GenesisState object
func NewGenesisState(params Params, entries []MsgFee) *GenesisState {
	return &GenesisState{
		Params:  params,
		MsgFees: entries,
	}
}

// Validate ensures all grants in the genesis state are valid
func (state GenesisState) Validate() error {
	for _, a := range state.MsgFees {
		if err := a.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

// DefaultGenesisState returns default state for msgfee module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:  DefaultParams(),
		MsgFees: []MsgFee{},
	}
}

// GetGenesisStateFromAppState returns x/msgfees GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
