package ibcratelimit

// DefaultGenesis creates a default GenesisState object.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}

// NewGenesisState returns a new instance of GenesisState object
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}
