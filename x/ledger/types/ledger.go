package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

const (
	MaxLedgerEntrySequence = 299
	MaxLenLedgerClassID    = 50
	MaxLenCorrelationID    = 50
	MaxLenCode             = 50
	MaxLenDescription      = 100
	MaxLenDenom            = 128
	MaxLenMemo             = 50
)

// Validate validates the LedgerClass type
func (lc *LedgerClass) Validate() error {
	var errs []error
	// Validate ledger class id format using asset class id validation
	if err := registrytypes.ValidateClassID(lc.LedgerClassId); err != nil {
		errs = append(errs, fmt.Errorf("ledger_class_id: %w", err))
	}

	if err := registrytypes.ValidateClassID(lc.AssetClassId); err != nil {
		errs = append(errs, fmt.Errorf("asset_class_id: %w", err))
	}

	// Check denom length first for nicer error messages.
	if err := registrytypes.ValidateStringLength(lc.Denom, 2, MaxLenDenom); err != nil {
		errs = append(errs, fmt.Errorf("denom: %w", err))
	}

	// Validate denom format (should be a valid coin denomination)
	if err := sdk.ValidateDenom(lc.Denom); err != nil {
		errs = append(errs, fmt.Errorf("denom must be a valid coin denomination: %w", err))
	}

	if _, err := sdk.AccAddressFromBech32(lc.MaintainerAddress); err != nil {
		errs = append(errs, fmt.Errorf("maintainer_address: %w", err))
	}

	return errors.Join(errs...)
}

// Validate validates the LedgerClassEntryType type
func (lcet *LedgerClassEntryType) Validate() error {
	var errs []error
	if lcet.Id < 0 {
		errs = append(errs, fmt.Errorf("id must be a non-negative integer"))
	}

	if err := lenCheck(lcet.Code, 1, MaxLenCode); err != nil {
		errs = append(errs, fmt.Errorf("code: %w", err))
	}

	if err := lenCheck(lcet.Description, 1, MaxLenDescription); err != nil {
		errs = append(errs, fmt.Errorf("description: %w", err))
	}

	return errors.Join(errs...)
}

// Validate validates the LedgerClassStatusType type
func (lcst *LedgerClassStatusType) Validate() error {
	var errs []error
	if lcst.Id < 0 {
		errs = append(errs, fmt.Errorf("id: must be a non-negative integer"))
	}

	if err := lenCheck(lcst.Code, 1, MaxLenCode); err != nil {
		errs = append(errs, fmt.Errorf("code: %w", err))
	}

	if err := lenCheck(lcst.Description, 1, MaxLenDescription); err != nil {
		errs = append(errs, fmt.Errorf("description: %w", err))
	}

	return errors.Join(errs...)
}

const (
	ledgerKeyHrp = "ledger"
)

func NewLedgerKey(assetClassID string, nftID string) *LedgerKey {
	return &LedgerKey{
		AssetClassId: assetClassID,
		NftId:        nftID,
	}
}

// Combine the asset class id and nft id into a bech32 string.
// Using bech32 here just allows us a readable identifier for the ledger.
func (lk LedgerKey) String() string {
	// Use null byte as delimiter
	joined := lk.AssetClassId + "\x00" + lk.NftId

	b32, err := bech32.ConvertAndEncode(ledgerKeyHrp, []byte(joined))
	if err != nil {
		panic(err)
	}

	return b32
}

// Convert a bech32 string to a LedgerKey.
func StringToLedgerKey(s string) (*LedgerKey, error) {
	hrp, b, err := bech32.DecodeAndConvert(s)
	if err != nil {
		return nil, err
	}

	if hrp != ledgerKeyHrp {
		return nil, fmt.Errorf("invalid hrp: %s", hrp)
	}

	// Split by null byte delimiter
	parts := strings.Split(string(b), "\x00")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key: %s", s)
	}

	return &LedgerKey{
		AssetClassId: parts[0],
		NftId:        parts[1],
	}, nil
}

// Implement Compare() for LedgerEntry
func (le *LedgerEntry) Compare(b *LedgerEntry) int {
	// First compare effective date (ISO8601 string)
	if le.EffectiveDate < b.EffectiveDate {
		return -1
	}
	if le.EffectiveDate > b.EffectiveDate {
		return 1
	}

	// Then compare sequence number
	if le.Sequence < b.Sequence {
		return -1
	}
	if le.Sequence > b.Sequence {
		return 1
	}

	// Equal
	return 0
}

func (lk LedgerKey) ToRegistryKey() *registrytypes.RegistryKey {
	return &registrytypes.RegistryKey{
		AssetClassId: lk.AssetClassId,
		NftId:        lk.NftId,
	}
}

