// Package wasm supports smart contract integration with the provenance attribute module.
package wasm

import (
	"encoding/json"
	"fmt"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/attribute/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile time interface check
var _ provwasm.Encoder = Encoder

// AttributeMsgParams are params for encoding []sdk.Msg types from the attribute module.
// Only one field should be set.
type AttributeMsgParams struct {
	// A request to encode a MsgAddAttribute
	Add *AddAttributeParams `json:"add_attribute"`
	// A request to encode a MsgDeleteAttribute
	Del *DeleteAttributeParams `json:"delete_attribute"`
	// A request to encode a MsgDeleteAttribute
	DelDistinct *DeleteDistinctAttributeParams `json:"delete_distinct_attribute"`
	// A request to encode a MsgUpdateAttribute
	Update *UpdateAttributeParams `json:"update_attribute"`
}

// AddAttributeParams are params for encoding a MsgAddAttribute
type AddAttributeParams struct {
	// The address of the account to add the attribute to.
	Address string `json:"address"`
	// The attribute name.
	Name string `json:"name"`
	// The attribute value.
	Value []byte `json:"value"`
	// The attribute value type.
	ValueType string `json:"value_type"`
}

// DeleteAttributeParams are params for encoding a MsgDeleteAttribute
type DeleteAttributeParams struct {
	// The address of the account to delete the attribute from.
	Address string `json:"address"`
	// The attribute name.
	Name string `json:"name"`
}

// DeleteDistinctAttributeParams are params for encoding a MsgDeleteDistinctAttribute
type DeleteDistinctAttributeParams struct {
	// The address of the account to delete the attribute from.
	Address string `json:"address"`
	// The attribute name.
	Name string `json:"name"`
	// The attribute value.
	Value []byte `json:"value"`
}

// UpdateAttributeParams are params for encoding a MsgUpdateAttributeRequest
type UpdateAttributeParams struct {
	// The address of the account on the attribute
	Address string `json:"address"`
	// The attribute name.
	Name string `json:"name"`
	// The original attribute value.
	OriginalValue []byte `json:"original_value"`
	// The original attribute value type.
	OriginalValueType string `json:"original_value_type"`
	// The new attribute value.
	UpdateValue []byte `json:"update_value"`
	// The new attribute value type.
	UpdateValueType string `json:"update_value_type"`
}

// Encoder returns a smart contract message encoder for the attribute module.
func Encoder(contract sdk.AccAddress, msg json.RawMessage, version string) ([]sdk.Msg, error) {
	wrapper := struct {
		Params *AttributeMsgParams `json:"attribute"`
	}{}
	if err := json.Unmarshal(msg, &wrapper); err != nil {
		return nil, fmt.Errorf("wasm: failed to unmarshal encode attribute request: %w", err)
	}
	params := wrapper.Params
	if params == nil {
		return nil, fmt.Errorf("wasm: nil attribute encode params")
	}
	switch {
	case params.Add != nil:
		return params.Add.Encode(contract)
	case params.Del != nil:
		return params.Del.Encode(contract)
	case params.DelDistinct != nil:
		return params.DelDistinct.Encode(contract)
	case params.Update != nil:
		return params.Update.Encode(contract)
	default:
		return nil, fmt.Errorf("wasm: invalid attribute encoder params: %s", string(msg))
	}
}

// Encode creates a MsgAddAttribute.
// INFO: The contract must be the owner of the name of the attribute being added.
func (params *AddAttributeParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := types.ValidateAttributeAddress(params.Address); err != nil {
		return nil, fmt.Errorf("wasm: invalid address: %w", err)
	}
	msg := types.NewMsgAddAttributeRequest(
		params.Address,
		contract,
		params.Name,
		encodeType(params.ValueType),
		params.Value,
	)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgDeleteAttribute.
// INFO: The contract must be the owner of the name of the attribute being deleted.
func (params *DeleteAttributeParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := types.ValidateAttributeAddress(params.Address); err != nil {
		return nil, fmt.Errorf("wasm: invalid address: %w", err)
	}
	msg := types.NewMsgDeleteAttributeRequest(params.Address, contract, params.Name)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgDeleteDistinctAttribute.
// INFO: The contract must be the owner of the name of the attribute being deleted.
func (params *DeleteDistinctAttributeParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := types.ValidateAttributeAddress(params.Address); err != nil {
		return nil, fmt.Errorf("wasm: invalid address: %w", err)
	}
	msg := types.NewMsgDeleteDistinctAttributeRequest(params.Address, contract, params.Name, params.Value)
	return []sdk.Msg{msg}, nil
}

// Encode creates a MsgUpdateAttribute.
// INFO: The contract must be the owner of the name of the attribute being updated.
func (params *UpdateAttributeParams) Encode(contract sdk.AccAddress) ([]sdk.Msg, error) {
	if err := types.ValidateAttributeAddress(params.Address); err != nil {
		return nil, fmt.Errorf("wasm: invalid address: %w", err)
	}
	msg := types.NewMsgUpdateAttributeRequest(
		params.Address,
		contract,
		params.Name,
		params.OriginalValue,
		params.UpdateValue,
		encodeType(params.OriginalValueType),
		encodeType(params.UpdateValueType),
	)
	return []sdk.Msg{msg}, nil
}

// Adapt the attribute type from a string passed in message encode params passed from a smart contract.
func encodeType(valueType string) types.AttributeType {
	switch valueType {
	case "bytes":
		return types.AttributeType_Bytes
	case "json":
		return types.AttributeType_JSON
	case "string":
		return types.AttributeType_String
	case "uuid":
		return types.AttributeType_UUID
	case "uri":
		return types.AttributeType_Uri
	case "int":
		return types.AttributeType_Int
	case "float":
		return types.AttributeType_Float
	case "proto":
		return types.AttributeType_Proto
	default:
		return types.AttributeType_Unspecified
	}
}
