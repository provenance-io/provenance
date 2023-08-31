package keeper

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// Keys for store prefixes.
// Items are stored with the following keys:
//
// Coin on hold:
// - 0x00<addr len (1 byte)><addr><denom> -> <amount>
var (
	// KeyPrefixHoldCoin is the prefix of a hold entry for an address and single denom.
	KeyPrefixHoldCoin = []byte{0x00}
)

// concatBzPlusCap creates a single byte slice consisting of the two provided byte slices with some extra capacity in the underlying array.
// The idea is that you can append(...) to the result of this without it needed a new underlying array.
func concatBzPlusCap(bz1, bz2 []byte, extraCap int) []byte {
	l1 := len(bz1)
	l2 := len(bz2)
	rv := make([]byte, l1+l2, l1+l2+extraCap)
	if l1 > 0 {
		copy(rv, bz1)
	}
	if l2 > 0 {
		copy(rv[l1:], bz2)
	}
	return rv
}

// parseLengthPrefixedBz parses a length-prefixed byte slice into those bytes and any leftover bytes.
func parseLengthPrefixedBz(bz []byte) ([]byte, []byte) {
	addrLen, addrLenEndIndex := sdk.ParseLengthPrefixedBytes(bz, 0, 1)
	addr, addrEndIndex := sdk.ParseLengthPrefixedBytes(bz, addrLenEndIndex+1, int(addrLen[0]))
	var remainder []byte
	if len(bz) > addrEndIndex+1 {
		remainder = bz[addrEndIndex+1:]
	}
	return addr, remainder
}

// createHoldCoinKeyAddrPrefixPlusCap creates a hold coin key prefix containing the provided address.
// The resulting slice will have the provided amount of extra capacity (in case you want to append something to it).
func createHoldCoinKeyAddrPrefixPlusCap(addr sdk.AccAddress, extraCap int) []byte {
	return concatBzPlusCap(KeyPrefixHoldCoin, address.MustLengthPrefix(addr), extraCap)
}

// CreateHoldCoinKeyAddrPrefix creates a hold coin key prefix containing the provided address.
// It's useful for iterating over all funds on hold for an address.
func CreateHoldCoinKeyAddrPrefix(addr sdk.AccAddress) []byte {
	return createHoldCoinKeyAddrPrefixPlusCap(addr, 0)
}

// CreateHoldCoinKey creates a hold coin key for the provided address and denom.
func CreateHoldCoinKey(addr sdk.AccAddress, denom string) []byte {
	rv := createHoldCoinKeyAddrPrefixPlusCap(addr, len(denom))
	rv = append(rv, []byte(denom)...)
	return rv
}

// ParseHoldCoinKey parses a full hold coin key into its address and denom.
func ParseHoldCoinKey(key []byte) (sdk.AccAddress, string) {
	return ParseHoldCoinKeyUnprefixed(key[1:])
}

// ParseHoldCoinKeyUnprefixed parses a hold coin key without the type prefix into its address and denom.
func ParseHoldCoinKeyUnprefixed(key []byte) (sdk.AccAddress, string) {
	addr, denom := parseLengthPrefixedBz(key)
	return addr, string(denom)
}

// UnmarshalHoldCoinValue parses the store value of a hold coin entry back into it's Int form.
func UnmarshalHoldCoinValue(value []byte) (sdkmath.Int, error) {
	if len(value) == 0 {
		return sdkmath.ZeroInt(), nil
	}
	var rv sdkmath.Int
	err := rv.Unmarshal(value)
	if err != nil {
		return sdkmath.ZeroInt(), err
	}
	return rv, nil
}
