package keeper

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

func (k Keeper) GetStore(ctx sdk.Context) storetypes.KVStore {
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

// GetFeeCollectorAddr is a TEST ONLY exposure of the feeCollectorAddr value.
func (k Keeper) GetFeeCollectorAddr() sdk.AccAddress {
	return k.feeCollectorAddr
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

// WithAttrKeeper is a TEST ONLY func that returns a copy of this marker keeper but with the provided attr keeper instead.
func (k Keeper) WithAttrKeeper(attrKeeper types.AttrKeeper) Keeper {
	k.attrKeeper = attrKeeper
	return k
}

// SetNewMarker is a TEST ONLY function that calls NewMarker, then SetMarker.
func (k Keeper) SetNewMarker(ctx sdk.Context, marker types.MarkerAccountI) {
	k.SetMarker(ctx, k.NewMarker(ctx, marker))
}
