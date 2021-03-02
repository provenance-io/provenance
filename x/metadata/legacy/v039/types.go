package v039

import (
	"fmt"
	"strings"
)

// GenesisState is the head state of all scopes with history.
type GenesisState struct {
	ScopeRecords []Scope `json:"scope_records,omitempty"`
	// NOTE: this comes from the v39 spec module that was merged in.  The migrate step should copy this over
	Specifications []ContractSpec `json:"specifications,omitempty"`
}

// SpecGenesisState is a sequence of contract specs (from the deprecated spec module)
type SpecGenesisState struct {
	ContractSpecRecords []ContractSpec `json:"contract_spec_records"`
}

// Validate ensures the genesis state is valid.
func (state GenesisState) Validate() error {
	for _, s := range state.ScopeRecords {
		if err := ValidateScope(s); err != nil {
			return err
		}
	}
	return nil
}

// ValidateScope ensures required scope fields are valid.
func ValidateScope(s Scope) error {
	if s.Uuid == nil {
		return fmt.Errorf("scope UUID cannot be nil")
	}
	if strings.TrimSpace(s.Uuid.Value) == "" {
		return fmt.Errorf("scope UUID value cannot be empty")
	}
	if len(s.Parties) == 0 {
		return fmt.Errorf("scope must have at least one party")
	}
	return nil
}

// DefaultGenesisState returns a zero-value genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}
