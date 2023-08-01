package types

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the module
	ModuleName = "marker"

	// StoreKey is string representation of the store key for marker
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for queries
	QuerierRoute = ModuleName

	// CoinPoolName to be used for coin pool associated with mint/burn activities.
	CoinPoolName = ModuleName

	// DefaultParamspace is the name used for the parameter subspace for this module.
	DefaultParamspace = ModuleName
)

var (
	// MarkerStoreKeyPrefix prefix for marker-address reference (improves iterator performance over auth accounts)
	MarkerStoreKeyPrefix = []byte{0x02}

	// DenySendKeyPrefix prefix for adding addresses that are denied send functionality on restricted markers
	DenySendKeyPrefix = []byte{0x03}

	// MarkerNetAssetValuePrefix prefix for net asset values of markers
	MarkerNetAssetValuePrefix = []byte{0x04}
)

// MarkerAddress returns the module account address for the given denomination
func MarkerAddress(denom string) (sdk.AccAddress, error) {
	if err := sdk.ValidateDenom(denom); err != nil {
		return nil, err
	}
	return sdk.AccAddress(crypto.AddressHash([]byte(fmt.Sprintf("%s/%s", ModuleName, denom)))), nil
}

// MustGetMarkerAddress returns the module account address for the given denomination, panics on error
func MustGetMarkerAddress(denom string) sdk.AccAddress {
	addr, err := MarkerAddress(denom)
	if err != nil {
		panic(err)
	}
	return addr
}

// MarkerStoreKey turn an address to key used to get it from the account store
func MarkerStoreKey(addr sdk.AccAddress) []byte {
	return append(MarkerStoreKeyPrefix, address.MustLengthPrefix(addr.Bytes())...)
}

// SplitMarkerStoreKey returns an account address given a store key, uses the length prefix to determine length of AccAddress
func SplitMarkerStoreKey(key []byte) sdk.AccAddress {
	return sdk.AccAddress(key[2 : key[1]+2])
}

// DenySendKey returns a key [prefix][denom addr][deny addr] for send deny list for restricted markers
func DenySendKey(markerAddr sdk.AccAddress, denyAddr sdk.AccAddress) []byte {
	key := DenySendKeyPrefix
	key = append(key, address.MustLengthPrefix(markerAddr.Bytes())...)
	return append(key, address.MustLengthPrefix(denyAddr.Bytes())...)
}

// MarkerNetAssetValueKey returns key [prefix][marker address] for marker net asset values
func MarkerNetAssetValueKeyPrefix(markerAddr sdk.AccAddress) []byte {
	return append(MarkerNetAssetValuePrefix, address.MustLengthPrefix(markerAddr.Bytes())...)
}

// MarkerNetAssetValueKey returns key [prefix][marker address][asset denom value] for marker net asset value by value denom
func MarkerNetAssetValueKey(markerAddr sdk.AccAddress, denom string) []byte {
	return append(MarkerNetAssetValueKeyPrefix(markerAddr), denom...)
}
