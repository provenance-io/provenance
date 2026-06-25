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

	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("entry: %w", err)
		}
		// An entry that references a registry class must reference one that exists in genesis;
		// otherwise its authorization tier would silently fall back to params/legacy at runtime.
		if entry.RegistryClassId != "" && !seenClasses[entry.RegistryClassId] {
			return fmt.Errorf("entry %q references unknown registry class id: %q", entry.Key.String(), entry.RegistryClassId)
		}
	}

	if err := m.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	return nil
}
