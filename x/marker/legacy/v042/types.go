package v042

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// MarkerStoreKeyPrefixLegacy legacy prefix for marker-address < v043
	MarkerStoreKeyPrefixLegacy = []byte{0x01}
)

// MarkerStoreKeyLegacy turn an address to key used to get it from the account store
func MarkerStoreKeyLegacy(addr sdk.AccAddress) []byte {
	return append(MarkerStoreKeyPrefixLegacy, addr.Bytes()...)
}
