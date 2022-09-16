package types

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	NhashDenom = "nhash"
)

// DefaultDuration is the default duration a module asset
// will live on chain before expiring. Defaults to 1 year
var DefaultDuration = 24 * 365 * time.Hour

var DefaultDeposit = sdk.Coin{
	Denom:  NhashDenom,
	Amount: sdk.NewInt(1905), // todo: set default required amount
}

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
	if that1 == nil {
		return p == nil
	} else if p == nil {
		return false
	}
	return true
}

func (p *Params) Validate() error {
	return validateDepositParam(p.Deposit)
}

func validateDepositParam(i interface{}) error {
	coin, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// validate appropriate Coin
	if coin.Validate() != nil {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
