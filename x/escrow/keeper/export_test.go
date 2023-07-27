package keeper

import (
	sdkmath "cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/escrow"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

// EscrowAccountBalancesInvariantHelper exposes the escrowAccountBalancesInvariantHelper function for unit tests.
var EscrowAccountBalancesInvariantHelper = escrowAccountBalancesInvariantHelper

// WithBankKeeper returns a new keeper that uses the provided bank keeper for unit tests.
func (k Keeper) WithBankKeeper(bk escrow.BankKeeper) Keeper {
	k.bankKeeper = bk
	return k
}

// GetStoreKey exposes this keeper's storekey for unit tests.
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// SetEscrowCoinAmount exposes this keeper's setEscrowCoinAmount function for unit tests.
func (k Keeper) SetEscrowCoinAmount(store sdk.KVStore, addr sdk.AccAddress, denom string, amount sdkmath.Int) error {
	return k.setEscrowCoinAmount(store, addr, denom, amount)
}
