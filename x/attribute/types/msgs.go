package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

const (
	TypeMsgAddAttribute    = "add_attribute"
	TypeMsgDeleteAttribute = "delete_attribute"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgAddAttributeRequest{}
	_ sdk.Msg = &MsgDeleteAttributeRequest{}
)

// NewMsgAddAttributeRequest creates a new add attribute message
func NewMsgAddAttributeRequest(account sdk.AccAddress, owner sdk.AccAddress, name string, attributeType AttributeType, value []byte) *MsgAddAttributeRequest { // nolint:interfacer
	return &MsgAddAttributeRequest{Account: account.String(), Name: strings.ToLower(strings.TrimSpace(name)), Owner: owner.String(), AttributeType: attributeType, Value: value}
}

// Route returns the name of the module.
func (msg MsgAddAttributeRequest) Route() string {
	return ModuleName
}

// Type returns the message action.
func (msg MsgAddAttributeRequest) Type() string { return TypeMsgAddAttribute }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgAddAttributeRequest) ValidateBasic() error {
	if len(msg.Account) == 0 {
		return fmt.Errorf("empty account address")
	}
	accAddr, err := sdk.AccAddressFromBech32(msg.Account)
	if err != nil {
		return err
	}
	if len(msg.Owner) == 0 {
		return fmt.Errorf("empty owner address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return err
	}
	a := NewAttribute(msg.Name, accAddr, msg.AttributeType, msg.Value)
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

// NewMsgDeleteAttributeRequest creates a new add attribute message
func NewMsgDeleteAttributeRequest(account sdk.AccAddress, owner sdk.AccAddress, name string) *MsgDeleteAttributeRequest { // nolint:interfacer
	return &MsgDeleteAttributeRequest{Account: account.String(), Name: strings.ToLower(strings.TrimSpace(name)), Owner: owner.String()}
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
	if len(msg.Account) == 0 {
		return fmt.Errorf("empty account address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return err
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
