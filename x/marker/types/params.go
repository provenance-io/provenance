package types

import (
	"fmt"
	"regexp"

	sdkmath "cosmossdk.io/math"
)

const (
	// DefaultEnableGovernance (true) indicates that governance proposals are allowed for managing markers
	DefaultEnableGovernance = true
	// DefaultMaxSupply is the upper bound to enforce on supply for markers.
	DefaultMaxSupply = "100000000000000000000"
	// DefaultUnrestrictedDenomRegex is a regex that denoms created by normal requests must pass.
	DefaultUnrestrictedDenomRegex = `[a-zA-Z][a-zA-Z0-9\-\.]{2,83}`
)

// NewParams creates a new parameter object
func NewParams(
	enableGovernance bool,
	unrestrictedDenomRegex string,
	maxSupply sdkmath.Int,
) Params {
	return Params{
		EnableGovernance:       enableGovernance,
		UnrestrictedDenomRegex: unrestrictedDenomRegex,
		MaxSupply:              maxSupply,
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(
		DefaultEnableGovernance,
		DefaultUnrestrictedDenomRegex,
		StringToBigInt(DefaultMaxSupply),
	)
}

func (p Params) Validate() error {
	exp := p.UnrestrictedDenomRegex
	if len(exp) > 0 && (exp[0:1] == "^" || exp[len(exp)-1:] == "$") {
		return fmt.Errorf("invalid parameter, validation regex must not contain anchors ^,$")
	}
	_, err := regexp.Compile(fmt.Sprintf(`^%s$`, exp))
	return err

}

func StringToBigInt(val string) sdkmath.Int {
	res, ok := sdkmath.NewIntFromString(val)
	if !ok {
		panic(fmt.Errorf("unable to create sdkmath.Int from string: %s", val))
	}
	return res
}
