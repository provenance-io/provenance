package types

import (
	fmt "fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

var _ sdk.Msg = &MsgCreateTriggerRequest{}

// GetSigners indicates that the message must have been signed by the parent.
func NewCreateTriggerRequest(authority string, event Event, msgs []sdk.Msg) *MsgCreateTriggerRequest {
	actions, err := sdktx.SetMsgs(msgs)
	if err != nil || len(actions) == 0 {
		fmt.Printf("unable to set messages : %s\n", err)
	}
	m := &MsgCreateTriggerRequest{
		Authority: authority,
		Event:     event,
	}
	m.Action = actions

	return m
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

// GetMsgs unpacks m.Messages Any's into sdk.Msg's
func (m *MsgCreateTriggerRequest) GetMsgs() ([]sdk.Msg, error) {
	return sdktx.GetMsgs(m.Action, "sdk.MsgCreateTriggerRequest")
}

// SetMsgs packs sdk.Msg's into m.Messages Any's
// NOTE: this will overwrite any existing messages
func (m *MsgCreateTriggerRequest) SetMsgs(msgs []sdk.Msg) error {
	anys, err := sdktx.SetMsgs(msgs)
	if err != nil {
		return err
	}

	m.Action = anys
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m Trigger) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, m.Action)
}

// GetMsgs unpacks m.Messages Any's into sdk.Msg's
func (m *Trigger) GetMsgs() ([]sdk.Msg, error) {
	return sdktx.GetMsgs(m.Action, "sdk.MsgCreateTriggerRequest")
}

// SetMsgs packs sdk.Msg's into m.Messages Any's
// NOTE: this will overwrite any existing messages
func (m *Trigger) SetMsgs(msgs []sdk.Msg) error {
	anys, err := sdktx.SetMsgs(msgs)
	if err != nil {
		return err
	}

	m.Action = anys
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateTriggerRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, m.Action)
}

// GetMsgs unpacks m.Messages Any's into sdk.Msg's
func (m *QueuedTrigger) GetMsgs() ([]sdk.Msg, error) {
	return sdktx.GetMsgs(m.Trigger.Action, "sdk.MsgCreateTriggerRequest")
}

// SetMsgs packs sdk.Msg's into m.Messages Any's
// NOTE: this will overwrite any existing messages
func (m *QueuedTrigger) SetMsgs(msgs []sdk.Msg) error {
	anys, err := sdktx.SetMsgs(msgs)
	if err != nil {
		return err
	}

	m.Trigger.Action = anys
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m QueuedTrigger) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, m.Trigger.Action)
}
