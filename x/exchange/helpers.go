package exchange

import sdk "github.com/cosmos/cosmos-sdk/types"

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

// CoinsEquals returns true if the two provided coins are equal.
//
// sdk.Coins.IsEqual will panic if the two have the same number of entries, but different denoms.
// This one will return false in that case instead of panicking.
func CoinsEquals(a, b sdk.Coins) (isEqual bool) {
	defer func() {
		if r := recover(); r != nil {
			isEqual = false
		}
	}()
	return a.IsEqual(b)
}

// CoinEquals returns true if the two provided coin entries are equal.
// Designed for use with intersection.
//
// We can't just provide sdk.Coin.IsEqual to intersection because that PANICS if the denoms are different.
// And we can't provide sdk.Coin.Equal to intersection because it takes in an interface{} (instead of sdk.Coin).
func CoinEquals(a, b sdk.Coin) bool {
	return a.Equal(b)
}

// IntersectionOfCoin returns each sdk.Coin entry that is in both lists.
func IntersectionOfCoin(list1, list2 []sdk.Coin) []sdk.Coin {
	return intersection(list1, list2, CoinEquals)
}
