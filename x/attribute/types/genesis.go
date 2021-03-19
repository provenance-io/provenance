package types

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, attributes []Attribute) *GenesisState {
	return &GenesisState{
		Params:     params,
		Attributes: attributes,
	}
}

// ValidateBasic ensures a genesis state is valid.
func (state GenesisState) ValidateBasic() error {
	for _, a := range state.Attributes {
		if err := a.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

// DefaultGenesisState returns the default module state at genesis.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		Attributes: []Attribute{},
	}
}
