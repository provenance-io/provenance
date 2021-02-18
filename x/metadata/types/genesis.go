package types

// Validate ensures the genesis state is valid.
func (state GenesisState) Validate() error {
	return nil
}

// NewGenesisState returns a new instance of GenesisState
func NewGenesisState(
	params Params,
	scopes []Scope,
	groups []RecordGroup,
	records []Record,
	scopeSpecs []ScopeSpecification,
	contracSpecs []ContractSpecification,
) *GenesisState {
	return &GenesisState{
		Params:                 params,
		Scopes:                 scopes,
		Groups:                 groups,
		Records:                records,
		ScopeSpecifications:    scopeSpecs,
		ContractSpecifications: contracSpecs,
	}
}

// DefaultGenesisState returns a zero-value genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}
