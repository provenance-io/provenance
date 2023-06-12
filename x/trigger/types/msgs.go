package types

import (
	fmt "fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

var _ sdk.Msg = &MsgCreateTriggerRequest{}
var _ sdk.Msg = &MsgDestroyTriggerRequest{}
var _ codectypes.UnpackInterfacesMessage = (*MsgCreateTriggerRequest)(nil)

// NewCreateTriggerRequest Creates a new trigger create request
func NewCreateTriggerRequest(authorities []string, event TriggerEventI, msgs []sdk.Msg) *MsgCreateTriggerRequest {
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
		Authorities: authorities,
		Event:       eventAny,
		Actions:     actions,
	}

	return m
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCreateTriggerRequest) ValidateBasic() error {
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
	actions, err := sdktx.GetMsgs(msg.Actions, "MsgCreateTriggerRequest - ValidateBasic")
	if err != nil {
		return err
	}

	authorities := make(map[string]bool)
	for _, authority := range msg.Authorities {
		var addr sdk.AccAddress
		addr, err = sdk.AccAddressFromBech32(authority)
		if err != nil {
			return fmt.Errorf("invalid address for trigger authority from address: %w", err)
		}
		authorities[string(addr)] = true
	}

	for idx, action := range actions {
		if err = action.ValidateBasic(); err != nil {
			return fmt.Errorf("action: %d: %w", idx, err)
		}
		if err = hasSigners(authorities, action.GetSigners()); err != nil {
			return fmt.Errorf("action: %d: %w", idx, err)
		}
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgCreateTriggerRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.GetAuthorities())
}

// hasSigners checks if the signers are all in the set of the entries
// The keys in the available map are a cast of an AccAddress to a string. It is not the result of AccAddress.String().
func hasSigners(available map[string]bool, signers []sdk.AccAddress) error {
	for i, signer := range signers {
		if !available[string(signer)] {
			return fmt.Errorf("signers[%d] %q is not a signer of the request message", i, signer.String())
		}
	}
	return nil
}

// stringsToAccAddresses converts an array of strings into an array of Acc Addresses.
// Panics if it can't convert one.
func stringsToAccAddresses(strings []string) []sdk.AccAddress {
	retval := make([]sdk.AccAddress, len(strings))

	for i, str := range strings {
		retval[i] = sdk.MustAccAddressFromBech32(str)
	}

	return retval
}

// accAddressesToStrings converts an array of sdk.AccAddress into an array of strings.
func accAddressesToStrings(addrs []sdk.AccAddress) []string {
	retval := make([]string, len(addrs))

	for i, addr := range addrs {
		retval[i] = addr.String()
	}

	return retval
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
		return nil, ErrNoTriggerEvent.Wrap("event is nil")
	}
	event, ok := msg.GetEvent().GetCachedValue().(TriggerEventI)
	if !ok {
		return nil, ErrNoTriggerEvent.Wrap("event is not a TriggerEventI")
	}

	return event, nil
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
