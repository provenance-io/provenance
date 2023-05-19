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
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid address for trigger authority from address: %w", err)
	}
	if len(msg.Actions) == 0 {
		return fmt.Errorf("trigger must contain actions")
	}
	event, err := msg.GetTriggerEventI()
	if err != nil {
		return err
	}
	if err = event.Validate(); err != nil {
		return err
	}
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

// GetTriggerEventI returns unpacked TriggerEvent
func (msg MsgCreateTriggerRequest) GetTriggerEventI() (TriggerEventI, error) {
	event, ok := msg.GetEvent().GetCachedValue().(TriggerEventI)
	if !ok {
		return nil, ErrNoTriggerEvent.Wrap("failed to get event")
	}

	return event, nil
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgDestroyTriggerRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid address for trigger authority from address: %w", err)
	}
	if msg.Id == 0 {
		return fmt.Errorf("invalid id for trigger")
	}
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
