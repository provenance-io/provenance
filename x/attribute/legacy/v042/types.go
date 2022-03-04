package v042

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

var (
	AttributeKeyPrefixLegacy  = []byte{0x01} // prefix for keys with address length 20 < v0.43
	AttributeAddrLengthLegacy = 20
)

// AccountAttributeKeyLegacy creates a legacy key for an account attribute
func AccountAttributeKeyLegacy(acc sdk.AccAddress, attr types.Attribute) []byte {
	if len(acc.Bytes()) != AttributeAddrLengthLegacy {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(acc.Bytes()), AttributeKeyPrefixLegacy))
	}
	key := AttributeKeyPrefixLegacy
	key = append(key, acc.Bytes()...)
	key = append(key, types.GetNameKeyBytes(attr.Name)...)
	return append(key, attr.Hash()...)
}
