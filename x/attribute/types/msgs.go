package types

import (
	"fmt"
	"strings"
	time "time"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgAddAttributeRequest)(nil),
	(*MsgUpdateAttributeRequest)(nil),
	(*MsgUpdateAttributeExpirationRequest)(nil),
	(*MsgDeleteAttributeRequest)(nil),
	(*MsgDeleteDistinctAttributeRequest)(nil),
	(*MsgSetAccountDataRequest)(nil),
}

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
	a := NewAttribute(msg.Name, msg.Account, msg.AttributeType, msg.Value, msg.ExpirationDate)
	return a.ValidateBasic()
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
	a := NewAttribute(msg.Name, msg.Account, msg.UpdateAttributeType, msg.UpdateValue, nil)
	return a.ValidateBasic()
}

// String implements stringer interface
func (msg MsgUpdateAttributeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// NewMsgUpdateAttributeRequest creates a new add attribute message
func NewMsgUpdateAttributeExpirationRequest(account, name, value string, expirationDate *time.Time, owner sdk.AccAddress) *MsgUpdateAttributeExpirationRequest {
	return &MsgUpdateAttributeExpirationRequest{
		Account:        account,
		Name:           strings.ToLower(strings.TrimSpace(name)),
		Value:          []byte(value),
		ExpirationDate: expirationDate,
		Owner:          owner.String(),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgUpdateAttributeExpirationRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	if strings.TrimSpace(msg.Name) == "" {
		return fmt.Errorf("invalid name: empty")
	}
	if msg.Value == nil {
		return fmt.Errorf("invalid value: nil")
	}

	err := ValidateAttributeAddress(msg.Account)
	if err != nil {
		return fmt.Errorf("invalid attribute address: %w", err)
	}
	return nil
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

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgSetAccountDataRequest) ValidateBasic() error {
	// This message is only for regular account addresses. No need to allow for scopes or others.
	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return fmt.Errorf("invalid account: %w", err)
	}
	return nil
}
