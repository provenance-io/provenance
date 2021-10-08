package types


// NewGenesisState creates new GenesisState object
func NewGenesisState(entries []MsgBasedFee) *GenesisState {
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

