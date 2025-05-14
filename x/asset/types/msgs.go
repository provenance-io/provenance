package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//var _ sdk.Msg = &MsgAddAsset{}
//var _ sdk.Msg = &MsgAddAssetClass{}

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgAddAsset)(nil),
	(*MsgAddAssetClass)(nil),
	(*MsgCreatePool)(nil),
	(*MsgCreateParticipation)(nil),
}

// ValidateBasic implements Msg
func (msg MsgAddAsset) ValidateBasic() error {
	if msg.Asset == nil {
		return fmt.Errorf("asset cannot be nil")
	}

	if msg.Asset.ClassId == "" {
		return fmt.Errorf("class_id cannot be empty")
	}

	if msg.Asset.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if msg.Asset.Data != "" {
		return validateJSON(msg.Asset.Data)
	}

	return nil
}

// ValidateBasic implements Msg
func (msg MsgAddAssetClass) ValidateBasic() error {
	if msg.AssetClass == nil {
		return fmt.Errorf("asset class cannot be nil")
	}

	if msg.AssetClass.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if msg.AssetClass.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if msg.AssetClass.Data != "" {
		return validateJSONSchema(msg.AssetClass.Data)
	}

	return nil
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

	// Check for $schema property which is common in JSON schemas
	if _, hasSchema := schemaMap["$schema"]; !hasSchema {
		// Not a strict requirement, but log a warning
		fmt.Println("Warning: Data does not contain $schema property, may not be a valid JSON schema")
	}

	// Check for type property which is required in JSON schemas
	if _, hasType := schemaMap["type"]; !hasType {
		return fmt.Errorf("data is missing required 'type' property for JSON schema")
	}

	return nil
}

// GetSigners implements Msg
func (msg MsgAddAsset) GetSigners() []sdk.AccAddress {
	// Since there's no owner field in the struct, we'll need to determine the signer
	// based on the asset's ownership or other business logic
	return []sdk.AccAddress{}
}

// GetSigners implements Msg
func (msg MsgAddAssetClass) GetSigners() []sdk.AccAddress {
	// Since there's no authority field in the struct, we'll need to determine the signer
	// based on the asset class's authority or other business logic
	return []sdk.AccAddress{}
}
