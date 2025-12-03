package keeper

import (
	sdkmath "cosmossdk.io/math"
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

// WithBankKeeper returns a new keeper that uses the provided account keeper for unit tests.
func (k Keeper) WithAccountKeeper(ak hold.AccountKeeper) Keeper {
	k.accountKeeper = ak
	return k
}

// SetHoldCoinAmount exposes this keeper's setHoldCoinAmount function for unit tests.
func (k Keeper) SetHoldCoinAmount(ctx sdk.Context, addr sdk.AccAddress, denom string, amount sdkmath.Int) error {
	return k.setHoldCoinAmount(ctx, addr, denom, amount)
}

// GetSpendableForDenoms exposes this keeper's getSpendableForDenoms function for unit tests.
func (k Keeper) GetSpendableForDenoms(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) sdk.Coins {
	return k.getSpendableForDenoms(ctx, addr, funds)
}