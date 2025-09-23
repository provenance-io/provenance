package types

import (
	"fmt"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

const (
	MaxLedgerEntrySequence = 99
	MaxLenLedgerClassID    = 50
	MaxLenCorrelationID    = 50
	MaxLenAssetClassID     = registrytypes.MaxLenAssetClassID
	MaxLenNFTID            = registrytypes.MaxLenNFTID
	MaxLenCode             = 50
	MaxLenDescription      = 100
	MaxLenDenom            = 128
	MaxLenMemo             = 50
)

var alNumDashRx = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// Validate validates the LedgerClass type
func (lc *LedgerClass) Validate() error {
	if err := lenCheck("ledger_class_id", lc.LedgerClassId, 1, MaxLenLedgerClassID); err != nil {
		return err
	}

	// Verify that the ledger class only contains alphanumeric and dashes
	if !alNumDashRx.MatchString(lc.LedgerClassId) {
		return NewErrCodeInvalidField("ledger_class_id", "must only contain alphanumeric and dashes")
	}

	if err := lenCheck("asset_class_id", lc.AssetClassId, 1, MaxLenAssetClassID); err != nil {
		return err
	}

	// Validate asset_class_id format (should be a valid UUID or similar format)
	if !alNumDashRx.MatchString(lc.AssetClassId) {
		return NewErrCodeInvalidField("asset_class_id", "must only contain alphanumeric and dashes")
	}

	// Check denom length first for nicer error messages.
	if err := lenCheck("denom", lc.Denom, 2, MaxLenDenom); err != nil {
		return err
	}

	// Validate denom format (should be a valid coin denomination)
	if err := sdk.ValidateDenom(lc.Denom); err != nil {
		return NewErrCodeInvalidField("denom", fmt.Sprintf("must be a valid coin denomination: %v", err))
	}

	if err := validateAccAddress("maintainer_address", lc.MaintainerAddress); err != nil {
		return err
	}

	return nil
}

// Validate validates the LedgerClassEntryType type
func (lcet *LedgerClassEntryType) Validate() error {
	if lcet.Id < 0 {
		return NewErrCodeInvalidField("id", "must be a non-negative integer")
	}

	if err := lenCheck("code", lcet.Code, 1, MaxLenCode); err != nil {
		return err
	}

	if err := lenCheck("description", lcet.Description, 1, MaxLenDescription); err != nil {
		return err
	}

	return nil
}

// Validate validates the LedgerClassStatusType type
func (lcst *LedgerClassStatusType) Validate() error {
	if lcst.Id < 0 {
		return NewErrCodeInvalidField("id", "must be a non-negative integer")
	}

	if err := lenCheck("code", lcst.Code, 1, MaxLenCode); err != nil {
		return err
	}

	if err := lenCheck("description", lcst.Description, 1, MaxLenDescription); err != nil {
		return err
	}

	return nil
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
		return NewErrCodeMissingField("key")
	}

	// Verify that the nft_id and asset_class_id do not contain a null byte
	if strings.Contains(lk.NftId, "\x00") {
		return NewErrCodeInvalidField("nft_id", "must not contain a null byte")
	}

	if err := lenCheck("nft_id", lk.NftId, 1, MaxLenNFTID); err != nil {
		return err
	}

	if err := lenCheck("asset_class_id", lk.AssetClassId, 1, MaxLenAssetClassID); err != nil {
		return err
	}

	if strings.Contains(lk.AssetClassId, "\x00") {
		return NewErrCodeInvalidField("asset_class_id", "must not contain a null byte")
	}

	return nil
}

// Validate validates the Ledger type
func (l *Ledger) Validate() error {
	if l.Key == nil {
		return NewErrCodeMissingField("key")
	}

	if err := l.Key.Validate(); err != nil {
		return err
	}

	// Validate the LedgerClassId field
	if err := lenCheck("ledger_class_id", l.LedgerClassId, 1, MaxLenLedgerClassID); err != nil {
		return err
	}

	// Validate status_type_id is positive
	if l.StatusTypeId <= 0 {
		return NewErrCodeInvalidField("status_type_id", "must be a positive integer")
	}

	// Validate next payment date format if provided
	if l.NextPmtDate <= 0 {
		return NewErrCodeInvalidField("next_pmt_date", "must be after 1970-01-01")
	}

	// Validate next payment amount if provided
	if l.NextPmtAmt.IsNegative() {
		return NewErrCodeInvalidField("next_pmt_amt", "must be a non-negative integer")
	}

	// Validate interest rate if provided (reasonable bounds: 0-100000000 for 0-100%)
	if l.InterestRate < 0 || l.InterestRate > 100_000_000 {
		return NewErrCodeInvalidField("interest_rate", "must be between 0 and 100,000,000 (0-100%)")
	}

	// Validate maturity date format if provided
	if l.MaturityDate < 0 {
		return NewErrCodeInvalidField("maturity_date", "must be after 1970-01-01")
	}

	if err := l.InterestDayCountConvention.Validate(); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := l.InterestAccrualMethod.Validate(); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := l.PaymentFrequency.Validate(); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return nil
}

// Validate validates the LedgerClassBucketType type
func (lcbt *LedgerClassBucketType) Validate() error {
	if lcbt.Id < 0 {
		return NewErrCodeInvalidField("id", "must be a non-negative integer")
	}

	if err := lenCheck("code", lcbt.Code, 1, MaxLenCode); err != nil {
		return err
	}

	if err := lenCheck("description", lcbt.Description, 1, MaxLenDescription); err != nil {
		return err
	}

	return nil
}

// Validate validates the LedgerEntry type
func (le *LedgerEntry) Validate() error {
	if err := lenCheck("correlation_id", le.CorrelationId, 1, MaxLenCorrelationID); err != nil {
		return err
	}

	// Validate reverses_correlation_id if provided
	if err := lenCheck("reverses_correlation_id", le.ReversesCorrelationId, 0, MaxLenCorrelationID); err != nil {
		return err
	}

	// Validate sequence number (should be < 100 as per proto comment)
	if le.Sequence >= MaxLedgerEntrySequence {
		return NewErrCodeInvalidField("sequence", fmt.Sprintf("must be less than %d", MaxLedgerEntrySequence))
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
	if le.TotalAmt.IsNil() || le.TotalAmt.IsNegative() {
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

// Validate validates the LedgerBucketAmount type
func (lba *LedgerBucketAmount) Validate() error {
	if lba.BucketTypeId <= 0 {
		return NewErrCodeInvalidField("bucket_type_id", "must be a positive integer")
	}

	if lba.AppliedAmt.IsNil() || lba.AppliedAmt.IsNegative() {
		return NewErrCodeInvalidField("applied_amt", "must be a non-negative integer")
	}

	return nil
}

// Validate validates the BucketBalance type
func (bb *BucketBalance) Validate() error {
	if bb.BucketTypeId <= 0 {
		return NewErrCodeInvalidField("bucket_type_id", "must be a positive integer")
	}

	if bb.BalanceAmt.IsNil() || bb.BalanceAmt.IsNegative() {
		return NewErrCodeInvalidField("balance_amt", "must be a non-negative integer")
	}

	return nil
}

// Validate validates the LedgerAndEntries type
func (lte *LedgerAndEntries) Validate() error {
	if err := lte.LedgerKey.Validate(); err != nil {
		return err
	}

	if err := lte.Ledger.Validate(); err != nil {
		return NewErrCodeInvalidField("ledger", err.Error())
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
