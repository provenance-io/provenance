package v042

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

var (
	// MarkerStoreKeyPrefixLegacy legacy prefix for marker-address < v043
	MarkerStoreKeyPrefixLegacy = []byte{0x01}
)

// MarkerStoreKeyLegacy turn an address to key used to get it from the account store
func MarkerStoreKeyLegacy(addr sdk.AccAddress) []byte {
	return append(MarkerStoreKeyPrefixLegacy, addr.Bytes()...)
}

// MarkerKeeperI is a minimal set of marker keeper operations required for store migrations
type MarkerKeeperI interface {
	// Set a marker in the auth account store
	SetMarker(sdk.Context, types.MarkerAccountI)
	// IterateMarker processes all markers with the given handler function.
	IterateMarkers(sdk.Context, func(types.MarkerAccountI) bool)
}
