package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	NhashDenom = "nhash"
)

// DefaultDeposit is defined  as 0nhash, so that it's only meaningful if set in params
var DefaultDeposit = sdk.NewInt64Coin(NhashDenom, 0)

// DefaultDuration is the period upon which an expiration will stay on chain
// when the period elapses the module asset is up for extension or deletion
//
// the accepted values for a duration are n{h,d,w,y}
// where 1h = 60m, 1d = 24h, 1w = 7d (or 168h), 1y = 365d (or 8760h)
//
// DefaultDuration is defined as "1y", this can also be set in params
var DefaultDuration = "1y"

var (
	ParamStoreKeyDeposit  = []byte("Deposit")
	ParamStoreKeyDuration = []byte("Expiration")
)

// ParamKeyTable for marker module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter object
func NewParams(deposit sdk.Coin, duration string) Params {
	return Params{Deposit: deposit, Duration: duration}
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyDeposit, &p.Deposit, validateDepositParam),
		paramtypes.NewParamSetPair(ParamStoreKeyDuration, &p.Duration, validateDurationParam),
	}
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return NewParams(DefaultDeposit, DefaultDuration)
}

// Validate validates the deposit parameter
func (p *Params) Validate() error {
	if err := validateDepositParam(p.Deposit); err != nil {
		return err
	}

	if err := validateDurationParam(p.Duration); err != nil {
		return err
	}

	return nil
}

// Private method that runs validation on the deposit parameter
func validateDepositParam(i interface{}) error {
	coin, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// validate appropriate Coin
	if err := coin.Validate(); err != nil {
		return fmt.Errorf("invalid coin: %w", err)
	}

	return nil
}

// Private method that runs validation on the duration parameter
func validateDurationParam(i interface{}) error {
	duration, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	_, err := ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	return nil
}
