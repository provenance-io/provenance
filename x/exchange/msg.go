package exchange

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func (m MsgGovCreateMarketRequest) ValidateBasic() error {
	errs := make([]error, 0, 2)
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		errs = append(errs, fmt.Errorf("invalid authority: %w", err))
	}
	errs = append(errs, m.Market.Validate())
	return errors.Join(errs...)
}

func (m MsgGovCreateMarketRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

func (m MsgGovManageFeesRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		errs = append(errs, fmt.Errorf("invalid authority: %w", err))
	}

	if m.HasUpdates() {
		errs = append(errs,
			ValidateAddRemoveFeeOptions("create-ask flat fee", m.AddFeeCreateAskFlat, m.RemoveFeeCreateAskFlat),
			ValidateAddRemoveFeeOptions("create-bid flat fee", m.AddFeeCreateBidFlat, m.RemoveFeeCreateBidFlat),
			ValidateAddRemoveFeeOptions("seller settlement flat fee", m.AddFeeSettlementSellerFlat, m.RemoveFeeSettlementSellerFlat),
			ValidateSellerFeeRatios(m.AddFeeSettlementSellerRatios),
			ValidateDisjointFeeRatios("seller settlement fee", m.AddFeeSettlementSellerRatios, m.RemoveFeeSettlementSellerRatios),
			ValidateAddRemoveFeeOptions("buyer settlement flat fee", m.AddFeeSettlementBuyerFlat, m.RemoveFeeSettlementBuyerFlat),
			ValidateBuyerFeeRatios(m.AddFeeSettlementBuyerRatios),
			ValidateDisjointFeeRatios("buyer settlement fee", m.AddFeeSettlementBuyerRatios, m.RemoveFeeSettlementBuyerRatios),
		)
	} else {
		errs = append(errs, errors.New("no updates"))
	}

	return errors.Join(errs...)
}

// HasUpdates returns true if this has at least one fee change, false if devoid of updates.
func (m MsgGovManageFeesRequest) HasUpdates() bool {
	return len(m.AddFeeCreateAskFlat) > 0 || len(m.RemoveFeeCreateAskFlat) > 0 ||
		len(m.AddFeeCreateBidFlat) > 0 || len(m.RemoveFeeCreateBidFlat) > 0 ||
		len(m.AddFeeSettlementSellerFlat) > 0 || len(m.RemoveFeeSettlementSellerFlat) > 0 ||
		len(m.AddFeeSettlementSellerRatios) > 0 || len(m.RemoveFeeSettlementSellerRatios) > 0 ||
		len(m.AddFeeSettlementBuyerFlat) > 0 || len(m.RemoveFeeSettlementBuyerFlat) > 0 ||
		len(m.AddFeeSettlementBuyerRatios) > 0 || len(m.RemoveFeeSettlementBuyerRatios) > 0
}

func (m MsgGovManageFeesRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

func (m MsgGovUpdateParamsRequest) ValidateBasic() error {
	errs := make([]error, 0, 2)
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		errs = append(errs, fmt.Errorf("invalid authority: %w", err))
	}
	errs = append(errs, m.Params.Validate())
	return errors.Join(errs...)
}

func (m MsgGovUpdateParamsRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
