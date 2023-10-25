package types

import (
	"fmt"

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
	return Params{
		ContractAddress: "",
	}
}

// Validate verifies all params are correct
func (p Params) Validate() error {
	return validateContractAddress(p.ContractAddress)
}

func validateContractAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Empty strings are valid for unsetting the param
	if v == "" {
		return nil
	}

	// Checks that the contract address is valid
	bech32, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return err
	}

	err = sdk.VerifyAddressFormat(bech32)
	if err != nil {
		return err
	}

	return nil
}
