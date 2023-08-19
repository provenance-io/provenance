package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/hold/keeper"
)

// assertKeyPrefixHoldCoinValue asserts that the KeyPrefixHoldCoin value
// is still exactly as it should be.
// Returns true if everything's okay, false if something has gone horribly wrong.
func assertKeyPrefixHoldCoinValue(t *testing.T) bool {
	t.Helper()
	rv := true
	rv = assert.Equal(t, []byte{0x00}, keeper.KeyPrefixHoldCoin, "KeyPrefixHoldCoin value") && rv
	rv = assert.Len(t, keeper.KeyPrefixHoldCoin, 1, "KeyPrefixHoldCoin length") && rv
	rv = assert.Equal(t, 1, cap(keeper.KeyPrefixHoldCoin), "KeyPrefixHoldCoin capacity") && rv
	return rv
}

// concatBzs creates a new slice by concatenating all the provided slices together.
func concatBzs(bzs ...[]byte) []byte {
	rv := make([]byte, 0)
	for _, bz := range bzs {
		rv = append(rv, bz...)
	}
	return rv
}

func TestCreateHoldCoinKeyAddrPrefix(t *testing.T) {
	addr20 := sdk.AccAddress("addr_with_20_bytes__")
	addr32 := sdk.AccAddress("longer__address__with__32__bytes")
	addr20WLen, err := address.LengthPrefix(addr20)
	require.NoError(t, err, "LengthPrefix(addr20)")
	addr32WLen, err := address.LengthPrefix(addr32)
	require.NoError(t, err, "LengthPrefix(addr32)")

	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  []byte
	}{
		{
			name: "nil address",
			addr: nil,
			exp:  keeper.KeyPrefixHoldCoin,
		},
		{
			name: "empty address",
			addr: sdk.AccAddress{},
			exp:  keeper.KeyPrefixHoldCoin,
		},
		{
			name: "20 byte address",
			addr: addr20,
			exp:  append(keeper.KeyPrefixHoldCoin, addr20WLen...),
		},
		{
			name: "32 byte address",
			addr: addr32,
			exp:  append(keeper.KeyPrefixHoldCoin, addr32WLen...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateHoldCoinKeyAddrPrefix(tc.addr)
			}
			require.NotPanics(t, testFunc, "CreateHoldCoinKeyAddrPrefix")
			if assert.Equal(t, tc.exp, actual, "result") {
				assert.Equal(t, len(actual), cap(actual), "result length (expected) vs capacity (actual)")
			}
			// Change the first byte and make sure KeyPrefixHoldCoin is still the same.
			actual[0] = 0xDD
			assertKeyPrefixHoldCoinValue(t)
		})
	}
}

func TestCreateHoldCoinKey(t *testing.T) {
	addr20 := sdk.AccAddress("addr_with_20_bytes__")
	addr32 := sdk.AccAddress("longer__address__with__32__bytes")
	addr20WLen, err := address.LengthPrefix(addr20)
	require.NoError(t, err, "LengthPrefix(addr20)")
	addr32WLen, err := address.LengthPrefix(addr32)
	require.NoError(t, err, "LengthPrefix(addr32)")

	tests := []struct {
		name  string
		addr  sdk.AccAddress
		denom string
		exp   []byte
	}{
		{
			name:  "20 byte address",
			addr:  addr20,
			denom: "foocoin",
			exp:   concatBzs(keeper.KeyPrefixHoldCoin, addr20WLen, []byte("foocoin")),
		},
		{
			name:  "32 byte address",
			addr:  addr32,
			denom: "barcoin",
			exp:   concatBzs(keeper.KeyPrefixHoldCoin, addr32WLen, []byte("barcoin")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateHoldCoinKey(tc.addr, tc.denom)
			}
			require.NotPanics(t, testFunc, "CreateHoldCoinKey")
			if assert.Equal(t, tc.exp, actual, "result") {
				assert.Equal(t, len(actual), cap(actual), "result length (expected) vs capacity (actual)")
			}
			// Change the first byte and make sure KeyPrefixHoldCoin is still the same.
			actual[0] = 0xAF
			assertKeyPrefixHoldCoinValue(t)
		})
	}
}

