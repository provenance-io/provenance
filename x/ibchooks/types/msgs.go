package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgEmitIBCAck)(nil),
}

func (m MsgEmitIBCAck) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	return err
}

// NewMsgUpdateParamsRequest creates a new MsgUpdateParamsRequest instance
func NewMsgUpdateParamsRequest(allowedAsyncAckContracts []string, authority string) *MsgUpdateParamsRequest {
	return &MsgUpdateParamsRequest{
		Params: Params{
			AllowedAsyncAckContracts: allowedAsyncAckContracts,
		},
		Authority: authority,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUpdateParamsRequest) ValidateBasic() error {
	for _, contract := range msg.Params.AllowedAsyncAckContracts {
		if _, err := sdk.AccAddressFromBech32(contract); err != nil {
			return fmt.Errorf("invalid contract address: %s", contract)
		}
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority address: %s", msg.Authority)
	}

	return nil
}
