package types

// Validate ensures the genesis state is valid.
func (state GenesisState) Validate() error {
	return nil
}

// NewGenesisState returns a new instance of GenesisState
func NewGenesisState(
	params Params,
	scopes []Scope,
	sessions []Session,
	records []Record,
	scopeSpecs []ScopeSpecification,
	contracSpecs []ContractSpecification,
	recordSpecs []RecordSpecification,
) *GenesisState {
	return &GenesisState{
		Params:                 params,
		Scopes:                 scopes,
		Sessions:               sessions,
		Records:                records,
		ScopeSpecifications:    scopeSpecs,
		ContractSpecifications: contracSpecs,
		RecordSpecifications:   recordSpecs,
	}
}

// DefaultGenesisState returns a zero-value genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}
