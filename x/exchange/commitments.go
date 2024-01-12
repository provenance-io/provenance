package exchange

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// String returns a string representation of this AccountAmount.
func (a AccountAmount) String() string {
	return fmt.Sprintf("%s:%q", a.Account, a.Amount)
}

// Validate returns an error if this AccountAmount is invalid.
func (a AccountAmount) Validate() error {
	if _, err := sdk.AccAddressFromBech32(a.Account); err != nil {
		return fmt.Errorf("invalid account %q: %w", a.Account, err)
	}
	if err := a.Amount.Validate(); err != nil {
		return fmt.Errorf("invalid amount %q: %w", a.Amount, err)
	}
	return nil
}

// String returns a string representation of this MarketAmount.
func (m MarketAmount) String() string {
	return fmt.Sprintf("%d:%q", m.MarketId, m.Amount)
}

// String returns a string representation of this NetAssetPrice.
func (n NetAssetPrice) String() string {
	return fmt.Sprintf("%q=%q", n.Assets, n.Price)
}

// Validate returns an error if this NetAssetPrice is invalid.
func (n NetAssetPrice) Validate() error {
	if err := n.Assets.Validate(); err != nil {
		return fmt.Errorf("invalid assets %q: %w", n.Assets, err)
	}
	if n.Assets.IsZero() {
		return fmt.Errorf("invalid assets %q: cannot be zero", n.Assets)
	}

	if err := n.Price.Validate(); err != nil {
		return fmt.Errorf("invalid price %q: %w", n.Price, err)
	}
	if n.Price.IsZero() {
		return fmt.Errorf("invalid price %q: cannot be zero", n.Price)
	}

	return nil
}
