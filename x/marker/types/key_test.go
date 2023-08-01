package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkerAddressLength(t *testing.T) {
	addr, err := MarkerAddress("nhash")
	assert.NoError(t, err)
	assert.Equal(t, 20, len(addr), "marker address should always be length of 20")
}

func TestSplitMarkerStoreKey(t *testing.T) {
	addr, err := MarkerAddress("nhash")
	largerLengthAddr := sdk.AccAddress("FFFFFFFFFFFFFFFFFFFFFFFF")
	assert.NoError(t, err)
	assert.Equal(t, addr, SplitMarkerStoreKey(MarkerStoreKey(addr)), "should parse a marker of length 20 from key")
	assert.Equal(t, largerLengthAddr, SplitMarkerStoreKey(MarkerStoreKey(largerLengthAddr)), "should parse a marker of length 24 from key")
}

func TestDenySendKey(t *testing.T) {
	addr, err := MarkerAddress("nhash")
	require.NoError(t, err)
	denyAddr := sdk.AccAddress("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	denyKey := DenySendKey(addr, denyAddr)
	assert.Equal(t, uint8(3), denyKey[0], "should have correct prefix for deny send key")
	denomArrLen := int32(denyKey[1])
	assert.Equal(t, addr.Bytes(), denyKey[2:denomArrLen+2], "should match denom key")
	denyAddrLen := int32(denyKey[denomArrLen+2])
	assert.Equal(t, denyAddr.Bytes(), denyKey[denomArrLen+3:denomArrLen+3+denyAddrLen], "should match deny key")
	assert.Len(t, denyKey, int(3+denomArrLen+denyAddrLen), "should have key of length of sum 1 for prefix 2 length bytes and length of denom and deny address")
}

func TestMarkerNetAssetValueKey(t *testing.T) {
	addr, err := MarkerAddress("nhash")
	require.NoError(t, err)
	navKey := MarkerNetAssetValueKey(addr, "nhash")
	assert.Equal(t, uint8(4), navKey[0], "should have correct prefix for nav key")
	denomArrLen := int32(navKey[1])
	assert.Equal(t, addr.Bytes(), navKey[2:denomArrLen+2], "should match denom key")
	assert.Equal(t, "nhash", string(navKey[denomArrLen+2:]))
}
