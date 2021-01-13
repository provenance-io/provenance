package types

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, markers []MarkerAccount) *GenesisState {
	return &GenesisState{
		Params:  params,
		Markers: markers,
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
	return &GenesisState{}
}
