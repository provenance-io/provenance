package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/hold"
)

var _ banktypes.GetLockedCoinsFn = Keeper{}.GetLockedCoins

// GetLockedCoins gets all the coins that are on hold for the given address.
func (k Keeper) GetLockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	if hold.HasBypass(ctx) {
		return nil
	}
	rv, err := k.GetHoldCoins(sdk.UnwrapSDKContext(ctx), addr)
	if err != nil {
		panic(err)
	}
	return rv
}
