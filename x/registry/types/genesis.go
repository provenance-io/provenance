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
	seenClasses := make(map[string]string, len(m.RegistryClasses))
	for _, class := range m.RegistryClasses {
		if err := class.Validate(); err != nil {
			return fmt.Errorf("registry class: %w", err)
		}
		if _, ok := seenClasses[class.RegistryClassId]; ok {
			return fmt.Errorf("duplicate registry class id: %q", class.RegistryClassId)
		}
		seenClasses[class.RegistryClassId] = class.AssetClassId
	}

	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("entry: %w", err)
		}
		// An entry that references a registry class must reference one that exists in genesis;
		// otherwise its authorization tier would silently fall back to params/legacy at runtime.
		// The referenced class must also be for the same asset class as the entry, since a class
		// only governs entries within its own asset class.
		if entry.RegistryClassId != "" {
			classAssetClassID, ok := seenClasses[entry.RegistryClassId]
			if !ok {
				return fmt.Errorf("entry %q references unknown registry class id: %q", entry.Key.String(), entry.RegistryClassId)
			}
			if classAssetClassID != entry.Key.AssetClassId {
				return fmt.Errorf("entry %q: %w", entry.Key.String(),
					NewErrCodeRegistryClassAssetMismatch(entry.RegistryClassId, classAssetClassID, entry.Key.AssetClassId))
			}
		}
	}

	if err := m.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	return nil
}
