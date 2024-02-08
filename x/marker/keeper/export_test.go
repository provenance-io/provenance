package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/marker/types"
)

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

func (k Keeper) GetStore(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(k.storeKey)
}

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

// WithBankKeeper is a TEST ONLY func that returns a copy of this marker keeper but with the provided bank keeper instead.
func (k Keeper) WithBankKeeper(bankKeeper types.BankKeeper) Keeper {
	k.bankKeeper = bankKeeper
	return k
}

// WithAuthzKeeper is a TEST ONLY func that returns a copy of this marker keeper but with the provided authz keeper instead.
func (k Keeper) WithAuthzKeeper(authzKeeper types.AuthzKeeper) Keeper {
	k.authzKeeper = authzKeeper
	return k
}
