package types

import errorsmod "cosmossdk.io/errors"

// ValidateBasic ensures a genesis state is valid.
func (state GenesisState) ValidateBasic() error {
	if err := state.Params.Validate(); err != nil {
		return errorsmod.Wrap(err, "params")
	}
	for _, a := range state.Accounts {
		if err := a.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}
