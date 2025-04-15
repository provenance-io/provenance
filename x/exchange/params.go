package exchange

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
)

const (
	// DefaultDefaultSplit is the default value used for the DefaultSplit parameter.
	DefaultDefaultSplit = uint32(500)
	// DefaultFeeCreatePaymentFlatAmount is the default amount for creating a payment. The denom is the chain's FeeDenom.
	DefaultFeeCreatePaymentFlatAmount = int64(10_000_000_000)
	// DefaultFeeAcceptPaymentFlatAmount is the default amount for accepting a payment. The denom is the chain's FeeDenom.
	DefaultFeeAcceptPaymentFlatAmount = int64(8_000_000_000)

	// MaxSplit is the maximum split value. 10,000 basis points = 100%.
	MaxSplit = uint32(10_000)
)

// DefaultParams returns the default exchange module params.
func DefaultParams() *Params {
	feeDenom := pioconfig.GetProvConfig().FeeDenom
	return &Params{
		DefaultSplit:         DefaultDefaultSplit,
		DenomSplits:          nil,
		FeeCreatePaymentFlat: []sdk.Coin{sdk.NewInt64Coin(feeDenom, DefaultFeeCreatePaymentFlatAmount)},
		FeeAcceptPaymentFlat: []sdk.Coin{sdk.NewInt64Coin(feeDenom, DefaultFeeAcceptPaymentFlatAmount)},
	}
}

// Validate returns an error if there's something wrong with these params.
func (p Params) Validate() error {
	var errs []error
	if err := validateSplit("default", p.DefaultSplit); err != nil {
		errs = append(errs, err)
	}
	for _, split := range p.DenomSplits {
		if err := split.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	if err := validatePaymentFlatFee("create", p.FeeCreatePaymentFlat); err != nil {
		errs = append(errs, err)
	}
	if err := validatePaymentFlatFee("accept", p.FeeAcceptPaymentFlat); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func NewDenomSplit(denom string, split uint32) *DenomSplit {
	return &DenomSplit{
		Denom: denom,
		Split: split,
	}
}

// Validate returns an error if there's something wrong with this DenomSplit.
func (d DenomSplit) Validate() error {
	if err := sdk.ValidateDenom(d.Denom); err != nil {
		return err
	}
	return validateSplit(d.Denom, d.Split)
}

// validateSplit makes sure that the provided split value is in the valid range.
func validateSplit(name string, split uint32) error {
	if split > MaxSplit {
		return fmt.Errorf("%s split %d cannot be greater than %d", name, split, MaxSplit)
	}
	return nil
}

// validatePaymentFlatFee makes sure that the provided fee options are valid for a payment-related flat fee field.
func validatePaymentFlatFee(name string, feeOpts []sdk.Coin) error {
	switch len(feeOpts) {
	case 0:
		return nil
	case 1:
		if err := feeOpts[0].Validate(); err != nil {
			return fmt.Errorf("invalid %s payment flat fee %q: %w", name, feeOpts[0], err)
		}
		if feeOpts[0].IsZero() {
			return fmt.Errorf("invalid %s payment flat fee %q: zero amount not allowed", name, feeOpts[0])
		}
		return nil
	default:
		return fmt.Errorf("invalid %s payment flat fee %q: max entries is 1", name, sdk.Coins(feeOpts).String())
	}
}
