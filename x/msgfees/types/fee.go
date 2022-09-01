package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	UsdDenom   string = "usd"
	NhashDenom string = "nhash"
)

// SplitCoinByBips returns split to recipient and fee module based on basis points for recipient
func SplitCoinByBips(coin sdk.Coin, bips uint32) (recipientCoin sdk.Coin, feePayoutCoin sdk.Coin, err error) {
	if bips > 10_000 {
		return recipientCoin, feePayoutCoin, ErrInvalidBipsValue.Wrapf("invalid: %v", bips)
	}
	numerator := sdk.NewDec(int64(bips))
	denominator := sdk.NewDec(10_000)
	decAmount := sdk.NewDec(coin.Amount.Int64())
	percentage := numerator.Quo(denominator)
	bipsAmount := decAmount.Mul(percentage).TruncateInt()
	feePayoutAmount := coin.Amount.Sub(bipsAmount)

	recipientCoin = sdk.NewCoin(coin.Denom, bipsAmount)
	feePayoutCoin = sdk.NewCoin(coin.Denom, feePayoutAmount)
	return recipientCoin, feePayoutCoin, nil
}
