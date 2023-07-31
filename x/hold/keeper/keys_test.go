package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/provenance-io/provenance/x/hold/keeper"
)

// assertKeyPrefixEscrowCoinValue asserts that the KeyPrefixEscrowCoin value
// is still exactly as it should be.
// Returns true if everything's okay, false if something has gone horribly wrong.
func assertKeyPrefixEscrowCoinValue(t *testing.T) bool {
	t.Helper()
	rv := true
	rv = assert.Equal(t, []byte{0x00}, keeper.KeyPrefixEscrowCoin, "KeyPrefixEscrowCoin value") && rv
	rv = assert.Len(t, keeper.KeyPrefixEscrowCoin, 1, "KeyPrefixEscrowCoin length") && rv
	rv = assert.Equal(t, 1, cap(keeper.KeyPrefixEscrowCoin), "KeyPrefixEscrowCoin capacity") && rv
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

func TestCreateEscrowCoinKeyAddrPrefix(t *testing.T) {
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
			exp:  keeper.KeyPrefixEscrowCoin,
		},
		{
			name: "empty address",
			addr: sdk.AccAddress{},
			exp:  keeper.KeyPrefixEscrowCoin,
		},
		{
			name: "20 byte address",
			addr: addr20,
			exp:  append(keeper.KeyPrefixEscrowCoin, addr20WLen...),
		},
		{
			name: "32 byte address",
			addr: addr32,
			exp:  append(keeper.KeyPrefixEscrowCoin, addr32WLen...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateEscrowCoinKeyAddrPrefix(tc.addr)
			}
			require.NotPanics(t, testFunc, "CreateEscrowCoinKeyAddrPrefix")
			if assert.Equal(t, tc.exp, actual, "result") {
				assert.Equal(t, len(actual), cap(actual), "result length (expected) vs capacity (actual)")
			}
			// Change the first byte and make sure KeyPrefixEscrowCoin is still the same.
			actual[0] = 0xDD
			assertKeyPrefixEscrowCoinValue(t)
		})
	}
}

func TestCreateEscrowCoinKey(t *testing.T) {
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
			exp:   concatBzs(keeper.KeyPrefixEscrowCoin, addr20WLen, []byte("foocoin")),
		},
		{
			name:  "32 byte address",
			addr:  addr32,
			denom: "barcoin",
			exp:   concatBzs(keeper.KeyPrefixEscrowCoin, addr32WLen, []byte("barcoin")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateEscrowCoinKey(tc.addr, tc.denom)
			}
			require.NotPanics(t, testFunc, "CreateEscrowCoinKey")
			if assert.Equal(t, tc.exp, actual, "result") {
				assert.Equal(t, len(actual), cap(actual), "result length (expected) vs capacity (actual)")
			}
			// Change the first byte and make sure KeyPrefixEscrowCoin is still the same.
			actual[0] = 0xAF
			assertKeyPrefixEscrowCoinValue(t)
		})
	}
}

func TestParseEscrowCoinKey(t *testing.T) {
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
			key:      concatBzs(keeper.KeyPrefixEscrowCoin, addr20WLen, []byte("bananacoin")),
			expAddr:  addr20,
			expDenom: "bananacoin",
		},
		{
			name:     "32 byte address",
			key:      concatBzs(keeper.KeyPrefixEscrowCoin, addr32WLen, []byte("grapegrape")),
			expAddr:  addr32,
			expDenom: "grapegrape",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var denom string
			testFunc := func() {
				addr, denom = keeper.ParseEscrowCoinKey(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseEscrowCoinKey")
			assert.Equal(t, tc.expAddr, addr, "address")
			assert.Equal(t, tc.expDenom, denom, "denom")
		})
	}
}

func TestParseEscrowCoinKeyUnprefixed(t *testing.T) {
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
				addr, denom = keeper.ParseEscrowCoinKeyUnprefixed(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseEscrowCoinKeyUnprefixed")
			assert.Equal(t, tc.expAddr, addr, "address")
			assert.Equal(t, tc.expDenom, denom, "denom")
		})
	}
}

func TestUnmarshalEscrowCoinValue(t *testing.T) {
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
				amount, err = keeper.UnmarshalEscrowCoinValue(tc.value)
			}
			require.NotPanics(t, testFunc, "UnmarshalEscrowCoinValue")
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "UnmarshalEscrowCoinValue")
			} else {
				assert.NoError(t, err, "UnmarshalEscrowCoinValue")
			}
			assert.Equal(t, tc.expAmt.String(), amount.String(), "result amount")
		})
	}
}
