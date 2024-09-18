package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func emptySdkContext() sdk.Context {
	return sdk.Context{}.WithContext(context.Background())
}

func TestSignersWrapper(t *testing.T) {
	addr1Acc := sdk.AccAddress("address_one_________")
	addr2Acc := sdk.AccAddress("address_one_________")
	addr1 := addr1Acc.String()
	addr2 := addr2Acc.String()

	strz := func(strings ...string) []string {
		rv := make([]string, 0, len(strings))
		rv = append(rv, strings...)
		return rv
	}
	accz := func(accs ...sdk.AccAddress) []sdk.AccAddress {
		rv := make([]sdk.AccAddress, 0, len(accs))
		rv = append(rv, accs...)
		return rv
	}

	tests := []struct {
		name       string
		wrapper    *keeper.SignersWrapper
		expStrings []string
		expAccs    []sdk.AccAddress
	}{
		{
			name:       "nil strings",
			wrapper:    keeper.NewSignersWrapper(nil),
			expStrings: nil,
			expAccs:    accz(),
		},
		{
			name:       "empty strings",
			wrapper:    keeper.NewSignersWrapper(strz()),
			expStrings: strz(),
			expAccs:    accz(),
		},
		{
			name:       "two valid address",
			wrapper:    keeper.NewSignersWrapper(strz(addr1, addr2)),
			expStrings: strz(addr1, addr2),
			expAccs:    accz(addr1Acc, addr2Acc),
		},
		{
			name:       "two invalid addresses",
			wrapper:    keeper.NewSignersWrapper(strz("bad1", "bad2")),
			expStrings: strz("bad1", "bad2"),
			expAccs:    accz(),
		},
		{
			name:       "three addresses first invalid",
			wrapper:    keeper.NewSignersWrapper(strz("bad1", addr1, addr2)),
			expStrings: strz("bad1", addr1, addr2),
			expAccs:    accz(addr1Acc, addr2Acc),
		},
		{
			name:       "three addresses second invalid",
			wrapper:    keeper.NewSignersWrapper(strz(addr1, "bad2", addr2)),
			expStrings: strz(addr1, "bad2", addr2),
			expAccs:    accz(addr1Acc, addr2Acc),
		},
		{
			name:       "three addresses third invalid",
			wrapper:    keeper.NewSignersWrapper(strz(addr1, addr2, "bad3")),
			expStrings: strz(addr1, addr2, "bad3"),
			expAccs:    accz(addr1Acc, addr2Acc),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualStrings := tc.wrapper.Strings()
			assert.Equal(t, tc.expStrings, actualStrings, ".String()")
			actualAccs := tc.wrapper.Accs()
			assert.Equal(t, tc.expAccs, actualAccs, ".Accs()")
		})
	}
}

func TestUnwrapMetadataContext(t *testing.T) {
	origCtx := emptySdkContext()
	var ctx sdk.Context
	testUnwrap := func() {
		ctx = keeper.UnwrapMetadataContext(origCtx)
	}
	require.NotPanics(t, testUnwrap, "UnwrapMetadataContext")
	var cache *types.AuthzCache
	testGet := func() {
		cache = types.GetAuthzCache(ctx)
	}
	require.NotPanics(t, testGet, "GetAuthzCache")
	assert.NotNil(t, cache, "cache returned by GetAuthzCache")
	assert.Empty(t, cache.GetAcceptableMap(), "cache acceptable map")
}

func TestSafeBech32ToAccAddresses(t *testing.T) {
	tests := []struct {
		name    string
		bech32s []string
		exp     []sdk.AccAddress
	}{
		{
			name:    "nil",
			bech32s: nil,
			exp:     []sdk.AccAddress{},
		},
		{
			name:    "empty",
			bech32s: []string{},
			exp:     []sdk.AccAddress{},
		},
		{
			name:    "one good",
			bech32s: []string{sdk.AccAddress("one_good_one________").String()},
			exp:     []sdk.AccAddress{sdk.AccAddress("one_good_one________")},
		},
		{
			name:    "one bad",
			bech32s: []string{"one_bad_one_________"},
			exp:     []sdk.AccAddress{},
		},
		{
			name:    "one empty",
			bech32s: []string{""},
			exp:     []sdk.AccAddress{},
		},
		{
			name: "three good",
			bech32s: []string{
				sdk.AccAddress("first_is_good_______").String(),
				sdk.AccAddress("second_is_good______").String(),
				sdk.AccAddress("third_is_good_______").String(),
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("first_is_good_______"),
				sdk.AccAddress("second_is_good______"),
				sdk.AccAddress("third_is_good_______"),
			},
		},
		{
			name: "three with first bad",
			bech32s: []string{
				"bad_first___________",
				sdk.AccAddress("second_is_good______").String(),
				sdk.AccAddress("third_is_good_______").String(),
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("second_is_good______"),
				sdk.AccAddress("third_is_good_______"),
			},
		},
		{
			name: "three with bad second",
			bech32s: []string{
				sdk.AccAddress("first_is_good_______").String(),
				"bad_second__________",
				sdk.AccAddress("third_is_good_______").String(),
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("first_is_good_______"),
				sdk.AccAddress("third_is_good_______"),
			},
		},
		{
			name: "three with bad third",
			bech32s: []string{
				sdk.AccAddress("first_is_good_______").String(),
				sdk.AccAddress("second_is_good______").String(),
				"bad_third___________",
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("first_is_good_______"),
				sdk.AccAddress("second_is_good______"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := keeper.SafeBech32ToAccAddresses(tc.bech32s)
			assert.Equal(t, tc.exp, actual, "safeBech32ToAccAddresses")
		})
	}
}