func TestParseHoldCoinKey(t *testing.T) {
	addr20 := sdk.AccAddress("addr_with_20_bytes__")
	addr32 := sdk.AccAddress("longer__address__with__32__bytes")
	addr20WLen, err := address.LengthPrefix(addr20)
	require.NoError(t, err, "LengthPrefix(addr20)")
	addr32WLen, err := address.LengthPrefix(addr32)
	require.NoError(t, err, "LengthPrefix(addr32)")

	tests := []struct {
		name     string
		key      []byte
		expAddr  sdk.AccAddress
		expDenom string
	}{
		{
			name:     "20 byte address",
			key:      concatBzs(keeper.KeyPrefixHoldCoin, addr20WLen, []byte("bananacoin")),
			expAddr:  addr20,
			expDenom: "bananacoin",
		},
		{
			name:     "32 byte address",
			key:      concatBzs(keeper.KeyPrefixHoldCoin, addr32WLen, []byte("grapegrape")),
			expAddr:  addr32,
			expDenom: "grapegrape",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var denom string
			testFunc := func() {
				addr, denom = keeper.ParseHoldCoinKey(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseHoldCoinKey")
			assert.Equal(t, tc.expAddr, addr, "address")
			assert.Equal(t, tc.expDenom, denom, "denom")
		})
	}
}

func TestParseHoldCoinKeyUnprefixed(t *testing.T) {
	addr20 := sdk.AccAddress("addr_with_20_bytes__")
	addr32 := sdk.AccAddress("longer__address__with__32__bytes")
	addr20WLen, err := address.LengthPrefix(addr20)
	require.NoError(t, err, "LengthPrefix(addr20)")
	addr32WLen, err := address.LengthPrefix(addr32)
	require.NoError(t, err, "LengthPrefix(addr32)")

	tests := []struct {
		name     string
		key      []byte
		expAddr  sdk.AccAddress
		expDenom string
	}{
		{
			name:     "20 byte address",
			key:      concatBzs(addr20WLen, []byte("bananacoin")),
			expAddr:  addr20,
			expDenom: "bananacoin",
		},
		{
			name:     "32 byte address",
			key:      concatBzs(addr32WLen, []byte("grapegrape")),
			expAddr:  addr32,
			expDenom: "grapegrape",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var denom string
			testFunc := func() {
				addr, denom = keeper.ParseHoldCoinKeyUnprefixed(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseHoldCoinKeyUnprefixed")
			assert.Equal(t, tc.expAddr, addr, "address")
			assert.Equal(t, tc.expDenom, denom, "denom")
		})
	}
}

func TestUnmarshalHoldCoinValue(t *testing.T) {
	newInt := func(amount string) sdkmath.Int {
		rv, ok := sdkmath.NewIntFromString(amount)
		require.True(t, ok, "NewIntFromString(%q)", amount)
		return rv
	}
	mustMarshall := func(amt sdkmath.Int) []byte {
		rv, err := amt.Marshal()
		require.NoError(t, err, "%q.Marshal()", amt.String())
		return rv
	}

	tests := []struct {
		name   string
		value  []byte
		expAmt sdkmath.Int
		expErr string
	}{
		{
			name:   "nil",
			value:  nil,
			expAmt: sdkmath.ZeroInt(),
		},
		{
			name:   "empty",
			value:  nil,
			expAmt: sdkmath.ZeroInt(),
		},
		{
			name:   "zero",
			value:  []byte("0"),
			expAmt: sdkmath.ZeroInt(),
		},
		{
			name:   "one",
			value:  []byte("1"),
			expAmt: sdkmath.OneInt(),
		},
		{
			name:   "int64 max",
			value:  []byte("9223372036854775807"),
			expAmt: newInt("9223372036854775807"),
		},
		{
			name:   "uint64 max + 1",
			value:  []byte("18446744073709551616"),
			expAmt: newInt("18446744073709551616"),
		},
		{
			name:   "uint64 max * 100",
			value:  []byte("1844674407370955161500"),
			expAmt: newInt("1844674407370955161500"),
		},
		{
			name:   "actually marshalled big number",
			value:  mustMarshall(newInt("9876543210123456789")),
			expAmt: newInt("9876543210123456789"),
		},
		{
			name:   "bad value",
			value:  []byte("bad"),
			expAmt: sdkmath.ZeroInt(),
			expErr: "math/big: cannot unmarshal \"bad\" into a *big.Int",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var amount sdkmath.Int
			var err error
			testFunc := func() {
				amount, err = keeper.UnmarshalHoldCoinValue(tc.value)
			}
			require.NotPanics(t, testFunc, "UnmarshalHoldCoinValue")
			testutil.AssertErrorValue(t, err, tc.expErr, "UnmarshalHoldCoinValue")
			assert.Equal(t, tc.expAmt.String(), amount.String(), "result amount")
		})
	}
}
