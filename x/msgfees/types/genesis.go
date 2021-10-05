package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
)

var _ types.UnpackInterfacesMessage = &GenesisState{}

// NewGenesisState creates new GenesisState object
func NewGenesisState(entries []MsgFees) *GenesisState {
	return &GenesisState{
		MsgFees: entries,
	}
}

// ValidateGenesis ensures all grants in the genesis state are valid
func ValidateGenesis(data GenesisState) error {
	return nil
}

// DefaultGenesisState returns default state for feegrant module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func (data *GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, a := range data.MsgFees {
		err := a.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}

