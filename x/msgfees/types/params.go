package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// DefaultFloorGasPrice to differentiate between base fee and additional fee when additional fee is in same denom as default base denom i.e nhash
	DefaultFloorGasPrice = 1905
)

var (
	// ParamStoreKeyFloorGasPrice if msg fees are paid in the same denom as base default gas is paid, then use this to differentiate between base price
	// and additional fees.
	ParamStoreKeyFloorGasPrice = []byte("FloorGasPrice")
)

// ParamKeyTable for marker module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter object
func NewParams(
	floorGasPrice uint32,
) Params {
	return Params{
		FloorGasPrice: floorGasPrice,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyFloorGasPrice, &p.FloorGasPrice, validateIntParam),
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(
		DefaultFloorGasPrice,
	)
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

func validateIntParam(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
