package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

const (
	// PrefixScope is the address human readable prefix used with bech32 encoding of Scope IDs
	PrefixScope = "scope"
	// PrefixSession is the address human readable prefix used with bech32 encoding of Session IDs
	PrefixSession = "session"
	// PrefixRecord is the address human readable prefix used with bech32 encoding of Record IDs
	PrefixRecord = "record"

	// PrefixScopeSpecification is the address human readable prefix used with bech32 encoding of ScopeSpecification IDs
	PrefixScopeSpecification = "scopespec"
	// PrefixContractSpecification is the address human readable prefix used with bech32 encoding of ContractSpecification IDs
	PrefixContractSpecification = "contractspec"
	// PrefixRecordSpecification is the address human readable prefix used with bech32 encoding of RecordSpecification IDs
	PrefixRecordSpecification = "recspec"
)

var (
	// Ensure MetadataAddress implements the sdk.Address interface
	_ sdk.Address = MetadataAddress{}
)

// MetadataAddress is a blockchain compliant address based on UUIDs
type MetadataAddress []byte

// VerifyMetadataAddressFormat checks a sequence of bytes for proper format as a MetadataAddress instance
// returns the associated bech32 hrp/type name or any errors encountered during verification
func VerifyMetadataAddressFormat(bz []byte) (string, error) {
	hrp := ""
	requiredLength := 1 + 16 // type byte plus size of one uuid
	if len(bz) < requiredLength {
		return hrp, fmt.Errorf("incorrect address length (must be at least 17, actual: %d)", len(bz))
	}
	checkSecondaryUUID := false
	switch bz[0] {
	case ScopeKeyPrefix[0]:
		hrp = PrefixScope
		requiredLength = 1 + 16 // type byte plus size of one uuid
	case SessionKeyPrefix[0]:
		hrp = PrefixSession
		requiredLength = 1 + 16 + 16 // type byte plus size of two uuids
		checkSecondaryUUID = true
	case RecordKeyPrefix[0]:
		hrp = PrefixRecord
		requiredLength = 1 + 16 + 16 // type byte plus size of one uuid and one half sha256 hash

	case ScopeSpecificationKeyPrefix[0]:
		hrp = PrefixScopeSpecification
		requiredLength = 1 + 16 // type byte plus size of one uuid
	case ContractSpecificationKeyPrefix[0]:
		hrp = PrefixContractSpecification
		requiredLength = 1 + 16 // type byte plus size of one uuid
	case RecordSpecificationKeyPrefix[0]:
		hrp = PrefixRecordSpecification
		requiredLength = 1 + 16 + 16 // type byte plus size of one uuid plus one-half sha256 hash

	default:
		return hrp, fmt.Errorf("invalid metadata address type: %d", bz[0])
	}
	if len(bz) != requiredLength {
		return hrp, fmt.Errorf("incorrect address length (expected: %d, actual: %d)", requiredLength, len(bz))
	}
	// all valid metdata address have at least one uuid
	if _, err := uuid.FromBytes(bz[1:17]); err != nil {
		return hrp, fmt.Errorf("invalid address bytes of uuid, expected uuid compliant: %w", err)
	}
	if checkSecondaryUUID {
		if _, err := uuid.FromBytes(bz[17:33]); err != nil {
			return hrp, fmt.Errorf("invalid address bytes of secondary uuid, expected uuid compliant: %w", err)
		}
	}
	return hrp, nil
}

// ConvertHashToAddress constructs a MetadataAddress using the provided type code and the first 16 bytes of the
// base64 decoded hash.  Resulting Address is not guaranteed to contain a valid V4 UUID (random only)
func ConvertHashToAddress(typeCode []byte, hash string) (addr MetadataAddress, err error) {
	var raw []byte
	raw, err = base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return addr, err
	}
	if len(raw) < 16 {
		return addr, fmt.Errorf("invalid specification identifier, expected at least 16 bytes, found %d", len(raw))
	}
	// The codes 0,3,4 all start with a code followed by uuid bytes meaning we can create a valid address with them
	if len(typeCode) > 1 || !(typeCode[0] == 0x00 || typeCode[0] == 0x03 || typeCode[0] == 0x04) {
		return addr, fmt.Errorf("invalid address type code 0x%X, expected 0x00, 0x03, or 0x04", typeCode)
	}
	err = addr.Unmarshal(append(typeCode, raw[0:16]...))
	return
}

