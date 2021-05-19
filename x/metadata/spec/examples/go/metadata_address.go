package provenance

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/google/uuid"
)

const (
	PrefixScope                 = "scope"
	PrefixSession               = "session"
	PrefixRecord                = "record"
	PrefixScopeSpecification    = "scopespec"
	PrefixContractSpecification = "contractspec"
	PrefixRecordSpecification   = "recspec"

	KeyScope                 = byte(0x00)
	KeySession               = byte(0x01)
	KeyRecord                = byte(0x02)
	KeyScopeSpecification    = byte(0x04) // Note that this is not in numerical order.
	KeyContractSpecification = byte(0x03)
	KeyRecordSpecification   = byte(0x05)
)

// MetadataAddress is a type that helps create ids for the various types objects stored by the metadata module.
type MetadataAddress []byte

// MetadataAddressForScope creates a MetadataAddress instance for the given scope by its uuid
func MetadataAddressForScope(scopeUUID uuid.UUID) MetadataAddress {
	return buildBytes(KeyScope, uuidMustMarshalBinary(scopeUUID))
}

// MetadataAddressForSession creates a MetadataAddress instance for a session within a scope by uuids
func MetadataAddressForSession(scopeUUID uuid.UUID, sessionUUID uuid.UUID) MetadataAddress {
	return buildBytes(KeySession, uuidMustMarshalBinary(scopeUUID), uuidMustMarshalBinary(sessionUUID))
}

// MetadataAddressForRecord creates a MetadataAddress instance for a record within a scope by scope uuid/record name
func MetadataAddressForRecord(scopeUUID uuid.UUID, recordName string) MetadataAddress {
	if stringIsBlank(recordName) {
		panic("invalid recordName: cannot be empty or blank")
	}
	return buildBytes(KeyRecord, uuidMustMarshalBinary(scopeUUID), stringAsHashedBytes(recordName))
}

// MetadataAddressForScopeSpecification creates a MetadataAddress instance for a scope specification
func MetadataAddressForScopeSpecification(scopeSpecUUID uuid.UUID) MetadataAddress {
	return buildBytes(KeyScopeSpecification, uuidMustMarshalBinary(scopeSpecUUID))
}

// MetadataAddressForContractSpecification creates a MetadataAddress instance for a contract specification
func MetadataAddressForContractSpecification(contractSpecUUID uuid.UUID) MetadataAddress {
	return buildBytes(KeyContractSpecification, uuidMustMarshalBinary(contractSpecUUID))
}

// MetadataAddressForRecordSpecification creates a MetadataAddress instance for a record specification
func MetadataAddressForRecordSpecification(contractSpecUUID uuid.UUID, recordSpecName string) MetadataAddress {
	if stringIsBlank(recordSpecName) {
		panic("invalid recordSpecName: cannot be empty or blank")
	}
	return buildBytes(KeyRecordSpecification, uuidMustMarshalBinary(contractSpecUUID), stringAsHashedBytes(recordSpecName))
}

// MetadataAddressFromBech32 creates a MetadataAddress from a Bech32 string.  The encoded data is checked against the
// provided bech32 hrp along with an overall verification of the byte format.
func MetadataAddressFromBech32(address string) (MetadataAddress, error) {
	hrp, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return nil, err
	}
	err = validateBytes(bz)
	if err != nil {
		return nil, err
	}
	expectedHrp := getPrefixFromKey(bz[0])
	if hrp != expectedHrp {
		return nil, fmt.Errorf("incorrect hrp: expected %s, actual %s", expectedHrp, hrp)
	}
	return bz, nil
}

