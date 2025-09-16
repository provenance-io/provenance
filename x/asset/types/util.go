package types

import (
	"encoding/json"
	"fmt"

	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"

	codec "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	proto "github.com/cosmos/gogoproto/proto"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// AnyToString extracts a string value from an Any type that contains a StringValue using the provided codec.
func AnyToString(cdc codec.BinaryCodec, anyMsg *cdctypes.Any) (string, error) {
	if anyMsg == nil {
		return "", nil
	}

	var unpacked proto.Message
	if err := cdc.UnpackAny(anyMsg, &unpacked); err != nil {
		return "", err
	}
	sv, ok := unpacked.(*wrapperspb.StringValue)
	if !ok {
		return "", NewErrCodeInvalidField("any_message", fmt.Sprintf("expected StringValue, got %T", unpacked))
	}
	return sv.Value, nil
}

// StringToAny converts a string to an Any type by wrapping it in a StringValue
func StringToAny(str string) (*cdctypes.Any, error) {
	strMsg := wrapperspb.String(str)
	anyMsg, err := cdctypes.NewAnyWithValue(strMsg)
	if err != nil {
		return nil, NewErrCodeInternal(fmt.Sprintf("failed to create Any from string: %v", err))
	}
	return anyMsg, nil
}

func AnyToJSONSchema(cdc codec.BinaryCodec, anyValue *cdctypes.Any) (map[string]interface{}, error) {
	if anyValue == nil {
		return nil, nil
	}

	schemaString, err := AnyToString(cdc, anyValue)
	if err != nil {
		return nil, err
	}

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaString), &schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func ValidateJSONSchema(schema map[string]interface{}, data []byte) error {
	// Decode the JSON instance.
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return NewErrCodeInvalidField("data", fmt.Sprintf("invalid JSON data: %v", err))
	}
	return validateAgainstSchema(schema, value)
}

// validateAgainstSchema provides a minimal, dependency-free subset of JSON Schema validation
// sufficient for our use cases/tests:
// - Supported types: object, array, string, integer
// - Object keywords: properties, required
// - Array keywords: items
// - Integer keywords: minimum
func validateAgainstSchema(schema map[string]interface{}, value interface{}) error {
	if schema == nil {
		return nil
	}

	schemaType, _ := schema["type"].(string)
	switch schemaType {
	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			return NewErrCodeInvalidField("schema_validation", "expected object")
		}

		// required
		if reqVal, exists := schema["required"]; exists {
			required, err := toStringSlice(reqVal)
			if err != nil {
				return NewErrCodeInvalidField("required", err.Error())
			}
			for _, key := range required {
				if _, present := obj[key]; !present {
					return NewErrCodeMissingField(key)
				}
			}
		}

		// properties
		if propsVal, exists := schema["properties"]; exists {
			props, ok := propsVal.(map[string]interface{})
			if !ok {
				return NewErrCodeInvalidField("properties", "invalid properties definition")
			}
			for key, sub := range props {
				subSchema, ok := sub.(map[string]interface{})
				if !ok {
					return NewErrCodeInvalidField("property_schema", fmt.Sprintf("invalid schema for property: %s", key))
				}
				if val, present := obj[key]; present {
					if err := validateAgainstSchema(subSchema, val); err != nil {
						return NewErrCodeInvalidField(fmt.Sprintf("property_%s", key), err.Error())
					}
				}
			}
		}
		return nil

	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			return NewErrCodeInvalidField("schema_validation", "expected array")
		}
		if itemsVal, exists := schema["items"]; exists {
			itemSchema, ok := itemsVal.(map[string]interface{})
			if !ok {
				return NewErrCodeInvalidField("items", "invalid items definition")
			}
			for i, item := range arr {
				if err := validateAgainstSchema(itemSchema, item); err != nil {
					return NewErrCodeInvalidField(fmt.Sprintf("item_%d", i), err.Error())
				}
			}
		}
		return nil

	case "string":
		if _, ok := value.(string); !ok {
			return NewErrCodeInvalidField("schema_validation", "expected string")
		}
		return nil

	case "integer":
		f, ok := numericToUint64(value)
		if !ok {
			return NewErrCodeInvalidField("schema_validation", "expected integer")
		}
		// minimum
		if minVal, ok := schema["minimum"]; ok {
			minn, ok := numericToUint64(minVal)
			if !ok {
				return NewErrCodeInvalidField("minimum", "invalid minimum")
			}
			if f < minn {
				return NewErrCodeInvalidField("value", fmt.Sprintf("value %v is less than minimum %v", f, minn))
			}
		}
		return nil

	default:
		// Unknown/unspecified type: accept to remain permissive
		return nil
	}
}

// toStringSlice converts a value to a slice of strings
func toStringSlice(v interface{}) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	if ss, ok := v.([]string); ok {
		return ss, nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil, NewErrCodeInvalidField("string_array", "expected array of strings")
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		s, ok := it.(string)
		if !ok {
			return nil, NewErrCodeInvalidField("array_element", "expected string in array")
		}
		out = append(out, s)
	}
	return out, nil
}

// numericToUint64 safely converts a numeric value to a uint64
func numericToUint64(v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case uint64:
		return n, true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int32:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case uint:
		return uint64(n), true
	case uint32:
		return uint64(n), true
	default:
		return 0, false
	}
}

// NewDefaultMarker creates a new default marker account for a given token and address
func NewDefaultMarker(token sdk.Coin, addr string) (*markertypes.MarkerAccount, error) {
	// Get the from address
	fromAcc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, NewErrCodeInvalidField("address", err.Error())
	}

	// Get the address of the new marker.
	markerAddr, err := markertypes.MarkerAddress(token.Denom)
	if err != nil {
		return nil, NewErrCodeInternal(fmt.Sprintf("failed to create marker address: %v", err))
	}

	// Create a new marker account
	marker := markertypes.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(markerAddr),
		token,
		fromAcc,
		[]markertypes.AccessGrant{
			{
				Address: fromAcc.String(),
				Permissions: markertypes.AccessList{
					markertypes.Access_Admin,
					markertypes.Access_Mint,
					markertypes.Access_Burn,
					markertypes.Access_Withdraw,
					markertypes.Access_Transfer,
				},
			},
		},
		markertypes.StatusProposed,
		markertypes.MarkerType_RestrictedCoin,
		true,       // Supply fixed
		false,      // Allow governance control
		false,      // Don't allow forced transfer
		[]string{}, // No required attributes
	)

	return marker, nil
}

func validateJSON(data string) error {
	if data == "" {
		return nil // Empty data is valid
	}

	var jsonData any
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return NewErrCodeInvalidField("data", fmt.Sprintf("invalid JSON data: %v", err))
	}

	return nil
}

// validateJSONSchema validates that the provided string is a well-formed JSON schema
func validateJSONSchema(data string) error {
	if data == "" {
		return nil // Empty data is valid
	}

	// Try to parse the data as JSON
	var jsonData any
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return NewErrCodeInvalidField("data", fmt.Sprintf("invalid JSON data: %v", err))
	}

	// Check if it's a JSON schema by looking for required schema properties
	schemaMap, ok := jsonData.(map[string]any)
	if !ok {
		return NewErrCodeInvalidField("data", "data is not a JSON object")
	}

	// Check for type property which is required in JSON schemas
	if _, hasType := schemaMap["type"]; !hasType {
		return NewErrCodeMissingField("type")
	}

	return nil
}
