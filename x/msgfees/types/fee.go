package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	UsdDenom   string = "usd"
	NhashDenom string = "nhash"
)

// SplitAmount returns split of Amount to be used for coin recipient and one for payout of fee, NOTE: this should only be used if a Recipient address exists
func SplitAmount(coin sdk.Coin) (recipientCoin sdk.Coin, feePayoutCoin sdk.Coin) {
	addFeeToPay := coin.Amount.Uint64()
	addFeeToPay /= 2
	feePayoutCoin = sdk.NewCoin(coin.Denom, sdk.NewIntFromUint64(addFeeToPay))
	recipientCoin = coin.Sub(feePayoutCoin)
	return recipientCoin, feePayoutCoin
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
