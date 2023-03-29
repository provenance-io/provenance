package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	longKey, err := GetAddressFromKey(AttributeNameAttrKeyPrefix(attr1))
	assert.NoError(t, err)
	assert.Equal(t, attr1.GetAddressBytes(), longKey.Bytes())

	shortKey, err := GetAddressFromKey(AttributeNameAttrKeyPrefix(attr2))
	assert.NoError(t, err)
	assert.Equal(t, attr2.GetAddressBytes(), shortKey.Bytes())

}
