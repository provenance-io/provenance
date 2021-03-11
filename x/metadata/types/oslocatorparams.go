package types

import (
	"fmt"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter namespace
const (
	DefaultMaxURILength = 2048
)

// Parameter store keys
var (
	ParamStoreKeyMaxValueLength = []byte("MaxUriCharacters")
)

// NewParams creates a new parameter object
func NewOSLocatorParams(maxUriCharacters uint64) OSLocatorParams {
	return OSLocatorParams{MaxUriCharacters: maxUriCharacters}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
func (p *OSLocatorParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMaxValueLength, &p.MaxUriCharacters, validateMaxURILength),
	}
}

// DefaultParams defines the parameters for this module
func DefaultOSLocatorParams() OSLocatorParams {
	return NewOSLocatorParams(DefaultMaxURILength)
}

func validateMaxURILength(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
