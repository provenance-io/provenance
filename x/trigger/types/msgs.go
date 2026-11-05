package types

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgCreateTriggerRequest)(nil),
	(*MsgDestroyTriggerRequest)(nil),
}

var _ codectypes.UnpackInterfacesMessage = (*MsgCreateTriggerRequest)(nil)

var errMsgRemoved = fmt.Errorf("the trigger module has been removed")

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgCreateTriggerRequest) ValidateBasic() error {
	return errMsgRemoved
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateTriggerRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if msg.Event != nil {
		var event TriggerEventI
		if err := unpacker.UnpackAny(msg.Event, &event); err != nil {
			return err
		}
	}
	return sdktx.UnpackInterfaces(unpacker, msg.Actions)
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgDestroyTriggerRequest) ValidateBasic() error {
	return errMsgRemoved
}
