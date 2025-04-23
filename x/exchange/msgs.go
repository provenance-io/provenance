package exchange

import (
	"errors"
	"fmt"

	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgCreateAskRequest)(nil),
	(*MsgCreateBidRequest)(nil),
	(*MsgCommitFundsRequest)(nil),
	(*MsgCancelOrderRequest)(nil),
	(*MsgFillBidsRequest)(nil),
	(*MsgFillAsksRequest)(nil),
	(*MsgMarketSettleRequest)(nil),
	(*MsgMarketCommitmentSettleRequest)(nil),
	(*MsgMarketReleaseCommitmentsRequest)(nil),
	(*MsgMarketTransferCommitmentsRequest)(nil),
	(*MsgMarketSetOrderExternalIDRequest)(nil),
	(*MsgMarketWithdrawRequest)(nil),
	(*MsgMarketUpdateDetailsRequest)(nil),
	(*MsgMarketUpdateEnabledRequest)(nil),
	(*MsgMarketUpdateAcceptingOrdersRequest)(nil),
	(*MsgMarketUpdateUserSettleRequest)(nil),
	(*MsgMarketUpdateAcceptingCommitmentsRequest)(nil),
	(*MsgMarketUpdateIntermediaryDenomRequest)(nil),
	(*MsgMarketManagePermissionsRequest)(nil),
	(*MsgMarketManageReqAttrsRequest)(nil),
	(*MsgCreatePaymentRequest)(nil),
	(*MsgAcceptPaymentRequest)(nil),
	(*MsgRejectPaymentRequest)(nil),
	(*MsgRejectPaymentsRequest)(nil),
	(*MsgCancelPaymentsRequest)(nil),
	(*MsgChangePaymentTargetRequest)(nil),
	(*MsgGovCreateMarketRequest)(nil),
	(*MsgGovManageFeesRequest)(nil),
	(*MsgGovCloseMarketRequest)(nil),
	(*MsgGovUpdateParamsRequest)(nil),
	(*MsgUpdateParamsRequest)(nil),
}

// createPaymentGetSignersFunc returns a custom GetSigners function for a Msg that has a signer in a Payment.
// The Payment must be in a field named "payment", and must not be nullable.
// The provided getter will be used to get the string of the signer address,
// which will be decoded using the decoder in the provided signing.Options.
func createPaymentGetSignersFunc(options *signing.Options, fieldName string) signing.GetSignersFunc {
	return func(msgIn protov2.Message) (addrs [][]byte, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic (recovered) getting %s.payment.%s as a signer: %v", protov2.MessageName(msgIn), fieldName, r)
			}
		}()

		msg := msgIn.ProtoReflect()
		pmtDesc := msg.Descriptor().Fields().ByName("payment")
		if pmtDesc == nil {
			return nil, fmt.Errorf("no payment field found in %s", protov2.MessageName(msgIn))
		}

		pmt := msg.Get(pmtDesc).Message()
		fieldDesc := pmt.Descriptor().Fields().ByName(protoreflect.Name(fieldName))
		if fieldDesc == nil {
			return nil, fmt.Errorf("no payment.%s field found in %s", fieldName, protov2.MessageName(msgIn))
		}

		b32 := pmt.Get(fieldDesc).Interface().(string)
		addr, err := options.AddressCodec.StringToBytes(b32)
		if err != nil {
			return nil, fmt.Errorf("error decoding payment.%s address %q: %w", fieldName, b32, err)
		}
		return [][]byte{addr}, nil
	}
}

// DefineCustomGetSigners registers all the exchange module custom GetSigners functions with the provided signing options.
func DefineCustomGetSigners(options *signing.Options) {
	options.DefineCustomGetSigners(
		protov2.MessageName(protoadapt.MessageV2Of(&MsgCreatePaymentRequest{})),
		createPaymentGetSignersFunc(options, "source"))
	options.DefineCustomGetSigners(
		protov2.MessageName(protoadapt.MessageV2Of(&MsgAcceptPaymentRequest{})),
		createPaymentGetSignersFunc(options, "target"))
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

func (m MsgCommitFundsRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		errs = append(errs, fmt.Errorf("invalid account %q: %w", m.Account, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: cannot be zero"))
	}

	if m.Amount.IsZero() {
		errs = append(errs, fmt.Errorf("invalid amount %q: cannot be zero", m.Amount))
	} else if err := m.Amount.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid amount %q: %w", m.Amount, err))
	}

	if m.CreationFee != nil {
		if err := m.CreationFee.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid creation fee %q: %w", m.CreationFee, err))
		}
	}

	if err := ValidateEventTag(m.EventTag); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
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

func (m MsgMarketCommitmentSettleRequest) Validate(requireInputs bool) error {
	var errs []error

	if requireInputs {
		if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
			errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
		}
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	inputsOk := true
	if len(m.Inputs) == 0 {
		if requireInputs {
			errs = append(errs, errors.New("no inputs provided"))
			inputsOk = false
		}
	} else {
		for i, input := range m.Inputs {
			if err := input.Validate(); err != nil {
				errs = append(errs, fmt.Errorf("inputs[%d]: %w", i, err))
				inputsOk = false
			}
		}
	}

	outputsOk := true
	if len(m.Outputs) == 0 {
		if requireInputs {
			errs = append(errs, errors.New("no outputs provided"))
			outputsOk = false
		}
	} else {
		for i, output := range m.Outputs {
			if err := output.Validate(); err != nil {
				errs = append(errs, fmt.Errorf("outputs[%d]: %w", i, err))
				outputsOk = false
			}
		}
	}

	if inputsOk && outputsOk {
		inputTot := SumAccountAmounts(m.Inputs)
		outputTot := SumAccountAmounts(m.Outputs)
		if !inputTot.Equal(outputTot) {
			errs = append(errs, fmt.Errorf("input total %q does not equal output total %q", inputTot, outputTot))
		}
	}

	for i, fee := range m.Fees {
		if err := fee.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("fees[%d]: %w", i, err))
		}
	}

	for i, nav := range m.Navs {
		if err := nav.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("navs[%d]: %w", i, err))
		}
	}

	if err := ValidateEventTag(m.EventTag); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (m MsgMarketCommitmentSettleRequest) ValidateBasic() error {
	return m.Validate(true)
}

func (m MsgMarketReleaseCommitmentsRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, fmt.Errorf("invalid market id: cannot be zero"))
	}

	if len(m.ToRelease) == 0 {
		errs = append(errs, errors.New("nothing to release"))
	} else {
		for i, toRelease := range m.ToRelease {
			if err := toRelease.ValidateWithOptionalAmount(); err != nil {
				errs = append(errs, fmt.Errorf("to release[%d]: %w", i, err))
			}
		}
	}

	if err := ValidateEventTag(m.EventTag); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (m MsgMarketTransferCommitmentsRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.CurrentMarketId == 0 {
		errs = append(errs, errors.New("invalid current market id: cannot be zero"))
	}

	if m.NewMarketId == 0 {
		errs = append(errs, errors.New("invalid new market id: cannot be zero"))
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		errs = append(errs, fmt.Errorf("invalid to address %q: %w", m.Account, err))
	}

	if err := m.Amount.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid amount %q: %w", m.Amount, err))
	} else if m.Amount.IsZero() {
		errs = append(errs, fmt.Errorf("invalid amount %q: cannot be zero", m.Amount))
	}

	return errors.Join(errs...)
}

func (m MsgMarketSetOrderExternalIDRequest) ValidateBasic() error {
	var errs []error

	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}

	if m.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: cannot be zero"))
	}

	if err := ValidateExternalID(m.ExternalId); err != nil {
		errs = append(errs, err)
	}

	if m.OrderId == 0 {
		errs = append(errs, errors.New("invalid order id: cannot be zero"))
	}

	return errors.Join(errs...)
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

func (m MsgMarketUpdateEnabledRequest) ValidateBasic() error {
	return errors.New("the MarketUpdateEnabled endpoint has been replaced by the MarketUpdateAcceptingOrders endpoint")
}

func (m MsgMarketUpdateAcceptingOrdersRequest) ValidateBasic() error {
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

func (m MsgMarketUpdateAcceptingCommitmentsRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}
	if m.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: cannot be zero"))
	}
	return errors.Join(errs...)
}

func (m MsgMarketUpdateIntermediaryDenomRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		errs = append(errs, fmt.Errorf("invalid administrator %q: %w", m.Admin, err))
	}
	if m.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: cannot be zero"))
	}
	if err := ValidateIntermediaryDenom(m.IntermediaryDenom); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
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

		toRevokeByAddr := make(map[string]AccessGrant, len(m.ToRevoke))
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
			ValidateAddRemoveReqAttrs("create-commitment", m.CreateCommitmentToAdd, m.CreateCommitmentToRemove),
		)
	} else {
		errs = append(errs, errors.New("no updates"))
	}

	return errors.Join(errs...)
}

// HasUpdates returns true if this has at least one required attribute change, false if devoid of updates.
func (m MsgMarketManageReqAttrsRequest) HasUpdates() bool {
	return len(m.CreateAskToAdd) > 0 || len(m.CreateAskToRemove) > 0 ||
		len(m.CreateBidToAdd) > 0 || len(m.CreateBidToRemove) > 0 ||
		len(m.CreateCommitmentToAdd) > 0 || len(m.CreateCommitmentToRemove) > 0
}

func (m MsgCreatePaymentRequest) ValidateBasic() error {
	return m.Payment.Validate()
}

func (m MsgAcceptPaymentRequest) ValidateBasic() error {
	var errs []error
	if err := m.Payment.Validate(); err != nil {
		errs = append(errs, err)
	}
	if len(m.Payment.Target) == 0 {
		errs = append(errs, fmt.Errorf("invalid target: empty address string is not allowed"))
	}
	return errors.Join(errs...)
}

