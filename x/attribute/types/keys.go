package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the module
	ModuleName = "attribute"

	// StoreKey is the store key string for account
	StoreKey = ModuleName

	// RouterKey is the message route for account
	RouterKey = ModuleName

	// AccountDataName is the name of the attribute used to store account data.
	AccountDataName = "accountdata"

	// NameTerminator ends the (variable length) reversed name in a store key so that whatever
	// follows it can't be confused for part of the name. Names can't contain this byte.
	NameTerminator = byte(0x00)
)

var (
	// Legacy amino encoded objects use this key prefix
	AttributeKeyPrefixAmino      = []byte{0x00}
	AttributeKeyPrefix           = []byte{0x02}
	AttributeAddrLookupKeyPrefix = []byte{0x03}
	AttributeExpirationKeyPrefix = []byte{0x04}
	AttributeParamPrefix         = []byte{0x05}
)

// AddrAttributeKey creates a key for an account attribute
// Format: [0x02][addr_len][addr][reversed name][0x00][hash(value)]
func AddrAttributeKey(addr []byte, attr Attribute) []byte {
	key := AddrAttributesNameKeyPrefix(addr, attr.Name)
	return append(key, attr.Hash()...)
}

// GetAddrAttributeKeyFromExpireKey returns the AddrAttribute key from attribute expiration key
func GetAddrAttributeKeyFromExpireKey(key []byte) []byte {
	return append(AttributeKeyPrefix, key[9:]...)
}

// AttributeExpireKey returns a key for expiration [AttributeExpirationKeyPrefix][epoch][AccAddress bytes][name hash][attribute hash]
func AttributeExpireKey(attr Attribute) []byte {
	if attr.ExpirationDate == nil {
		return nil
	}
	key := GetAttributeExpireTimePrefix(*attr.ExpirationDate)
	key = append(key, address.MustLengthPrefix(attr.GetAddressBytes())...)
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
	key := AddrAttributesKeyPrefix(addr)
	return append(key, GetNameKeyBytes(attributeName)...)
}

// AddrStrAttributesNameKeyPrefix is the same as AddrAttributesNameKeyPrefix but takes in the address as a string.
func AddrStrAttributesNameKeyPrefix(addr string, attributeName string) []byte {
	return AddrAttributesNameKeyPrefix(GetAttributeAddressBytes(addr), attributeName)
}

// AttributeNameKeyPrefix returns a prefix key for all addresses with attribute name
func AttributeNameKeyPrefix(attributeName string) []byte {
	key := AttributeAddrLookupKeyPrefix
	return append(key, GetNameKeyBytes(attributeName)...)
}

// AttributeNameAddrKeyPrefix returns a prefix key for attribute and address
func AttributeNameAddrKeyPrefix(attributeName string, addr []byte) []byte {
	key := AttributeNameKeyPrefix(attributeName)
	return append(key, address.MustLengthPrefix(addr)...)
}

// GetAddressFromKey returns the AccAddress from full attribute address key ([prefix][name hash][length + AccAddress bytes][attribute hash])
// ([0x03][reversed name][0x00][addr_len][addr]).
func GetAddressFromKey(nameAddrKey []byte) (sdk.AccAddress, error) {
	// Everything up to (and including) the terminator is the prefix byte and reversed name.
	end := bytes.IndexByte(nameAddrKey, NameTerminator)
	if end < 0 {
		return nil, fmt.Errorf("invalid attribute address lookup key: no name terminator found")
	}
	// Skip the terminator and the address length byte.
	if len(nameAddrKey) < end+2 {
		return nil, fmt.Errorf("invalid attribute address lookup key: no address found")
	}
	addressBytes := nameAddrKey[end+2:]
	if err := sdk.VerifyAddressFormat(addressBytes); err != nil {
		return nil, err
	}
	return addressBytes, nil
}

// GetNameKeyBytes returns a set of bytes that uniquely identifies the given name
func GetNameKeyBytes(name string) []byte {
	attrName := ReverseName(name)
	if len(attrName) == 0 {
		panic(fmt.Sprintf("invalid account attribute name %s", name))
	}
	return append([]byte(attrName), NameTerminator)
}

// ReverseName normalizes the provided attribute name and reverses the order of its segments.
// E.g. "id.sso.provenance.io" becomes "io.provenance.sso.id".
func ReverseName(name string) string {
	return reverse(strings.ToLower(strings.TrimSpace(name)))
}

// GetAttributeExpireTimePrefix returns a prefix for expired time [AttributeExpirationKeyPrefix][epoch]
func GetAttributeExpireTimePrefix(expireTime time.Time) []byte {
	key := AttributeExpirationKeyPrefix
	expireTimeBz := make([]byte, 8)
	binary.BigEndian.PutUint64(expireTimeBz, uint64(expireTime.Unix())) //nolint:gosec // Unix timestamp cast is safe and non-security-sensitive.
	return append(key, expireTimeBz...)
}

// Reverse the component order of a name for better scan support.
// For example, the name "id.sso.provenance.io" will become "io.provenance.sso.id",
// which allows us to easily scan all account attributes under "io.provenance.sso".
func reverse(name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
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
