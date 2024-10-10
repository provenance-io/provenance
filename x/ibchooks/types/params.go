package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewParams(allowedAsyncAckContracts []string) Params {
	return Params{
		AllowedAsyncAckContracts: allowedAsyncAckContracts,
	}
}

// DefaultParams returns default concentrated-liquidity module parameters.
func DefaultParams() Params {
	return Params{
		AllowedAsyncAckContracts: []string{},
	}
}

// Validate params.
func (p Params) Validate() error {
	for _, contract := range p.AllowedAsyncAckContracts {
		if _, err := sdk.AccAddressFromBech32(contract); err != nil {
			return err
		}
	}
	return nil
}
