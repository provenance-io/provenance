package types

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgAddAttributeRequest{}
	_ sdk.Msg = &MsgUpdateAttributeRequest{}
	_ sdk.Msg = &MsgUpdateAttributeExpirationRequest{}
	_ sdk.Msg = &MsgDeleteAttributeRequest{}
	_ sdk.Msg = &MsgDeleteDistinctAttributeRequest{}
)

// NewMsgAddAttributeRequest creates a new add attribute message
func NewMsgAddAttributeRequest(account string, owner sdk.AccAddress, name string, attributeType AttributeType, value []byte) *MsgAddAttributeRequest {
	return &MsgAddAttributeRequest{
		Account:       account,
		Name:          strings.ToLower(strings.TrimSpace(name)),
		Owner:         owner.String(),
		AttributeType: attributeType,
		Value:         value,
	}
}

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
func NewMsgUpdateAttributeRequest(account string, owner sdk.AccAddress, name string, originalValue []byte, updateValue []byte, origAttrType AttributeType, updatedAttrType AttributeType) *MsgUpdateAttributeRequest {
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

// NewMsgUpdateAttributeRequest creates a new add attribute message
func NewMsgUpdateAttributeExpirationRequest(account string, owner sdk.AccAddress, name string) *MsgUpdateAttributeExpirationRequest {
	return &MsgUpdateAttributeExpirationRequest{
		Account: account,
		Name:    strings.ToLower(strings.TrimSpace(name)),
		Owner:   owner.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgUpdateAttributeExpirationRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the name owner.
func (msg MsgUpdateAttributeExpirationRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

// NewMsgDeleteAttributeRequest deletes all attributes with specific name
func NewMsgDeleteAttributeRequest(account string, owner sdk.AccAddress, name string) *MsgDeleteAttributeRequest {
	return &MsgDeleteAttributeRequest{
		Account: account,
		Name:    strings.ToLower(strings.TrimSpace(name)),
		Owner:   owner.String(),
	}
}

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

// GetSigners indicates that the message must have been signed by the name owner.
func (msg MsgDeleteAttributeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(fmt.Errorf("invalid owner value on message: %w", err))
	}
	return []sdk.AccAddress{addr}
}

// NewMsgDeleteDistinctAttributeRequest deletes a attribute with specific value and type
func NewMsgDeleteDistinctAttributeRequest(account string, owner sdk.AccAddress, name string, value []byte) *MsgDeleteDistinctAttributeRequest {
	return &MsgDeleteDistinctAttributeRequest{
		Account: account,
		Name:    strings.ToLower(strings.TrimSpace(name)),
		Owner:   owner.String(),
		Value:   value,
	}
}

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

// GetSigners indicates that the message must have been signed by the name owner.
func (msg MsgDeleteDistinctAttributeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(fmt.Errorf("invalid owner value on message: %w", err))
	}
	return []sdk.AccAddress{addr}
}
