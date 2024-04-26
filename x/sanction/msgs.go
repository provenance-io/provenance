package sanction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/sanction/errors"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgSanction)(nil),
	(*MsgUnsanction)(nil),
	(*MsgUpdateParams)(nil),
}

func NewMsgSanction(authority string, addrs ...sdk.AccAddress) *MsgSanction {
	rv := &MsgSanction{
		Authority: authority,
	}
	for _, addr := range addrs {
		rv.Addresses = append(rv.Addresses, addr.String())
	}
	return rv
}

func (m MsgSanction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("authority, %q: %v", m.Authority, err)
	}
	for i, addr := range m.Addresses {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("addresses[%d], %q: %v", i, addr, err)
		}
	}
	return nil
}

func NewMsgUnsanction(authority string, addrs ...sdk.AccAddress) *MsgUnsanction {
	rv := &MsgUnsanction{
		Authority: authority,
	}
	for _, addr := range addrs {
		rv.Addresses = append(rv.Addresses, addr.String())
	}
	return rv
}

func (m MsgUnsanction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("authority, %q: %v", m.Authority, err)
	}
	for i, addr := range m.Addresses {
		_, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("addresses[%d], %q: %v", i, addr, err)
		}
	}
	return nil
}

func NewMsgUpdateParams(authority string, minDepSanction, minDepUnsanction sdk.Coins) *MsgUpdateParams {
	rv := &MsgUpdateParams{
		Authority: authority,
		Params: &Params{
			ImmediateSanctionMinDeposit:   minDepSanction,
			ImmediateUnsanctionMinDeposit: minDepUnsanction,
		},
	}
	return rv
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("authority, %q: %v", m.Authority, err)
	}
	if m.Params != nil {
		err = m.Params.ValidateBasic()
		if err != nil {
			return errors.ErrInvalidParams.Wrap(err.Error())
		}
	}
	return nil
}
