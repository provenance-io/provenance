package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

// AddRecord is a TEST ONLY exposure of addRecord.
func (k Keeper) AddRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict, isModifiable bool) error {
	return k.addRecord(ctx, name, addr, restrict, isModifiable)
}
