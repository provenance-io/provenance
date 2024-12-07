package keeper

import (
	"cosmossdk.io/collections"
	"github.com/provenance-io/provenance/x/nav"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

// NAVs exposes the keeper.navs field.
func (k Keeper) NAVs() collections.Map[collections.Pair[string, string], nav.NetAssetValueRecord] {
	return k.navs
}
