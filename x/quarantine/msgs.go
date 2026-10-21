package quarantine

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgOptIn)(nil),
	(*MsgOptOut)(nil),
	(*MsgAccept)(nil),
	(*MsgDecline)(nil),
	(*MsgUpdateAutoResponses)(nil),
}

var errMsgRemoved = fmt.Errorf("the quarantine module has been removed")

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgOptIn) ValidateBasic() error {
	return errMsgRemoved
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgOptOut) ValidateBasic() error {
	return errMsgRemoved
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgAccept) ValidateBasic() error {
	return errMsgRemoved
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgDecline) ValidateBasic() error {
	return errMsgRemoved
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgUpdateAutoResponses) ValidateBasic() error {
	return errMsgRemoved
}
