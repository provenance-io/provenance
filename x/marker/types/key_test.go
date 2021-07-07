package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, addr, SplitMarkerStoreKey(MarkerStoreKey(addr)), "marker address should always be length of 20")
	assert.Equal(t, largerLengthAddr, SplitMarkerStoreKey(MarkerStoreKey(largerLengthAddr)), "marker address should always be length of 20")
}
