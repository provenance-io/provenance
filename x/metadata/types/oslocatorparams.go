package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default parameter namespace
const (
	DefaultMaxURILength = 2048
)

// ParamKeyTable for metadata module
func OSParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&OSLocatorParams{})
}

// Parameter store keys
var (
	ParamStoreKeyMaxValueLength = []byte("MaxUriLength")
)

// NewParams creates a new parameter object
func NewOSLocatorParams(maxURILength uint32) OSLocatorParams {
	return OSLocatorParams{MaxUriLength: maxURILength}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
func (p *OSLocatorParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMaxValueLength, &p.MaxUriLength, validateMaxURILength),
	}
}

// DefaultParams defines the parameters for this module
func DefaultOSLocatorParams() OSLocatorParams {
	return NewOSLocatorParams(DefaultMaxURILength)
}

func validateMaxURILength(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
