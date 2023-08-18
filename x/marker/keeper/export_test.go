package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

// GetMarkerModuleAddr is a TEST ONLY exposure of the markerModuleAddr value.
func (k Keeper) GetMarkerModuleAddr() sdk.AccAddress {
	return k.markerModuleAddr
}

// GetIbcTransferModuleAddr is a TEST ONLY exposure of the ibcTransferModuleAddr value.
func (k Keeper) GetIbcTransferModuleAddr() sdk.AccAddress {
	return k.ibcTransferModuleAddr
}

// CanForceTransferFrom is a TEST ONLY exposure of the canForceTransferFrom value.
func (k Keeper) CanForceTransferFrom(ctx sdk.Context, from sdk.AccAddress) bool {
	return k.canForceTransferFrom(ctx, from)
}
