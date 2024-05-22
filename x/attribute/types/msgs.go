package types

import (
	"fmt"
	"strings"
	time "time"

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
	(*MsgUpdateParamsRequest)(nil),
}

func NewMsgAddAttributeRequest(account string, owner sdk.AccAddress, name string, attributeType AttributeType, value []byte) *MsgAddAttributeRequest {
	return &MsgAddAttributeRequest{
		Account:       account,
		Name:          strings.ToLower(strings.TrimSpace(name)),
		Owner:         owner.String(),
		AttributeType: attributeType,
		Value:         value,
	}
}

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

func NewMsgUpdateAttributeExpirationRequest(account, name, value string, expirationDate *time.Time, owner sdk.AccAddress) *MsgUpdateAttributeExpirationRequest {
	return &MsgUpdateAttributeExpirationRequest{
		Account:        account,
		Name:           strings.ToLower(strings.TrimSpace(name)),
		Value:          []byte(value),
		ExpirationDate: expirationDate,
		Owner:          owner.String(),
	}
}

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

func NewMsgDeleteAttributeRequest(account string, owner sdk.AccAddress, name string) *MsgDeleteAttributeRequest {
	return &MsgDeleteAttributeRequest{
		Account: account,
		Name:    strings.ToLower(strings.TrimSpace(name)),
		Owner:   owner.String(),
	}
}

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

func NewMsgDeleteDistinctAttributeRequest(account string, owner sdk.AccAddress, name string, value []byte) *MsgDeleteDistinctAttributeRequest {
	return &MsgDeleteDistinctAttributeRequest{
		Account: account,
		Name:    strings.ToLower(strings.TrimSpace(name)),
		Owner:   owner.String(),
		Value:   value,
	}
}

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

func (msg MsgSetAccountDataRequest) ValidateBasic() error {
	// This message is only for regular account addresses. No need to allow for scopes or others.
	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return fmt.Errorf("invalid account: %w", err)
	}
	return nil
}

// NewMsgUpdateParamsRequest creates a new UpdateParamsRequest message.
func NewMsgUpdateParamsRequest(authority string, maxValueLength uint32) *MsgUpdateParamsRequest {
	return &MsgUpdateParamsRequest{
		Authority: authority,
		Params:    NewParams(maxValueLength),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (m MsgUpdateParamsRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	return nil
}
