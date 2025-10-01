package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgBindNameRequest)(nil),
	(*MsgDeleteNameRequest)(nil),
	(*MsgModifyNameRequest)(nil),
	(*MsgCreateRootNameRequest)(nil),
	(*MsgUpdateParamsRequest)(nil),
}

func NewMsgBindNameRequest(record, parent NameRecord) *MsgBindNameRequest {
	return &MsgBindNameRequest{
		Parent: parent,
		Record: record,
	}
}

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

func NewMsgDeleteNameRequest(record NameRecord) *MsgDeleteNameRequest {
	return &MsgDeleteNameRequest{
		Record: record,
	}
}

func (msg MsgDeleteNameRequest) ValidateBasic() error {
	if strings.TrimSpace(msg.Record.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.TrimSpace(msg.Record.Address) == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return nil
}

func NewMsgModifyNameRequest(authority string, name string, owner sdk.AccAddress, restricted bool) *MsgModifyNameRequest {
	return &MsgModifyNameRequest{
		Authority: authority,
		Record:    NewNameRecord(name, owner, restricted),
	}
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

func (msg MsgCreateRootNameRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress
	}

	err := msg.Record.Validate()
	if err != nil {
		return err
	}

	return nil
}

func NewMsgUpdateParamsRequest(
	maxSegmentLength uint32,
	minSegmentLength uint32,
	maxNameLevels uint32,
	allowUnrestrictedNames bool,
	authority string,
) *MsgUpdateParamsRequest {
	return &MsgUpdateParamsRequest{
		Authority: authority,
		Params: NewParams(
			maxSegmentLength,
			minSegmentLength,
			maxNameLevels,
			allowUnrestrictedNames,
		),
	}
}

func (msg MsgUpdateParamsRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	return err
}
