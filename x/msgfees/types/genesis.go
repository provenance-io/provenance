package types

// NewGenesisState creates new GenesisState object
func NewGenesisState(params Params, entries []MsgBasedFee) *GenesisState {
	return &GenesisState{
		Params:       params,
		MsgBasedFees: entries,
	}
}

// Validate ensures all grants in the genesis state are valid
func (state GenesisState) Validate() error {
	for _, a := range state.MsgBasedFees {
		if err := a.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

// DefaultGenesisState returns default state for msgfee module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		MsgBasedFees: []MsgBasedFee{},
	}
}
