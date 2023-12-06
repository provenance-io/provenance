package types

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	UsdDenom string = "usd"
)

// SplitCoinByBips returns split to recipient and fee module based on basis points for recipient
// if bips set to 100bips recipient gets all the fees.
func SplitCoinByBips(coin sdk.Coin, bips uint32) (recipientCoin sdk.Coin, feePayoutCoin sdk.Coin, err error) {
	if bips > 10_000 {
		return recipientCoin, feePayoutCoin, ErrInvalidBipsValue.Wrapf("invalid: %v", bips)
	}
	// nothing to calculate if recipient gets 10_000 bips or 100%, short circuit
	if bips == 10_000 {
		recipientCoin = coin
		feePayoutCoin = sdk.NewInt64Coin(coin.Denom, 0)
		return recipientCoin, feePayoutCoin, nil
	}
	numerator := sdkmath.LegacyNewDec(int64(bips))
	denominator := sdkmath.LegacyNewDec(10_000)
	decAmount := sdkmath.LegacyNewDec(coin.Amount.Int64())
	percentage := numerator.Quo(denominator)
	bipsAmount := decAmount.Mul(percentage).TruncateInt()
	feePayoutAmount := coin.Amount.Sub(bipsAmount)

	recipientCoin = sdk.NewCoin(coin.Denom, bipsAmount)
	feePayoutCoin = sdk.NewCoin(coin.Denom, feePayoutAmount)
	return recipientCoin, feePayoutCoin, nil
}

// MsgFeesDistribution holds information on message based fees that should be collected.
type MsgFeesDistribution struct {
	// TotalAdditionalFees is the total of all additional fees.
	TotalAdditionalFees sdk.Coins
	// AdditionalModuleFees is just the additional fees to send to the module.
	AdditionalModuleFees sdk.Coins
	// RecipientDistributions is just the additional specific distribution fees.
	RecipientDistributions map[string]sdk.Coins
}

// Increase adds the provided coin to be distributed (as long as it's positive).
// If there's no recipient, it all goes to the module. Otherwise, it's split using bips between recipient and module.
func (d *MsgFeesDistribution) Increase(coin sdk.Coin, bips uint32, recipient string) error {
	if !coin.IsPositive() {
		return nil
	}

	d.TotalAdditionalFees = d.TotalAdditionalFees.Add(coin)

	if len(recipient) == 0 {
		d.AdditionalModuleFees = d.AdditionalModuleFees.Add(coin)
		return nil
	}

	recipientCoin, feePayoutCoin, err := SplitCoinByBips(coin, bips)
	if err != nil {
		return err
	}

	d.RecipientDistributions[recipient] = d.RecipientDistributions[recipient].Add(recipientCoin)
	// fee payout for module for now will be zero, keeping it still here if we goto  bips at a message level or a param this will come back into play.
	if !feePayoutCoin.IsZero() {
		d.AdditionalModuleFees = d.AdditionalModuleFees.Add(feePayoutCoin)
	}

	return nil
}
