package types

import (
	"regexp"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Validate validates the Ledger type
func (l *Ledger) Validate() error {
	if l.Key == nil {
		return NewErrCodeMissingField("key")
	}

	if err := l.Key.Validate(); err != nil {
		return err
	}

	// Validate the LedgerClassId field
	if err := lenCheck("ledger_class_id", &l.LedgerClassId, 1, 50); err != nil {
		return err
	}

	// Validate status_type_id is positive
	if l.StatusTypeId <= 0 {
		return NewErrCodeInvalidField("status_type_id", "must be a positive integer")
	}

	// Validate next payment date format if provided
	if l.NextPmtDate < 0 {
		return NewErrCodeInvalidField("next_pmt_date", "must be after 1970-01-01")
	}

	// Validate next payment amount if provided
	if l.NextPmtAmt < 0 {
		return NewErrCodeInvalidField("next_pmt_amt", "must be a non-negative integer")
	}

	// Validate interest rate if provided (reasonable bounds: 0-100000000 for 0-1000%)
	if l.InterestRate < 0 || l.InterestRate > 100000000 {
		return NewErrCodeInvalidField("interest_rate", "must be between 0 and 100000000 (0-1000%)")
	}

	// Validate maturity date format if provided
	if l.MaturityDate < 0 {
		return NewErrCodeInvalidField("maturity_date", "must be after 1970-01-01")
	}

	if _, ok := DayCountConvention_name[int32(l.InterestDayCountConvention)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid interest day count convention: %d", l.InterestDayCountConvention)
	}

	if _, ok := InterestAccrualMethod_name[int32(l.InterestAccrualMethod)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid interest accrual method: %d", l.InterestAccrualMethod)
	}

	if _, ok := PaymentFrequency_name[int32(l.PaymentFrequency)]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid payment frequency: %d", l.PaymentFrequency)
	}

	return nil
}

// Validate validates the LedgerKey type
func (lk *LedgerKey) Validate() error {
	if lk == nil {
		return NewErrCodeMissingField("key")
	}

	// Verify that the nft_id and asset_class_id do not contain a null byte
	if strings.Contains(lk.NftId, "\x00") {
		return NewErrCodeInvalidField("nft_id", "must not contain a null byte")
	}

	if err := lenCheck("nft_id", &lk.NftId, 1, 128); err != nil {
		return err
	}

	if err := lenCheck("asset_class_id", &lk.AssetClassId, 1, 128); err != nil {
		return err
	}

	if strings.Contains(lk.AssetClassId, "\x00") {
		return NewErrCodeInvalidField("asset_class_id", "must not contain a null byte")
	}

	return nil
}

// Validate validates the LedgerEntry type
func (le *LedgerEntry) Validate() error {
	if err := lenCheck("correlation_id", &le.CorrelationId, 1, 50); err != nil {
		return err
	}

	// Validate reverses_correlation_id if provided
	if le.ReversesCorrelationId != "" {
		if err := lenCheck("reverses_correlation_id", &le.ReversesCorrelationId, 1, 50); err != nil {
			return err
		}
	}

	// Validate sequence number (should be < 100 as per proto comment)
	if le.Sequence >= 100 {
		return NewErrCodeInvalidField("sequence", "must be less than 100")
	}

	// Validate entry_type_id is positive
	if le.EntryTypeId <= 0 {
		return NewErrCodeInvalidField("entry_type_id", "must be a positive integer")
	}

	if le.PostedDate <= 0 {
		return NewErrCodeInvalidField("posted_date", "must be a valid integer")
	}

	if le.EffectiveDate <= 0 {
		return NewErrCodeInvalidField("effective_date", "must be a valid integer")
	}

	// Validate amounts are non-negative
	if le.TotalAmt.LT(math.NewInt(0)) {
		return NewErrCodeInvalidField("total_amt", "must be a non-negative integer")
	}

	// Validate applied_amounts
	if len(le.AppliedAmounts) == 0 {
		return NewErrCodeInvalidField("applied_amounts", "cannot be empty")
	}

	for _, applied := range le.AppliedAmounts {
		if applied.BucketTypeId <= 0 {
			return NewErrCodeInvalidField("bucket_type_id", "must be a positive integer")
		}
	}

	if err := validateEntryAmounts(le.TotalAmt, le.AppliedAmounts); err != nil {
		return err
	}

	return nil
}

