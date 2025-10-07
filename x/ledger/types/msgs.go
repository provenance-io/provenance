package types

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgCreateLedgerRequest)(nil),
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

// Note: Signer address validation is performed in the message server to avoid duplicate bech32 conversions.

// ValidateBasic implements the sdk.Msg interface for MsgCreateLedgerRequest
func (m MsgCreateLedgerRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Ledger.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("ledger", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateStatusRequest
func (m MsgUpdateStatusRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	if m.StatusTypeId <= 0 {
		errs = append(errs, NewErrCodeInvalidField("status_type_id", "must be a positive integer"))
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateInterestRateRequest
func (m MsgUpdateInterestRateRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	// Validate interest rate bounds (reasonable bounds: 0-100000000 for 0-100%)
	if m.InterestRate < 0 || m.InterestRate > 100_000_000 {
		errs = append(errs, NewErrCodeInvalidField("interest_rate", "must be between 0 and 100,000,000 (0-100%%)"))
	}

	if err := m.InterestDayCountConvention.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("interest_day_count_convention", "%s", err))
	}

	if err := m.InterestAccrualMethod.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("interest_accrual_method", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdatePaymentRequest
func (m MsgUpdatePaymentRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	// Validate next payment amount bounds
	if m.NextPmtAmt.IsNegative() {
		errs = append(errs, NewErrCodeInvalidField("next_pmt_amt", "cannot be negative"))
	}

	// Validate next payment date
	if m.NextPmtDate <= 0 {
		errs = append(errs, NewErrCodeInvalidField("next_pmt_date", "must be a positive integer"))
	}

	if err := m.PaymentFrequency.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("payment_frequency", "invalid payment frequency"))
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateMaturityDateRequest
func (m MsgUpdateMaturityDateRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	if m.MaturityDate <= 0 {
		errs = append(errs, NewErrCodeInvalidField("maturity_date", "must be a positive integer"))
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgAppendRequest
func (m MsgAppendRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	if len(m.Entries) == 0 {
		errs = append(errs, NewErrCodeInvalidField("entries", "cannot be empty"))
	}

	corIDs := make(map[string]int)
	for i, e := range m.Entries {
		if err := e.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("entries", "%s", err))
		}
		if j, known := corIDs[e.CorrelationId]; known {
			errs = append(errs, NewErrCodeInvalidField("entries",
				"correlation id %q duplicated at indexes %d and %d", e.CorrelationId, j, i))
		}
		corIDs[e.CorrelationId] = i
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateBalancesRequest
func (m MsgUpdateBalancesRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	if err := lenCheck(m.CorrelationId, 1, MaxLenCorrelationID); err != nil {
		errs = append(errs, NewErrCodeInvalidField("correlation_id", "%s", err))
	}

	if len(m.BalanceAmounts) == 0 {
		errs = append(errs, NewErrCodeInvalidField("balance_amounts", "cannot be empty"))
	}
	if len(m.AppliedAmounts) == 0 {
		errs = append(errs, NewErrCodeInvalidField("applied_amounts", "cannot be empty"))
	}

	for i, balanceAmount := range m.BalanceAmounts {
		if err := balanceAmount.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField(fmt.Sprintf("balance_amounts[%d]", i), "%s", err))
		}
	}

	// Validate applied_amounts bucket_type_ids
	for i, applied := range m.AppliedAmounts {
		if err := applied.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField(fmt.Sprintf("applied_amounts[%d]", i), "%s", err))
		}
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgTransferFundsWithSettlementRequest
func (m MsgTransferFundsWithSettlementRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if len(m.Transfers) == 0 {
		errs = append(errs, NewErrCodeInvalidField("transfers", "cannot be empty"))
	}

	for _, ft := range m.Transfers {
		if err := ft.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("transfers", "%s", err))
		}
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgDestroyRequest
func (m MsgDestroyRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateLedgerClassRequest
func (m MsgCreateLedgerClassRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	// Validate that the maintainer in the ledger class is the same as the maintainer address.
	// We force them to be the same for now so that a ledger class isn't locked out.
	if m.LedgerClass.MaintainerAddress != m.Signer {
		errs = append(errs, NewErrCodeUnauthorized("maintainer address is not the same as the signer"))
	}

	if err := m.LedgerClass.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("ledger_class", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassStatusTypeRequest
func (m MsgAddLedgerClassStatusTypeRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := lenCheck(m.LedgerClassId, 1, MaxLenLedgerClassID); err != nil {
		errs = append(errs, NewErrCodeInvalidField("ledger_class_id", "%s", err))
	}

	if m.StatusType == nil {
		errs = append(errs, NewErrCodeInvalidField("status_type", "cannot be nil"))
	} else {
		if err := m.StatusType.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("status_type", "%s", err))
		}
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassEntryTypeRequest
func (m MsgAddLedgerClassEntryTypeRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := lenCheck(m.LedgerClassId, 1, MaxLenLedgerClassID); err != nil {
		errs = append(errs, NewErrCodeInvalidField("ledger_class_id", "%s", err))
	}

	if m.EntryType == nil {
		errs = append(errs, NewErrCodeInvalidField("entry_type", "cannot be nil"))
	} else if err := m.EntryType.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("entry_type", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgAddLedgerClassBucketTypeRequest
func (m MsgAddLedgerClassBucketTypeRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := lenCheck(m.LedgerClassId, 1, MaxLenLedgerClassID); err != nil {
		errs = append(errs, NewErrCodeInvalidField("ledger_class_id", "%s", err))
	}

	if m.BucketType == nil {
		errs = append(errs, NewErrCodeInvalidField("bucket_type", "cannot be nil"))
	} else {
		if err := m.BucketType.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("bucket_type", "%s", err))
		}
	}

	return errors.Join(errs...)
}

// ValidateBasic implements the sdk.Msg interface for MsgBulkCreateRequest
func (m MsgBulkCreateRequest) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	for _, ledgerAndEntries := range m.LedgerAndEntries {
		if err := ledgerAndEntries.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("ledger_and_entries", "%s", err))
		}
	}

	return errors.Join(errs...)
}

// validateEntryAmounts checks if the amounts are valid
func validateEntryAmounts(totalAmt math.Int, appliedAmounts []*LedgerBucketAmount) error {
	// Check if total amount matches sum of applied amounts
	totalApplied := math.NewInt(0)
	for _, applied := range appliedAmounts {
		totalApplied = totalApplied.Add(applied.AppliedAmt.Abs())
	}

	if !totalAmt.Equal(totalApplied) {
		return fmt.Errorf("total amount must equal sum of abs(applied amounts)")
	}

	return nil
}

// lenCheck returns an error if the provided str is shorter or longer than the provided bounds.
// If minLength > 0 and str is empty, a missing field error is returned. Otherwise, if str
// is shorter than minLength or longer than maxLength, an invalid field error is returned.
func lenCheck(str string, minLength int, maxLength int) error {
	if len(str) > maxLength || len(str) < minLength {
		return fmt.Errorf("must be between %d and %d characters", minLength, maxLength)
	}

	return nil
}
