package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter namespace
const (
	DefaultMinSegmentLength       = 2
	DefaultMaxSegmentLength       = 32
	DefaultMaxSegments            = 16
	DefaultAllowUnrestrictedNames = true
)

// Parameter store keys
var (
	// maximum length of name segment to allow
	ParamStoreKeyMaxSegmentLength = []byte("MaxSegmentLength")
	// minimum length of name segment to allow
	ParamStoreKeyMinSegmentLength = []byte("MinSegmentLength")
	// maximum number of name segments to allow.  Example: `foo.bar.baz` would be 3
	ParamStoreKeyMaxNameLevels = []byte("MaxNameLevels")
	// determines if unrestricted name keys are allowed or not
	ParamStoreKeyAllowUnrestrictedNames = []byte("AllowUnrestrictedNames")
)

// ParamKeyTable for slashing module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter object
func NewParams(
	maxSegmentLength uint32,
	minSegmentLength uint32,
	maxNameLevels uint32,
	allowUnrestrictedNames bool,
) Params {
	return Params{
		MaxSegmentLength:       maxSegmentLength,
		MinSegmentLength:       minSegmentLength,
		MaxNameLevels:          maxNameLevels,
		AllowUnrestrictedNames: allowUnrestrictedNames,
	}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMaxSegmentLength, &p.MaxSegmentLength, validateIntParam),
		paramtypes.NewParamSetPair(ParamStoreKeyMinSegmentLength, &p.MinSegmentLength, validateIntParam),
		paramtypes.NewParamSetPair(ParamStoreKeyMaxNameLevels, &p.MaxNameLevels, validateIntParam),
		paramtypes.NewParamSetPair(ParamStoreKeyAllowUnrestrictedNames, &p.AllowUnrestrictedNames, validateAllowUnrestrictedNames),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		DefaultMaxSegmentLength,
		DefaultMinSegmentLength,
		DefaultMaxSegments,
		DefaultAllowUnrestrictedNames,
	)
}

func validateIntParam(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 1 {
		return fmt.Errorf("value must be greater than zero: %d", v)
	}

	return nil
}

func validateAllowUnrestrictedNames(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
