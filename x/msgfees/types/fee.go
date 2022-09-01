package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	UsdDenom   string = "usd"
	NhashDenom string = "nhash"
)

// SplitCoinByPercentage returns split of Percentage (0 - 100) for recipient and fee collector
func SplitCoinByPercentage(coin sdk.Coin, split uint32) (recipientCoin sdk.Coin, feePayoutCoin sdk.Coin) {
	numerator := sdk.NewDec(int64(split))
	denominator := sdk.NewDec(10_000)
	decAmount := sdk.NewDec(coin.Amount.Int64())
	percentage := numerator.Quo(denominator)
	bipsAmount := decAmount.Mul(percentage).TruncateInt()
	feePayoutAmount := coin.Amount.Sub(bipsAmount)

	recipientCoin = sdk.NewCoin(coin.Denom, bipsAmount)
	feePayoutCoin = sdk.NewCoin(coin.Denom, feePayoutAmount)
	return recipientCoin, feePayoutCoin
}
