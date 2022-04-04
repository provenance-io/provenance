package types

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the module
	ModuleName = "attribute"

	// StoreKey is the store key string for account
	StoreKey = ModuleName

	// RouterKey is the message route for account
	RouterKey = ModuleName

	// QuerierRoute is the querier route for account
	QuerierRoute = ModuleName
)

var (
	// Legacy amino encoded objects use this key prefix
	AttributeKeyPrefixAmino = []byte{0x00}
	AttributeKeyPrefix      = []byte{0x02}
)

// AddrAttributeKey creates a key for an account attribute
func AddrAttributeKey(addr []byte, attr Attribute) []byte {
	key := AttributeKeyPrefix
	key = append(key, address.MustLengthPrefix(addr)...)
	key = append(key, GetNameKeyBytes(attr.Name)...)
	return append(key, attr.Hash()...)
}

// AddrAttributesKeyPrefix returns a prefix key for all attributes on an account
func AddrAttributesKeyPrefix(addr []byte) []byte {
	return append(AttributeKeyPrefix, address.MustLengthPrefix(addr)...)
}

// AddrStrAttributesKeyPrefix is the same as AddrAttributesKeyPrefix but takes in the address as a string.
func AddrStrAttributesKeyPrefix(addr string) []byte {
	return AddrAttributesKeyPrefix(GetAttributeAddressBytes(addr))
}

// AddrAttributesNameKeyPrefix returns a prefix key for all attributes with a given name on an account
func AddrAttributesNameKeyPrefix(addr []byte, attributeName string) []byte {
	key := AttributeKeyPrefix
	key = append(key, address.MustLengthPrefix(addr)...)
	return append(key, GetNameKeyBytes(attributeName)...)
}

// AddrStrAttributesNameKeyPrefix is the same as AddrAttributesNameKeyPrefix but takes in the address as a string.
func AddrStrAttributesNameKeyPrefix(addr string, attributeName string) []byte {
	return AddrAttributesNameKeyPrefix(GetAttributeAddressBytes(addr), attributeName)
}

// GetNameKeyBytes returns a set of bytes that uniquely identifies the given name
func GetNameKeyBytes(name string) []byte {
	attrName := strings.ToLower(strings.TrimSpace(name))
	attrName = reverse(attrName)
	if len(attrName) == 0 {
		panic(fmt.Sprintf("invalid account attribute name %s", name))
	}
	hash := sha256.Sum256([]byte(attrName))
	return hash[:]
}

// Reverse the component order of a name for better scan support.
// For example, the name "id.sso.provenance.io" will become "io.provenance.sso.id",
// which allows us to easily scan all account attributes under "io.provenance.sso".
func reverse(name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
	// check if there is nothing to reverse (root name)
	if !strings.Contains(name, ".") {
		return name
	}
	comps := strings.Split(name, ".")
	for i := len(comps)/2 - 1; i >= 0; i-- {
		j := len(comps) - 1 - i
		comps[i], comps[j] = comps[j], comps[i]
	}
	return strings.Join(comps, ".")
}
