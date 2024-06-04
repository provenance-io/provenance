package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	simappparams "github.com/provenance-io/provenance/app/params"
	internalsdk "github.com/provenance-io/provenance/internal/sdk"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgCreateTriggerRequest)(nil),
	(*MsgDestroyTriggerRequest)(nil),
}

var _ codectypes.UnpackInterfacesMessage = (*MsgCreateTriggerRequest)(nil)

// NewCreateTriggerRequest Creates a new trigger create request
func NewCreateTriggerRequest(authorities []string, event TriggerEventI, msgs []sdk.Msg) (*MsgCreateTriggerRequest, error) {
	actions, err := sdktx.SetMsgs(msgs)
	if err != nil {
		return nil, fmt.Errorf("unable to set messages: %w", err)
	}

	eventAny, err := codectypes.NewAnyWithValue(event)
	if err != nil {
		return nil, fmt.Errorf("unable to set event: %w", err)
	}

	m := &MsgCreateTriggerRequest{
		Authorities: authorities,
		Event:       eventAny,
		Actions:     actions,
	}

	return m, nil
}

func MustNewCreateTriggerRequest(authorities []string, event TriggerEventI, msgs []sdk.Msg) *MsgCreateTriggerRequest {
	m, err := NewCreateTriggerRequest(authorities, event, msgs)
	if err != nil {
		panic(err)
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

	cdc := simappparams.AppEncodingConfig.Marshaler
	for idx, action := range actions {
		if err = internalsdk.ValidateBasic(action); err != nil {
			return fmt.Errorf("action: %d: %w", idx, err)
		}
		if err = hasSigners(cdc, authorities, action); err != nil {
			return fmt.Errorf("action: %d: %w", idx, err)
		}
	}
	return nil
}

// hasSigners checks if the signers are all in the set of the entries
// The keys in the available map are a cast of an AccAddress to a string. It is not the result of AccAddress.String().
func hasSigners(codec codec.Codec, available map[string]bool, action sdk.Msg) error {
	signers, _, err := codec.GetMsgV1Signers(action)
	if err != nil {
		return fmt.Errorf("could not get signers of %T: %w", action, err)
	}
	for i, signer := range signers {
		if !available[string(signer)] {
			return fmt.Errorf("%T signers[%d] %q is not a signer of the request message", action, i, sdk.AccAddress(signer).String())
		}
	}
	return nil
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
