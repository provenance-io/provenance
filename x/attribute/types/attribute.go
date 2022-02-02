package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	fmt "fmt"
	"math/big"
	"net/url"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

// NewAttribute creates a new instance of an Attribute
func NewAttribute(name string, address string, attrType AttributeType, value []byte) Attribute { // nolint:interfacer
	// Ensure string type values are trimmed.
	if attrType != AttributeType_Bytes && attrType != AttributeType_Proto {
		trimmed := strings.TrimSpace(string(value))
		value = []byte(trimmed)
	}
	return Attribute{
		Name:          name,
		Address:       address,
		AttributeType: attrType,
		Value:         value,
	}
}

// String implements fmt.Stringer
func (a Attribute) String() string {
	value := base64.StdEncoding.EncodeToString(a.Value)
	return fmt.Sprintf("Name: %s, Type: %s, Value: %s", a.Name, a.AttributeType, value)
}

// Hash returns the SHA256 hash of the attribute value.
func (a Attribute) Hash() []byte {
	sum := sha256.Sum256(a.Value)
	return sum[:]
}

// ValidateBasic ensures an attribute is valid.
func (a Attribute) ValidateBasic() error {
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("invalid name: empty")
	}
	if a.Value == nil {
		return fmt.Errorf("invalid value: nil")
	}

	err := ValidateAttributeAddress(a.Address)
	if err != nil {
		return fmt.Errorf("invalid attribute address: %w", err)
	}

	if !ValidAttributeType(a.AttributeType) {
		return fmt.Errorf("invalid attribute type")
	}
	if !isValidValueForType(a.AttributeType, a.Value) {
		return fmt.Errorf("invalid attribute value for assigned type: %s", a.AttributeType)
	}
	return nil
}

// GetAddressBytes Gets the bytes of this attribute's address.
// If the address is neither an account address nor metadata address (or is an empty string), an empty byte slice is returned.
func (a Attribute) GetAddressBytes() []byte {
	return GetAttributeAddressBytes(a.Address)
}

// GetAttributeAddressBytes Gets the bytes of an address used in an attribute.
// If the address is neither an account address nor metadata address (or is an empty string), an empty byte slice is returned.
func GetAttributeAddressBytes(addr string) []byte {
	if len(strings.TrimSpace(addr)) == 0 {
		return []byte{}
	}
	accAddr, accErr := sdk.AccAddressFromBech32(addr)
	if accErr == nil {
		return accAddr.Bytes()
	}
	mdAddr, mdErr := metadatatypes.MetadataAddressFromBech32(addr)
	if mdErr == nil {
		return mdAddr.Bytes()
	}
	return []byte{}
}

// ValidateAttributeAddress validates that the provide string is a valid address for an attribute.
// Failures:
//  * The provided address is empty
//  * The provided address is neither an account address nor scope metadata address.
func ValidateAttributeAddress(addr string) error {
	if len(strings.TrimSpace(addr)) == 0 {
		return errors.New("must not be empty")
	}
	_, accErr := sdk.AccAddressFromBech32(addr)
	if accErr == nil {
		return nil
	}
	mdAddr, mdErr := metadatatypes.MetadataAddressFromBech32(addr)
	if mdErr == nil && mdAddr.IsScopeAddress() {
		return nil
	}
	return fmt.Errorf("must be either an account address or scope metadata address: %q", addr)
}

// Determines whether a byte array value is valid for the given type.
func isValidValueForType(attrType AttributeType, value []byte) bool {
	switch attrType {
	case AttributeType_UUID:
		return isValidUUID(value)
	case AttributeType_JSON:
		return json.Valid(value)
	case AttributeType_String:
		return isValidString(value)
	case AttributeType_Uri:
		return isValidURI(value)
	case AttributeType_Int:
		return isValidInt(value)
	case AttributeType_Float:
		return isValidFloat(value)
	case AttributeType_Proto:
		return true // Treat proto as just a special tag for bytes
	case AttributeType_Bytes:
		return true
	default:
		return false
	}
}

// Ensure byte array can be parsed into a UUID.
func isValidUUID(value []byte) bool {
	s := strings.TrimSpace(string(value))
	if _, err := uuid.Parse(s); err != nil {
		return false
	}
	return true
}

// Ensure a byte array is a non-empty string.
func isValidString(value []byte) bool {
	return strings.TrimSpace(string(value)) != ""
}

// Ensure a byte array is a URI.
func isValidURI(value []byte) bool {
	if _, err := url.ParseRequestURI(string(value)); err != nil {
		return false
	}
	return true
}

// Ensure a byte array is a string that can be converted to an base-10 big.Int.
func isValidInt(value []byte) bool {
	s := strings.TrimSpace(string(value))
	_, ok := new(big.Int).SetString(s, 10)
	return ok
}

// Ensure a byte array is a string that can be converted to big.Float.
func isValidFloat(value []byte) bool {
	s := strings.TrimSpace(string(value))
	_, ok := new(big.Float).SetString(s)
	return ok
}

// AttributeTypeFromString returns a AttributeType from a string. It returns an error
// if the string is invalid.
func AttributeTypeFromString(str string) (AttributeType, error) {
	str = "ATTRIBUTE_TYPE_" + strings.ToUpper(str)
	option, ok := AttributeType_value[str]
	if !ok {
		return AttributeType_Unspecified, fmt.Errorf("'%s' is not a valid attribute type option", str)
	}
	return AttributeType(option), nil
}

// ValidAttributeType returns true if the attribute type option is valid and false otherwise.
func ValidAttributeType(attributeType AttributeType) bool {
	if attributeType == AttributeType_UUID ||
		attributeType == AttributeType_JSON ||
		attributeType == AttributeType_String ||
		attributeType == AttributeType_Uri ||
		attributeType == AttributeType_Int ||
		attributeType == AttributeType_Float ||
		attributeType == AttributeType_Proto ||
		attributeType == AttributeType_Bytes {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility.
func (at AttributeType) Marshal() ([]byte, error) {
	return []byte{byte(at)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (at *AttributeType) Unmarshal(data []byte) error {
	*at = AttributeType(data[0])
	return nil
}

// Format implements the fmt.Formatter interface.
func (at AttributeType) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(at.String()))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(at))))
	}
}
