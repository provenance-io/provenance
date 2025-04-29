package types

import (
	"encoding/json"
	"fmt"
	
	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/xeipuuv/gojsonschema"
	"google.golang.org/protobuf/types/known/wrapperspb"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// AnyToString extracts a string value from an Any type that contains a StringValue
func AnyToString(cdc codec.BinaryCodec, anyMsg *cdctypes.Any) (string, error) {
	// Convert to proto Any so we can use UnmarshalTo
	stdAny := &wrapperspb.StringValue{}
	err := proto.Unmarshal(anyMsg.Value, stdAny)
	if err != nil {
		return "", err
	}
	return stdAny.Value, nil
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

func AnyToJSONSchema(cdc codec.BinaryCodec, any *cdctypes.Any) (map[string]interface{}, error) {
	if any == nil {
		return nil, nil
	}

	schemaString, err := AnyToString(cdc, any)
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
	// Convert schema to JSON string
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Load the schema and document
	schemaLoader := gojsonschema.NewStringLoader(string(schemaBytes))
	documentLoader := gojsonschema.NewBytesLoader(data)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("failed to validate schema: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, err.String())
		}
		return fmt.Errorf("validation failed: %s", errMsgs)
	}

	return nil
}
