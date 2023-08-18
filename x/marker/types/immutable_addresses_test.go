package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// newAddr creates a new sdk.AccAddress for the given index.
func newAddr(i int) sdk.AccAddress {
	return sdk.AccAddress(fmt.Sprintf("addr_%d______________", i))[:20]
}

// addrz creates a new slice of addresses with the given number of entries.
func addrz(count int) []sdk.AccAddress {
	rv := make([]sdk.AccAddress, count)
	for i := range rv {
		rv[i] = newAddr(i + 1)
	}
	return rv
}

func TestImmutableAccAddresses(t *testing.T) {
	newGetTests := []struct {
		name  string
		addrs []sdk.AccAddress
	}{
		{name: "nil", addrs: nil},
		{name: "empty", addrs: []sdk.AccAddress{}},
		{name: "one addr", addrs: addrz(1)},
		{name: "two addr", addrs: addrz(2)},
		{name: "five addrs", addrs: addrz(5)},
	}

	for _, tc := range newGetTests {
		t.Run("new+get:"+tc.name, func(t *testing.T) {
			var iAddrs ImmutableAccAddresses
			testFunc := func() {
				iAddrs = NewImmutableAccAddresses(tc.addrs)
			}
			require.NotPanics(t, testFunc, "NewImmutableAccAddresses")

			// Get the addrs and make sure the result is equal to what we provided,
			// but isn't the exact same slice that we provided. Also make sure each
			// address was copied to a new slice too.
			addrs := iAddrs.GetSlice()
			require.Equal(t, tc.addrs, addrs, "GetSlice() compared to expected")
			require.NotSame(t, tc.addrs, addrs, "GetSlice() compared to expected")
			for i := range tc.addrs {
				require.NotSame(t, tc.addrs[i], addrs[i], "GetSlice()[%d] compared to expected", i)
			}

			// Get the slice five more times and make sure each one is a fresh copy.
			for x := 1; x <= 5; x++ {
				addrs2 := iAddrs.GetSlice()
				require.Equal(t, addrs, addrs2, "[%d]: GetSlice() compared to previous result", x)
				require.NotSame(t, addrs, addrs2, "[%d]: GetSlice() compared to previous result", x)
				for i := range tc.addrs {
					require.NotSame(t, addrs[i], addrs2[i], "[%d]: GetSlice()[%d] compared to previous result", x, i)
				}
				addrs = addrs2
			}
		})
	}

	firstWDiffFirst := newAddr(1)
	firstWDiffFirst[0] = 'b'
	firstWDiffLast := newAddr(1)
	firstWDiffLast[len(firstWDiffLast)-1]++
	firstWDiffMiddle := newAddr(1)
	firstWDiffMiddle[12]--

	iAddrsHasTester := NewImmutableAccAddresses(addrz(5))

	hasTests := []struct {
		name string
		addr sdk.AccAddress
		exp  bool
	}{
		{name: "nil", addr: nil, exp: false},
		{name: "empty", addr: sdk.AccAddress{}, exp: false},
		{name: "zeroth address", addr: newAddr(0), exp: false},
		{name: "first address", addr: newAddr(1), exp: true},
		{name: "second address", addr: newAddr(2), exp: true},
		{name: "third address", addr: newAddr(3), exp: true},
		{name: "fourth address", addr: newAddr(4), exp: true},
		{name: "fifth address", addr: newAddr(5), exp: true},
		{name: "sixth address", addr: newAddr(6), exp: false},
		{name: "first address without last byte", addr: newAddr(1)[:19], exp: false},
		{name: "first address without first byte", addr: newAddr(1)[1:], exp: false},
		{name: "first address with first byte changed", addr: firstWDiffFirst, exp: false},
		{name: "first address with last byte changed", addr: firstWDiffLast, exp: false},
		{name: "first address with a middle byte changed", addr: firstWDiffMiddle, exp: false},
		{name: "first address with extra byte at end", addr: append(newAddr(1), '_'), exp: false},
		{name: "first address with extra byte at start", addr: append(sdk.AccAddress{'_'}, newAddr(1)...), exp: false},
	}

	for _, tc := range hasTests {
		t.Run("has:"+tc.name, func(t *testing.T) {
			var has bool
			testFunc := func() {
				has = iAddrsHasTester.Has(tc.addr)
			}
			require.NotPanics(t, testFunc, "Has")
			assert.Equal(t, tc.exp, has, "Has")
		})
	}

	t.Run("change to slice provided to NewImmutableAccAddresses", func(t *testing.T) {
		// This test makes sure that a change to the slice provided to NewImmutableAccAddresses
		// does not alter anything inside the resulting ImmutableAccAddresses.
		input := addrz(3)
		expected00 := input[0][0]
		iaddrs := NewImmutableAccAddresses(input)
		input[0][0] = 'b'
		addrs := iaddrs.GetSlice()
		actual00 := addrs[0][0]
		assert.Equal(t, expected00, actual00, "first byte of first address returned by GetSlice()")
		hasChanged := iaddrs.Has(input[0])
		assert.False(t, hasChanged, "Has(address that was changed)")
		hasOrig := iaddrs.Has(newAddr(1))
		assert.True(t, hasOrig, "Has(address before it was changed)")
	})

	t.Run("change to slice returned by GetSlice", func(t *testing.T) {
		// This test makes sure that a change to the result of a GetSlice()
		// doesn't somehow change anything in the ImmutableAccAddresses.
		iaddrs := NewImmutableAccAddresses(addrz(4))
		addrs := iaddrs.GetSlice()
		expected00 := addrs[0][0]
		addrs[0][0] = 'b'
		actual := iaddrs.GetSlice()
		actual00 := actual[0][0]
		assert.Equal(t, expected00, actual00, "first byte of first address returned by GetSlice()")
		hasChanged := iaddrs.Has(addrs[0])
		assert.False(t, hasChanged, "Has(address that was changed)")
		hasOrig := iaddrs.Has(newAddr(1))
		assert.True(t, hasOrig, "Has(address before it was changed)")
	})
}

func TestDeepCopyAccAddresses(t *testing.T) {
	tests := []struct {
		name string
		orig []sdk.AccAddress
	}{
		{name: "nil", orig: nil},
		{name: "empty", orig: []sdk.AccAddress{}},
		{name: "one address", orig: addrz(1)},
		{name: "two address", orig: addrz(2)},
		{name: "five address", orig: addrz(5)},
		{name: "same address 3 times", orig: []sdk.AccAddress{newAddr(1), newAddr(1), newAddr(1)}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expCap := len(tc.orig) // Result should not have extra capacity.
			var actual []sdk.AccAddress
			testFunc := func() {
				actual = deepCopyAccAddresses(tc.orig)
			}
			require.NotPanics(t, testFunc, "deepCopyAccAddresses")

			// Make sure the result is equal to what was provided, but in a different slice.
			// Also make sure each entry slice was also copied.
			if assert.Equal(t, tc.orig, actual, "deepCopyAccAddresses result") {
				if assert.NotSame(t, tc.orig, actual, "deepCopyAccAddresses result") {
					for i := range tc.orig {
						assert.NotSame(t, tc.orig[i], actual[i], "deepCopyAccAddresses result[%d]", i)
					}
				}
			}

			// Make sure there isn't any extra capacity
			actualCap := cap(actual)
			assert.Equal(t, expCap, actualCap, "capacity of deepCopyAccAddresses result")
		})
	}
}
