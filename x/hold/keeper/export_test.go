package keeper

import (
	sdkmath "cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

// HoldAccountBalancesInvariantHelper exposes the holdAccountBalancesInvariantHelper function for unit tests.
var HoldAccountBalancesInvariantHelper = holdAccountBalancesInvariantHelper

// WithBankKeeper returns a new keeper that uses the provided bank keeper for unit tests.
func (k Keeper) WithBankKeeper(bk hold.BankKeeper) Keeper {
	k.bankKeeper = bk
	return k
}

// GetStoreKey exposes this keeper's storekey for unit tests.
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// SetHoldCoinAmount exposes this keeper's setHoldCoinAmount function for unit tests.
func (k Keeper) SetHoldCoinAmount(store sdk.KVStore, addr sdk.AccAddress, denom string, amount sdkmath.Int) error {
	return k.setHoldCoinAmount(store, addr, denom, amount)
}
