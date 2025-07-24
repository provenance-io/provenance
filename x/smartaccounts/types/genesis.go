package types

// ValidateBasic ensures a genesis state is valid.
func (state GenesisState) ValidateBasic() error {
	for _, a := range state.Accounts {
		if err := a.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}