// Validate validates the LedgerClass type
func (lc *LedgerClass) Validate() error {
	if err := lenCheck("ledger_class_id", &lc.LedgerClassId, 1, 50); err != nil {
		return err
	}

	// Verify that the ledger class only contains alphanumeric and dashes
	if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(lc.LedgerClassId) {
		return NewErrCodeInvalidField("ledger_class_id", "must only contain alphanumeric and dashes")
	}

	if err := lenCheck("asset_class_id", &lc.AssetClassId, 1, 128); err != nil {
		return err
	}

	// Validate asset_class_id format (should be a valid UUID or similar format)
	if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(lc.AssetClassId) {
		return NewErrCodeInvalidField("asset_class_id", "must only contain alphanumeric and dashes")
	}

	if err := lenCheck("denom", &lc.Denom, 1, 128); err != nil {
		return err
	}

	// Validate denom format (should be a valid coin denomination)
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9/]{2,127}$`).MatchString(lc.Denom) {
		return NewErrCodeInvalidField("denom", "must be a valid coin denomination")
	}

	maintainerAddress := lc.MaintainerAddress
	if err := lenCheck("maintainer_address", &maintainerAddress, 1, 256); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(lc.MaintainerAddress); err != nil {
		return NewErrCodeInvalidField("maintainer_address", "must be a valid bech32 address")
	}

	return nil
}

// Validate validates the LedgerClassStatusType type
func (lcst *LedgerClassStatusType) Validate() error {
	if lcst.Id < 0 {
		return NewErrCodeInvalidField("id", "must be a non-negative integer")
	}

	if err := lenCheck("code", &lcst.Code, 1, 50); err != nil {
		return err
	}

	if err := lenCheck("description", &lcst.Description, 1, 100); err != nil {
		return err
	}

	return nil
}

// Validate validates the LedgerClassEntryType type
func (lcet *LedgerClassEntryType) Validate() error {
	if lcet.Id < 0 {
		return NewErrCodeInvalidField("id", "must be a non-negative integer")
	}

	if err := lenCheck("code", &lcet.Code, 1, 50); err != nil {
		return err
	}

	if err := lenCheck("description", &lcet.Description, 1, 100); err != nil {
		return err
	}

	return nil
}

// Validate validates the LedgerClassBucketType type
func (lcbt *LedgerClassBucketType) Validate() error {
	if lcbt.Id < 0 {
		return NewErrCodeInvalidField("id", "must be a non-negative integer")
	}

	if err := lenCheck("code", &lcbt.Code, 1, 50); err != nil {
		return err
	}

	if err := lenCheck("description", &lcbt.Description, 1, 100); err != nil {
		return err
	}

	return nil
}

// Validate validates the FundTransferWithSettlement type
func (ft *FundTransferWithSettlement) Validate() error {
	if ft.Key == nil {
		return NewErrCodeMissingField("key")
	}

	if err := ft.Key.Validate(); err != nil {
		return err
	}

	// Validate that the correlation ID is valid
	if err := lenCheck("ledger_entry_correlation_id", &ft.LedgerEntryCorrelationId, 1, 50); err != nil {
		return err
	}

	// Validate that the settlement instructions are valid
	for _, si := range ft.SettlementInstructions {
		if err := si.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the SettlementInstruction type
func (si *SettlementInstruction) Validate() error {
	if si.Amount.IsNegative() {
		return NewErrCodeInvalidField("amount", "must be a non-negative integer")
	}

	// Validate amount has reasonable bounds
	if si.Amount.Amount.GT(math.NewInt(1000000000000000)) { // 1 quadrillion as reasonable max
		return NewErrCodeInvalidField("amount", "amount too large")
	}

	if si.Memo != "" {
		if len(si.Memo) > 50 {
			return NewErrCodeInvalidField("memo", "must be less than 50 characters")
		}
	}

	if si.RecipientAddress == "" {
		return NewErrCodeMissingField("recipient_address")
	}

	// Validate recipient address format
	if _, err := sdk.AccAddressFromBech32(si.RecipientAddress); err != nil {
		return NewErrCodeInvalidField("recipient_address", "must be a valid bech32 address")
	}

	// Validate status enum
	if _, ok := FundingTransferStatus_name[int32(si.Status)]; !ok {
		return NewErrCodeInvalidField("status", "invalid funding transfer status")
	}

	return nil
}

// Validate validates the LedgerBucketAmount type
func (lba *LedgerBucketAmount) Validate() error {
	if lba.BucketTypeId <= 0 {
		return NewErrCodeInvalidField("bucket_type_id", "must be a positive integer")
	}

	if lba.AppliedAmt.LT(math.NewInt(0)) {
		return NewErrCodeInvalidField("applied_amt", "must be a non-negative integer")
	}

	return nil
}

// Validate validates the BucketBalance type
func (bb *BucketBalance) Validate() error {
	if bb.BucketTypeId <= 0 {
		return NewErrCodeInvalidField("bucket_type_id", "must be a positive integer")
	}

	if bb.BalanceAmt.LT(math.NewInt(0)) {
		return NewErrCodeInvalidField("balance_amt", "must be a non-negative integer")
	}

	return nil
}

// Validate validates the LedgerToEntries type
func (lte *LedgerToEntries) Validate() error {
	if err := lte.LedgerKey.Validate(); err != nil {
		return err
	}

	if len(lte.Entries) == 0 {
		return NewErrCodeMissingField("entries")
	}

	for _, entry := range lte.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}

	return nil
}
