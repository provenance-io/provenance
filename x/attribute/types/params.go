package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter namespace
const (
	DefaultMaxValueLength = 10000
)

// Parameter store keys
// TODO: remove with the umber (v1.19.x) handlers.
var (
	ParamStoreKeyMaxValueLength = []byte("MaxValueLength")
)

// ParamKeyTable for slashing module
// TODO: remove with the umber (v1.19.x) handlers.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams create a new Params object
func NewParams(
	maxValueLength uint32,
) Params {
	return Params{
		MaxValueLength: maxValueLength,
	}
}

// ParamSetPairs - Implements params.ParamSet
// TODO: remove with the umber (v1.19.x) handlers.
func (params *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMaxValueLength, &params.MaxValueLength, validateMaxValueLength),
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		DefaultMaxValueLength,
	)
}

func validateMaxValueLength(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
