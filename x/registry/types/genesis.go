package types

import "fmt"

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate validates the GenesisState.
func (m *GenesisState) Validate() error {
	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("entry: %w", err)
		}
	}

	seenClasses := make(map[string]bool, len(m.RegistryClasses))
	for _, class := range m.RegistryClasses {
		if err := class.Validate(); err != nil {
			return fmt.Errorf("registry class: %w", err)
		}
		if seenClasses[class.RegistryClassId] {
			return fmt.Errorf("duplicate registry class id: %q", class.RegistryClassId)
		}
		seenClasses[class.RegistryClassId] = true
	}

	if err := m.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	return nil
}
