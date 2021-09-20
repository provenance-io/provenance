package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Fee keeper calculates the additional fees to be charged
type FeeKeeper interface {
	GetFeeRate(ctx sdk.Context) (feeRate sdk.Dec)
}
