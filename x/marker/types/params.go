package types

import (
	"errors"
	"fmt"
	"regexp"

	sdkmath "cosmossdk.io/math"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// DefaultEnableGovernance (true) indicates that governance proposals are allowed for managing markers
	DefaultEnableGovernance = true
	// DefaultMaxTotalSupply is deprecated.
	DefaultMaxTotalSupply = uint64(100000000000)
	// DefaultMaxSupply is the upper bound to enforce on supply for markers.
	DefaultMaxSupply = "100000000000000000000"
	// DefaultUnrestrictedDenomRegex is a regex that denoms created by normal requests must pass.
	DefaultUnrestrictedDenomRegex = `[a-zA-Z][a-zA-Z0-9\-\.]{2,83}`
)

// TODO: remove with the umber (v1.19.x) handlers.
var (
	// ParamStoreKeyEnableGovernance indicates if governance proposal management of markers is enabled
	ParamStoreKeyEnableGovernance = []byte("EnableGovernance")
	// ParamStoreKeyMaxTotalSupply is deprecated.
	ParamStoreKeyMaxTotalSupply = []byte("MaxTotalSupply")
	// ParamStoreKeyMaxSupply is maximum supply to allow a marker to create
	ParamStoreKeyMaxSupply = []byte("MaxSupply")
	// ParamStoreKeyUnrestrictedDenomRegex is the validation regex for validating denoms supplied by users.
	ParamStoreKeyUnrestrictedDenomRegex = []byte("UnrestrictedDenomRegex")
)

// ParamKeyTable for marker module
// TODO: remove with the umber (v1.19.x) handlers.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter object
func NewParams(
	maxTotalSupply uint64,
	enableGovernance bool,
	unrestrictedDenomRegex string,
	maxSupply sdkmath.Int,
) Params {
	return Params{
		EnableGovernance:       enableGovernance,
		MaxTotalSupply:         maxTotalSupply,
		UnrestrictedDenomRegex: unrestrictedDenomRegex,
		MaxSupply:              maxSupply,
	}
}

// ParamSetPairs - Implements params.ParamSet
// TODO: remove with the umber (v1.19.x) handlers.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableGovernance, &p.EnableGovernance, validateEnableGovernance),
		paramtypes.NewParamSetPair(ParamStoreKeyMaxTotalSupply, &p.MaxTotalSupply, validateIntParam),
		paramtypes.NewParamSetPair(ParamStoreKeyUnrestrictedDenomRegex, &p.UnrestrictedDenomRegex, validateRegexParam),
		paramtypes.NewParamSetPair(ParamStoreKeyMaxSupply, &p.MaxSupply, validateBigIntParam),
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(
		DefaultMaxTotalSupply,
		DefaultEnableGovernance,
		DefaultUnrestrictedDenomRegex,
		StringToBigInt(DefaultMaxSupply),
	)
}

func (p Params) Validate() error {
	errs := []error{
		validateEnableGovernance(p.EnableGovernance),
		validateIntParam(p.MaxTotalSupply),
		validateRegexParam(p.UnrestrictedDenomRegex),
		validateBigIntParam(p.MaxSupply),
	}
	return errors.Join(errs...)
}

func validateIntParam(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBigIntParam(i interface{}) error {
	_, ok := i.(sdkmath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
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
