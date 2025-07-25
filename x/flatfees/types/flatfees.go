package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
)

const (
	// MaxMsgTypeURLLen is the maximum number of bytes that a MsgTypeUrl can have.
	// The MsgTypeUrl gets used in the key for an entry, so we need to limit its length.
	// As of writing this (Jan 2025), the longest known msg type is 80 characters:
	// "/ibc.applications.interchain_accounts.controller.v1.MsgRegisterInterchainAccount"
	// Due to the nature of these strings, its unlikely that there will be one that is much longer.
	// To be on the safe side, I doubled that for this value.
	MaxMsgTypeURLLen = 160
	// DefaultFeeDefinitionDenom is the denomination that, by default, the flat fees will be defined in (before being converted).
	DefaultFeeDefinitionDenom = "musd"
)

// DefaultParams is the default parameter configuration for the flatfees module.
func DefaultParams() Params {
	return Params{
		DefaultCost: sdk.NewInt64Coin(DefaultFeeDefinitionDenom, 1),
		ConversionFactor: ConversionFactor{
			DefinitionAmount: sdk.NewInt64Coin(DefaultFeeDefinitionDenom, 1),
			ConvertedAmount:  sdk.NewInt64Coin(pioconfig.GetProvConfig().FeeDenom, 1),
		},
	}
}

func (p Params) Validate() error {
	if err := p.DefaultCost.Validate(); err != nil {
		return fmt.Errorf("invalid default cost %q: %w", p.DefaultCost, err)
	}
	if err := p.ConversionFactor.Validate(); err != nil {
		return fmt.Errorf("invalid conversion factor: %w", err)
	}
	if p.DefaultCost.Denom != p.ConversionFactor.DefinitionAmount.Denom {
		return fmt.Errorf("default cost denom %q does not equal conversion factor base amount denom %q",
			p.DefaultCost.Denom, p.ConversionFactor.DefinitionAmount.Denom)
	}
	return nil
}

// DefaultCostCoins returns the default cost wrapped in a Coins.
func (p Params) DefaultCostCoins() sdk.Coins {
	if p.DefaultCost.IsNil() || p.DefaultCost.IsZero() {
		return nil
	}
	return sdk.Coins{p.DefaultCost}
}

func NewMsgFee(msgTypeURL string, cost ...sdk.Coin) *MsgFee {
	rv := &MsgFee{
		MsgTypeUrl: msgTypeURL,
	}
	// Adding each cost coin in so that the result is always valid and sorted.
	for _, coin := range cost {
		rv.Cost = rv.Cost.Add(coin)
	}
	return rv
}

func (m *MsgFee) String() string {
	if m == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s=%s", valueOr(m.MsgTypeUrl, `""`), valueOr(m.Cost.String(), "<free>"))
}

// valueOr returns the first string if it isn't empty, otherwise returns the second.
func valueOr(val string, or string) string {
	if len(val) > 0 {
		return val
	}
	return or
}

func (m *MsgFee) Validate() error {
	if m == nil {
		return errors.New("nil MsgFee not allowed")
	}

	if err := ValidateMsgTypeURL(m.MsgTypeUrl); err != nil {
		return err
	}

	if err := m.Cost.Validate(); err != nil {
		return fmt.Errorf("invalid %s cost %q: %w", m.MsgTypeUrl, m.Cost, err)
	}

	return nil
}

// ValidateMsgTypeURL returns an error if there's a problem with the provided msgTypeUrl.
func ValidateMsgTypeURL(msgTypeURL string) error {
	if len(msgTypeURL) == 0 {
		return errors.New("msg type url cannot be empty")
	}
	if len(msgTypeURL) > MaxMsgTypeURLLen {
		return fmt.Errorf("msg type url %q length (%d) exceeds max length (%d)",
			msgTypeURL[:5]+"..."+msgTypeURL[len(msgTypeURL)-5:], len(msgTypeURL), MaxMsgTypeURLLen)
	}
	return nil
}

func (c ConversionFactor) Validate() error {
	if err := c.DefinitionAmount.Validate(); err != nil {
		return fmt.Errorf("invalid base amount %q: %w", c.DefinitionAmount, err)
	}
	if c.DefinitionAmount.IsZero() {
		return fmt.Errorf("invalid base amount %q: cannot be zero", c.DefinitionAmount)
	}

	if err := c.ConvertedAmount.Validate(); err != nil {
		return fmt.Errorf("invalid converted amount %q: %w", c.ConvertedAmount, err)
	}
	if c.ConvertedAmount.IsZero() {
		return fmt.Errorf("invalid converted amount %q: cannot be zero", c.ConvertedAmount)
	}

	if c.DefinitionAmount.Denom == c.ConvertedAmount.Denom && !c.DefinitionAmount.Amount.Equal(c.ConvertedAmount.Amount) {
		return fmt.Errorf("base amount %q and converted amount %q cannot have different amounts when the denoms are the same",
			c.DefinitionAmount, c.ConvertedAmount)
	}
	return nil
}

func (c ConversionFactor) String() string {
	return fmt.Sprintf("*%s/%s", c.DefinitionAmount, c.ConvertedAmount)
}

// ConvertCoin converts the provided coin into the equivalent amount in this conversion factor's converted denom.
// If the provided coin doesn't have the same denom as the DefinitionAmount, the provided coin is returned unchanged.
// Otherwise, this is essentially ceil(toConvert * ConvertedAmount / DefinitionAmount).
// See also: ConvertMsgFee.
func (c ConversionFactor) ConvertCoin(toConvert sdk.Coin) sdk.Coin {
	// If the toConvert isn't in the convertable denom, just return it back.
	// If the conversion factor is 1-1 with the same denom, there's nothing to convert.
	if toConvert.Denom != c.DefinitionAmount.Denom || c.DefinitionAmount.Equal(c.ConvertedAmount) {
		return toConvert
	}
	// If the conversion factor is 1-1 with different denoms, we can use the base amount with the converted denom.
	if c.DefinitionAmount.Amount.Equal(c.ConvertedAmount.Amount) {
		return sdk.NewCoin(c.ConvertedAmount.Denom, toConvert.Amount)
	}

	// Gotta do it the math way.
	top := toConvert.Amount.Mul(c.ConvertedAmount.Amount)
	bot := c.DefinitionAmount.Amount
	amt := top.Quo(bot)
	if r := top.Mod(bot); !r.IsZero() {
		amt = amt.AddRaw(1)
	}
	return sdk.NewCoin(c.ConvertedAmount.Denom, amt)
}

// ConvertCoins returns the provided coins with any applicable coin being converted.
func (c ConversionFactor) ConvertCoins(toConvert sdk.Coins) sdk.Coins {
	var rv sdk.Coins
	for _, coin := range toConvert {
		rv = rv.Add(c.ConvertCoin(coin))
	}
	return rv
}

// ConvertMsgFee returns a new MsgFee with the coin fields converted into the equivalent amounts in this conversion factor's converted denom.
// See also: ConvertCoin.
func (c ConversionFactor) ConvertMsgFee(msgFee *MsgFee) *MsgFee {
	if msgFee == nil {
		return nil
	}
	return &MsgFee{
		MsgTypeUrl: msgFee.MsgTypeUrl,
		Cost:       c.ConvertCoins(msgFee.Cost),
	}
}
