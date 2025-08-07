package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreateAsset{}
var _ sdk.Msg = &MsgCreateAssetClass{}
var _ sdk.Msg = &MsgCreatePool{}
var _ sdk.Msg = &MsgCreateTokenization{}
var _ sdk.Msg = &MsgCreateSecuritization{}

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgCreateAsset)(nil),
	(*MsgCreateAssetClass)(nil),
	(*MsgCreatePool)(nil),
	(*MsgCreateTokenization)(nil),
	(*MsgCreateSecuritization)(nil),
}

// ValidateBasic implements Msg
func (msg MsgCreateAsset) ValidateBasic() error {
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

	if msg.FromAddress == "" {
		return fmt.Errorf("from_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return fmt.Errorf("invalid from_address: %w", err)
	}

	return nil
}

// ValidateBasic implements Msg
func (msg MsgCreateAssetClass) ValidateBasic() error {
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

	if msg.FromAddress == "" {
		return fmt.Errorf("from_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return fmt.Errorf("invalid from_address: %w", err)
	}

	return nil
}

// ValidateBasic implements Msg
func (msg MsgCreatePool) ValidateBasic() error {
	if msg.Pool == nil {
		return fmt.Errorf("pool cannot be nil")
	}

	if err := msg.Pool.Validate(); err != nil {
		return fmt.Errorf("invalid pool: %w", err)
	}

	if len(msg.Assets) == 0 {
		return fmt.Errorf("assets cannot be empty")
	}

	for i, asset := range msg.Assets {
		if asset == nil {
			return fmt.Errorf("asset at index %d cannot be nil", i)
		}
		if asset.ClassId == "" {
			return fmt.Errorf("asset at index %d class_id cannot be empty", i)
		}
		if asset.Id == "" {
			return fmt.Errorf("asset at index %d id cannot be empty", i)
		}
	}

	if msg.FromAddress == "" {
		return fmt.Errorf("from_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return fmt.Errorf("invalid from_address: %w", err)
	}

	return nil
}

// ValidateBasic implements Msg
func (msg MsgCreateTokenization) ValidateBasic() error {
	if err := msg.Denom.Validate(); err != nil {
		return fmt.Errorf("invalid denom: %w", err)
	}

	if msg.Asset == nil {
		return fmt.Errorf("asset cannot be nil")
	}

	if msg.Asset.ClassId == "" {
		return fmt.Errorf("asset class_id cannot be empty")
	}

	if msg.Asset.Id == "" {
		return fmt.Errorf("asset id cannot be empty")
	}

	if msg.FromAddress == "" {
		return fmt.Errorf("from_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return fmt.Errorf("invalid from_address: %w", err)
	}

	return nil
}

// ValidateBasic implements Msg
func (msg MsgCreateSecuritization) ValidateBasic() error {
	if msg.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if len(msg.Pools) == 0 {
		return fmt.Errorf("pools cannot be empty")
	}

	for i, pool := range msg.Pools {
		if pool == "" {
			return fmt.Errorf("pool at index %d cannot be empty", i)
		}
	}

	if len(msg.Tranches) == 0 {
		return fmt.Errorf("tranches cannot be empty")
	}

	for i, tranche := range msg.Tranches {
		if tranche == nil {
			return fmt.Errorf("tranche at index %d cannot be nil", i)
		}
		if err := tranche.Validate(); err != nil {
			return fmt.Errorf("invalid tranche at index %d: %w", i, err)
		}
	}

	if msg.FromAddress == "" {
		return fmt.Errorf("from_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return fmt.Errorf("invalid from_address: %w", err)
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
func (msg MsgCreateAsset) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// GetSigners implements Msg
func (msg MsgCreateAssetClass) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// GetSigners implements Msg
func (msg MsgCreatePool) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// GetSigners implements Msg
func (msg MsgCreateTokenization) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

// GetSigners implements Msg
func (msg MsgCreateSecuritization) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
