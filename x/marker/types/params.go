package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultEnableGovernance = true
	DefaultMaxTotalSupply   = 100000000000 // 100 billion
)

var (
	// ParamStoreKeyEnableGovernance indicates if governance proposal management of markers is enabled
	ParamStoreKeyEnableGovernance = []byte("EnableGovernance")
	// ParamStoreKeyMaxTotalSupply is maximum supply to allow a marker to create
	ParamStoreKeyMaxTotalSupply = []byte("MaxTotalSupply")
)

// ParamKeyTable for slashing module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter object
func NewParams(
	maxTotalSupply uint64,
	enableGovernance bool,
) Params {
	return Params{
		EnableGovernance: enableGovernance,
		MaxTotalSupply:   maxTotalSupply,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableGovernance, &p.EnableGovernance, validateEnableGovernance),
		paramtypes.NewParamSetPair(ParamStoreKeyMaxTotalSupply, &p.MaxTotalSupply, validateIntParam),
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(
		DefaultMaxTotalSupply,
		DefaultEnableGovernance,
	)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateIntParam(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 1 {
		return fmt.Errorf("value must be greater than zero: %d", v)
	}

	return nil
}

func validateEnableGovernance(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
