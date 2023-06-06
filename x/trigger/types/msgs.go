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
	if err != nil {
		fmt.Printf("unable to set messages: %s\n", err)
		return nil
	}

	eventAny, err := codectypes.NewAnyWithValue(event)
	if err != nil {
		fmt.Printf("unable to set event: %s\n", err)
		return nil
	}

	m := &MsgCreateTriggerRequest{
		Authority: authority,
		Event:     eventAny,
		Actions:   actions,
	}

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
	msgs, err := sdktx.GetMsgs(msg.Actions, "MsgCreateTriggerRequest - ValidateBasic")
	if err != nil {
		return err
	}

	for idx, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return fmt.Errorf("msg: %d, err: %w", idx, err)
		}
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgCreateTriggerRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.GetAuthority())}
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
	if msg.GetEvent() == nil {
		return nil, ErrNoTriggerEvent.Wrap("failed to get event")
	}
	event, ok := msg.GetEvent().GetCachedValue().(TriggerEventI)
	if !ok {
		return nil, ErrNoTriggerEvent.Wrap("event is nil")
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
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.GetAuthority())}
}