// Validate validates the LedgerKey type
func (lk *LedgerKey) Validate() error {
	if lk == nil {
		return fmt.Errorf("key cannot be nil")
	}
	var errs []error
	if err := registrytypes.ValidateClassID(lk.AssetClassId); err != nil {
		errs = append(errs, NewErrCodeInvalidField("asset_class_id", "%s", err))
	}

	if err := registrytypes.ValidateNftID(lk.NftId); err != nil {
		errs = append(errs, NewErrCodeInvalidField("nft_id", "%s", err))
	}

	return errors.Join(errs...)
}

// Equals returns true if this LedgerKey equals the provided one.
func (lk *LedgerKey) Equals(other *LedgerKey) bool {
	if lk == other {
		return true
	}
	if lk == nil || other == nil {
		return false
	}
	return lk.NftId == other.NftId && lk.AssetClassId == other.AssetClassId
}

// Validate validates the Ledger type
func (l *Ledger) Validate() error {
	if l.Key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	var errs []error
	if err := l.Key.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("key: %w", err))
	}

	// Validate ledger class id format using asset class id validation
	if err := registrytypes.ValidateClassID(l.LedgerClassId); err != nil {
		errs = append(errs, fmt.Errorf("ledger_class_id: %w", err))
	}

	// Validate status_type_id is positive
	if l.StatusTypeId <= 0 {
		errs = append(errs, fmt.Errorf("status_type_id: must be a positive integer"))
	}

	// Validate next payment date format if provided
	if l.NextPmtDate <= 0 {
		errs = append(errs, fmt.Errorf("next_pmt_date: must be after 1970-01-01"))
	}

	// Validate next payment amount if provided
	if l.NextPmtAmt.IsNil() || l.NextPmtAmt.IsNegative() {
		errs = append(errs, fmt.Errorf("next_pmt_amt: must be a non-negative integer"))
	}

	// Validate interest rate if provided (reasonable bounds: 0-100000000 for 0-100%)
	if l.InterestRate < 0 || l.InterestRate > 100_000_000 {
		errs = append(errs, fmt.Errorf("interest_rate: must be between 0 and 100,000,000 (0-100%%)"))
	}

	// Validate maturity date format if provided
	if l.MaturityDate < 0 {
		errs = append(errs, fmt.Errorf("maturity_date: must be after 1970-01-01"))
	}

	if err := l.InterestDayCountConvention.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("interest_day_count_convention: %w", err))
	}

	if err := l.InterestAccrualMethod.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("interest_accrual_method: %w", err))
	}

	if err := l.PaymentFrequency.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("payment_frequency: %w", err))
	}

	return errors.Join(errs...)
}

// Validate validates the LedgerClassBucketType type
func (lcbt *LedgerClassBucketType) Validate() error {
	var errs []error
	if lcbt.Id < 0 {
		errs = append(errs, fmt.Errorf("id: must be a non-negative integer"))
	}

	if err := lenCheck(lcbt.Code, 1, MaxLenCode); err != nil {
		errs = append(errs, fmt.Errorf("code: %w", err))
	}

	if err := lenCheck(lcbt.Description, 1, MaxLenDescription); err != nil {
		errs = append(errs, fmt.Errorf("description: %w", err))
	}

	return errors.Join(errs...)
}

// Validate validates the LedgerEntry type
func (le *LedgerEntry) Validate() error {
	var errs []error
	if err := lenCheck(le.CorrelationId, 1, MaxLenCorrelationID); err != nil {
		errs = append(errs, fmt.Errorf("correlation_id: %w", err))
	}

	// Validate reverses_correlation_id if provided
	if err := lenCheck(le.ReversesCorrelationId, 0, MaxLenCorrelationID); err != nil {
		errs = append(errs, fmt.Errorf("reverses_correlation_id: %w", err))
	}

	// Validate sequence number (should be < 100 as per proto comment)
	if err := ValidateSequence(le.Sequence); err != nil {
		errs = append(errs, err)
	}

	// Validate entry_type_id is positive
	if le.EntryTypeId <= 0 {
		errs = append(errs, fmt.Errorf("entry_type_id: must be a positive integer"))
	}

	if le.PostedDate <= 0 {
		errs = append(errs, fmt.Errorf("posted_date: must be a valid integer"))
	}

	if le.EffectiveDate <= 0 {
		errs = append(errs, fmt.Errorf("effective_date: must be a valid integer"))
	}

	// Validate amounts are non-negative
	totOK, amtsOK := true, true
	if le.TotalAmt.IsNil() || le.TotalAmt.IsNegative() {
		errs = append(errs, fmt.Errorf("total_amt: must be a non-negative integer"))
		totOK = false
	} else if !le.TotalAmt.IsZero() && len(le.AppliedAmounts) == 0 {
		errs = append(errs, fmt.Errorf("applied_amounts: cannot be empty"))
		amtsOK = false
	}

	for i, applied := range le.AppliedAmounts {
		if err := applied.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("applied_amounts[%d]: %w", i, err))
			amtsOK = false
		}
	}

	if totOK && amtsOK {
		if err := validateEntryAmounts(le.TotalAmt, le.AppliedAmounts); err != nil {
			errs = append(errs, fmt.Errorf("applied_amounts: %w", err))
		}
	}

	// Validate balance amounts
	for i, balance := range le.BalanceAmounts {
		if err := balance.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("balance_amounts[%d]: %w", i, err))
		}
	}

	return errors.Join(errs...)
}

