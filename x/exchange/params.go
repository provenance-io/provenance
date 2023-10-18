package exchange

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DefaultDefaultSplit is the default value used for the DefaultSplit parameter.
	// TODO[1658]: Discuss what this should be with someone who would know.
	DefaultDefaultSplit = uint32(500)

	// MaxSplit is the maximum split value. 10,000 basis points = 100%.
	MaxSplit = uint32(10_000)
)

func NewParams(defaultSplit uint32, denomSplits []DenomSplit) *Params {
	return &Params{
		DefaultSplit: defaultSplit,
		DenomSplits:  denomSplits,
	}
}

// DefaultParams returns the default exchange module params.
func DefaultParams() *Params {
	return NewParams(DefaultDefaultSplit, nil)
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