// MetadataAddressFromHex creates a MetadataAddress from a hex string.  NOTE: Does not perform validation on address,
// only performs basic HEX decoding checks.  This method matches the sdk.AccAddress approach
func MetadataAddressFromHex(address string) (MetadataAddress, error) {
	if len(address) == 0 {
		return MetadataAddress{}, errors.New("address decode failed: must provide an address")
	}
	bz, err := hex.DecodeString(address)
	return MetadataAddress(bz), err
}

// MetadataAddressFromBech32 creates a MetadataAddress from a Bech32 string.  The encoded data is checked against the
// provided bech32 hrp along with an overall verification of the byte format.
func MetadataAddressFromBech32(address string) (addr MetadataAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return MetadataAddress{}, errors.New("empty address string is not allowed")
	}

	hrp, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return nil, err
	}
	expectedHrp, err := VerifyMetadataAddressFormat(bz)
	if err != nil {
		return nil, err
	}
	if expectedHrp != hrp {
		return MetadataAddress{}, fmt.Errorf("invalid bech32 prefix; expected %s, got %s", expectedHrp, hrp)
	}

	return MetadataAddress(bz), nil
}

// ScopeMetadataAddress creates a MetadataAddress instance for the given scope by its uuid
func ScopeMetadataAddress(scopeUUID uuid.UUID) MetadataAddress {
	bz, err := scopeUUID.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return append(ScopeKeyPrefix, bz...)
}

// SessionMetadataAddress creates a MetadataAddress instance for a session within a scope by uuids
func SessionMetadataAddress(scopeUUID uuid.UUID, sessionUUID uuid.UUID) MetadataAddress {
	bz, err := scopeUUID.MarshalBinary()
	if err != nil {
		panic(err)
	}
	addr := append(SessionKeyPrefix, bz...)
	bz, err = sessionUUID.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return append(addr, bz...)
}

// RecordMetadataAddress creates a MetadataAddress instance for a record within a scope by scope uuid/record name
func RecordMetadataAddress(scopeUUID uuid.UUID, name string) MetadataAddress {
	bz, err := scopeUUID.MarshalBinary()
	if err != nil {
		panic(err)
	}
	addr := append(RecordKeyPrefix, bz...)
	name = strings.ToLower(strings.TrimSpace(name))
	if len(name) < 1 {
		panic("missing name value for record metadata address")
	}
	nameBytes := sha256.Sum256([]byte(name))
	return append(addr, nameBytes[0:16]...)
}

// ScopeSpecMetadataAddress creates a MetadataAddress instance for a scope specification
func ScopeSpecMetadataAddress(specUUID uuid.UUID) MetadataAddress {
	bz, err := specUUID.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return append(ScopeSpecificationKeyPrefix, bz...)
}

// ContractSpecMetadataAddress creates a MetadataAddress instance for a contract specification
func ContractSpecMetadataAddress(specUUID uuid.UUID) MetadataAddress {
	bz, err := specUUID.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return append(ContractSpecificationKeyPrefix, bz...)
}

// RecordSpecMetadataAddress creates a MetadataAddress instance for a record specification
func RecordSpecMetadataAddress(contractSpecUUID uuid.UUID, name string) MetadataAddress {
	bz, err := contractSpecUUID.MarshalBinary()
	if err != nil {
		panic(err)
	}
	addr := append(RecordSpecificationKeyPrefix, bz...)
	name = strings.ToLower(strings.TrimSpace(name))
	if len(name) < 1 {
		panic("missing name value for record spec metadata address")
	}
	nameBytes := sha256.Sum256([]byte(name))
	return append(addr, nameBytes[0:16]...)
}

// Equals determines if the current MetadataAddress is equal to another sdk.Address
func (ma MetadataAddress) Equals(ma2 sdk.Address) bool {
	if ma.Empty() && ma2.Empty() {
		return true
	}
	return bytes.Equal(ma.Bytes(), ma2.Bytes())
}

// Empty returns true if the MetadataAddress is uninitialized
func (ma MetadataAddress) Empty() bool {
	if ma == nil {
		return true
	}

	ma2 := MetadataAddress{}
	return bytes.Equal(ma.Bytes(), ma2.Bytes())
}

// Validate determines if the contained bytes form a valid MetadataAddress according to its type
func (ma MetadataAddress) Validate() (err error) {
	_, err = VerifyMetadataAddressFormat(ma)
	return
}

// Marshal returns the bytes underlying the MetadataAddress instance
func (ma MetadataAddress) Marshal() ([]byte, error) {
	return ma, nil
}

