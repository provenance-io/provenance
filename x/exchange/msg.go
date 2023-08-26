package exchange

import sdk "github.com/cosmos/cosmos-sdk/types"

var allRequestMsgs = []sdk.Msg{
	(*MsgCreateAskRequest)(nil),
	(*MsgCreateBidRequest)(nil),
	(*MsgCancelOrderRequest)(nil),
	(*MsgFillBidsRequest)(nil),
	(*MsgFillAsksRequest)(nil),
	(*MsgMarketSettleRequest)(nil),
	(*MsgMarketWithdrawRequest)(nil),
	(*MsgMarketUpdateDetailsRequest)(nil),
	(*MsgMarketUpdateEnabledRequest)(nil),
	(*MsgMarketUpdateUserSettleRequest)(nil),
	(*MsgMarketManagePermissionsRequest)(nil),
	(*MsgMarketManageReqAttrsRequest)(nil),
	(*MsgCreateMarketRequest)(nil),
	(*MsgGovCreateMarketRequest)(nil),
	(*MsgGovManageFeesRequest)(nil),
	(*MsgGovUpdateParamsRequest)(nil),
}

func (m MsgCreateAskRequest) ValidateBasic() error {
	// TODO[1658]: MsgCreateAskRequest.ValidateBasic()
	return nil
}

func (m MsgCreateAskRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgCreateAskRequest.GetSigners
	panic("not implemented")
}

func (m MsgCreateBidRequest) ValidateBasic() error {
	// TODO[1658]: MsgCreateBidRequest.ValidateBasic()
	return nil
}

func (m MsgCreateBidRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgCreateBidRequest.GetSigners
	panic("not implemented")
}

func (m MsgCancelOrderRequest) ValidateBasic() error {
	// TODO[1658]: MsgCancelOrderRequest.ValidateBasic()
	return nil
}

func (m MsgCancelOrderRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgCancelOrderRequest.GetSigners
	panic("not implemented")
}

func (m MsgFillBidsRequest) ValidateBasic() error {
	// TODO[1658]: MsgFillBidsRequest.ValidateBasic()
	return nil
}

func (m MsgFillBidsRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgFillBidsRequest.GetSigners
	panic("not implemented")
}

func (m MsgFillAsksRequest) ValidateBasic() error {
	// TODO[1658]: MsgFillAsksRequest.ValidateBasic()
	return nil
}

func (m MsgFillAsksRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgFillAsksRequest.GetSigners
	panic("not implemented")
}

func (m MsgMarketSettleRequest) ValidateBasic() error {
	// TODO[1658]: MsgMarketSettleRequest.ValidateBasic()
	return nil
}

func (m MsgMarketSettleRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgMarketSettleRequest.GetSigners
	panic("not implemented")
}

func (m MsgMarketWithdrawRequest) ValidateBasic() error {
	// TODO[1658]: MsgMarketWithdrawRequest.ValidateBasic()
	return nil
}

func (m MsgMarketWithdrawRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgMarketWithdrawRequest.GetSigners
	panic("not implemented")
}

func (m MsgMarketUpdateDetailsRequest) ValidateBasic() error {
	// TODO[1658]: MsgMarketUpdateDetailsRequest.ValidateBasic()
	return nil
}

func (m MsgMarketUpdateDetailsRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgMarketUpdateDetailsRequest.GetSigners
	panic("not implemented")
}

func (m MsgMarketUpdateEnabledRequest) ValidateBasic() error {
	// TODO[1658]: MsgMarketUpdateEnabledRequest.ValidateBasic()
	return nil
}

func (m MsgMarketUpdateEnabledRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgMarketUpdateEnabledRequest.GetSigners
	panic("not implemented")
}

func (m MsgMarketUpdateUserSettleRequest) ValidateBasic() error {
	// TODO[1658]: MsgMarketUpdateUserSettleRequest.ValidateBasic()
	return nil
}

func (m MsgMarketUpdateUserSettleRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgMarketUpdateUserSettleRequest.GetSigners
	panic("not implemented")
}

func (m MsgMarketManagePermissionsRequest) ValidateBasic() error {
	// TODO[1658]: MsgMarketManagePermissionsRequest.ValidateBasic()
	return nil
}

func (m MsgMarketManagePermissionsRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgMarketManagePermissionsRequest.GetSigners
	panic("not implemented")
}

func (m MsgMarketManageReqAttrsRequest) ValidateBasic() error {
	// TODO[1658]: MsgMarketManageReqAttrsRequest.ValidateBasic()
	return nil
}

func (m MsgMarketManageReqAttrsRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgMarketManageReqAttrsRequest.GetSigners
	panic("not implemented")
}

func (m MsgCreateMarketRequest) ValidateBasic() error {
	// TODO[1658]: MsgCreateMarketRequest.ValidateBasic()
	return nil
}

func (m MsgCreateMarketRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgCreateMarketRequest.GetSigners
	panic("not implemented")
}

func (m MsgGovCreateMarketRequest) ValidateBasic() error {
	// TODO[1658]: MsgGovCreateMarketRequest.ValidateBasic()
	return nil
}

func (m MsgGovCreateMarketRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgGovCreateMarketRequest.GetSigners
	panic("not implemented")
}

func (m MsgGovManageFeesRequest) ValidateBasic() error {
	// TODO[1658]: MsgGovManageFeesRequest.ValidateBasic()
	return nil
}

func (m MsgGovManageFeesRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgGovManageFeesRequest.GetSigners
	panic("not implemented")
}

func (m MsgGovUpdateParamsRequest) ValidateBasic() error {
	// TODO[1658]: MsgGovUpdateParamsRequest.ValidateBasic()
	return nil
}

func (m MsgGovUpdateParamsRequest) GetSigners() []sdk.AccAddress {
	// TODO[1658]: MsgGovUpdateParamsRequest.GetSigners
	panic("not implemented")
}
