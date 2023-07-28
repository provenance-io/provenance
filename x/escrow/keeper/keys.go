package keeper

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// Keys for store prefixes.
// Items are stored with the following keys:
//
// Coin in escrow:
// - 0x00<addr len (1 byte)><addr><denom> -> <amount>
var (
	// KeyPrefixEscrowCoin is the prefix of an escrow entry for an address and single denom
	KeyPrefixEscrowCoin = []byte{0x00}
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

// createEscrowCoinKeyAddrPrefixPlusCap creates an escrow coin key prefix containing the provided address.
// The resulting slice will have the provided amount of extra capacity (in case you want to append something to it).
func createEscrowCoinKeyAddrPrefixPlusCap(addr sdk.AccAddress, extraCap int) []byte {
	return concatBzPlusCap(KeyPrefixEscrowCoin, address.MustLengthPrefix(addr), extraCap)
}

// CreateEscrowCoinKeyAddrPrefix creates an escrow coin key prefix containing the provided address.
// It's useful for iterating over all funds in escrow for an address.
func CreateEscrowCoinKeyAddrPrefix(addr sdk.AccAddress) []byte {
	return createEscrowCoinKeyAddrPrefixPlusCap(addr, 0)
}

// CreateEscrowCoinKey creates an escrow coin key for the provided address and denom.
func CreateEscrowCoinKey(addr sdk.AccAddress, denom string) []byte {
	rv := createEscrowCoinKeyAddrPrefixPlusCap(addr, len(denom))
	rv = append(rv, []byte(denom)...)
	return rv
}

// ParseEscrowCoinKey parses a full escrow coin key into its address and denom.
func ParseEscrowCoinKey(key []byte) (sdk.AccAddress, string) {
	return ParseEscrowCoinKeyUnprefixed(key[1:])
}

// ParseEscrowCoinKeyUnprefixed parses an escrow coin key without the type prefix into its address and denom.
func ParseEscrowCoinKeyUnprefixed(key []byte) (sdk.AccAddress, string) {
	addr, denom := parseLengthPrefixedBz(key)
	return addr, string(denom)
}

// UnmarshalEscrowCoinValue parses the store value of an escrow coin entry back into it's Int form.
func UnmarshalEscrowCoinValue(value []byte) (sdkmath.Int, error) {
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
