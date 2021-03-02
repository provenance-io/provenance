package types

import (
	"fmt"
	"regexp"

	yaml "gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// DefaultEnableGovernance (true) indicates that governance proposals are allowed for managing markers
	DefaultEnableGovernance = true
	// DefaultMaxTotalSupply is the upper bound to enforce on supply for markers.
	DefaultMaxTotalSupply = 100000000000
	// DefaultUnrestrictedDenomRegex is a regex that denoms created by normal requests must pass.
	DefaultUnrestrictedDenomRegex = `[a-zA-Z][a-zA-Z0-9/]{2,64}`
)

var (
	// ParamStoreKeyEnableGovernance indicates if governance proposal management of markers is enabled
	ParamStoreKeyEnableGovernance = []byte("EnableGovernance")
	// ParamStoreKeyMaxTotalSupply is maximum supply to allow a marker to create
	ParamStoreKeyMaxTotalSupply = []byte("MaxTotalSupply")
	// ParamStoreKeyUnrestrictedDenomRegex is the validation regex for validating denoms supplied by users.
	ParamStoreKeyUnrestrictedDenomRegex = []byte("UnrestrictedDenomRegex")
)

// ParamKeyTable for slashing module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter object
func NewParams(
	maxTotalSupply uint64,
	enableGovernance bool,
	unrestrictedDenomRegex string,
) Params {
	return Params{
		EnableGovernance:       enableGovernance,
		MaxTotalSupply:         maxTotalSupply,
		UnrestrictedDenomRegex: unrestrictedDenomRegex,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableGovernance, &p.EnableGovernance, validateEnableGovernance),
		paramtypes.NewParamSetPair(ParamStoreKeyMaxTotalSupply, &p.MaxTotalSupply, validateIntParam),
		paramtypes.NewParamSetPair(ParamStoreKeyUnrestrictedDenomRegex, &p.UnrestrictedDenomRegex, validateRegexParam),
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(
		DefaultMaxTotalSupply,
		DefaultEnableGovernance,
		DefaultUnrestrictedDenomRegex,
	)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

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
	if p.MaxTotalSupply != that1.MaxTotalSupply {
		return false
	}
	if p.EnableGovernance != that1.EnableGovernance {
		return false
	}
	if p.UnrestrictedDenomRegex != that1.UnrestrictedDenomRegex {
		return false
	}
	return true
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

func validateRegexParam(i interface{}) error {
	exp, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	_, err := regexp.Compile(exp)

	return err
}