// Unmarshal initializes a MetadataAddress instance using the given bytes.  An error will be returned if the
// given bytes do not form a valid Address
func (ma *MetadataAddress) Unmarshal(data []byte) error {
	*ma = data
	if len(data) == 0 {
		return nil
	}
	_, err := VerifyMetadataAddressFormat(data)
	return err
}

// MarshalJSON returns a JSON representation for the current address using a bech32 encoded string
func (ma MetadataAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(ma.String())
}

// MarshalYAML returns a YAML representation for the current address using a bech32 encoded string
func (ma MetadataAddress) MarshalYAML() (interface{}, error) {
	return ma.String(), nil
}

// UnmarshalJSON creates a MetadataAddress instance from the given JSON data
func (ma *MetadataAddress) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*ma = MetadataAddress{}
		return nil
	}

	ma2, err := MetadataAddressFromBech32(s)
	if err != nil {
		return err
	}

	*ma = ma2
	return nil
}

// UnmarshalYAML creates a MetadataAddress instance from the given YAML data
func (ma *MetadataAddress) UnmarshalYAML(data []byte) error {
	var s string
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*ma = MetadataAddress{}
		return nil
	}

	ma2, err := MetadataAddressFromBech32(s)
	if err != nil {
		return err
	}

	*ma = ma2
	return nil
}

// Bytes implements Address interface, returns the raw bytes for this Address
func (ma MetadataAddress) Bytes() []byte {
	return ma
}

