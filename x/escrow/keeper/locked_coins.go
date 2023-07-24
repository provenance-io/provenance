package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ banktypes.GetLockedCoinsFn = Keeper{}.GetLockedCoins

// GetLockedCoins gets all the coins that are in escrow for the given address.
func (k Keeper) GetLockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	var rv sdk.Coins
	err := k.IterateEscrow(ctx, addr, func(coin sdk.Coin) bool {
		rv = rv.Add(coin)
		return false
	})
	if err != nil {
		panic(err)
	}
	return rv
}
