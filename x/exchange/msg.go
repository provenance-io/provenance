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
	if err := m.AskOrder.Validate(); err != nil {
		return err
	}
	if m.OrderCreationFee != nil {
		if err := m.OrderCreationFee.Validate(); err != nil {
			return fmt.Errorf("invalid order creation fee: %w", err)
		}
	}
	return nil
}

func (m MsgCreateAskRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.AskOrder.Seller)
	return []sdk.AccAddress{addr}
}

func (m MsgCreateBidRequest) ValidateBasic() error {
	if err := m.BidOrder.Validate(); err != nil {
		return err
	}
	if m.OrderCreationFee != nil {
		if err := m.OrderCreationFee.Validate(); err != nil {
			return fmt.Errorf("invalid order creation fee: %w", err)
		}
	}
	return nil
}

func (m MsgCreateBidRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.BidOrder.Buyer)
	return []sdk.AccAddress{addr}
}

func (m MsgCancelOrderRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return fmt.Errorf("invalid signer: %w", err)
	}
	if m.OrderId == 0 {
		return fmt.Errorf("invalid order id: cannot be zero")
	}
	return nil
}

func (m MsgCancelOrderRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Signer)
	return []sdk.AccAddress{addr}
}

func (m MsgFillBidsRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Seller); err != nil {
		errs = append(errs, fmt.Errorf("invalid seller: %w", err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	if err := m.TotalAssets.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid total assets: %w", err))
	} else if m.TotalAssets.IsZero() {
		errs = append(errs, fmt.Errorf("invalid total assets: cannot be zero"))
	}

	if err := ValidateOrderIDs("bid", m.BidOrderIds); err != nil {
		errs = append(errs, err)
	}

	if m.SellerSettlementFlatFee != nil {
		if err := m.SellerSettlementFlatFee.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid seller settlement flat fee: %w", err))
		} else if m.SellerSettlementFlatFee.IsZero() {
			errs = append(errs, fmt.Errorf("invalid seller settlement flat fee: %s amount cannot be zero", m.SellerSettlementFlatFee.Denom))
		}
	}

	if m.AskOrderCreationFee != nil {
		if err := m.AskOrderCreationFee.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid ask order creation fee: %w", err))
		} else if m.AskOrderCreationFee.IsZero() {
			errs = append(errs, fmt.Errorf("invalid ask order creation fee: %s amount cannot be zero", m.AskOrderCreationFee.Denom))
		}
	}

	return errors.Join(errs...)
}

func (m MsgFillBidsRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Seller)
	return []sdk.AccAddress{addr}
}

func (m MsgFillAsksRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		errs = append(errs, fmt.Errorf("invalid buyer: %w", err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	if err := m.TotalPrice.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid total price: %w", err))
	} else if m.TotalPrice.IsZero() {
		errs = append(errs, fmt.Errorf("invalid total price: cannot be zero"))
	}

	if err := ValidateOrderIDs("ask", m.AskOrderIds); err != nil {
		errs = append(errs, err)
	}

	if len(m.BuyerSettlementFees) > 0 {
		if err := m.BuyerSettlementFees.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid buyer settlement fees: %w", err))
		}
	}

	if m.BidOrderCreationFee != nil {
		if err := m.BidOrderCreationFee.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid bid order creation fee: %w", err))
		} else if m.BidOrderCreationFee.IsZero() {
			errs = append(errs, fmt.Errorf("invalid bid order creation fee: %s amount cannot be zero", m.BidOrderCreationFee.Denom))
		}
	}

	return errors.Join(errs...)
}

func (m MsgFillAsksRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Buyer)
	return []sdk.AccAddress{addr}
}

func (m MsgMarketSettleRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	if err := ValidateOrderIDs("ask", m.AskOrderIds); err != nil {
		errs = append(errs, err)
	}

	if err := ValidateOrderIDs("bid", m.BidOrderIds); err != nil {
		errs = append(errs, err)
	}

	inBoth := IntersectionUint64(m.AskOrderIds, m.BidOrderIds)
	if len(inBoth) > 0 {
		errs = append(errs, fmt.Errorf("order ids duplicated as both bid and ask: %v", inBoth))
	}

	// Nothing to validate now for the ExpectPartial flag.

	return errors.Join(errs...)
}

func (m MsgMarketSettleRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{addr}
}

func (m MsgMarketWithdrawRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: cannot be zero"))
	}

	if _, err := sdk.AccAddressFromBech32(m.ToAddress); err != nil {
		errs = append(errs, fmt.Errorf("invalid to address %q: %w", m.ToAddress, err))
	}

	if err := m.Amount.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid amount %q: %w", m.Amount, err))
	} else if m.Amount.IsZero() {
		errs = append(errs, fmt.Errorf("invalid amount %q: cannot be zero", m.Amount))
	}

	return errors.Join(errs...)
}

func (m MsgMarketWithdrawRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{addr}
}

func (m MsgMarketUpdateDetailsRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	if err := m.MarketDetails.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (m MsgMarketUpdateDetailsRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{addr}
}

func (m MsgMarketUpdateEnabledRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	// Nothing to validate for the AcceptingOrders field.

	return errors.Join(errs...)
}

func (m MsgMarketUpdateEnabledRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{addr}
}

func (m MsgMarketUpdateUserSettleRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	// Nothing to validate for the AllowUserSettlement field.

	return errors.Join(errs...)
}

func (m MsgMarketUpdateUserSettleRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{addr}
}

func (m MsgMarketManagePermissionsRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	if m.HasUpdates() {
		for _, addrStr := range m.RevokeAll {
			if _, err := sdk.AccAddressFromBech32(addrStr); err != nil {
				errs = append(errs, fmt.Errorf("invalid revoke-all address %q: %w", addrStr, err))
			}
		}

		if err := ValidateAccessGrantsField("to-revoke", m.ToRevoke); err != nil {
			errs = append(errs, err)
		}

		toRevokeByAddr := make(map[string]AccessGrant)
		for _, ag := range m.ToRevoke {
			if ContainsString(m.RevokeAll, ag.Address) {
				errs = append(errs, fmt.Errorf("address %s appears in both the revoke-all and to-revoke fields", ag.Address))
			}
			toRevokeByAddr[ag.Address] = ag
		}

		if err := ValidateAccessGrantsField("to-grant", m.ToGrant); err != nil {
			errs = append(errs, err)
		}

		for _, ag := range m.ToGrant {
			toRev, ok := toRevokeByAddr[ag.Address]
			if ok {
				for _, perm := range ag.Permissions {
					if toRev.Contains(perm) {
						errs = append(errs, fmt.Errorf("address %s has both revoke and grant %q", ag.Address, perm.SimpleString()))
					}
				}
			}
		}
	} else {
		errs = append(errs, errors.New("no updates"))
	}

	return errors.Join(errs...)
}

// HasUpdates returns true if this has at least one permission change, false if devoid of updates.
func (m MsgMarketManagePermissionsRequest) HasUpdates() bool {
	return len(m.RevokeAll) > 0 || len(m.ToRevoke) > 0 || len(m.ToGrant) > 0
}

func (m MsgMarketManagePermissionsRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{addr}
}

func (m MsgMarketManageReqAttrsRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	if m.HasUpdates() {
		errs = append(errs,
			ValidateAddRemoveReqAttrs("create-ask", m.CreateAskToAdd, m.CreateAskToRemove),
			ValidateAddRemoveReqAttrs("create-bid", m.CreateBidToAdd, m.CreateBidToRemove),
		)
	} else {
		errs = append(errs, errors.New("no updates"))
	}

	return errors.Join(errs...)
}

// HasUpdates returns true if this has at least one required attribute change, false if devoid of updates.
func (m MsgMarketManageReqAttrsRequest) HasUpdates() bool {
	return len(m.CreateAskToAdd) > 0 || len(m.CreateAskToRemove) > 0 ||
		len(m.CreateBidToAdd) > 0 || len(m.CreateBidToRemove) > 0
}

func (m MsgMarketManageReqAttrsRequest) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Admin)
	return []sdk.AccAddress{addr}
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

	if m.MarketId == 0 {
		errs = append(errs, errors.New("market id cannot be zero"))
	}

	if m.HasUpdates() {
		errs = append(errs,
			ValidateAddRemoveFeeOptions("create-ask flat fee", m.AddFeeCreateAskFlat, m.RemoveFeeCreateAskFlat),
			ValidateAddRemoveFeeOptions("create-bid flat fee", m.AddFeeCreateBidFlat, m.RemoveFeeCreateBidFlat),
			ValidateAddRemoveFeeOptions("seller settlement flat fee", m.AddFeeSellerSettlementFlat, m.RemoveFeeSellerSettlementFlat),
			ValidateSellerFeeRatios(m.AddFeeSellerSettlementRatios),
			ValidateDisjointFeeRatios("seller settlement fee", m.AddFeeSellerSettlementRatios, m.RemoveFeeSellerSettlementRatios),
			ValidateAddRemoveFeeOptions("buyer settlement flat fee", m.AddFeeBuyerSettlementFlat, m.RemoveFeeBuyerSettlementFlat),
			ValidateBuyerFeeRatios(m.AddFeeBuyerSettlementRatios),
			ValidateDisjointFeeRatios("buyer settlement fee", m.AddFeeBuyerSettlementRatios, m.RemoveFeeBuyerSettlementRatios),
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
		len(m.AddFeeSellerSettlementFlat) > 0 || len(m.RemoveFeeSellerSettlementFlat) > 0 ||
		len(m.AddFeeSellerSettlementRatios) > 0 || len(m.RemoveFeeSellerSettlementRatios) > 0 ||
		len(m.AddFeeBuyerSettlementFlat) > 0 || len(m.RemoveFeeBuyerSettlementFlat) > 0 ||
		len(m.AddFeeBuyerSettlementRatios) > 0 || len(m.RemoveFeeBuyerSettlementRatios) > 0
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