// String implements the stringer interface and encodes as a bech32
func (ma MetadataAddress) String() string {
	if ma.Empty() {
		return ""
	}

	hrp, err := VerifyMetadataAddressFormat(ma)
	if err != nil {
		panic(err)
	}

	bech32Addr, err := bech32.ConvertAndEncode(hrp, ma.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Size implements gogoproto custom type interface and returns the number of bytes in this instance
func (ma MetadataAddress) Size() int {
	return len(ma)
}

// MarshalTo implements gogoproto custom type interface and writes the current bytes into the provided data structure
func (ma *MetadataAddress) MarshalTo(data []byte) (int, error) {
	if len(*ma) == 0 {
		return 0, nil
	}
	copy(data, *ma)
	return len(*ma), nil
}

// Compare exists to fit gogoprotobuf custom type interface.
func (ma MetadataAddress) Compare(other MetadataAddress) int {
	return bytes.Compare(ma[0:], other[0:])
}

// ScopeUUID returns the scope uuid component of a MetadataAddress (if appropriate)
func (ma MetadataAddress) ScopeUUID() (uuid.UUID, error) {
	if !ma.isTypeOneOf(ScopeKeyPrefix, SessionKeyPrefix, RecordKeyPrefix) {
		return uuid.UUID{}, fmt.Errorf("this metadata address (%s) does not contain a scope uuid", ma)
	}
	return ma.PrimaryUUID()
}

// SessionUUID returns the session uuid component of a MetadataAddress (if appropriate)
func (ma MetadataAddress) SessionUUID() (uuid.UUID, error) {
	if len(ma) > 0 && ma[0] != SessionKeyPrefix[0] {
		return uuid.UUID{}, fmt.Errorf("this metadata address (%s) does not contain a session uuid", ma)
	}
	return ma.SecondaryUUID()
}

// ScopeSpecUUID returns the scope specification uuid component of a MetadataAddress (if appropriate)
func (ma MetadataAddress) ScopeSpecUUID() (uuid.UUID, error) {
	if len(ma) > 0 && ma[0] != ScopeSpecificationKeyPrefix[0] {
		return uuid.UUID{}, fmt.Errorf("this metadata address (%s) does not contain a scope specification uuid", ma)
	}
	return ma.PrimaryUUID()
}

// ContractSpecUUID returns the contract specification uuid component of a MetadataAddress (if appropriate)
func (ma MetadataAddress) ContractSpecUUID() (uuid.UUID, error) {
	if !ma.isTypeOneOf(ContractSpecificationKeyPrefix, RecordSpecificationKeyPrefix) {
		return uuid.UUID{}, fmt.Errorf("this metadata address (%s) does not contain a contract specification uuid", ma)
	}
	return ma.PrimaryUUID()
}

// Prefix returns the human readable part (prefix) of this MetadataAddress, e.g. "scope" or "contractspec"
// More accurately, this converts the 1st byte into its human readable string value.
func (ma MetadataAddress) Prefix() (string, error) {
	return VerifyMetadataAddressFormat(ma)
}

// PrimaryUUID returns the primary UUID from this MetadataAddress (if applicable).
// More accurately, this converts bytes 2 to 17 to a UUID.
// For example, if this MetadataAddress is for a scope specification, this will return the scope specification uuid.
// But if this MetadataAddress is for a record specification, this will return the contract specification
// (since that's the first part of those metadata addresses).
func (ma MetadataAddress) PrimaryUUID() (uuid.UUID, error) {
	if len(ma) < 1 {
		return uuid.UUID{}, fmt.Errorf("address empty")
	}
	// if we don't know this type
	if !ma.isTypeOneOf(ScopeKeyPrefix, SessionKeyPrefix, RecordKeyPrefix, ScopeSpecificationKeyPrefix, ContractSpecificationKeyPrefix, RecordSpecificationKeyPrefix) {
		return uuid.UUID{}, fmt.Errorf("invalid address type out of valid range (got: %d)", ma[0])
	}
	if len(ma) < 17 {
		return uuid.UUID{}, fmt.Errorf("incorrect address length (must be at least 17, actual: %d)", len(ma))
	}
	return uuid.FromBytes(ma[1:17])
}

// SecondaryUUID returns the secondary UUID from this MetadataAddress (if applicable).
// More accurately, this converts bytes 18 to 33 (inclusive) to a UUID.
func (ma MetadataAddress) SecondaryUUID() (uuid.UUID, error) {
	if len(ma) < 1 {
		return uuid.UUID{}, fmt.Errorf("address empty")
	}
	// if we don't know this type
	if !ma.isTypeOneOf(SessionKeyPrefix) {
		return uuid.UUID{}, fmt.Errorf("invalid address type out of valid range (got: %d)", ma[0])
	}
	if len(ma) < 33 {
		return uuid.UUID{}, fmt.Errorf("incorrect address length (must be at least 33, actual: %d)", len(ma))
	}
	return uuid.FromBytes(ma[17:33])
}

// NameHash returns the hashed name bytes from this MetadataAddress (if applicable).
// More accurately, this returns a copy of bytes 18 through 33 (inclusive).
func (ma MetadataAddress) NameHash() ([]byte, error) {
	namehash := make([]byte, 16)
	if len(ma) < 1 {
		return namehash, fmt.Errorf("address empty")
	}
	if !ma.isTypeOneOf(RecordKeyPrefix, RecordSpecificationKeyPrefix) {
		return namehash, fmt.Errorf("invalid address type out of valid range (got: %d)", ma[0])
	}
	if len(ma) < 33 {
		return namehash, fmt.Errorf("incorrect address length (must be at least 33, actual: %d)", len(ma))
	}
	copy(namehash, ma[17:])
	return namehash, nil
}

// AsScopeAddress returns the MetadataAddress for a scope using the scope UUID within the current context
func (ma MetadataAddress) AsScopeAddress() (MetadataAddress, error) {
	scopeUUID, err := ma.ScopeUUID()
	if err != nil {
		return MetadataAddress{}, err
	}
	return ScopeMetadataAddress(scopeUUID), nil
}

// AsRecordAddress returns the MetadataAddress for a record using the scope UUID within the current context and the provided name
func (ma MetadataAddress) AsRecordAddress(name string) (MetadataAddress, error) {
	scopeUUID, err := ma.ScopeUUID()
	if err != nil {
		return MetadataAddress{}, err
	}
	return RecordMetadataAddress(scopeUUID, name), nil
}

// AsRecordSpecAddress returns the MetadataAddress for a record spec using the contract spec UUID within the current context and the provided name
func (ma MetadataAddress) AsRecordSpecAddress(name string) (MetadataAddress, error) {
	contractSpecUUID, err := ma.ContractSpecUUID()
	if err != nil {
		return MetadataAddress{}, err
	}
	return RecordSpecMetadataAddress(contractSpecUUID, name), nil
}

// AsContractSpecAddress returns the MetadataAddress for a contract spec using the contract spec UUID within the current context
func (ma MetadataAddress) AsContractSpecAddress() (MetadataAddress, error) {
	contractSpecUUID, err := ma.ContractSpecUUID()
	if err != nil {
		return MetadataAddress{}, err
	}
	return ContractSpecMetadataAddress(contractSpecUUID), nil
}

// ScopeSessionIteratorPrefix returns an iterator prefix that finds all Sessions assigned to the scope designated in this MetadataAddress.
// If the current address is empty this returns a prefix to iterate through all sessions.
// If the current address is a scope, this returns a prefix to iterate through all sessions in this scope.
// If the current address is a session or record, this returns a prefix to iterate through all sessions in the scope that contains
// this session or record.
func (ma MetadataAddress) ScopeSessionIteratorPrefix() ([]byte, error) {
	if len(ma) < 1 {
		return SessionKeyPrefix, nil
	}
	// if we don't know this type
	if !ma.isTypeOneOf(ScopeKeyPrefix, SessionKeyPrefix, RecordKeyPrefix) {
		return []byte{}, fmt.Errorf("this metadata address does not contain a scope uuid")
	}
	return append(SessionKeyPrefix, ma[1:17]...), nil
}

// ScopeRecordIteratorPrefix returns an iterator prefix that finds all Records assigned to the scope designated in this MetadataAddress.
// If the current address is empty this returns a prefix to iterate through all records.
// If the current address is a scope, this returns a prefix to iterate through all records in this scope.
// If the current address is a session or record, this returns a prefix to iterate through all records in the scope that contains
// this session or record.
func (ma MetadataAddress) ScopeRecordIteratorPrefix() ([]byte, error) {
	if len(ma) < 1 {
		return RecordKeyPrefix, nil
	}
	// if we don't know this type
	if !ma.isTypeOneOf(ScopeKeyPrefix, SessionKeyPrefix, RecordKeyPrefix) {
		return []byte{}, fmt.Errorf("this metadata address does not contain a scope uuid")
	}
	return append(RecordKeyPrefix, ma[1:17]...), nil
}

// ContractSpecRecordSpecIteratorPrefix returns an iterator prefix that finds all record specifications
// for the contract specification designated in this MetadataAddress.
// If the current address is empty this returns a prefix to iterate through all record specifications
// If the current address is a contract specification, this returns a prefix to iterate through all record specifications
// associated with that contract specification.
// If the current address is a record specification, this returns a prefix to iterate through all record specifications
// associated with the contract specification that contains this record specification.
// If the current address is some other type, an error is returned.
func (ma MetadataAddress) ContractSpecRecordSpecIteratorPrefix() ([]byte, error) {
	if len(ma) < 1 {
		return RecordSpecificationKeyPrefix, nil
	}
	// if we don't know this type
	if !ma.isTypeOneOf(ContractSpecificationKeyPrefix, RecordSpecificationKeyPrefix) {
		return []byte{}, fmt.Errorf("this metadata address does not contain a contract spec uuid")
	}
	return append(RecordSpecificationKeyPrefix, ma[1:17]...), nil
}

// Format implements fmt.Format interface
func (ma MetadataAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(ma.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", ma)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(ma))))
	}
}

// IsScopeAddress returns true if this address is valid and has a scope type byte.
func (ma MetadataAddress) IsScopeAddress() bool {
	hrp, err := VerifyMetadataAddressFormat(ma)
	return (err == nil && hrp == PrefixScope)
}

// IsSessionAddress returns true if this address is valid and has a session type byte.
func (ma MetadataAddress) IsSessionAddress() bool {
	hrp, err := VerifyMetadataAddressFormat(ma)
	return (err == nil && hrp == PrefixSession)
}

// IsRecordAddress returns true if the address is valid and has a record type byte.
func (ma MetadataAddress) IsRecordAddress() bool {
	hrp, err := VerifyMetadataAddressFormat(ma)
	return (err == nil && hrp == PrefixRecord)
}

// IsScopeSpecificationAddress returns true if this address is valid and has a scope specification type byte.
func (ma MetadataAddress) IsScopeSpecificationAddress() bool {
	hrp, err := VerifyMetadataAddressFormat(ma)
	return (err == nil && hrp == PrefixScopeSpecification)
}

// IsContractSpecificationAddress returns true if this address is valid and has a contract specification type byte.
func (ma MetadataAddress) IsContractSpecificationAddress() bool {
	hrp, err := VerifyMetadataAddressFormat(ma)
	return (err == nil && hrp == PrefixContractSpecification)
}

// IsRecordSpecificationAddress returns true if this address is valid and has a record specification type byte.
func (ma MetadataAddress) IsRecordSpecificationAddress() bool {
	hrp, err := VerifyMetadataAddressFormat(ma)
	return (err == nil && hrp == PrefixRecordSpecification)
}

// isTypeOneOf returns true if the first byte is equal to the first byte in any provided options.
func (ma MetadataAddress) isTypeOneOf(options ...[]byte) bool {
	if len(ma) == 0 {
		return false
	}
	for _, o := range options {
		if len(o) > 0 && ma[0] == o[0] {
			return true
		}
	}
	return false
}
