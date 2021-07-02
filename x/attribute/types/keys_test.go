package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestAttributeKey(t *testing.T) {
	privKey, _ := secp256r1.GenPrivKey()
	var addr = sdk.AccAddress(privKey.PubKey().Address())
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
