package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgCreateRequest)(nil),
	(*MsgUpdateStatusRequest)(nil),
	(*MsgUpdateInterestRateRequest)(nil),
	(*MsgUpdatePaymentRequest)(nil),
	(*MsgUpdateMaturityDateRequest)(nil),
	(*MsgAppendRequest)(nil),
	(*MsgUpdateBalancesRequest)(nil),
	(*MsgFundAssetRequest)(nil),
	(*MsgFundAssetByRegistryRequest)(nil),
	(*MsgTransferFundsWithSettlementRequest)(nil),
	(*MsgDestroyRequest)(nil),
	(*MsgCreateLedgerClassRequest)(nil),
	(*MsgAddLedgerClassStatusTypeRequest)(nil),
	(*MsgAddLedgerClassEntryTypeRequest)(nil),
	(*MsgAddLedgerClassBucketTypeRequest)(nil),
}

// Note: Authority address validation is performed in the message server to avoid duplicate bech32 conversions.

// ValidateBasic implements the sdk.Msg interface for MsgCreateRequest
func (m *MsgCreateRequest) ValidateBasic() error {
	if err := ValidateLedgerBasic(m.Ledger); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateStatusRequest
func (m *MsgUpdateStatusRequest) ValidateBasic() error {
	if err := ValidateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.StatusTypeId == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("status type id cannot be zero")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateInterestRateRequest
func (m *MsgUpdateInterestRateRequest) ValidateBasic() error {
	if err := ValidateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.InterestRate < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("interest rate cannot be negative")
	}
	if m.InterestDayCountConvention == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("interest day count convention cannot be zero (0)")
	}
	if m.InterestAccrualMethod == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("interest accrual method cannot be zero")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdatePaymentRequest
func (m *MsgUpdatePaymentRequest) ValidateBasic() error {
	if err := ValidateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.NextPmtAmt < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("next payment amount cannot be negative")
	}
	if m.NextPmtDate == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("next payment date cannot be zero")
	}
	if m.PaymentFrequency == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("payment frequency cannot be zero")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateMaturityDateRequest
func (m *MsgUpdateMaturityDateRequest) ValidateBasic() error {
	if err := ValidateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.MaturityDate < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("maturity date cannot be negative")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAppendRequest
func (m *MsgAppendRequest) ValidateBasic() error {
	if err := ValidateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if len(m.Entries) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("entries cannot be empty")
	}

	for _, entry := range m.Entries {
		if err := ValidateLedgerEntryBasic(entry); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateBalancesRequest
func (m *MsgUpdateBalancesRequest) ValidateBasic() error {
	if err := ValidateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.CorrelationId == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("correlation id cannot be empty")
	}
	if len(m.BalanceAmounts) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("balance amounts cannot be empty")
	}
	if len(m.AppliedAmounts) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("applied amounts cannot be empty")
	}

	for _, balanceAmount := range m.BalanceAmounts {
		if err := ValidateBucketBalance(balanceAmount); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgFundAssetRequest
func (m *MsgFundAssetRequest) ValidateBasic() error {
	if len(m.Transfers) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("transfers cannot be empty")
	}

	for _, transfer := range m.Transfers {
		if err := ValidateFundTransferWithSettlementBasic(transfer); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgFundAssetByRegistryRequest
func (m *MsgFundAssetByRegistryRequest) ValidateBasic() error {
	if len(m.Transfers) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("transfers cannot be empty")
	}

	for _, transfer := range m.Transfers {
		if err := ValidateFundTransferWithSettlementBasic(transfer); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgTransferFundsWithSettlementRequest
func (m *MsgTransferFundsWithSettlementRequest) ValidateBasic() error {
	if len(m.Transfers) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("transfers cannot be empty")
	}

	for _, transfer := range m.Transfers {
		if err := ValidateFundTransferWithSettlementBasic(transfer); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgDestroyRequest
func (m *MsgDestroyRequest) ValidateBasic() error {
	if err := ValidateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateLedgerClassRequest
func (m *MsgCreateLedgerClassRequest) ValidateBasic() error {
	if err := ValidateLedgerClassBasic(m.LedgerClass); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassStatusTypeRequest
func (m *MsgAddLedgerClassStatusTypeRequest) ValidateBasic() error {
	if m.LedgerClassId == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("ledger class id cannot be empty")
	}
	if m.StatusType == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("status type cannot be nil")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassEntryTypeRequest
func (m *MsgAddLedgerClassEntryTypeRequest) ValidateBasic() error {
	if m.LedgerClassId == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("ledger class id cannot be empty")
	}
	if m.EntryType == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("entry type cannot be nil")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassBucketTypeRequest
func (m *MsgAddLedgerClassBucketTypeRequest) ValidateBasic() error {
	if m.LedgerClassId == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("ledger class id cannot be empty")
	}
	if m.BucketType == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("bucket type cannot be nil")
	}

	return nil
}
