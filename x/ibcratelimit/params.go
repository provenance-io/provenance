package ibcratelimit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params object.
func NewParams(contractAddress string) Params {
	return Params{
		ContractAddress: contractAddress,
	}
}

// DefaultParams creates default ibcratelimit module parameters.
func DefaultParams() Params {
	return NewParams("")
}

// Validate verifies all params are correct
func (p Params) Validate() error {
	return validateContractAddress(p.ContractAddress)
}

// validateContractAddress Checks if the supplied address is a valid contract address.
func validateContractAddress(addr string) error {
	// Empty strings are valid for unsetting the param
	if addr == "" {
		return nil
	}

	// Checks that the contract address is valid
	_, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return err
	}

	return nil
}
