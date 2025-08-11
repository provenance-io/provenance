package types

import (
	"strconv"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	(*MsgBulkCreateRequest)(nil),
}

// Note: Authority address validation is performed in the message server to avoid duplicate bech32 conversions.

// ValidateBasic implements the sdk.Msg interface for MsgCreateRequest
func (m MsgCreateRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := m.Ledger.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateStatusRequest
func (m MsgUpdateStatusRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	if m.StatusTypeId <= 0 {
		return NewErrCodeInvalidField("status_type_id", "must be a positive integer")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateInterestRateRequest
func (m MsgUpdateInterestRateRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	// Validate interest rate bounds (reasonable bounds: 0-100000000 for 0-1000%)
	if m.InterestRate < 0 || m.InterestRate > 100000000 {
		return NewErrCodeInvalidField("interest_rate", "must be between 0 and 100000000 (0-1000%)")
	}

	if _, ok := DayCountConvention_name[int32(m.InterestDayCountConvention)]; !ok {
		return NewErrCodeInvalidField("interest_day_count_convention", "invalid interest day count convention")
	}

	if _, ok := InterestAccrualMethod_name[int32(m.InterestAccrualMethod)]; !ok {
		return NewErrCodeInvalidField("interest_accrual_method", "invalid interest accrual method")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdatePaymentRequest
func (m MsgUpdatePaymentRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	// Validate next payment amount bounds
	if m.NextPmtAmt < 0 {
		return NewErrCodeInvalidField("next_pmt_amt", "cannot be negative")
	}

	// Validate next payment date
	if m.NextPmtDate <= 0 {
		return NewErrCodeInvalidField("next_pmt_date", "must be a positive integer")
	}

	if _, ok := PaymentFrequency_name[int32(m.PaymentFrequency)]; !ok {
		return NewErrCodeInvalidField("payment_frequency", "invalid payment frequency")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateMaturityDateRequest
func (m MsgUpdateMaturityDateRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	if m.MaturityDate <= 0 {
		return NewErrCodeInvalidField("maturity_date", "must be a positive integer")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAppendRequest
func (m MsgAppendRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	if len(m.Entries) == 0 {
		return NewErrCodeInvalidField("entries", "cannot be empty")
	}

	for _, e := range m.Entries {
		if err := e.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateBalancesRequest
func (m MsgUpdateBalancesRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	if err := lenCheck("correlation_id", &m.CorrelationId, 1, 50); err != nil {
		return err
	}

	if len(m.BalanceAmounts) == 0 {
		return NewErrCodeInvalidField("balance_amounts", "cannot be empty")
	}
	if len(m.AppliedAmounts) == 0 {
		return NewErrCodeInvalidField("applied_amounts", "cannot be empty")
	}

	for _, balanceAmount := range m.BalanceAmounts {
		if balanceAmount.BucketTypeId <= 0 {
			return NewErrCodeInvalidField("bucket_type_id", "must be a positive integer")
		}

		if err := validateEntryAmounts(balanceAmount.BalanceAmt, m.AppliedAmounts); err != nil {
			return err
		}
	}

	// Validate applied_amounts bucket_type_ids
	for _, applied := range m.AppliedAmounts {
		if applied.BucketTypeId <= 0 {
			return NewErrCodeInvalidField("applied_amounts.bucket_type_id", "must be a positive integer")
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgTransferFundsWithSettlementRequest
func (m MsgTransferFundsWithSettlementRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if len(m.Transfers) == 0 {
		return NewErrCodeInvalidField("transfers", "cannot be empty")
	}

	for _, ft := range m.Transfers {
		if err := ft.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgDestroyRequest
func (m MsgDestroyRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateLedgerClassRequest
func (m MsgCreateLedgerClassRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	// Validate that the maintainer in the ledger class is the same as the maintainer address.
	// We force them to be the same for now so that a ledger class isn't locked out.
	if m.LedgerClass.MaintainerAddress != m.Authority {
		return NewErrCodeUnauthorized("maintainer address is not the same as the authority")
	}

	if err := m.LedgerClass.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassStatusTypeRequest
func (m MsgAddLedgerClassStatusTypeRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := lenCheck("ledger_class_id", &m.LedgerClassId, 1, 50); err != nil {
		return err
	}

	if m.StatusType == nil {
		return NewErrCodeInvalidField("status_type", "cannot be nil")
	}

	if err := m.StatusType.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassEntryTypeRequest
func (m MsgAddLedgerClassEntryTypeRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := lenCheck("ledger_class_id", &m.LedgerClassId, 1, 50); err != nil {
		return err
	}

	if m.EntryType == nil {
		return NewErrCodeInvalidField("entry_type", "cannot be nil")
	}

	if err := m.EntryType.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassBucketTypeRequest
func (m MsgAddLedgerClassBucketTypeRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	if err := lenCheck("ledger_class_id", &m.LedgerClassId, 1, 50); err != nil {
		return err
	}

	if m.BucketType == nil {
		return NewErrCodeInvalidField("bucket_type", "cannot be nil")
	}

	if err := m.BucketType.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgBulkCreateRequest
func (m MsgBulkCreateRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return err
	}

	for _, ledgerToEntries := range m.LedgerToEntries {
		if err := ledgerToEntries.Validate(); err != nil {
			return err
		}
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
