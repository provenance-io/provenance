package exchange

import (
	"math/big"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// contains returns true if the provided toFind is present in the provided vals.
func contains[T any](vals []T, toFind T, equals func(T, T) bool) bool {
	for _, v := range vals {
		if equals(toFind, v) {
			return true
		}
	}
	return false
}

// intersection returns each entry that is in both lists.
func intersection[T any](list1, list2 []T, equals func(T, T) bool) []T {
	var rv []T
	for _, a := range list1 {
		if contains(list2, a, equals) && !contains(rv, a, equals) {
			rv = append(rv, a)
		}
	}
	return rv
}

// EqualsUint64 returns true if the two uint64 values provided are equal.
func EqualsUint64(a, b uint64) bool {
	return a == b
}

// ContainsUint64 returns true if the uint64 to find is in the vals slice.
func ContainsUint64(vals []uint64, toFind uint64) bool {
	return contains(vals, toFind, EqualsUint64)
}

// IntersectionUint64 returns each uint64 that is in both lists.
func IntersectionUint64(a, b []uint64) []uint64 {
	return intersection(a, b, EqualsUint64)
}

// ContainsString returns true if the string to find is in the vals slice.
func ContainsString(vals []string, toFind string) bool {
	return contains(vals, toFind, func(a, b string) bool {
		return a == b
	})
}

// CoinEquals returns true if the two provided coin entries are equal.
// Designed for use with intersection.
//
// We can't provide sdk.Coin.Equal to intersection because it takes in an interface{} (instead of sdk.Coin).
func CoinEquals(a, b sdk.Coin) bool {
	return a.Equal(b)
}

// IntersectionOfCoin returns each sdk.Coin entry that is in both lists.
func IntersectionOfCoin(list1, list2 []sdk.Coin) []sdk.Coin {
	return intersection(list1, list2, CoinEquals)
}

// ContainsCoin returns true if the coin to find is in the vals slice.
func ContainsCoin(vals []sdk.Coin, toFind sdk.Coin) bool {
	return contains(vals, toFind, CoinEquals)
}

// ContainsCoinWithSameDenom returns true if there's an entry in vals with the same denom as the denom to find.
func ContainsCoinWithSameDenom(vals []sdk.Coin, toFind sdk.Coin) bool {
	return contains(vals, toFind, func(a, b sdk.Coin) bool {
		return a.Denom == b.Denom
	})
}

// MinSDKInt returns the lesser of the two provided ints.
func MinSDKInt(a, b sdkmath.Int) sdkmath.Int {
	if a.LTE(b) {
		return a
	}
	return b
}

// QuoRemInt does a/b returning the integer result and remainder such that a = quo * b + rem
// If y == 0, a division-by-zero run-time panic occurs.
//
// QuoRem implements T-division and modulus (like Go):
//
//	quo = x/y      with the result truncated to zero
//	rem = x - y*q
//
// (See Daan Leijen, “Division and Modulus for Computer Scientists”.)
func QuoRemInt(a, b sdkmath.Int) (quo sdkmath.Int, rem sdkmath.Int) {
	var q, r big.Int
	q.QuoRem(a.BigInt(), b.BigInt(), &r)
	quo = sdkmath.NewIntFromBigInt(&q)
	rem = sdkmath.NewIntFromBigInt(&r)
	return
}
