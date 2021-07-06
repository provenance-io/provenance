package v042

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/provenance-io/provenance/x/attribute/types"
)

var (
	AttributeKeyPrefixLegacy  = []byte{0x01} // prefix for keys with address length 20 < v0.43
	AttributeAddrLengthLegacy = 20
	AttributeKeyLengthLegacy  = 1 + AttributeAddrLengthLegacy + 32 + 32 // prefix length + address (20) + name-hash + value-hash
)

// AccountAttributeKeyLegacy creates a key for an account attribute
func AccountAttributeKeyLegacy(acc sdk.AccAddress, attr types.Attribute) []byte {
	if len(acc.Bytes()) != AttributeAddrLengthLegacy {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(acc.Bytes()), AttributeKeyPrefixLegacy))
	}
	key := append(AttributeKeyPrefixLegacy, address.MustLengthPrefix(acc.Bytes())...)
	key = append(key, types.GetNameKeyBytes(attr.Name)...)
	return append(key, attr.Hash()...)
}
