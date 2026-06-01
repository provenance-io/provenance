package types

import (
	"errors"
)

func NewParams() Params {
	return Params{
		AllowedAsyncAckContracts: []string{},
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
	if len(p.AllowedAsyncAckContracts) != 0 {
		return errors.New("the allowed_async_ack_contracts field must be empty")
	}
	return nil
}
