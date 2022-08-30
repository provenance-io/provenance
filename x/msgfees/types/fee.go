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
	amount := sdk.NewDec(coin.Amount.Int64())
	percentage := numerator.Quo(denominator)
	splitAmount1 := amount.Mul(percentage).TruncateInt()
	splitAmount2 := coin.Amount.Sub(splitAmount1)
	feePayoutCoin = sdk.NewCoin(coin.Denom, splitAmount1)
	recipientCoin = sdk.NewCoin(coin.Denom, splitAmount2)
	return recipientCoin, feePayoutCoin
}
