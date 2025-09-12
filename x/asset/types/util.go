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
		return "", fmt.Errorf("expected StringValue, got %T", unpacked)
	}
	return sv.Value, nil
}

// StringToAny converts a string to an Any type by wrapping it in a StringValue
func StringToAny(str string) (*cdctypes.Any, error) {
	strMsg := wrapperspb.String(str)
	anyMsg, err := cdctypes.NewAnyWithValue(strMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Any from string: %w", err)
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
		return fmt.Errorf("invalid JSON data: %w", err)
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
			return fmt.Errorf("expected object")
		}

		// required
		if reqVal, exists := schema["required"]; exists {
			required, err := toStringSlice(reqVal)
			if err != nil {
				return fmt.Errorf("invalid required: %w", err)
			}
			for _, key := range required {
				if _, present := obj[key]; !present {
					return fmt.Errorf("missing required field: %s", key)
				}
			}
		}

		// properties
		if propsVal, exists := schema["properties"]; exists {
			props, ok := propsVal.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid properties definition")
			}
			for key, sub := range props {
				subSchema, ok := sub.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid schema for property: %s", key)
				}
				if val, present := obj[key]; present {
					if err := validateAgainstSchema(subSchema, val); err != nil {
						return fmt.Errorf("property %s: %w", key, err)
					}
				}
			}
		}
		return nil

	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array")
		}
		if itemsVal, exists := schema["items"]; exists {
			itemSchema, ok := itemsVal.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid items definition")
			}
			for i, item := range arr {
				if err := validateAgainstSchema(itemSchema, item); err != nil {
					return fmt.Errorf("item %d: %w", i, err)
				}
			}
		}
		return nil

	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string")
		}
		return nil

	case "integer":
		f, ok := numericToUint64(value)
		if !ok {
			return fmt.Errorf("expected integer")
		}
		// minimum
		if minVal, ok := schema["minimum"]; ok {
			minn, ok := numericToUint64(minVal)
			if !ok {
				return fmt.Errorf("invalid minimum")
			}
			if f < minn {
				return fmt.Errorf("value %v is less than minimum %v", f, minn)
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
		return nil, fmt.Errorf("expected array of strings")
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		s, ok := it.(string)
		if !ok {
			return nil, fmt.Errorf("expected string in array")
		}
		out = append(out, s)
	}
	return out, nil
}

// numericToUint64 converts a numeric value to a uint64
func numericToUint64(v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case uint64:
		return n, true
	case int:
		return uint64(n), true
	case int32:
		return uint64(n), true
	case int64:
		return uint64(n), true
	case uint:
		return uint64(n), true
	case uint32:
		return uint64(n), true
	default:
		return 0, false
	}
}

// NewDefaultMarker creates a new default marker account for a given denom and from address
func NewDefaultMarker(token sdk.Coin, fromAddr string) (*markertypes.MarkerAccount, error) {
	// Get the from address
	fromAcc, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	// Get the address of the new marker.
	markerAddr, err := markertypes.MarkerAddress(token.Denom)
	if err != nil {
		return nil, fmt.Errorf("failed to create marker address: %w", err)
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
		return fmt.Errorf("invalid JSON data: %w", err)
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
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	// Check if it's a JSON schema by looking for required schema properties
	schemaMap, ok := jsonData.(map[string]any)
	if !ok {
		return fmt.Errorf("data is not a JSON object")
	}

	// Check for type property which is required in JSON schemas
	if _, hasType := schemaMap["type"]; !hasType {
		return fmt.Errorf("data is missing required 'type' property for JSON schema")
	}

	return nil
}
