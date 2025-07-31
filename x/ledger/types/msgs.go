package types

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/provenance-io/provenance/x/ledger/helper"
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
	(*MsgTransferFundsWithSettlementRequest)(nil),
	(*MsgDestroyRequest)(nil),
	(*MsgCreateLedgerClassRequest)(nil),
	(*MsgAddLedgerClassStatusTypeRequest)(nil),
	(*MsgAddLedgerClassEntryTypeRequest)(nil),
	(*MsgAddLedgerClassBucketTypeRequest)(nil),
	(*MsgBulkImportRequest)(nil),
}

// Note: Authority address validation is performed in the message server to avoid duplicate bech32 conversions.

// ValidateBasic implements the sdk.Msg interface for MsgCreateRequest
func (m MsgCreateRequest) ValidateBasic() error {
	if err := validateLedgerKeyBasic(m.Ledger.Key); err != nil {
		return err
	}

	// Validate the LedgerClassId field
	if err := lenCheck("ledger_class_id", &m.Ledger.LedgerClassId, 1, 50); err != nil {
		return err
	}

	epochTime, _ := time.Parse("2006-01-02", "1970-01-01")
	epoch := helper.DaysSinceEpoch(epochTime.UTC())

	// Validate next payment date format if provided
	if m.Ledger.NextPmtDate < epoch {
		return NewErrCodeInvalidField("next_pmt_date", "must be after 1970-01-01")
	}

	// Validate next payment amount if provided
	if m.Ledger.NextPmtAmt < 0 {
		return NewErrCodeInvalidField("next_pmt_amt", "must be a non-negative integer")
	}

	// Validate interest rate if provided
	if m.Ledger.InterestRate < 0 {
		return NewErrCodeInvalidField("interest_rate", "must be a non-negative integer")
	}

	// Validate maturity date format if provided
	if m.Ledger.MaturityDate < epoch {
		return NewErrCodeInvalidField("maturity_date", "must be after 1970-01-01")
	}

	if _, ok := DayCountConvention_name[int32(m.Ledger.InterestDayCountConvention)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid interest day count convention: %d", m.Ledger.InterestDayCountConvention)
	}

	if _, ok := InterestAccrualMethod_name[int32(m.Ledger.InterestAccrualMethod)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid interest accrual method: %d", m.Ledger.InterestAccrualMethod)
	}

	if _, ok := PaymentFrequency_name[int32(m.Ledger.PaymentFrequency)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid payment frequency: %d", m.Ledger.PaymentFrequency)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateStatusRequest
func (m MsgUpdateStatusRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.StatusTypeId == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("status type id cannot be zero")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateInterestRateRequest
func (m MsgUpdateInterestRateRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.InterestRate < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("interest rate cannot be negative")
	}

	if _, ok := DayCountConvention_name[int32(m.InterestDayCountConvention)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid interest day count convention: %d", m.InterestDayCountConvention)
	}

	if _, ok := InterestAccrualMethod_name[int32(m.InterestAccrualMethod)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid interest accrual method: %d", m.InterestAccrualMethod)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdatePaymentRequest
func (m MsgUpdatePaymentRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.NextPmtAmt < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("next payment amount cannot be negative")
	}
	if m.NextPmtDate == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("next payment date cannot be zero")
	}

	if _, ok := PaymentFrequency_name[int32(m.PaymentFrequency)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid payment frequency: %d", m.PaymentFrequency)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateMaturityDateRequest
func (m MsgUpdateMaturityDateRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.MaturityDate < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("maturity date cannot be negative")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAppendRequest
func (m MsgAppendRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if len(m.Entries) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("entries cannot be empty")
	}

	for _, e := range m.Entries {
		if err := lenCheck("correlation_id", &e.CorrelationId, 1, 50); err != nil {
			return err
		}

		if e.PostedDate <= 0 {
			return NewErrCodeInvalidField("posted_date", "must be a valid integer")
		}

		if e.EffectiveDate <= 0 {
			return NewErrCodeInvalidField("effective_date", "must be a valid integer")
		}

		// Validate amounts are non-negative
		if e.TotalAmt.LT(math.NewInt(0)) {
			return NewErrCodeInvalidField("total_amt", "must be a non-negative integer")
		}

		if err := validateEntryAmounts(e.TotalAmt, e.AppliedAmounts); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateBalancesRequest
func (m MsgUpdateBalancesRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if err := validateLedgerKeyBasic(m.Key); err != nil {
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
		if balanceAmount.BucketTypeId <= 0 {
			return NewErrCodeInvalidField("bucket_type_id", "must be a positive integer")
		}

		if err := validateEntryAmounts(balanceAmount.BalanceAmt, m.AppliedAmounts); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgTransferFundsWithSettlementRequest
func (m MsgTransferFundsWithSettlementRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if len(m.Transfers) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("transfers cannot be empty")
	}

	for _, ft := range m.Transfers {
		if err := validateLedgerKeyBasic(ft.Key); err != nil {
			return err
		}

		// Validate that the correlation ID is valid
		if err := lenCheck("ledger_entry_correlation_id", &ft.LedgerEntryCorrelationId, 1, 50); err != nil {
			return err
		}

		// Validate that the settlement instructions are valid
		for _, si := range ft.SettlementInstructions {
			if si.Amount.IsNegative() {
				return NewErrCodeInvalidField("amount", "must be a non-negative integer")
			}

			if si.Memo != "" {
				if len(si.Memo) > 50 {
					return NewErrCodeInvalidField("memo", "must be less than 50 characters")
				}
			}

			if si.RecipientAddress == "" {
				return NewErrCodeMissingField("recipient_address")
			}
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgDestroyRequest
func (m MsgDestroyRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateLedgerClassRequest
func (m MsgCreateLedgerClassRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	// Validate that the maintainer in the ledger class is the same as the maintainer address.
	// We force them to be the same for now so that a ledger class isn't locked out.
	if m.LedgerClass.MaintainerAddress != m.Authority {
		return NewErrCodeUnauthorized("maintainer address is not the same as the authority")
	}

	if err := lenCheck("ledger_class_id", &m.LedgerClass.LedgerClassId, 1, 50); err != nil {
		return err
	}

	// Verify that the ledger class only contains alphanumeric and dashes
	if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(m.LedgerClass.LedgerClassId) {
		return NewErrCodeInvalidField("ledger_class_id", "must only contain alphanumeric and dashes")
	}

	if err := lenCheck("asset_class_id", &m.LedgerClass.AssetClassId, 1, 128); err != nil {
		return err
	}

	if err := lenCheck("denom", &m.LedgerClass.Denom, 1, 128); err != nil {
		return err
	}

	maintainerAddress := m.LedgerClass.MaintainerAddress
	if err := lenCheck("maintainer_address", &maintainerAddress, 1, 256); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(m.LedgerClass.MaintainerAddress); err != nil {
		return NewErrCodeInvalidField("maintainer_address", "must be a valid bech32 address")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassStatusTypeRequest
func (m MsgAddLedgerClassStatusTypeRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if m.LedgerClassId == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("ledger class id cannot be empty")
	}

	if m.StatusType == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("status type cannot be nil")
	}

	if m.StatusType.Id < 0 {
		return NewErrCodeInvalidField("status_type_id", "must be a positive integer")
	}

	if err := lenCheck("status_type_code", &m.StatusType.Code, 1, 50); err != nil {
		return err
	}

	if err := lenCheck("status_type_description", &m.StatusType.Description, 1, 100); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassEntryTypeRequest
func (m MsgAddLedgerClassEntryTypeRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if m.LedgerClassId == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("ledger class id cannot be empty")
	}
	if m.EntryType == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("entry type cannot be nil")
	}

	if m.EntryType.Id < 0 {
		return NewErrCodeInvalidField("entry_type_id", "must be a positive integer")
	}

	if err := lenCheck("entry_type_code", &m.EntryType.Code, 1, 50); err != nil {
		return err
	}

	if err := lenCheck("entry_type_description", &m.EntryType.Description, 1, 100); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassBucketTypeRequest
func (m MsgAddLedgerClassBucketTypeRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if m.LedgerClassId == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("ledger class id cannot be empty")
	}
	if m.BucketType == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("bucket type cannot be nil")
	}

	if m.BucketType.Id < 0 {
		return NewErrCodeInvalidField("bucket_type_id", "must be a positive integer")
	}

	if err := lenCheck("bucket_type_code", &m.BucketType.Code, 1, 50); err != nil {
		return err
	}

	if err := lenCheck("bucket_type_description", &m.BucketType.Description, 1, 100); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgBulkImportRequest
func (m MsgBulkImportRequest) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return err
	}

	if m.GenesisState == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("genesis state cannot be nil")
	}

	return nil
}

func validateLedgerKeyBasic(key *LedgerKey) error {
	if err := lenCheck("nft_id", &key.NftId, 1, 128); err != nil {
		return err
	}

	if err := lenCheck("asset_class_id", &key.AssetClassId, 1, 128); err != nil {
		return err
	}

	return nil
}

// validateEntryAmounts checks if the amounts are valid
func validateEntryAmounts(totalAmt math.Int, appliedAmounts []*LedgerBucketAmount) error {
	// Check if total amount matches sum of applied amounts
	totalApplied := math.NewInt(0)
	for _, applied := range appliedAmounts {
		totalApplied = totalApplied.Add(applied.AppliedAmt.Abs())
	}

	if !totalAmt.Equal(totalApplied) {
		return NewErrCodeInvalidField("total_amt", "total amount must equal sum of abs(applied amounts)")
	}

	return nil
}

// lenCheck checks if the string is nil or empty and if it is, returns a missing field error.
// It also checks if the string is less than the minimum length or greater than the maximum length and returns an invalid field error.
func lenCheck(field string, s *string, minLength int, maxLength int) error {
	if s == nil {
		return NewErrCodeMissingField(field)
	}

	trimmed := strings.TrimSpace(*s)

	// empty string
	if trimmed == "" {
		return NewErrCodeMissingField(field)
	}

	if len(trimmed) < minLength {
		return NewErrCodeInvalidField(field, "must be greater than or equal to "+strconv.Itoa(minLength)+" characters")
	}

	if len(trimmed) > maxLength {
		return NewErrCodeInvalidField(field, "must be less than or equal to "+strconv.Itoa(maxLength)+" characters")
	}

	return nil
}
