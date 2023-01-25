package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Compile time interface checks.
var _, _, _, _ sdk.Msg = &MsgBindNameRequest{}, &MsgDeleteNameRequest{}, &MsgModifyNameRequest{}, &MsgCreateRootNameRequest{}

// NewMsgBindNameRequest creates a new bind name request
func NewMsgBindNameRequest(record, parent NameRecord) *MsgBindNameRequest {
	return &MsgBindNameRequest{
		Parent: parent,
		Record: record,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgBindNameRequest) ValidateBasic() error {
	if strings.TrimSpace(msg.Parent.Name) == "" {
		return fmt.Errorf("parent name cannot be empty")
	}
	if strings.TrimSpace(msg.Parent.Address) == "" {
		return fmt.Errorf("parent address cannot be empty")
	}
	if strings.TrimSpace(msg.Record.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.Contains(msg.Record.Name, ".") {
		return fmt.Errorf("invalid name: \".\" is reserved")
	}
	if strings.TrimSpace(msg.Record.Address) == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgBindNameRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Parent.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgDeleteNameRequest creates a new Delete Name Request
func NewMsgDeleteNameRequest(record NameRecord) *MsgDeleteNameRequest {
	return &MsgDeleteNameRequest{
		Record: record,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDeleteNameRequest) ValidateBasic() error {
	if strings.TrimSpace(msg.Record.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.TrimSpace(msg.Record.Address) == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the record owner.
func (msg MsgDeleteNameRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Record.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// NewMsgCreateRootNameRequest creates a new Create Root Name Request
func NewMsgCreateRootNameRequest(authority string, name string, address string, restricted bool) *MsgCreateRootNameRequest {
	return &MsgCreateRootNameRequest{
		Authority: authority,
		Record: &NameRecord{
			Name:       name,
			Address:    address,
			Restricted: restricted,
		},
	}
}

// NewMsgModifyNameRequest modifies an existing name record
func NewMsgModifyNameRequest(authority string, name string, owner sdk.AccAddress, restricted bool) *MsgModifyNameRequest {
	return &MsgModifyNameRequest{
		Authority: authority,
		Record:    NewNameRecord(name, owner, restricted),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCreateRootNameRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress
	}

	if _, err := sdk.AccAddressFromBech32(msg.Record.Address); err != nil {
		return ErrInvalidAddress
	}

	if strings.TrimSpace(msg.Record.Name) == "" {
		return ErrInvalidLengthName
	}

	if strings.Contains(msg.Record.Name, ".") {
		return ErrNameContainsSegments
	}

	return nil
}

// GetSigners Implements Msg.
func (msg MsgCreateRootNameRequest) GetSigners() []sdk.AccAddress {
	fromAddress, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{fromAddress}
}

func (msg MsgModifyNameRequest) ValidateBasic() error {
	if strings.TrimSpace(msg.Record.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Record.Address); err != nil {
		return fmt.Errorf("invalid record address: %w", err)
	}
	if strings.TrimSpace(msg.GetAuthority()) == "" {
		return govtypes.ErrInvalidSigner
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the gov module.
func (msg MsgModifyNameRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.GetAuthority())
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
