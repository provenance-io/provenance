package types

import (
	"encoding/json"
	"fmt"

	codec "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// AnyToString extracts a string value from an Any type that contains an AssetData using the provided codec.
func AnyToString(cdc codec.BinaryCodec, anyMsg *cdctypes.Any) (string, error) {
	if anyMsg == nil {
		return "", fmt.Errorf("nil Any")
	}

	var msg proto.Message
	if err := cdc.UnpackAny(anyMsg, &msg); err != nil {
		return "", fmt.Errorf("failed to unpack Any (type_url=%q) as %s: %w", anyMsg.TypeUrl, proto.MessageName(&AssetData{}), err)
	}
	ad, ok := msg.(*AssetData)
	if !ok {
		return "", fmt.Errorf("unexpected Any concrete type: got type_url=%q, expected %s", anyMsg.TypeUrl, proto.MessageName(&AssetData{}))
	}
	return ad.Value, nil
}

// StringToAny converts a string to an Any type by wrapping it in an AssetData
func StringToAny(str string) (*cdctypes.Any, error) {
	strMsg := AssetData{Value: str}
	anyMsg, err := cdctypes.NewAnyWithValue(&strMsg)
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

func ValidateDataWithJSONSchema(schema map[string]interface{}, data []byte) error {
	// Decode the JSON instance.
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("data is not valid JSON: %w", err)
	}
	return validateAgainstSchema(schema, jsonData)
}

// validateAgainstSchema provides a minimal, dependency-free subset of JSON Schema validation
// sufficient for our use cases/tests:
// - Supported types: object, array, string
// - Object keywords: properties, required
// - Array keywords: items
func validateAgainstSchema(schema map[string]interface{}, jsonData interface{}) error {
	if schema == nil {
		return nil
	}

	schemaType, _ := schema["type"].(string)
	switch schemaType {
	case "object":
		obj, ok := jsonData.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid object definition")
		}

		// required
		if reqVal, exists := schema["required"]; exists {
			required, err := toStringSlice(reqVal)
			if err != nil {
				return fmt.Errorf("invalid required definition: %w", err)
			}
			for _, key := range required {
				if _, present := obj[key]; !present {
					return fmt.Errorf("required %s cannot be empty", key)
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
					return fmt.Errorf("invalid property %s definition", key)
				}
				if val, present := obj[key]; present {
					if err := validateAgainstSchema(subSchema, val); err != nil {
						return fmt.Errorf("invalid property %s definition: %w", key, err)
					}
				}
			}
		}
		return nil

	case "array":
		arr, ok := jsonData.([]interface{})
		if !ok {
			return fmt.Errorf("invalid array definition")
		}
		if itemsVal, exists := schema["items"]; exists {
			itemSchema, ok := itemsVal.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid array items definition")
			}
			for i, item := range arr {
				if err := validateAgainstSchema(itemSchema, item); err != nil {
					return fmt.Errorf("invalid array item definition at index %d: %w", i, err)
				}
			}
		}
		return nil

	case "string":
		if _, ok := jsonData.(string); !ok {
			return fmt.Errorf("invalid string definition")
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
		return nil, fmt.Errorf("string_array: expected array of strings")
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		s, ok := it.(string)
		if !ok {
			return nil, fmt.Errorf("array_element: expected string in array")
		}
		out = append(out, s)
	}
	return out, nil
}

// NewDefaultMarker creates a new default marker account for a given token and address
func NewDefaultMarker(token sdk.Coin, addr string) (*markertypes.MarkerAccount, error) {
	// Get the signer address
	signerAcc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Get the address of the new marker.
	markerAddr, err := markertypes.MarkerAddress(token.Denom)
	if err != nil {
		return nil, fmt.Errorf("failed to create marker address from denom: %w", err)
	}

	// Create a new marker account
	marker := markertypes.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(markerAddr),
		token,
		signerAcc,
		[]markertypes.AccessGrant{
			{
				Address: signerAcc.String(),
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
		false,      // Disallow governance control
		false,      // Don't allow forced transfer
		[]string{}, // No required attributes
	)

	return marker, nil
}