func MetadataAddressFromBytes(bz []byte) (MetadataAddress, error) {
	err := validateBytes(bz)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

// GetKey gets the key byte for this MetadataAddress.
func (m MetadataAddress) GetKey() byte {
	return m[0]
}

// GetPrefix gets the prefix string for this MetadataAddress, e.g. "scope".
func (m MetadataAddress) GetPrefix() string {
	return getPrefixFromKey(m[0])
}

// GetPrimaryUUID gets the set of bytes for the primary uuid part of this MetadataAddress as a UUID.
func (m MetadataAddress) GetPrimaryUUID() uuid.UUID {
	retval, err := uuid.FromBytes(m[1:17])
	if err != nil {
		panic(err)
	}
	return retval
}

// GetSecondaryBytes gets a copy of the bytes that make up the secondary part of this MetadataAddress.
func (m MetadataAddress) GetSecondaryBytes() []byte {
	if len(m) <= 17 {
		return []byte{}
	}
	retval := make([]byte, len(m)-17)
	copy(retval, m[17:])
	return retval
}

// Bytes gets all the bytes of this MetadataAddress.
func (m MetadataAddress) Bytes() []byte {
	return m
}

// String implements the stringer interface and encodes as a bech32.
func (m MetadataAddress) String() string {
	if len(m) == 0 {
		return ""
	}
	bech32Addr, err := bech32.ConvertAndEncode(getPrefixFromKey(m[0]), m)
	if err != nil {
		panic(err)
	}
	return bech32Addr
}

// Equals implementation for comparing MetadataAddress values.
func (m MetadataAddress) Equals(m2 MetadataAddress) bool {
	return (m == nil && m2 == nil) || bytes.Equal(m, m2)
}

// Format implements fmt.Format interface
// %s formats as bech32 address string (same as m.String()).
// %p formats as the address of 0th element in base 16 notation, with leading 0x.
// all others format as base 16, upper-case, two characters per byte.
func (m MetadataAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(m.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", m.Bytes())))
	default:
		s.Write([]byte(fmt.Sprintf("%X", m.Bytes())))
	}
}

// uuidMustMarshalBinary gets the bytes of a UUID or panics.
func uuidMustMarshalBinary(id uuid.UUID) []byte {
	bz, err := id.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return bz
}

// stringIsBlank returns true if the string is empty or all whitespace.
func stringIsBlank(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// stringAsHashedBytes hashes a string and gets the bytes desired for a MetadataAddress.
func stringAsHashedBytes(str string) []byte {
	bz := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(str))))
	return bz[0:16]
}

// buildBytes creates a new slice with the provided bytes.
func buildBytes(key byte, parts ...[]byte) []byte {
	l := 1
	for _, p := range parts {
		l += len(p)
	}
	retval := make([]byte, 0, l)
	retval = append(retval, key)
	for _, p := range parts {
		retval = append(retval, p...)
	}
	return retval
}

// getPrefixFromKey gets the prefix that corresponds to the provided key byte.
func getPrefixFromKey(key byte) string {
	switch key {
	case KeyScope:
		return PrefixScope
	case KeySession:
		return PrefixSession
	case KeyRecord:
		return PrefixRecord
	case KeyScopeSpecification:
		return PrefixScopeSpecification
	case KeyContractSpecification:
		return PrefixContractSpecification
	case KeyRecordSpecification:
		return PrefixRecordSpecification
	default:
		panic(fmt.Errorf("invalid key: %d", key))
	}
}

// validateBytes makes sure the provided bytes have a correct key and length.
func validateBytes(bz []byte) error {
	if len(bz) == 0 {
		return fmt.Errorf("no bytes found in metadata address")
	}
	expectedLength := 0
	switch bz[0] {
	case KeyScope:
		expectedLength = 17
	case KeySession:
		expectedLength = 33
	case KeyRecord:
		expectedLength = 33
	case KeyScopeSpecification:
		expectedLength = 17
	case KeyContractSpecification:
		expectedLength = 17
	case KeyRecordSpecification:
		expectedLength = 33
	default:
		return fmt.Errorf("invalid key: %d", bz[0])
	}
	if expectedLength != len(bz) {
		return fmt.Errorf("incorrect data length for %s address: expected %d, actual %d", getPrefixFromKey(bz[0]), expectedLength, len(bz))
	}
	return nil
}
