package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter namespace
const (
	DefaultMinSegmentLength       = uint32(2)
	DefaultMaxSegmentLength       = uint32(32)
	DefaultMaxSegments            = uint32(16)
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
	if p.AllowUnrestrictedNames != that1.AllowUnrestrictedNames {
		return false
	}
	if p.MaxNameLevels != that1.MaxNameLevels {
		return false
	}
	if p.MaxSegmentLength != that1.MaxSegmentLength {
		return false
	}
	if p.MinSegmentLength != that1.MinSegmentLength {
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

func validateAllowUnrestrictedNames(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