// ValidateSequence returns an error if the sequence number is too large.
func ValidateSequence(seq uint32) error {
	if seq >= MaxLedgerEntrySequence {
		return fmt.Errorf("sequence: must be less than %d", MaxLedgerEntrySequence)
	}
	return nil
}

// Validate validates the LedgerBucketAmount type
func (lba *LedgerBucketAmount) Validate() error {
	var errs []error
	if lba.BucketTypeId <= 0 {
		errs = append(errs, fmt.Errorf("bucket_type_id: must be a positive integer"))
	}

	if lba.AppliedAmt.IsNil() {
		errs = append(errs, fmt.Errorf("applied_amt: must not be nil"))
	}

	return errors.Join(errs...)
}

// Validate validates the BucketBalance type
func (bb *BucketBalance) Validate() error {
	var errs []error
	if bb.BucketTypeId <= 0 {
		errs = append(errs, fmt.Errorf("bucket_type_id: must be a positive integer"))
	}

	if bb.BalanceAmt.IsNil() {
		errs = append(errs, fmt.Errorf("balance_amt: must not be nil"))
	}

	return errors.Join(errs...)
}

// Validate validates the LedgerAndEntries type
func (lte *LedgerAndEntries) Validate() error {
	// Validate the ledger key and ledger.
	//  - One or both must be non-nil.
	//  - If both are non-nil: validate each.
	//  - key and ledger.key must be the same.
	var errs []error
	switch {
	case lte.LedgerKey == nil && lte.Ledger == nil:
		return fmt.Errorf("ledger_key, ledger: one or both must be non-nil")
	case lte.LedgerKey != nil && lte.Ledger == nil:
		if err := lte.LedgerKey.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("ledger_key: %w", err))
		}
	case lte.LedgerKey == nil && lte.Ledger != nil:
		if err := lte.Ledger.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("ledger: %w", err))
		}
	case lte.LedgerKey != nil && lte.Ledger != nil:
		if err := lte.LedgerKey.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("ledger_key: %w", err))
		}
		if err := lte.Ledger.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("ledger: %w", err))
		}
		if lte.LedgerKey.String() != lte.Ledger.Key.String() {
			errs = append(errs, fmt.Errorf("ledger_key, ledger.key: must be the same value"))
		}
	}

	for i, entry := range lte.Entries {
		if err := entry.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("entries[%d]: %w", i, err))
		}
	}

	return errors.Join(errs...)
}

// UnmarshalJSON implements json.Unmarshaler for DayCount.
func (d *DayCountConvention) UnmarshalJSON(data []byte) error {
	value, err := enumUnmarshalJSON(data, DayCountConvention_value, DayCountConvention_name)
	if err != nil {
		return err
	}
	*d = DayCountConvention(value)
	return nil
}

// Validate returns an error if this DayCountConvention isn't a defined enum entry.
func (d DayCountConvention) Validate() error {
	return enumValidateExists(d, DayCountConvention_name)
}

// UnmarshalJSON implements json.Unmarshaler for InterestAccrual.
func (i *InterestAccrualMethod) UnmarshalJSON(data []byte) error {
	value, err := enumUnmarshalJSON(data, InterestAccrualMethod_value, InterestAccrualMethod_name)
	if err != nil {
		return err
	}
	*i = InterestAccrualMethod(value)
	return nil
}

// Validate returns an error if this InterestAccrualMethod isn't a defined enum entry.
func (i InterestAccrualMethod) Validate() error {
	return enumValidateExists(i, InterestAccrualMethod_name)
}

// UnmarshalJSON implements json.Unmarshaler for PaymentFrequency.
func (p *PaymentFrequency) UnmarshalJSON(data []byte) error {
	value, err := enumUnmarshalJSON(data, PaymentFrequency_value, PaymentFrequency_name)
	if err != nil {
		return err
	}
	*p = PaymentFrequency(value)
	return nil
}

// Validate returns an error if this PaymentFrequency isn't a defined enum entry.
func (p PaymentFrequency) Validate() error {
	return enumValidateExists(p, PaymentFrequency_name)
}
