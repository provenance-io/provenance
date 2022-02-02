package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

const (
	TypeMsgAddAttribute            = "add_attribute"
	TypeMsgUpdateAttribute         = "update_attribute"
	TypeMsgDeleteAttribute         = "delete_attribute"
	TypeMsgDeleteDistinctAttribute = "delete_distinct_attribute"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgAddAttributeRequest{}
	_ sdk.Msg = &MsgUpdateAttributeRequest{}
	_ sdk.Msg = &MsgDeleteAttributeRequest{}
	_ sdk.Msg = &MsgDeleteDistinctAttributeRequest{}
)

// NewMsgAddAttributeRequest creates a new add attribute message
func NewMsgAddAttributeRequest(account string, owner sdk.AccAddress, name string, attributeType AttributeType, value []byte) *MsgAddAttributeRequest { // nolint:interfacer
	return &MsgAddAttributeRequest{
		Account:       account,
		Name:          strings.ToLower(strings.TrimSpace(name)),
		Owner:         owner.String(),
		AttributeType: attributeType,
		Value:         value,
	}
}

// Route returns the name of the module.
func (msg MsgAddAttributeRequest) Route() string {
	return ModuleName
}

// Type returns the message action.
func (msg MsgAddAttributeRequest) Type() string { return TypeMsgAddAttribute }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgAddAttributeRequest) ValidateBasic() error {
	if len(msg.Owner) == 0 {
		return fmt.Errorf("empty owner address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	a := NewAttribute(msg.Name, msg.Account, msg.AttributeType, msg.Value)
	return a.ValidateBasic()
}

// GetSignBytes encodes the message for signing
func (msg MsgAddAttributeRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the name owner.
func (msg MsgAddAttributeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(fmt.Errorf("invalid owner value on message: %w", err))
	}
	return []sdk.AccAddress{addr}
}

// String implements stringer interface
func (msg MsgAddAttributeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// NewMsgUpdateAttributeRequest creates a new add attribute message
func NewMsgUpdateAttributeRequest(account string, owner sdk.AccAddress, name string, originalValue []byte, updateValue []byte, origAttrType AttributeType, updatedAttrType AttributeType) *MsgUpdateAttributeRequest { // nolint:interfacer
	return &MsgUpdateAttributeRequest{
		Account:               account,
		Name:                  strings.ToLower(strings.TrimSpace(name)),
		Owner:                 owner.String(),
		OriginalValue:         originalValue,
		UpdateValue:           updateValue,
		OriginalAttributeType: origAttrType,
		UpdateAttributeType:   updatedAttrType,
	}
}

// Route returns the name of the module.
func (msg MsgUpdateAttributeRequest) Route() string {
	return ModuleName
}

// Type returns the message action.
func (msg MsgUpdateAttributeRequest) Type() string { return TypeMsgUpdateAttribute }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgUpdateAttributeRequest) ValidateBasic() error {
	if len(msg.Owner) == 0 {
		return fmt.Errorf("empty owner address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	a := NewAttribute(msg.Name, msg.Account, msg.UpdateAttributeType, msg.UpdateValue)
	return a.ValidateBasic()
}

// GetSignBytes encodes the message for signing
func (msg MsgUpdateAttributeRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the name owner.
func (msg MsgUpdateAttributeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(fmt.Errorf("invalid owner value on message: %w", err))
	}
	return []sdk.AccAddress{addr}
}

// String implements stringer interface
func (msg MsgUpdateAttributeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// NewMsgDeleteAttributeRequest deletes all attributes with specific name
func NewMsgDeleteAttributeRequest(account string, owner sdk.AccAddress, name string) *MsgDeleteAttributeRequest { // nolint:interfacer
	return &MsgDeleteAttributeRequest{
		Account: account,
		Name:    strings.ToLower(strings.TrimSpace(name)),
		Owner:   owner.String(),
	}
}

// Route returns the name of the module.
func (msg MsgDeleteAttributeRequest) Route() string {
	return ModuleName
}

// Type returns the message action.
func (msg MsgDeleteAttributeRequest) Type() string { return TypeMsgDeleteAttribute }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteAttributeRequest) ValidateBasic() error {
	if strings.TrimSpace(msg.Name) == "" {
		return fmt.Errorf("empty name")
	}
	if err := ValidateAttributeAddress(msg.Account); err != nil {
		return fmt.Errorf("invalid account address: %w", err)
	}
	if len(msg.Owner) == 0 {
		return fmt.Errorf("empty owner address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	return nil
}

// String implements stringer interface
func (msg MsgDeleteAttributeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// GetSignBytes encodes the message for signing
func (msg MsgDeleteAttributeRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the name owner.
func (msg MsgDeleteAttributeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(fmt.Errorf("invalid owner value on message: %w", err))
	}
	return []sdk.AccAddress{addr}
}

// NewMsgDeleteDistinctAttributeRequest deletes a attribute with specific value and type
func NewMsgDeleteDistinctAttributeRequest(account string, owner sdk.AccAddress, name string, value []byte) *MsgDeleteDistinctAttributeRequest { // nolint:interfacer
	return &MsgDeleteDistinctAttributeRequest{
		Account: account,
		Name:    strings.ToLower(strings.TrimSpace(name)),
		Owner:   owner.String(),
		Value:   value,
	}
}

// Route returns the name of the module.
func (msg MsgDeleteDistinctAttributeRequest) Route() string {
	return ModuleName
}

// Type returns the message action.
func (msg MsgDeleteDistinctAttributeRequest) Type() string { return TypeMsgDeleteDistinctAttribute }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteDistinctAttributeRequest) ValidateBasic() error {
	if strings.TrimSpace(msg.Name) == "" {
		return fmt.Errorf("empty name")
	}
	if len(msg.Value) == 0 {
		return fmt.Errorf("empty value")
	}
	if err := ValidateAttributeAddress(msg.Account); err != nil {
		return fmt.Errorf("invalid account address: %w", err)
	}
	if len(msg.Owner) == 0 {
		return fmt.Errorf("empty owner address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	return nil
}

// String implements stringer interface
func (msg MsgDeleteDistinctAttributeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// GetSignBytes encodes the message for signing
func (msg MsgDeleteDistinctAttributeRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners indicates that the message must have been signed by the name owner.
func (msg MsgDeleteDistinctAttributeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(fmt.Errorf("invalid owner value on message: %w", err))
	}
	return []sdk.AccAddress{addr}
}
