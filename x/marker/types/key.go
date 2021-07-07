package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/tendermint/tendermint/crypto"
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
