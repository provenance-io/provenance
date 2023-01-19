package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile time interface checks.
var _, _ sdk.Msg = &MsgBindNameRequest{}, &MsgDeleteNameRequest{}

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
