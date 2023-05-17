package types

import (
	fmt "fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

var _ sdk.Msg = &MsgCreateTriggerRequest{}
var _ sdk.Msg = &MsgDestroyTriggerRequest{}

// NewCreateTriggerRequest Creates a new trigger create request
func NewCreateTriggerRequest(authority string, event TriggerEventI, msgs []sdk.Msg) *MsgCreateTriggerRequest {
	actions, err := sdktx.SetMsgs(msgs)
	if err != nil || len(actions) == 0 {
		fmt.Printf("unable to set messages : %s\n", err)
		return nil
	}

	// This is where we would convert the TriggerEventI into Any
	eventAny, err := codectypes.NewAnyWithValue(event)
	if err != nil {
		fmt.Printf("unable to set messages : %s\n", err)
		return nil
	}

	m := &MsgCreateTriggerRequest{
		Authority: authority,
		Event:     eventAny,
	}
	m.Actions = actions

	return m
}

// NewDestroyTriggerRequest Creates a new trigger destroy request
func NewDestroyTriggerRequest(authority string, id TriggerID) *MsgDestroyTriggerRequest {
	msg := &MsgDestroyTriggerRequest{
		Authority: authority,
		Id:        id,
	}
	return msg
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCreateTriggerRequest) ValidateBasic() error {
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgCreateTriggerRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.GetAuthority())
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateTriggerRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if msg.Event != nil {
		var event TriggerEventI
		err := unpacker.UnpackAny(msg.Event, &event)
		if err != nil {
			return err
		}
	}
	return sdktx.UnpackInterfaces(unpacker, msg.Actions)
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDestroyTriggerRequest) ValidateBasic() error {
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgDestroyTriggerRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.GetAuthority())
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
