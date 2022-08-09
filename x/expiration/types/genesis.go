package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewGenesisState creates a new genesis object
func NewGenesisState(params Params, expirations []Expiration) *GenesisState {
	return &GenesisState{
		Params:      params,
		Expirations: expirations,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}

}

// ValidateGenesis ensures the genesis state is valid
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	for _, expiration := range data.Expirations {
		if err := expiration.ValidateBasic(); err != nil {
			return err
		}
		if _, err := sdk.AccAddressFromBech32(expiration.ModuleAssetId); err != nil {
			return err
		}
		if _, err := sdk.AccAddressFromBech32(expiration.Owner); err != nil {
			return err
		}
	}
	return nil
}
