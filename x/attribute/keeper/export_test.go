package keeper

import "github.com/provenance-io/provenance/x/attribute/types"

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

// GetMarkerModuleAddr is a TEST ONLY way of getting a new attribute keeper with an injected name keeper.
func (k Keeper) WithNameKeeper(nameK types.NameKeeper) Keeper {
	k.nameKeeper = nameK
	return k
}
