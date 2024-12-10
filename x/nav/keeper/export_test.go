package keeper

import (
	"cosmossdk.io/collections"

	"github.com/provenance-io/provenance/x/nav"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

// NAVs is a test-only exposure of the Keeper.navs field.
func (k Keeper) NAVs() collections.Map[collections.Pair[string, string], nav.NetAssetValueRecord] {
	return k.navs
}