func (m MsgRejectPaymentRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Target); err != nil {
		errs = append(errs, fmt.Errorf("invalid target %q: %w", m.Target, err))
	}
	if _, err := sdk.AccAddressFromBech32(m.Source); err != nil {
		errs = append(errs, fmt.Errorf("invalid source %q: %w", m.Source, err))
	}
	if err := ValidateExternalID(m.ExternalId); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (m MsgRejectPaymentsRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Target); err != nil {
		errs = append(errs, fmt.Errorf("invalid target %q: %w", m.Target, err))
	}
	if len(m.Sources) == 0 {
		errs = append(errs, errors.New("at least one source is required"))
	}
	known := make(map[string]bool)
	bad := make(map[string]bool)
	for i, source := range m.Sources {
		if bad[source] {
			continue
		}
		if known[source] {
			errs = append(errs, fmt.Errorf("invalid sources: duplicate entry %s", source))
			bad[source] = true
			continue
		}
		known[source] = true
		if _, err := sdk.AccAddressFromBech32(source); err != nil {
			errs = append(errs, fmt.Errorf("invalid sources[%d] %q: %w", i, source, err))
			bad[source] = true
		}
	}
	return errors.Join(errs...)
}

func (m MsgCancelPaymentsRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Source); err != nil {
		errs = append(errs, fmt.Errorf("invalid source %q: %w", m.Source, err))
	}
	if len(m.ExternalIds) == 0 {
		errs = append(errs, errors.New("at least one external id is required"))
	}
	known := make(map[string]bool)
	bad := make(map[string]bool)
	for i, externalID := range m.ExternalIds {
		if bad[externalID] {
			continue
		}
		if known[externalID] {
			errs = append(errs, fmt.Errorf("invalid external ids: duplicate entry %q", externalID))
			bad[externalID] = true
			continue
		}
		known[externalID] = true
		if err := ValidateExternalID(externalID); err != nil {
			errs = append(errs, fmt.Errorf("invalid external ids[%d]: %w", i, err))
			bad[externalID] = true
		}
	}
	return errors.Join(errs...)
}

func (m MsgChangePaymentTargetRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Source); err != nil {
		errs = append(errs, fmt.Errorf("invalid source %q: %w", m.Source, err))
	}
	if err := ValidateExternalID(m.ExternalId); err != nil {
		errs = append(errs, err)
	}
	if len(m.NewTarget) > 0 {
		if _, err := sdk.AccAddressFromBech32(m.NewTarget); err != nil {
			errs = append(errs, fmt.Errorf("invalid new target %q: %w", m.NewTarget, err))
		}
	}
	return errors.Join(errs...)
}

func (m MsgGovCreateMarketRequest) ValidateBasic() error {
	errs := make([]error, 0, 2)
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		errs = append(errs, fmt.Errorf("invalid authority: %w", err))
	}
	errs = append(errs, m.Market.Validate())
	return errors.Join(errs...)
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
			ValidateAddRemoveFeeOptions("create-commitment flat fee", m.AddFeeCreateCommitmentFlat, m.RemoveFeeCreateCommitmentFlat),
			ValidateAddRemoveFeeOptions("seller settlement flat fee", m.AddFeeSellerSettlementFlat, m.RemoveFeeSellerSettlementFlat),
			ValidateSellerFeeRatios(m.AddFeeSellerSettlementRatios),
			ValidateDisjointFeeRatios("seller settlement fee", m.AddFeeSellerSettlementRatios, m.RemoveFeeSellerSettlementRatios),
			ValidateAddRemoveFeeOptions("buyer settlement flat fee", m.AddFeeBuyerSettlementFlat, m.RemoveFeeBuyerSettlementFlat),
			ValidateBuyerFeeRatios(m.AddFeeBuyerSettlementRatios),
			ValidateDisjointFeeRatios("buyer settlement fee", m.AddFeeBuyerSettlementRatios, m.RemoveFeeBuyerSettlementRatios),
			ValidateBips("commitment settlement", m.SetFeeCommitmentSettlementBips),
		)

		if m.UnsetFeeCommitmentSettlementBips && m.SetFeeCommitmentSettlementBips > 0 {
			errs = append(errs, fmt.Errorf(
				"invalid commitment settlement bips %d: must be zero when unset_fee_commitment_settlement_bips is true",
				m.SetFeeCommitmentSettlementBips))
		}
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
		len(m.AddFeeBuyerSettlementRatios) > 0 || len(m.RemoveFeeBuyerSettlementRatios) > 0 ||
		len(m.AddFeeCreateCommitmentFlat) > 0 || len(m.RemoveFeeCreateCommitmentFlat) > 0 ||
		m.SetFeeCommitmentSettlementBips != 0 || m.UnsetFeeCommitmentSettlementBips
}

func (m MsgGovCloseMarketRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		errs = append(errs, fmt.Errorf("invalid authority %q: %w", m.Authority, err))
	}
	if m.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: cannot be zero"))
	}
	return errors.Join(errs...)
}

func (m MsgGovUpdateParamsRequest) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (m MsgUpdateParamsRequest) ValidateBasic() error {
	errs := make([]error, 0, 2)
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		errs = append(errs, fmt.Errorf("invalid authority: %w", err))
	}
	errs = append(errs, m.Params.Validate())
	return errors.Join(errs...)
}
