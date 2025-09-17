package types

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Validate validates the GenesisState.
func (m *GenesisState) Validate() error {
	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return NewErrCodeInvalidField("entry", err.Error())
		}
	}

	return nil
}
