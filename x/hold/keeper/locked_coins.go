package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	// banktypes "github.com/cosmos/cosmos-sdk/x/bank/types" // TODO[1760]: locked-coins

	"github.com/provenance-io/provenance/x/hold"
)

// var _ banktypes.GetLockedCoinsFn = Keeper{}.GetLockedCoins // TODO[1760]: locked-coins

// GetLockedCoins gets all the coins that are on hold for the given address.
func (k Keeper) GetLockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	if hold.HasBypass(ctx) {
		return nil
	}
	rv, err := k.GetHoldCoins(ctx, addr)
	if err != nil {
		panic(err)
	}
	return rv
}
