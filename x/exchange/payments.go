package exchange

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate returns an error if any of this Payment's info is invalid.
func (p Payment) Validate() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(p.Source); err != nil {
		errs = append(errs, fmt.Errorf("invalid source %q: %w", p.Source, err))
	}
	if len(p.Target) > 0 {
		if _, err := sdk.AccAddressFromBech32(p.Target); err != nil {
			errs = append(errs, fmt.Errorf("invalid target %q: %w", p.Target, err))
		}
	}

	amountsOK := true
	if err := p.SourceAmount.Validate(); err != nil {
		amountsOK = false
		errs = append(errs, fmt.Errorf("invalid source amount %q: %w", p.SourceAmount, err))
	}
	if err := p.TargetAmount.Validate(); err != nil {
		amountsOK = false
		errs = append(errs, fmt.Errorf("invalid target amount %q: %w", p.TargetAmount, err))
	}
	if amountsOK && p.SourceAmount.IsZero() && p.TargetAmount.IsZero() {
		errs = append(errs, errors.New("source amount and target amount cannot both be zero"))
	}

	if err := ValidateExternalID(p.ExternalId); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// String returns a string representing this Payment.
func (p Payment) String() string {
	return fmt.Sprintf("%q;%s:%q<->%s:%q", p.ExternalId, p.Source, p.SourceAmount, p.Target, p.TargetAmount)
}
