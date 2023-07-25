package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/escrow"
)

var _ banktypes.GetLockedCoinsFn = Keeper{}.GetLockedCoins

// GetLockedCoins gets all the coins that are in escrow for the given address.
func (k Keeper) GetLockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	if escrow.HasBypass(ctx) {
		return nil
	}
	rv, err := k.GetEscrowCoins(ctx, addr)
	if err != nil {
		panic(err)
	}
	return rv
}
