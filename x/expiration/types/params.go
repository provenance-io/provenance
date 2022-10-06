package types

import (
	"fmt"
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	NhashDenom = "nhash"
)

// DefaultDeposit is defined as 1000000000nhash (1 hash)
var DefaultDeposit = sdk.NewInt64Coin(NhashDenom, 1000000000)

var (
	ParamStoreKeyDeposit = []byte("Deposit")
)

// ParamKeyTable for marker module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter object
func NewParams(deposit sdk.Coin) Params {
	return Params{Deposit: deposit}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyDeposit, &p.Deposit, validateDepositParam),
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(DefaultDeposit)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Equal returns true if the given value is equivalent to the current instance of params
func (p *Params) Equal(that interface{}) bool {
	if that == nil {
		return p == nil
	}

	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}

	return that1.Deposit.IsEqual(p.Deposit)
}

// Validate validates the deposit parameter
func (p *Params) Validate() error {
	return validateDepositParam(p.Deposit)
}

// Private method that runs validation on the deposit parameter
func validateDepositParam(i interface{}) error {
	coin, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// validate appropriate Coin
	if err := coin.Validate(); err != nil {
		return fmt.Errorf("invalid parameter: %T", err)
	}

	return nil
}
