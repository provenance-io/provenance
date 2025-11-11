package types

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/provenance-io/provenance/internal/provutils"
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
	if lc == nil {
		return fmt.Errorf("ledger class cannot be nil")
	}

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
	} else if err := sdk.ValidateDenom(lc.Denom); err != nil {
		// Validate denom format (should be a valid coin denomination)
		errs = append(errs, fmt.Errorf("denom must be a valid coin denomination: %w", err))
	}

	if _, err := sdk.AccAddressFromBech32(lc.MaintainerAddress); err != nil {
		errs = append(errs, fmt.Errorf("maintainer_address: %w", err))
	}

	return errors.Join(errs...)
}

// Validate validates the LedgerClassEntryType type
func (lcet *LedgerClassEntryType) Validate() error {
	if lcet == nil {
		return fmt.Errorf("ledger class entry type cannot be nil")
	}

	var errs []error
	if lcet.Id < 0 {
		errs = append(errs, fmt.Errorf("id: %d must be a non-negative integer", lcet.Id))
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
	if lcst == nil {
		return fmt.Errorf("ledger class status type cannot be nil")
	}

	var errs []error
	if lcst.Id < 0 {
		errs = append(errs, fmt.Errorf("id: %d must be a non-negative integer", lcst.Id))
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

// StringToLedgerKey converts a bech32 string to a LedgerKey.
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
	if l == nil {
		return fmt.Errorf("ledger cannot be nil")
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

	// Validate the next payment date and amount and the payment frequency.
	if err := ValidatePmtFields(l.NextPmtDate, l.NextPmtAmt); err != nil {
		errs = append(errs, err)
	}

	if err := l.PaymentFrequency.ValidateSpecified(); err != nil {
		errs = append(errs, fmt.Errorf("payment_frequency: %w", err))
	}

	// Validate interest rate if provided (reasonable bounds: 0-100,000,000 for 0-100%)
	if l.InterestRate < 0 || l.InterestRate > 100_000_000 {
		errs = append(errs, fmt.Errorf("interest_rate: must be between 0 and 100,000,000 (0-100%%)"))
	}

	// Validate maturity date format if provided
	if l.MaturityDate < 0 {
		errs = append(errs, fmt.Errorf("maturity_date: must be after 1970-01-01"))
	}

	if err := l.InterestDayCountConvention.ValidateSpecified(); err != nil {
		errs = append(errs, fmt.Errorf("interest_day_count_convention: %w", err))
	}

	if err := l.InterestAccrualMethod.ValidateSpecified(); err != nil {
		errs = append(errs, fmt.Errorf("interest_accrual_method: %w", err))
	}

	return errors.Join(errs...)
}

// ValidatePmtFields returns an error if any of the provided fields have invalid values.
func ValidatePmtFields(nextPmtDate int32, nextPmtAmt sdkmath.Int) error {
	var errs []error
	// Validate the next payment date. Allow zero to indicate "not provided."
	if nextPmtDate < 0 {
		errs = append(errs, fmt.Errorf("next_pmt_date: must be after 1970-01-01"))
	}

	// NextPmtAmt is allowed to be nil (not provided), zero, or positive; but not negative.
	if !nextPmtAmt.IsNil() && nextPmtAmt.IsNegative() {
		errs = append(errs, fmt.Errorf("next_pmt_amt: must be a non-negative integer"))
	}

	return errors.Join(errs...)
}

// Validate validates the LedgerClassBucketType type
func (lcbt *LedgerClassBucketType) Validate() error {
	if lcbt == nil {
		return fmt.Errorf("ledger class bucket type cannot be nil")
	}

	var errs []error
	if lcbt.Id < 0 {
		errs = append(errs, fmt.Errorf("id: %d must be a non-negative integer", lcbt.Id))
	}

	if err := lenCheck(lcbt.Code, 1, MaxLenCode); err != nil {
		errs = append(errs, fmt.Errorf("code: %w", err))
	}

	if err := lenCheck(lcbt.Description, 1, MaxLenDescription); err != nil {
		errs = append(errs, fmt.Errorf("description: %w", err))
	}

	return errors.Join(errs...)
}

// Compare returns -1 if this is < b, 0 if this == b, and 1 if this is > b.
func (le *LedgerEntry) Compare(b *LedgerEntry) int {
	if le == b {
		return 0
	}
	// nils are greatest (sorts them to the end).
	if b == nil {
		return -1
	}
	if le == nil {
		return 1
	}

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

// Validate validates the LedgerEntry type
func (le *LedgerEntry) Validate() error {
	if le == nil {
		return fmt.Errorf("ledger entry cannot be nil")
	}

	var errs []error
	if err := lenCheck(le.CorrelationId, 1, MaxLenCorrelationID); err != nil {
		errs = append(errs, fmt.Errorf("correlation_id: %w", err))
	}

	// Validate reverses_correlation_id if provided
	if err := lenCheck(le.ReversesCorrelationId, 0, MaxLenCorrelationID); err != nil {
		errs = append(errs, fmt.Errorf("reverses_correlation_id: %w", err))
	}

	// Validate sequence number (should be < 300 as per proto comment)
	if err := ValidateSequence(le.Sequence); err != nil {
		errs = append(errs, err)
	}

	// Validate entry_type_id is positive
	if le.EntryTypeId <= 0 {
		errs = append(errs, fmt.Errorf("entry_type_id: must be a positive integer"))
	}

	if le.PostedDate <= 0 {
		errs = append(errs, fmt.Errorf("posted_date: must be a positive integer"))
	}

	if le.EffectiveDate <= 0 {
		errs = append(errs, fmt.Errorf("effective_date: must be a positive integer"))
	}

	if err := ValidateLedgerEntryAmounts(le.TotalAmt, le.AppliedAmounts, le.BalanceAmounts); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// ValidateSequence returns an error if the sequence number is too large.
func ValidateSequence(seq uint32) error {
	if seq > MaxLedgerEntrySequence {
		return fmt.Errorf("sequence: cannot be more than %d", MaxLedgerEntrySequence)
	}
	return nil
}

// ValidateLedgerEntryAmounts returns an error if there is anything wrong with the provided amounts.
func ValidateLedgerEntryAmounts(totalAmt sdkmath.Int, appliedAmounts []*LedgerBucketAmount, balanceAmounts []*BucketBalance) error {
	var errs []error

	// Make sure the total is valid and if not zero, that there are applied amounts.
	if totalAmt.IsNil() || totalAmt.IsNegative() {
		errs = append(errs, fmt.Errorf("total_amt: must be a non-negative integer"))
	} else if !totalAmt.IsZero() && len(appliedAmounts) == 0 {
		errs = append(errs, fmt.Errorf("applied_amounts: cannot be empty"))
	}

	// Make sure all the applied amounts are valid.
	if err := validateSlice(appliedAmounts, "applied_amounts"); err != nil {
		errs = append(errs, err)
	}

	// If there isn't anything wrong yet, verify that the total matches the applied amounts.
	if len(errs) == 0 {
		if err := ValidateEntryAmounts(totalAmt, appliedAmounts); err != nil {
			errs = append(errs, fmt.Errorf("applied_amounts: %w", err))
		}
	}

	// Make sure all the balance amounts are valid.
	if err := validateSlice(balanceAmounts, "balance_amounts"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ValidateEntryAmounts checks if the amounts are valid
func ValidateEntryAmounts(totalAmt sdkmath.Int, appliedAmounts []*LedgerBucketAmount) error {
	// Check if the total amount matches the sum of applied amounts.
	totalApplied := sdkmath.NewInt(0)
	for _, applied := range appliedAmounts {
		totalApplied = totalApplied.Add(applied.AppliedAmt)
	}

	if !totalAmt.Equal(totalApplied.Abs()) {
		return fmt.Errorf("total amount must equal abs(sum of applied amounts)")
	}

	return nil
}

// validateSlice runs the Validate() method on each element of the slice and returns all errors it encounters.
// The name parameter is used in the error messages.
func validateSlice[S ~[]E, E interface{ Validate() error }](vals S, name string) error {
	var errs []error
	for i, val := range vals {
		if err := val.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("%s[%d]: %w", name, i, err))
		}
	}
	return errors.Join(errs...)
}

// Validate validates the LedgerBucketAmount type
func (lba *LedgerBucketAmount) Validate() error {
	if lba == nil {
		return fmt.Errorf("ledger bucket amount cannot be nil")
	}

	var errs []error
	if lba.BucketTypeId < 0 {
		errs = append(errs, fmt.Errorf("bucket_type_id: must be a non-negative integer"))
	}

	if lba.AppliedAmt.IsNil() {
		errs = append(errs, fmt.Errorf("applied_amt: must not be nil"))
	}

	return errors.Join(errs...)
}

// Validate validates the BucketBalance type
func (bb *BucketBalance) Validate() error {
	if bb == nil {
		return fmt.Errorf("bucket balance cannot be nil")
	}
	var errs []error
	if bb.BucketTypeId < 0 {
		errs = append(errs, fmt.Errorf("bucket_type_id: must be a non-negative integer"))
	}

	if bb.BalanceAmt.IsNil() {
		errs = append(errs, fmt.Errorf("balance_amt: must not be nil"))
	}

	return errors.Join(errs...)
}

// Validate validates the LedgerAndEntries type
func (lte *LedgerAndEntries) Validate() error {
	if lte == nil {
		return fmt.Errorf("ledger and entries cannot be nil")
	}

	// Validate the ledger key and ledger.
	//  - One or both must be non-nil.
	//  - If both are non-nil: validate each.
	//  - key and ledger.key must be the same.
	var errs []error
	switch {
	case lte.LedgerKey == nil && lte.Ledger == nil:
		return fmt.Errorf("a ledger or ledger_key is required")
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
		if len(errs) == 0 && !lte.LedgerKey.Equals(lte.Ledger.Key) {
			errs = append(errs, fmt.Errorf("ledger_key and ledger.key must be the same"))
		}
	}

	if err := validateSlice(lte.Entries, "entries"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// UnmarshalJSON implements json.Unmarshaler for DayCount.
func (d *DayCountConvention) UnmarshalJSON(data []byte) error {
	value, err := provutils.EnumUnmarshalJSON(data, DayCountConvention_value, DayCountConvention_name)
	if err != nil {
		return err
	}
	*d = DayCountConvention(value)
	return nil
}

// Validate returns an error if this DayCountConvention isn't a defined enum entry.
func (d DayCountConvention) Validate() error {
	return provutils.EnumValidateExists(d, DayCountConvention_name)
}

// ValidateSpecified returns an error if this DayCountConvention isn't a defined enum entry or is the zero (UNSPECIFIED) value).
func (d DayCountConvention) ValidateSpecified() error {
	return provutils.EnumValidateSpecified(d, DayCountConvention_name)
}

// UnmarshalJSON implements json.Unmarshaler for InterestAccrual.
func (i *InterestAccrualMethod) UnmarshalJSON(data []byte) error {
	value, err := provutils.EnumUnmarshalJSON(data, InterestAccrualMethod_value, InterestAccrualMethod_name)
	if err != nil {
		return err
	}
	*i = InterestAccrualMethod(value)
	return nil
}

// Validate returns an error if this InterestAccrualMethod isn't a defined enum entry.
func (i InterestAccrualMethod) Validate() error {
	return provutils.EnumValidateExists(i, InterestAccrualMethod_name)
}

// ValidateSpecified returns an error if this InterestAccrualMethod isn't a defined enum entry or is the zero (UNSPECIFIED) value).
func (i InterestAccrualMethod) ValidateSpecified() error {
	return provutils.EnumValidateSpecified(i, InterestAccrualMethod_name)
}

// UnmarshalJSON implements json.Unmarshaler for PaymentFrequency.
func (p *PaymentFrequency) UnmarshalJSON(data []byte) error {
	value, err := provutils.EnumUnmarshalJSON(data, PaymentFrequency_value, PaymentFrequency_name)
	if err != nil {
		return err
	}
	*p = PaymentFrequency(value)
	return nil
}

// Validate returns an error if this PaymentFrequency isn't a defined enum entry.
func (p PaymentFrequency) Validate() error {
	return provutils.EnumValidateExists(p, PaymentFrequency_name)
}

// ValidateSpecified returns an error if this PaymentFrequency isn't a defined enum entry or is the zero (UNSPECIFIED) value).
func (p PaymentFrequency) ValidateSpecified() error {
	return provutils.EnumValidateSpecified(p, PaymentFrequency_name)
}
