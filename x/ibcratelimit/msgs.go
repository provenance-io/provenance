package ibcratelimit

import (
	"errors"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgGovUpdateParamsRequest)(nil),
	(*MsgUpdateParamsRequest)(nil),
}

// ValidateBasic runs stateless validation checks on the message.
func (m MsgGovUpdateParamsRequest) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

// NewUpdateParamsRequest creates a new GovUpdateParams message.
func NewUpdateParamsRequest(authority, ratelimiter string) *MsgUpdateParamsRequest {
	return &MsgUpdateParamsRequest{
		Authority: authority,
		Params:    NewParams(ratelimiter),
	}
}

func (m MsgUpdateParamsRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	return m.Params.Validate()
}
