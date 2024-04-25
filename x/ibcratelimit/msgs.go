package ibcratelimit

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgGovUpdateParamsRequest)(nil),
}

// NewMsgGovUpdateParamsRequest creates a new GovUpdateParams message.
func NewMsgGovUpdateParamsRequest(authority, ratelimiter string) *MsgGovUpdateParamsRequest {
	return &MsgGovUpdateParamsRequest{
		Authority: authority,
		Params:    NewParams(ratelimiter),
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (m MsgGovUpdateParamsRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	return m.Params.Validate()
}

// GetSigners indicates that the message must have been signed by the address provided.
func (m MsgGovUpdateParamsRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
