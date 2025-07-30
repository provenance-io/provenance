package types

import (
	"regexp"
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
func (m *MsgCreateRequest) ValidateBasic() error {
	if err := validateLedgerBasic(m.Ledger); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateStatusRequest
func (m *MsgUpdateStatusRequest) ValidateBasic() error {
	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.StatusTypeId == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("status type id cannot be zero")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateInterestRateRequest
func (m *MsgUpdateInterestRateRequest) ValidateBasic() error {
	if err := validateLedgerKeyBasic(m.Key); err != nil {
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
	if err := validateLedgerKeyBasic(m.Key); err != nil {
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
	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if m.MaturityDate < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("maturity date cannot be negative")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAppendRequest
func (m *MsgAppendRequest) ValidateBasic() error {
	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	if len(m.Entries) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("entries cannot be empty")
	}

	for _, entry := range m.Entries {
		if err := validateLedgerEntryBasic(entry); err != nil {
			return err
		}

	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateBalancesRequest
func (m *MsgUpdateBalancesRequest) ValidateBasic() error {
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
		if err := validateBucketBalance(balanceAmount); err != nil {
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
		if err := validateFundTransferWithSettlementBasic(transfer); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgDestroyRequest
func (m *MsgDestroyRequest) ValidateBasic() error {
	if err := validateLedgerKeyBasic(m.Key); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateLedgerClassRequest
func (m *MsgCreateLedgerClassRequest) ValidateBasic() error {
	if err := validateLedgerClassBasic(m.LedgerClass); err != nil {
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

// ValidateBasic implements the sdk.Msg interface for MsgBulkImportRequest
func (m *MsgBulkImportRequest) ValidateBasic() error {
	if m.Authority == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("authority cannot be empty")
	}
	if m.GenesisState == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("genesis state cannot be nil")
	}

	return nil
}

func validateLedgerClassBasic(l *LedgerClass) error {
	if emptyString(&l.LedgerClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "ledger_class_id")
	}

	// Verify that the ledger class id is less than 50 characters
	if len(l.LedgerClassId) > 50 {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger_class_id", "must be less than 50 characters")
	}

	// Verify that the ledger class only contains alphanumeric and dashes
	if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(l.LedgerClassId) {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger_class_id", "must only contain alphanumeric and dashes")
	}

	if emptyString(&l.AssetClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "asset_class_id")
	}

	if emptyString(&l.Denom) {
		return NewLedgerCodedError(ErrCodeMissingField, "denom")
	}

	maintainerAddress := l.MaintainerAddress
	if emptyString(&maintainerAddress) {
		return NewLedgerCodedError(ErrCodeMissingField, "maintainer_address")
	}

	return nil
}

func validateLedgerKeyBasic(key *LedgerKey) error {
	if emptyString(&key.NftId) {
		return NewLedgerCodedError(ErrCodeMissingField, "nft_id")
	}
	if emptyString(&key.AssetClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "asset_class_id")
	}
	return nil
}

func validateLedgerBasic(l *Ledger) error {
	if l == nil {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger", "ledger cannot be nil")
	}

	if err := validateLedgerKeyBasic(l.Key); err != nil {
		return err
	}

	// Validate the LedgerClassId field
	if emptyString(&l.LedgerClassId) {
		return NewLedgerCodedError(ErrCodeMissingField, "ledger_class_id")
	}

	keyError := validateLedgerKeyBasic(l.Key)
	if keyError != nil {
		return keyError
	}

	epochTime, _ := time.Parse("2006-01-02", "1970-01-01")
	epoch := helper.DaysSinceEpoch(epochTime.UTC())

	// Validate next payment date format if provided
	if l.NextPmtDate < epoch {
		return NewLedgerCodedError(ErrCodeInvalidField, "next_pmt_date", "must be after 1970-01-01")
	}

	// Validate next payment amount if provided
	if l.NextPmtAmt < 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "next_pmt_amt", "must be a non-negative integer")
	}

	// Validate interest rate if provided
	if l.InterestRate < 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "interest_rate", "must be a non-negative integer")
	}

	// Validate maturity date format if provided
	if l.MaturityDate < epoch {
		return NewLedgerCodedError(ErrCodeInvalidField, "maturity_date", "must be after 1970-01-01")
	}

	return nil
}

func validateLedgerEntryBasic(e *LedgerEntry) error {
	if emptyString(&e.CorrelationId) {
		return NewLedgerCodedError(ErrCodeMissingField, "correlation_id")
	} else {
		if !isCorrelationIDValid(e.CorrelationId) {
			return NewLedgerCodedError(ErrCodeInvalidField, "correlation_id", "must be a valid string that is less than 50 characters")
		}
	}

	if e.PostedDate <= 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "posted_date", "must be a valid integer")
	}

	if e.EffectiveDate <= 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "effective_date", "must be a valid integer")
	}

	// Validate amounts are non-negative
	if e.TotalAmt.LT(math.NewInt(0)) {
		return NewLedgerCodedError(ErrCodeInvalidField, "total_amt", "must be a non-negative integer")
	}

	return nil
}

func validateBucketBalance(bb *BucketBalance) error {
	if bb.BucketTypeId <= 0 {
		return NewLedgerCodedError(ErrCodeInvalidField, "bucket_type_id", "must be a positive integer")
	}

	return nil
}

func validateFundTransferWithSettlementBasic(ft *FundTransferWithSettlement) error {
	if err := validateLedgerKeyBasic(ft.Key); err != nil {
		return err
	}

	// Validate that the correlation ID is valid
	if !isCorrelationIDValid(ft.LedgerEntryCorrelationId) {
		return NewLedgerCodedError(ErrCodeInvalidField, "ledger_entry_correlation_id", "must be a valid string that is less than 50 characters")
	}

	// Validate that the settlement instructions are valid
	for _, settlementInstruction := range ft.SettlementInstructions {
		if err := validateSettlementInstructionBasic(settlementInstruction); err != nil {
			return err
		}
	}

	return nil
}

func validateSettlementInstructionBasic(si *SettlementInstruction) error {
	if si.Amount.IsNegative() {
		return NewLedgerCodedError(ErrCodeInvalidField, "amount", "must be a non-negative integer")
	}

	if si.Memo != "" {
		if len(si.Memo) > 50 {
			return NewLedgerCodedError(ErrCodeInvalidField, "memo", "must be less than 50 characters")
		}
	}

	if si.RecipientAddress == "" {
		return NewLedgerCodedError(ErrCodeMissingField, "recipient_address")
	}

	return nil
}

// Returns true if the string is nil or empty(TrimSpace(*s))
func emptyString(s *string) bool {
	if s == nil || strings.TrimSpace(*s) == "" {
		return true
	}
	return false
}

// validateCorrelationID validates that the provided string is a valid correlation ID.
// Returns true if the string is a valid correlation ID (non-empty and max 50 characters), false otherwise.
func isCorrelationIDValid(correlationID string) bool {
	if len(correlationID) == 0 || len(correlationID) > 50 {
		return false
	}
	return true
}
