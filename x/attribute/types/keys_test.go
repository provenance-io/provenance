package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestModuleAddresses(t *testing.T) {
	// This test is primarily just to know what the bech32 strings are for the attribute
	// module account so that they can be used in scripts/initialize.sh and unit tests.

	// Get the attribute module address bytes so we can make bech32 strings of it with various HRPs.
	moduleAddr := authtypes.NewModuleAddress(ModuleName)

	// Not using sdk.MustAccAddressFromBech32 here because we want different HRPs.
	// Not using the app/prefix.go variables directly because it causes a circular import.
	// AccountAddressPrefixMainNet = "pb"
	// AccountAddressPrefixTestNet = "tp"
	// The expected values came from running the test once and copying the actual values from the results.

	tests := []struct {
		name string
		hrp  string
		exp  string
	}{
		{name: "mainnet", hrp: "pb", exp: "pb14y4l6qky2zhsxcx7540ejqvtye7fr66dczq2f9"},
		{name: "testnet", hrp: "tp", exp: "tp14y4l6qky2zhsxcx7540ejqvtye7fr66dtf9xt0"},
		{name: "cosmos", hrp: "cosmos", exp: "cosmos14y4l6qky2zhsxcx7540ejqvtye7fr66d3vt0ye"},
	}

	for _, tc := range tests {
		// Skipping the t.Run construct since a) there's only 3, b) none of the other tests
		// in here do it, and c) having it all in a single failure might be useful in this case.
		actual, err := bech32.ConvertAndEncode(tc.hrp, moduleAddr)
		if assert.NoError(t, err, "%s ConvertAndEncode error", tc.name) {
			assert.Equal(t, tc.exp, actual, "%s bech32 string of attribute module account address", tc.name)
		}
	}
}

func TestAttributeNameProcessing(t *testing.T) {
	require.Equal(t, "", reverse(""), "an empty name reversed is empty")
	require.Equal(t, "", reverse(" "), "an empty name with whitespace reversed is empty")
	require.Equal(t, "root", reverse("root"), "a root name reversed is a root name")
	require.Equal(t, "root.domain.sub", reverse("sub.domain.root"), "a domain name can be reversed correctly")
}

func TestGetAddressFromKey(t *testing.T) {
	attr1 := Attribute{
		Name:          "long.address.name",
		Value:         []byte("0123456789"),
		Address:       "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqs2m6sx4",
		AttributeType: AttributeType_String,
	}
	attr2 := Attribute{
		Name:          "short.address.name",
		Value:         []byte("0123456789"),
		Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		AttributeType: AttributeType_String,
	}

	longKey, err := GetAddressFromKey(AttributeNameAddrKeyPrefix(attr1.Name, attr1.GetAddressBytes()))
	assert.NoError(t, err)
	assert.Equal(t, attr1.GetAddressBytes(), longKey.Bytes())

	shortKey, err := GetAddressFromKey(AttributeNameAddrKeyPrefix(attr2.Name, attr2.GetAddressBytes()))
	assert.NoError(t, err)
	assert.Equal(t, attr2.GetAddressBytes(), shortKey.Bytes())
}
