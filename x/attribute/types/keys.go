package types

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	AttributeKeyPrefix = []byte{0x00}
	AttributeKeyLength = 1 + sdk.AddrLen + 32 + 32 // prefix + address + name-hash + value-hash
)

// AccountAttributeKey creates a key for an account attribute
func AccountAttributeKey(acc sdk.AccAddress, attr Attribute) []byte {
	key := append(AttributeKeyPrefix, acc.Bytes()...)
	key = append(key, GetNameKeyBytes(attr.Name)...)
	return append(key, attr.Hash()...)
}

// AccountAttributesKeyPrefix returns a prefix key for all attributes on an account
func AccountAttributesKeyPrefix(acc sdk.AccAddress) []byte {
	return append(AttributeKeyPrefix, acc.Bytes()...)
}

// AccountAttributesNameKeyPrefix returns a prefix key for all attributes with a given name on an account
func AccountAttributesNameKeyPrefix(acc sdk.AccAddress, attributeName string) []byte {
	key := append(AttributeKeyPrefix, acc.Bytes()...)
	return append(key, GetNameKeyBytes(attributeName)...)
}

func SplitAccountAttributeKey(key []byte) (addr sdk.AccAddress, nameID []byte, valueID []byte) {
	if len(key) != AttributeKeyLength {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(key), AttributeKeyLength))
	}
	// first byte is key prefix for AttributeKey
	addr = sdk.AccAddress(key[1 : sdk.AddrLen+1])
	nameID = key[1+sdk.AddrLen : sdk.AddrLen+32]
	valueID = key[1+sdk.AddrLen+32 : 1+sdk.AddrLen+64]
	return
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
	if strings.TrimSpace(name) == "" || !strings.Contains(name, ".") {
		return ""
	}
	comps := strings.Split(name, ".")
	for i := len(comps)/2 - 1; i >= 0; i-- {
		j := len(comps) - 1 - i
		comps[i], comps[j] = comps[j], comps[i]
	}
	return strings.Join(comps, ".")
}
