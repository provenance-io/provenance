package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var addr = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

func TestAttributeKey(t *testing.T) {
	// key proposal
	key := AccountAttributeKey(addr, Attribute{Name: "test.attribute", AttributeType: AttributeType_String, Value: []byte("test")})
	keyAddr, _, _ := SplitAccountAttributeKey(key)
	require.Equal(t, keyAddr.String(), addr.String())

	// invalid key
	require.Panics(t, func() { SplitAccountAttributeKey([]byte("test")) })
}

func TestAttributeNameProcessing(t *testing.T) {
	require.Equal(t, "", reverse(""), "an empty name reversed is empty")
	require.Equal(t, "", reverse(" "), "an empty name with whitespace reversed is empty")
	require.Equal(t, "root", reverse("root"), "a root name reversed is a root name")
	require.Equal(t, "root.domain.sub", reverse("sub.domain.root"), "a domain name can be reversed correctly")
}
