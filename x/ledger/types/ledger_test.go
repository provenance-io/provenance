package types_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"

	. "github.com/provenance-io/provenance/x/ledger/types"
)

// joinErrs returns a string with the provided strings joined by newlines.
// This mimics what errors.Join produces.
func joinErrs(errs ...string) string {
	return strings.Join(errs, "\n")
}

func TestLedgerClass_Validate(t *testing.T) {
	tests := []struct {
		name        string
		ledgerClass *LedgerClass
		expErr      string
	}{
		{
			name:        "nil",
			ledgerClass: nil,
			expErr:      "ledger class cannot be nil",
		},
		{
			name: "valid",
			ledgerClass: &LedgerClass{
				LedgerClassId:     "abc123def",
				AssetClassId:      "ghi456jkl",
				Denom:             "banana",
				MaintainerAddress: sdk.AccAddress("MaintainerAddress___").String(),
			},
		},
		{
			name: "invalid ledger class id",
			ledgerClass: &LedgerClass{
				LedgerClassId:     "abc 123 def",
				AssetClassId:      "ghi456jkl",
				Denom:             "banana",
				MaintainerAddress: sdk.AccAddress("MaintainerAddress___").String(),
			},
			expErr: "ledger_class_id: \"abc 123 def\" must only contain alphanumeric, '-', '.' characters",
		},
		{
			name: "invalid asset class id",
			ledgerClass: &LedgerClass{
				LedgerClassId:     "abc123def",
				AssetClassId:      "ghi 456 jkl",
				Denom:             "banana",
				MaintainerAddress: sdk.AccAddress("MaintainerAddress___").String(),
			},
			expErr: "asset_class_id: \"ghi 456 jkl\" must only contain alphanumeric, '-', '.' characters",
		},
		{
			name: "invalid denom: too long",
			ledgerClass: &LedgerClass{
				LedgerClassId:     "abc123def",
				AssetClassId:      "ghi456jkl",
				Denom:             strings.Repeat("v", 129),
				MaintainerAddress: sdk.AccAddress("MaintainerAddress___").String(),
			},
			expErr: "denom: must be between 2 and 128 characters",
		},
		{
			name: "invalid denom: invalid chars",
			ledgerClass: &LedgerClass{
				LedgerClassId:     "abc123def",
				AssetClassId:      "ghi456jkl",
				Denom:             "ban ana",
				MaintainerAddress: sdk.AccAddress("MaintainerAddress___").String(),
			},
			expErr: "denom must be a valid coin denomination: invalid denom: ban ana",
		},
		{
			name: "invalid maintainer address",
			ledgerClass: &LedgerClass{
				LedgerClassId:     "abc123def",
				AssetClassId:      "ghi456jkl",
				Denom:             "banana",
				MaintainerAddress: "MaintainerAddress",
			},
			expErr: "maintainer_address: decoding bech32 failed: string not all lowercase or all uppercase",
		},
		{
			name: "multiple errors",
			ledgerClass: &LedgerClass{
				LedgerClassId:     "abc 123 def",
				AssetClassId:      "ghi 456 jkl",
				Denom:             "ban ana",
				MaintainerAddress: "MaintainerAddress",
			},
			expErr: joinErrs("ledger_class_id: \"abc 123 def\" must only contain alphanumeric, '-', '.' characters",
				"asset_class_id: \"ghi 456 jkl\" must only contain alphanumeric, '-', '.' characters",
				"denom must be a valid coin denomination: invalid denom: ban ana",
				"maintainer_address: decoding bech32 failed: string not all lowercase or all uppercase"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.ledgerClass.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.ledgerClass)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.ledgerClass)
		})
	}
}

func TestLedgerClassEntryType_Validate(t *testing.T) {
	tests := []struct {
		name   string
		entry  *LedgerClassEntryType
		expErr string
	}{
		{
			name:   "nil",
			entry:  nil,
			expErr: "ledger class entry type cannot be nil",
		},
		{
			name: "valid: normal id",
			entry: &LedgerClassEntryType{
				Id:          3,
				Code:        "just-some-code",
				Description: "description for the code",
			},
		},
		{
			name: "valid: 0 id",
			entry: &LedgerClassEntryType{
				Id:          0,
				Code:        "just-some-code",
				Description: "description for the code",
			},
		},
		{
			name: "negative Id",
			entry: &LedgerClassEntryType{
				Id:          -2,
				Code:        "just-some-code",
				Description: "description for the code",
			},
			expErr: "id: -2 must be a non-negative integer",
		},
		{
			name: "code empty",
			entry: &LedgerClassEntryType{
				Id:          3,
				Code:        "",
				Description: "description for the code",
			},
			expErr: "code: must be between 1 and 50 characters",
		},
		{
			name: "code too long",
			entry: &LedgerClassEntryType{
				Id:          3,
				Code:        strings.Repeat("c", 51),
				Description: "description for the code",
			},
			expErr: "code: must be between 1 and 50 characters",
		},
		{
			name: "description empty",
			entry: &LedgerClassEntryType{
				Id:          3,
				Code:        "just-some-code",
				Description: "",
			},
			expErr: "description: must be between 1 and 100 characters",
		},
		{
			name: "description too long",
			entry: &LedgerClassEntryType{
				Id:          3,
				Code:        "just-some-code",
				Description: strings.Repeat("D", 101),
			},
			expErr: "description: must be between 1 and 100 characters",
		},
		{
			name: "multiple errors",
			entry: &LedgerClassEntryType{
				Id:          -7,
				Code:        "",
				Description: "",
			},
			expErr: joinErrs("id: -7 must be a non-negative integer",
				"code: must be between 1 and 50 characters",
				"description: must be between 1 and 100 characters"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.entry.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.entry)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.entry)
		})
	}
}

func TestLedgerClassStatusType_Validate(t *testing.T) {
	tests := []struct {
		name   string
		entry  *LedgerClassStatusType
		expErr string
	}{
		{
			name:   "nil",
			entry:  nil,
			expErr: "ledger class status type cannot be nil",
		},
		{
			name: "valid: normal id",
			entry: &LedgerClassStatusType{
				Id:          3,
				Code:        "just-some-code",
				Description: "description for the code",
			},
		},
		{
			name: "valid: 0 id",
			entry: &LedgerClassStatusType{
				Id:          0,
				Code:        "just-some-code",
				Description: "description for the code",
			},
		},
		{
			name: "negative Id",
			entry: &LedgerClassStatusType{
				Id:          -2,
				Code:        "just-some-code",
				Description: "description for the code",
			},
			expErr: "id: -2 must be a non-negative integer",
		},
		{
			name: "code empty",
			entry: &LedgerClassStatusType{
				Id:          3,
				Code:        "",
				Description: "description for the code",
			},
			expErr: "code: must be between 1 and 50 characters",
		},
		{
			name: "code too long",
			entry: &LedgerClassStatusType{
				Id:          3,
				Code:        strings.Repeat("c", 51),
				Description: "description for the code",
			},
			expErr: "code: must be between 1 and 50 characters",
		},
		{
			name: "description empty",
			entry: &LedgerClassStatusType{
				Id:          3,
				Code:        "just-some-code",
				Description: "",
			},
			expErr: "description: must be between 1 and 100 characters",
		},
		{
			name: "description too long",
			entry: &LedgerClassStatusType{
				Id:          3,
				Code:        "just-some-code",
				Description: strings.Repeat("D", 101),
			},
			expErr: "description: must be between 1 and 100 characters",
		},
		{
			name: "multiple errors",
			entry: &LedgerClassStatusType{
				Id:          -7,
				Code:        "",
				Description: "",
			},
			expErr: joinErrs("id: -7 must be a non-negative integer",
				"code: must be between 1 and 50 characters",
				"description: must be between 1 and 100 characters"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.entry.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.entry)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.entry)
		})
	}
}

func TestLedgerKey_String(t *testing.T) {
	uuid1, err := uuid.Parse("12345678-1234-1234-1234-123456789abc")
	require.NoError(t, err, "uuid.Parse(\"12345678-1234-1234-1234-123456789abc\")")
	uuid2, err := uuid.Parse("87654321-4321-4321-4321-cba987654321")
	require.NoError(t, err, "uuid.Parse(\"87654321-4321-4321-4321-cba987654321\")")

	tests := []struct {
		name string
		key  LedgerKey
		exp  string
	}{
		{
			name: "scope spec and scope",
			key: LedgerKey{
				NftId:        metadatatypes.ScopeMetadataAddress(uuid1).String(),
				AssetClassId: metadatatypes.ScopeSpecMetadataAddress(uuid2).String(),
			},
			exp: "ledger1wd3k7ur9wdcx2ce3w948y6ejwdjhqemkwv6hseejwfunswfkdecx6wt8weehxump0pnkxacqwd3k7ur9x9chzenjvu6xucm6vum8q7tyw948sumxwfnngmnrdcerwutc8948wmrnkujj7q",
		},
		{
			name: "nft class and nft",
			key:  LedgerKey{NftId: "this-is-the-nft-id", AssetClassId: "but-this-is-the-asset-class-id"},
			exp:  "ledger1vf6hgtt5dp5hxttfwvkhg6r994shxum9wskkxmrpwdej66tyqp6xs6tn945hxtt5dpjj6mnxwskkjeqwkyvue",
		},
		{
			name: "empty asset class id",
			key:  LedgerKey{NftId: "", AssetClassId: "this-only-has-an-asset-class-id"},
			exp:  "ledger1w35xjueddahxc7fddpshxttpdckkzumnv46z6cmvv9ehxttfvsqqzg64fc",
		},
		{
			name: "empty nft id",
			key:  LedgerKey{NftId: "this-one-only-has-an-nft-id", AssetClassId: ""},
			exp:  "ledger1qp6xs6tn94hkuefddahxc7fddpshxttpdckkuen5945kged9d95",
		},
		{
			name: "both empty",
			key:  LedgerKey{NftId: "", AssetClassId: ""},
			exp:  "ledger1qqkam643",
		},
		{
			name: "simple",
			key: LedgerKey{
				NftId:        "def",
				AssetClassId: "abc",
			},
			exp: "ledger1v93xxqryv4nqcgu2nz",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.key.String()
			}
			require.NotPanics(t, testFunc, "%T.String()", tc.key)
			assert.Equal(t, tc.exp, act, "%T.String() result", tc.key)

			// Make sure that it converts back to the exact same thing.
			var key *LedgerKey
			var err error
			testRetFunc := func() {
				key, err = StringToLedgerKey(act)
			}
			require.NotPanics(t, testRetFunc, "StringToLedgerKey(%q)", act)
			require.NoError(t, err, "StringToLedgerKey(%q) error", act)
			assert.Equal(t, &tc.key, key, "StringToLedgerKey(%q) result", act)
		})
	}
}

func TestStringToLedgerKey(t *testing.T) {
	uuid1, err := uuid.Parse("12345678-1234-1234-1234-123456789abc")
	require.NoError(t, err, "uuid.Parse(\"12345678-1234-1234-1234-123456789abc\")")
	uuid2, err := uuid.Parse("87654321-4321-4321-4321-cba987654321")
	require.NoError(t, err, "uuid.Parse(\"87654321-4321-4321-4321-cba987654321\")")

	tests := []struct {
		name   string
		s      string
		exp    *LedgerKey
		expErr string
	}{
		{
			name:   "not a bech32",
			s:      "oops",
			expErr: "decoding bech32 failed: invalid bech32 string length 4",
		},
		{
			name:   "wrong hrp",
			s:      "lodger1vf6hgtt5dp5hxttfwvkhg6r994shxum9wskkxmrpwdej66tyqp6xs6tn945hxtt5dpjj6mnxwskkjeqvfdlxa",
			expErr: "invalid hrp: lodger",
		},
		{
			name:   "no null byte",
			s:      "ledger1v93xxer9vcarn0xz",
			expErr: "invalid key: ledger1v93xxer9vcarn0xz",
		},
		{
			name:   "two null bytes",
			s:      "ledger1v93xxqryv4nqqemgdyvj5q4p",
			expErr: "invalid key: ledger1v93xxqryv4nqqemgdyvj5q4p",
		},
		{
			name: "okay: scope",
			s:    "ledger1wd3k7ur9wdcx2ce3w948y6ejwdjhqemkwv6hseejwfunswfkdecx6wt8weehxump0pnkxacqwd3k7ur9x9chzenjvu6xucm6vum8q7tyw948sumxwfnngmnrdcerwutc8948wmrnkujj7q",
			exp: &LedgerKey{
				NftId:        metadatatypes.ScopeMetadataAddress(uuid1).String(),
				AssetClassId: metadatatypes.ScopeSpecMetadataAddress(uuid2).String(),
			},
			expErr: "",
		},
		{
			name: "okay: nft",
			s:    "ledger1v93xxqryv4nqcgu2nz",
			exp: &LedgerKey{
				NftId:        "def",
				AssetClassId: "abc",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *LedgerKey
			var err error
			testFunc := func() {
				act, err = StringToLedgerKey(tc.s)
			}
			require.NotPanics(t, testFunc, "StringToLedgerKey(%q)", tc.s)
			assertions.AssertErrorValue(t, err, tc.expErr, "StringToLedgerKey(%q) error", tc.s)
			assert.Equal(t, tc.exp, act, "StringToLedgerKey(%q) result", tc.s)

			// If it was supposed to convert okay (and did), make sure it converts back to the same thing.
			if act == nil || tc.exp == nil || len(tc.expErr) > 0 {
				return
			}
			var str string
			testRetFunc := func() {
				str = act.String()
			}
			require.NotPanics(t, testRetFunc, "act.String()")
			assert.Equal(t, tc.s, str, "act.String() result")

		})
	}
}

func TestLedgerKey_ToRegistryKey(t *testing.T) {
	uuid1, err := uuid.Parse("12345678-1234-1234-1234-123456789abc")
	require.NoError(t, err, "uuid.Parse(\"12345678-1234-1234-1234-123456789abc\")")
	uuid2, err := uuid.Parse("87654321-4321-4321-4321-cba987654321")
	require.NoError(t, err, "uuid.Parse(\"87654321-4321-4321-4321-cba987654321\")")

	tests := []struct {
		name string
		lk   LedgerKey
		exp  *registrytypes.RegistryKey
	}{
		{
			name: "empty",
			lk:   LedgerKey{},
			exp:  &registrytypes.RegistryKey{},
		},
		{
			name: "scopes",
			lk: LedgerKey{
				NftId:        metadatatypes.ScopeMetadataAddress(uuid1).String(),
				AssetClassId: metadatatypes.ScopeSpecMetadataAddress(uuid2).String(),
			},
			exp: &registrytypes.RegistryKey{
				NftId:        metadatatypes.ScopeMetadataAddress(uuid1).String(),
				AssetClassId: metadatatypes.ScopeSpecMetadataAddress(uuid2).String(),
			},
		},
		{
			name: "nft",
			lk: LedgerKey{
				NftId:        "the-nft-id",
				AssetClassId: "the-asset-class-id",
			},
			exp: &registrytypes.RegistryKey{
				NftId:        "the-nft-id",
				AssetClassId: "the-asset-class-id",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *registrytypes.RegistryKey
			testFunc := func() {
				act = tc.lk.ToRegistryKey()
			}
			require.NotPanics(t, testFunc, "%T.ToRegistryKey()", tc.lk)
			assert.Equal(t, tc.exp, act, "%T.ToRegistryKey() result", tc.lk)
		})
	}
}

func TestLedgerKey_Validate(t *testing.T) {
	uuid1, err := uuid.Parse("12345678-1234-1234-1234-123456789abc")
	require.NoError(t, err, "uuid.Parse(\"12345678-1234-1234-1234-123456789abc\")")
	uuid2, err := uuid.Parse("87654321-4321-4321-4321-cba987654321")
	require.NoError(t, err, "uuid.Parse(\"87654321-4321-4321-4321-cba987654321\")")

	tests := []struct {
		name   string
		lk     *LedgerKey
		expErr string
	}{
		{
			name:   "nil",
			lk:     nil,
			expErr: "key cannot be nil",
		},
		{
			name: "valid nft",
			lk: &LedgerKey{
				NftId:        "the-nft",
				AssetClassId: "the-class",
			},
		},
		{
			name: "valid scope",
			lk: &LedgerKey{
				NftId:        metadatatypes.ScopeMetadataAddress(uuid1).String(),
				AssetClassId: metadatatypes.ScopeSpecMetadataAddress(uuid2).String(),
			},
		},
		{
			name: "invalid class id",
			lk: &LedgerKey{
				NftId:        "the-nft",
				AssetClassId: "",
			},
			expErr: "invalid asset_class_id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "invalid nft id",
			lk: &LedgerKey{
				NftId:        "",
				AssetClassId: "the-class",
			},
			expErr: "invalid nft_id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "multiple errors",
			lk:   &LedgerKey{},
			expErr: joinErrs("invalid asset_class_id: must be between 1 and 128 characters: invalid field",
				"invalid nft_id: must be between 1 and 128 characters: invalid field"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.lk.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.lk)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.lk)
		})
	}
}

func TestLedgerKey_Equals(t *testing.T) {
	lk := LedgerKey{
		NftId:        "the-nft-1",
		AssetClassId: "the-class-1",
	}

	tests := []struct {
		name  string
		lk    *LedgerKey
		other *LedgerKey
		exp   bool
	}{
		{name: "both nil", lk: nil, other: nil, exp: true},
		{name: "nil receiver", lk: nil, other: &lk, exp: false},
		{name: "nil other", lk: &lk, other: nil, exp: false},
		{name: "same objects", lk: &lk, other: &lk, exp: true},
		{
			name:  "different objects with same values",
			lk:    &LedgerKey{NftId: "the-nft-1", AssetClassId: "the-class-1"},
			other: &LedgerKey{NftId: "the-nft-1", AssetClassId: "the-class-1"},
			exp:   true,
		},
		{
			name:  "same nft, different class",
			lk:    &LedgerKey{NftId: "the-nft-1", AssetClassId: "the-class-1"},
			other: &LedgerKey{NftId: "the-nft-1", AssetClassId: "the-class-2"},
			exp:   false,
		},
		{
			name:  "same class, different nft",
			lk:    &LedgerKey{NftId: "the-nft-1", AssetClassId: "the-class-1"},
			other: &LedgerKey{NftId: "the-nft-2", AssetClassId: "the-class-1"},
			exp:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act bool
			testFunc := func() {
				act = tc.lk.Equals(tc.other)
			}
			require.NotPanics(t, testFunc, "Equals")
			assert.Equal(t, tc.exp, act, "Equals result")
		})
	}
}

func TestLedger_Validate(t *testing.T) {
	tests := []struct {
		name   string
		l      *Ledger
		expErr string
	}{
		{
			name:   "nil",
			l:      nil,
			expErr: "ledger cannot be nil",
		},
		{
			name: "valid",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "",
		},
		{
			name: "invalid key",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "key: invalid nft_id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "invalid ledger class id",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "ledger_class_id: must be between 1 and 128 characters",
		},
		{
			name: "negative status type",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               -1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "status_type_id: must be a positive integer",
		},
		{
			name: "invalid next pmt date",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                -2,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "next_pmt_date: must be after 1970-01-01",
		},
		{
			name: "invalid next pmt amount",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(-3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "next_pmt_amt: must be a non-negative integer",
		},
		{
			name: "invalid payment frequency",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           -1,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "payment_frequency: unknown payment_frequency enum value: -1",
		},
		{
			name: "negative interest rate",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               -5,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "interest_rate: must be between 0 and 100,000,000 (0-100%)",
		},
		{
			name: "too large interest rate",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               100_000_001,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "interest_rate: must be between 0 and 100,000,000 (0-100%)",
		},
		{
			name: "negative maturity date",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               -5,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "maturity_date: must be after 1970-01-01",
		},
		{
			name: "invalid interest day count convention",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: -1,
				InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING,
			},
			expErr: "interest_day_count_convention: unknown day_count_convention enum value: -1",
		},
		{
			name: "invalid interest accrual method",
			l: &Ledger{
				Key:                        &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               3_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      -1,
			},
			expErr: "interest_accrual_method: unknown interest_accrual_method enum value: -1",
		},
		{
			name: "multiple errors",
			l: &Ledger{
				Key:                        nil,
				LedgerClassId:              "ledger-class-id",
				StatusTypeId:               1,
				NextPmtDate:                20384,
				NextPmtAmt:                 sdkmath.NewInt(3),
				PaymentFrequency:           PAYMENT_FREQUENCY_DAILY,
				InterestRate:               103_400_000,
				MaturityDate:               21000,
				InterestDayCountConvention: DAY_COUNT_CONVENTION_ACTUAL_365,
				InterestAccrualMethod:      15,
			},
			expErr: joinErrs("key: key cannot be nil",
				"interest_rate: must be between 0 and 100,000,000 (0-100%)",
				"interest_accrual_method: unknown interest_accrual_method enum value: 15"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.l.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.l)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.l)
		})
	}
}

func TestValidatePmtFields(t *testing.T) {
	tests := []struct {
		name             string
		nextPmtDate      int32
		nextPmtAmt       sdkmath.Int
		paymentFrequency PaymentFrequency
		expErr           string
	}{
		{
			name:             "valid: baseline",
			nextPmtDate:      20444,
			nextPmtAmt:       sdkmath.NewInt(5000),
			paymentFrequency: PAYMENT_FREQUENCY_ANNUALLY,
		},
		{
			name:             "valid: nil next pmt amt",
			nextPmtDate:      20445,
			nextPmtAmt:       sdkmath.Int{},
			paymentFrequency: PAYMENT_FREQUENCY_QUARTERLY,
		},
		{
			name:             "valid: zero next pmt amt",
			nextPmtDate:      20,
			nextPmtAmt:       sdkmath.ZeroInt(),
			paymentFrequency: PAYMENT_FREQUENCY_MONTHLY,
		},
		{
			name:             "valid: unspecified payment frequency",
			nextPmtDate:      10444,
			nextPmtAmt:       sdkmath.NewInt(43382),
			paymentFrequency: PAYMENT_FREQUENCY_UNSPECIFIED,
		},
		{
			name:             "negative next pmt date",
			nextPmtDate:      -1,
			nextPmtAmt:       sdkmath.NewInt(123),
			paymentFrequency: PAYMENT_FREQUENCY_WEEKLY,
			expErr:           "next_pmt_date: must be after 1970-01-01",
		},
		{
			name:             "negative next pmt amt",
			nextPmtDate:      20432,
			nextPmtAmt:       sdkmath.NewInt(-1),
			paymentFrequency: PAYMENT_FREQUENCY_DAILY,
			expErr:           "next_pmt_amt: must be a non-negative integer",
		},
		{
			name:             "invalid payment frequency",
			nextPmtDate:      12345,
			nextPmtAmt:       sdkmath.NewInt(1),
			paymentFrequency: 12,
			expErr:           "payment_frequency: unknown payment_frequency enum value: 12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidatePmtFields(tc.nextPmtDate, tc.nextPmtAmt, tc.paymentFrequency)
			}
			require.NotPanics(t, testFunc, "ValidatePmtFields(%d, %s, %s)", tc.nextPmtDate, tc.nextPmtAmt, tc.paymentFrequency)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidatePmtFields(%d, %s, %s) error", tc.nextPmtDate, tc.nextPmtAmt, tc.paymentFrequency)
		})
	}
}

func TestLedgerClassBucketType_Validate(t *testing.T) {
	tests := []struct {
		name   string
		entry  *LedgerClassBucketType
		expErr string
	}{
		{
			name:   "nil",
			entry:  nil,
			expErr: "ledger class bucket type cannot be nil",
		},
		{
			name: "valid: normal id",
			entry: &LedgerClassBucketType{
				Id:          3,
				Code:        "just-some-code",
				Description: "description for the code",
			},
		},
		{
			name: "valid: 0 id",
			entry: &LedgerClassBucketType{
				Id:          0,
				Code:        "just-some-code",
				Description: "description for the code",
			},
		},
		{
			name: "negative Id",
			entry: &LedgerClassBucketType{
				Id:          -2,
				Code:        "just-some-code",
				Description: "description for the code",
			},
			expErr: "id: -2 must be a non-negative integer",
		},
		{
			name: "code empty",
			entry: &LedgerClassBucketType{
				Id:          3,
				Code:        "",
				Description: "description for the code",
			},
			expErr: "code: must be between 1 and 50 characters",
		},
		{
			name: "code too long",
			entry: &LedgerClassBucketType{
				Id:          3,
				Code:        strings.Repeat("c", 51),
				Description: "description for the code",
			},
			expErr: "code: must be between 1 and 50 characters",
		},
		{
			name: "description empty",
			entry: &LedgerClassBucketType{
				Id:          3,
				Code:        "just-some-code",
				Description: "",
			},
			expErr: "description: must be between 1 and 100 characters",
		},
		{
			name: "description too long",
			entry: &LedgerClassBucketType{
				Id:          3,
				Code:        "just-some-code",
				Description: strings.Repeat("D", 101),
			},
			expErr: "description: must be between 1 and 100 characters",
		},
		{
			name: "multiple errors",
			entry: &LedgerClassBucketType{
				Id:          -7,
				Code:        "",
				Description: "",
			},
			expErr: joinErrs("id: -7 must be a non-negative integer",
				"code: must be between 1 and 50 characters",
				"description: must be between 1 and 100 characters"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.entry.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.entry)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.entry)
		})
	}
}

func TestLedgerEntry_Compare(t *testing.T) {
	le := LedgerEntry{EffectiveDate: 20000, Sequence: 5}

	tests := []struct {
		name string
		le   *LedgerEntry
		b    *LedgerEntry
		exp  int
	}{
		{name: "both nil", le: nil, b: nil, exp: 0},
		{name: "this nil", le: nil, b: &le, exp: 1},
		{name: "that nil", le: &le, b: nil, exp: -1},
		{name: "same refs", le: &le, b: &le, exp: 0},
		{
			name: "this has lesser effective date",
			le:   &LedgerEntry{EffectiveDate: 20000, Sequence: 6},
			b:    &LedgerEntry{EffectiveDate: 20001, Sequence: 5},
			exp:  -1,
		},
		{
			name: "that has lesser effective date",
			le:   &LedgerEntry{EffectiveDate: 20001, Sequence: 5},
			b:    &LedgerEntry{EffectiveDate: 20000, Sequence: 6},
			exp:  1,
		},
		{
			name: "same effective date: this has smaller sequence",
			le:   &LedgerEntry{EffectiveDate: 1, Sequence: 5},
			b:    &LedgerEntry{EffectiveDate: 1, Sequence: 6},
			exp:  -1,
		},
		{
			name: "same effective date: that has smaller sequence",
			le:   &LedgerEntry{EffectiveDate: 1, Sequence: 7},
			b:    &LedgerEntry{EffectiveDate: 1, Sequence: 6},
			exp:  1,
		},
		{
			name: "same effective date and sequence",
			le:   &LedgerEntry{EffectiveDate: 1, Sequence: 6},
			b:    &LedgerEntry{EffectiveDate: 1, Sequence: 6},
			exp:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act int
			testFunc := func() {
				act = tc.le.Compare(tc.b)
			}
			require.NotPanics(t, testFunc, "%T.Compare(...)", tc.le)
			assert.Equal(t, tc.exp, act, "%T.Compare(...)", tc.le)
		})
	}
}

func TestLedgerEntry_Validate(t *testing.T) {
	tests := []struct {
		name   string
		le     *LedgerEntry
		expErr string
	}{
		{
			name:   "nil",
			le:     nil,
			expErr: "ledger entry cannot be nil",
		},
		{
			name: "valid: base",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
		},
		{
			name: "no correlation id",
			le: &LedgerEntry{
				CorrelationId:  "",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "correlation_id: must be between 1 and 50 characters",
		},
		{
			name: "correlation id too long",
			le: &LedgerEntry{
				CorrelationId:  strings.Repeat("c", 51),
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "correlation_id: must be between 1 and 50 characters",
		},
		{
			name: "reverse correlation id too long",
			le: &LedgerEntry{
				CorrelationId:         "abcdefgh-ijkl",
				ReversesCorrelationId: strings.Repeat("r", 51),
				Sequence:              1,
				EntryTypeId:           4,
				PostedDate:            20200,
				EffectiveDate:         20199,
				TotalAmt:              sdkmath.NewInt(4),
				AppliedAmounts:        []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts:        []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "reverses_correlation_id: must be between 0 and 50 characters",
		},
		{
			name: "invalid sequence",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       300,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "sequence: cannot be more than 299",
		},
		{
			name: "no entry type",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    0,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "entry_type_id: must be a positive integer",
		},
		{
			name: "negative entry type",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    -1,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "entry_type_id: must be a positive integer",
		},
		{
			name: "no posted date",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     0,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "posted_date: must be a positive integer",
		},
		{
			name: "negative posted date",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     -1,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "posted_date: must be a positive integer",
		},
		{
			name: "no effective date",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  0,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "effective_date: must be a positive integer",
		},
		{
			name: "negative effective date",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  -1,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "effective_date: must be a positive integer",
		},
		{
			name: "invalid total amt",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.Int{},
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "total_amt: must be a non-negative integer",
		},
		{
			name: "invalid applied amount",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.Int{}}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "applied_amounts[0]: applied_amt: must not be nil",
		},
		{
			name: "invalid balance amount",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.Int{}}},
			},
			expErr: "balance_amounts[0]: balance_amt: must not be nil",
		},
		{
			name: "total does not equal sum of applied amounts",
			le: &LedgerEntry{
				CorrelationId:  "abcdefgh-ijkl",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(100),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.NewInt(4)}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.NewInt(9996)}},
			},
			expErr: "applied_amounts: total amount must equal abs(sum of applied amounts)",
		},
		{
			name: "multiple errors",
			le: &LedgerEntry{
				CorrelationId:  "",
				Sequence:       1,
				EntryTypeId:    4,
				PostedDate:     20200,
				EffectiveDate:  20199,
				TotalAmt:       sdkmath.NewInt(4),
				AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 3, AppliedAmt: sdkmath.Int{}}},
				BalanceAmounts: []*BucketBalance{{BucketTypeId: 3, BalanceAmt: sdkmath.Int{}}},
			},
			expErr: joinErrs("correlation_id: must be between 1 and 50 characters",
				"applied_amounts[0]: applied_amt: must not be nil",
				"balance_amounts[0]: balance_amt: must not be nil"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.le.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.le)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.le)
		})
	}
}

func TestValidateSequence(t *testing.T) {
	tests := []struct {
		name   string
		seq    uint32
		expErr string
	}{
		{name: "zero", seq: 0},
		{name: "max", seq: MaxLedgerEntrySequence},
		{name: "max + 1", seq: MaxLedgerEntrySequence + 1, expErr: "sequence: cannot be more than 299"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateSequence(tc.seq)
			}
			require.NotPanics(t, testFunc, "ValidateSequence(%d)", tc.seq)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateSequence(%d) error", tc.seq)
		})
	}
}

// TODO: func TestValidateLedgerEntryAmounts(t *testing.T) {}

// TODO: func TestValidateEntryAmounts(t *testing.T) {}

func TestLedgerBucketAmount_Validate(t *testing.T) {
	lba := func(bucketTypeID int32, appliedAmt int64) *LedgerBucketAmount {
		return &LedgerBucketAmount{
			BucketTypeId: bucketTypeID,
			AppliedAmt:   sdkmath.NewInt(appliedAmt),
		}
	}

	tests := []struct {
		name   string
		lba    *LedgerBucketAmount
		expErr string
	}{
		{
			name:   "nil",
			lba:    nil,
			expErr: "ledger bucket amount cannot be nil",
		},
		{name: "valid: base", lba: lba(4, 12345)},
		{
			name: "zero bucket id",
			lba:  lba(0, 12345),
		},
		{name: "zero applied amount", lba: lba(1, 0)},
		{name: "negative applied amount", lba: lba(1, -54321)},
		{name: "positive applied amount", lba: lba(2, 12345)},
		{
			name:   "negative bucket id",
			lba:    lba(-1, 12345),
			expErr: "bucket_type_id: must be a non-negative integer",
		},
		{
			name:   "nil applied amount",
			lba:    &LedgerBucketAmount{BucketTypeId: 3, AppliedAmt: sdkmath.Int{}},
			expErr: "applied_amt: must not be nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.lba.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.lba)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.lba)
		})
	}
}

func TestBucketBalance_Validate(t *testing.T) {
	bb := func(bucketTypeID int32, balanceAmt int64) *BucketBalance {
		return &BucketBalance{
			BucketTypeId: bucketTypeID,
			BalanceAmt:   sdkmath.NewInt(balanceAmt),
		}
	}

	tests := []struct {
		name   string
		lba    *BucketBalance
		expErr string
	}{
		{
			name:   "nil",
			lba:    nil,
			expErr: "bucket balance cannot be nil",
		},
		{name: "valid: base", lba: bb(4, 12345)},
		{
			name: "zero bucket id",
			lba:  bb(0, 12345),
		},
		{name: "zero balance amount", lba: bb(1, 0)},
		{name: "negative balance amount", lba: bb(1, -54321)},
		{name: "positive balance amount", lba: bb(2, 12345)},
		{
			name:   "negative bucket id",
			lba:    bb(-1, 12345),
			expErr: "bucket_type_id: must be a non-negative integer",
		},
		{
			name:   "nil balance amount",
			lba:    &BucketBalance{BucketTypeId: 3, BalanceAmt: sdkmath.Int{}},
			expErr: "balance_amt: must not be nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.lba.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.lba)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.lba)
		})
	}
}

// TODO: func TestLedgerAndEntries_Validate(t *testing.T) {}

func TestDayCountConvention_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		exp    DayCountConvention
		expErr string
	}{
		{
			name:   "empty string",
			data:   "",
			expErr: "day_count_convention must be a string or integer, got: \"\"",
		},
		{
			name:   "unknown data type",
			data:   "not-right",
			expErr: "day_count_convention must be a string or integer, got: \"not-right\"",
		},
		{
			name:   "unknown int: negative",
			data:   "-1",
			expErr: "unknown day_count_convention integer value: -1",
		},
		{
			name:   "unknown name",
			data:   `"unknown"`,
			expErr: "unknown day_count_convention string value: \"unknown\"",
		},

		// DAY_COUNT_CONVENTION_UNSPECIFIED
		{name: "unspecified: long: upper case", data: `"DAY_COUNT_CONVENTION_UNSPECIFIED"`, exp: DAY_COUNT_CONVENTION_UNSPECIFIED},
		{name: "unspecified: long: lower case", data: `"day_count_convention_unspecified"`, exp: DAY_COUNT_CONVENTION_UNSPECIFIED},
		{name: "unspecified: long: mixed case", data: `"Day_CounT_ConvenTion_unspeCified"`, exp: DAY_COUNT_CONVENTION_UNSPECIFIED},
		{name: "unspecified: short: upper case", data: `"UNSPECIFIED"`, exp: DAY_COUNT_CONVENTION_UNSPECIFIED},
		{name: "unspecified: short: lower case", data: `"unspecified"`, exp: DAY_COUNT_CONVENTION_UNSPECIFIED},
		{name: "unspecified: short: mixed case", data: `"uNspecIFied"`, exp: DAY_COUNT_CONVENTION_UNSPECIFIED},
		{name: "unspecified: int", data: "0", exp: DAY_COUNT_CONVENTION_UNSPECIFIED},

		// DAY_COUNT_CONVENTION_ACTUAL_365
		{name: "actual_365: long: lower case", data: `"day_count_convention_actual_365"`, exp: DAY_COUNT_CONVENTION_ACTUAL_365},
		{name: "actual_365: long: upper case", data: `"DAY_COUNT_CONVENTION_ACTUAL_365"`, exp: DAY_COUNT_CONVENTION_ACTUAL_365},
		{name: "actual_365: long: mixed case", data: `"DAy_CouNT_CoNVeNTioN_ACTuaL_365"`, exp: DAY_COUNT_CONVENTION_ACTUAL_365},
		{name: "actual_365: short: upper case", data: `"ACTUAL_365"`, exp: DAY_COUNT_CONVENTION_ACTUAL_365},
		{name: "actual_365: short: lower case", data: `"actual_365"`, exp: DAY_COUNT_CONVENTION_ACTUAL_365},
		{name: "actual_365: short: mixed case", data: `"ActuAl_365"`, exp: DAY_COUNT_CONVENTION_ACTUAL_365},
		{name: "actual_365: int", data: "1", exp: DAY_COUNT_CONVENTION_ACTUAL_365},

		// DAY_COUNT_CONVENTION_ACTUAL_360
		{name: "actual_360: long: upper case", data: `"DAY_COUNT_CONVENTION_ACTUAL_360"`, exp: DAY_COUNT_CONVENTION_ACTUAL_360},
		{name: "actual_360: long: lower case", data: `"day_count_convention_actual_360"`, exp: DAY_COUNT_CONVENTION_ACTUAL_360},
		{name: "actual_360: long: mixed case", data: `"dAY_cOUNT_cONVENTION_aCTUAL_360"`, exp: DAY_COUNT_CONVENTION_ACTUAL_360},
		{name: "actual_360: short: upper case", data: `"ACTUAL_360"`, exp: DAY_COUNT_CONVENTION_ACTUAL_360},
		{name: "actual_360: short: lower case", data: `"actual_360"`, exp: DAY_COUNT_CONVENTION_ACTUAL_360},
		{name: "actual_360: short: mixed case", data: `"aCTUal_360"`, exp: DAY_COUNT_CONVENTION_ACTUAL_360},
		{name: "actual_360: int", data: "2", exp: DAY_COUNT_CONVENTION_ACTUAL_360},

		// DAY_COUNT_CONVENTION_THIRTY_360
		{name: "thirty_360: long: upper case", data: `"DAY_COUNT_CONVENTION_THIRTY_360"`, exp: DAY_COUNT_CONVENTION_THIRTY_360},
		{name: "thirty_360: long: lower case", data: `"day_count_convention_thirty_360"`, exp: DAY_COUNT_CONVENTION_THIRTY_360},
		{name: "thirty_360: long: mixed case", data: `"Day_Count_Convention_Thirty_360"`, exp: DAY_COUNT_CONVENTION_THIRTY_360},
		{name: "thirty_360: short: upper case", data: `"THIRTY_360"`, exp: DAY_COUNT_CONVENTION_THIRTY_360},
		{name: "thirty_360: short: lower case", data: `"thirty_360"`, exp: DAY_COUNT_CONVENTION_THIRTY_360},
		{name: "thirty_360: short: mixed case", data: `"thirTy_360"`, exp: DAY_COUNT_CONVENTION_THIRTY_360},
		{name: "thirty_360: int", data: "3", exp: DAY_COUNT_CONVENTION_THIRTY_360},

		// DAY_COUNT_CONVENTION_ACTUAL_ACTUAL
		{name: "actual_actual: long: upper case", data: `"DAY_COUNT_CONVENTION_ACTUAL_ACTUAL"`, exp: DAY_COUNT_CONVENTION_ACTUAL_ACTUAL},
		{name: "actual_actual: long: lower case", data: `"day_count_convention_actual_actual"`, exp: DAY_COUNT_CONVENTION_ACTUAL_ACTUAL},
		{name: "actual_actual: long: mixed case", data: `"day_COUNT_convention_ACTUAL_actual"`, exp: DAY_COUNT_CONVENTION_ACTUAL_ACTUAL},
		{name: "actual_actual: short: upper case", data: `"ACTUAL_ACTUAL"`, exp: DAY_COUNT_CONVENTION_ACTUAL_ACTUAL},
		{name: "actual_actual: short: lower case", data: `"actual_actual"`, exp: DAY_COUNT_CONVENTION_ACTUAL_ACTUAL},
		{name: "actual_actual: short: mixed case", data: `"actual_ACTUAL"`, exp: DAY_COUNT_CONVENTION_ACTUAL_ACTUAL},
		{name: "actual_actual: int", data: "4", exp: DAY_COUNT_CONVENTION_ACTUAL_ACTUAL},

		// DAY_COUNT_CONVENTION_DAYS_365
		{name: "days_365: long: upper case", data: `"DAY_COUNT_CONVENTION_DAYS_365"`, exp: DAY_COUNT_CONVENTION_DAYS_365},
		{name: "days_365: long: lower case", data: `"day_count_convention_days_365"`, exp: DAY_COUNT_CONVENTION_DAYS_365},
		{name: "days_365: long: mixed case", data: `"day_couNT_COnvention_days_365"`, exp: DAY_COUNT_CONVENTION_DAYS_365},
		{name: "days_365: short: upper case", data: `"DAYS_365"`, exp: DAY_COUNT_CONVENTION_DAYS_365},
		{name: "days_365: short: lower case", data: `"days_365"`, exp: DAY_COUNT_CONVENTION_DAYS_365},
		{name: "days_365: short: mixed case", data: `"dAYs_365"`, exp: DAY_COUNT_CONVENTION_DAYS_365},
		{name: "days_365: int", data: "5", exp: DAY_COUNT_CONVENTION_DAYS_365},

		// DAY_COUNT_CONVENTION_DAYS_360
		{name: "days_360: long: upper case", data: `"DAY_COUNT_CONVENTION_DAYS_360"`, exp: DAY_COUNT_CONVENTION_DAYS_360},
		{name: "days_360: long: lower case", data: `"day_count_convention_days_360"`, exp: DAY_COUNT_CONVENTION_DAYS_360},
		{name: "days_360: long: mixed case", data: `"day_COUNT_convention_daYs_360"`, exp: DAY_COUNT_CONVENTION_DAYS_360},
		{name: "days_360: short: upper case", data: `"DAYS_360"`, exp: DAY_COUNT_CONVENTION_DAYS_360},
		{name: "days_360: short: lower case", data: `"days_360"`, exp: DAY_COUNT_CONVENTION_DAYS_360},
		{name: "days_360: short: mixed case", data: `"dayS_360"`, exp: DAY_COUNT_CONVENTION_DAYS_360},
		{name: "days_360: int", data: "6", exp: DAY_COUNT_CONVENTION_DAYS_360},

		{
			name:   "unknown int: too large",
			data:   "7",
			expErr: "unknown day_count_convention integer value: 7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var val DayCountConvention
			var err error
			testFunc := func() {
				err = val.UnmarshalJSON([]byte(tc.data))
			}
			require.NotPanics(t, testFunc, "%T.UnmarshalJSON(%q)", val, tc.data)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.UnmarshalJSON(%q) error", val, tc.data)
			assert.Equal(t, tc.exp, val, "%T.UnmarshalJSON(%q) resulting value", val, tc.data)
		})
	}
}

func TestDayCountConvention_Validate(t *testing.T) {
	tests := []struct {
		name   string
		d      DayCountConvention
		expErr string
	}{
		{name: "negative one", d: -1, expErr: "unknown day_count_convention enum value: -1"},
		{name: "unspecified", d: DAY_COUNT_CONVENTION_UNSPECIFIED},
		{name: "actual_365", d: DAY_COUNT_CONVENTION_ACTUAL_365},
		{name: "actual_360", d: DAY_COUNT_CONVENTION_ACTUAL_360},
		{name: "thirty_360", d: DAY_COUNT_CONVENTION_THIRTY_360},
		{name: "actual_actual", d: DAY_COUNT_CONVENTION_ACTUAL_ACTUAL},
		{name: "days_365", d: DAY_COUNT_CONVENTION_DAYS_365},
		{name: "days_360", d: DAY_COUNT_CONVENTION_DAYS_360},
		{name: "seven", d: 7, expErr: "unknown day_count_convention enum value: 7"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.d.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.d)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.d)
		})
	}
}

func TestInterestAccrualMethod_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		exp    InterestAccrualMethod
		expErr string
	}{
		{
			name:   "empty string",
			data:   "",
			expErr: "interest_accrual_method must be a string or integer, got: \"\"",
		},
		{
			name:   "unknown data type",
			data:   "not-right",
			expErr: "interest_accrual_method must be a string or integer, got: \"not-right\"",
		},
		{
			name:   "unknown int: negative",
			data:   "-1",
			expErr: "unknown interest_accrual_method integer value: -1",
		},
		{
			name:   "unknown name",
			data:   `"unknown"`,
			expErr: "unknown interest_accrual_method string value: \"unknown\"",
		},

		// INTEREST_ACCRUAL_METHOD_UNSPECIFIED
		{name: "unspecified: long: upper case", data: `"INTEREST_ACCRUAL_METHOD_UNSPECIFIED"`, exp: INTEREST_ACCRUAL_METHOD_UNSPECIFIED},
		{name: "unspecified: long: lower case", data: `"interest_accrual_method_unspecified"`, exp: INTEREST_ACCRUAL_METHOD_UNSPECIFIED},
		{name: "unspecified: long: mixed case", data: `"intereST_Accrual_metHOD_Unspecified"`, exp: INTEREST_ACCRUAL_METHOD_UNSPECIFIED},
		{name: "unspecified: short: upper case", data: `"UNSPECIFIED"`, exp: INTEREST_ACCRUAL_METHOD_UNSPECIFIED},
		{name: "unspecified: short: lower case", data: `"unspecified"`, exp: INTEREST_ACCRUAL_METHOD_UNSPECIFIED},
		{name: "unspecified: short: mixed case", data: `"unsPEcified"`, exp: INTEREST_ACCRUAL_METHOD_UNSPECIFIED},
		{name: "unspecified: int", data: "0", exp: INTEREST_ACCRUAL_METHOD_UNSPECIFIED},

		// INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST
		{name: "simple_interest: long: upper case", data: `"INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST"`, exp: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST},
		{name: "simple_interest: long: lower case", data: `"interest_accrual_method_simple_interest"`, exp: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST},
		{name: "simple_interest: long: mixed case", data: `"INTErest_accRUAL_METHOD_SIMple_inteREST"`, exp: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST},
		{name: "simple_interest: short: upper case", data: `"SIMPLE_INTEREST"`, exp: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST},
		{name: "simple_interest: short: lower case", data: `"simple_interest"`, exp: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST},
		{name: "simple_interest: short: mixed case", data: `"siMPLE_INterest"`, exp: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST},
		{name: "simple_interest: int", data: "1", exp: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST},

		// INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST
		{name: "compound_interest: long: upper case", data: `"INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST"`, exp: INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST},
		{name: "compound_interest: long: lower case", data: `"interest_accrual_method_compound_interest"`, exp: INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST},
		{name: "compound_interest: long: mixed case", data: `"inTEREST_ACCRual_method_compounD_INterEst"`, exp: INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST},
		{name: "compound_interest: short: upper case", data: `"COMPOUND_INTEREST"`, exp: INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST},
		{name: "compound_interest: short: lower case", data: `"compound_interest"`, exp: INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST},
		{name: "compound_interest: short: mixed case", data: `"CompoUNd_inTereST"`, exp: INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST},
		{name: "compound_interest: int", data: "2", exp: INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST},

		// INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING
		{name: "daily_compounding: long: upper case", data: `"INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING},
		{name: "daily_compounding: long: lower case", data: `"interest_accrual_method_daily_compounding"`, exp: INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING},
		{name: "daily_compounding: long: mixed case", data: `"inteRest_accrual_method_daiLY_COmpounding"`, exp: INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING},
		{name: "daily_compounding: short: upper case", data: `"DAILY_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING},
		{name: "daily_compounding: short: lower case", data: `"daily_compounding"`, exp: INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING},
		{name: "daily_compounding: short: mixed case", data: `"DailY_COmpoundiNG"`, exp: INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING},
		{name: "daily_compounding: int", data: "3", exp: INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING},

		// INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING
		{name: "monthly_compounding: long: upper case", data: `"INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING},
		{name: "monthly_compounding: long: short case", data: `"interest_accrual_method_monthly_compounding"`, exp: INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING},
		{name: "monthly_compounding: long: mixed case", data: `"inteRest_accrual_method_monthLy_compouNding"`, exp: INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING},
		{name: "monthly_compounding: short: upper case", data: `"MONTHLY_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING},
		{name: "monthly_compounding: short: short case", data: `"monthly_compounding"`, exp: INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING},
		{name: "monthly_compounding: short: mixed case", data: `"Monthly_Compounding"`, exp: INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING},
		{name: "monthly_compounding: int", data: "4", exp: INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING},

		// INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING
		{name: "quarterly_compounding: long: upper case", data: `"INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING},
		{name: "quarterly_compounding: long: lower case", data: `"interest_accrual_method_quarterly_compounding"`, exp: INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING},
		{name: "quarterly_compounding: long: mixed case", data: `"INTerest_ACCrual_METhod_QUArterly_COMpounding"`, exp: INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING},
		{name: "quarterly_compounding: short: upper case", data: `"QUARTERLY_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING},
		{name: "quarterly_compounding: short: lower case", data: `"quarterly_compounding"`, exp: INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING},
		{name: "quarterly_compounding: short: mixed case", data: `"Quarterly_comPOUndinG"`, exp: INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING},
		{name: "quarterly_compounding: int", data: "5", exp: INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING},

		// INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING
		{name: "annual_compounding: long: upper case", data: `"INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING},
		{name: "annual_compounding: long: lower case", data: `"interest_accrual_method_annual_compounding"`, exp: INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING},
		{name: "annual_compounding: long: mixed case", data: `"iNTEREST_aCCRUAL_mETHOD_aNNUAL_cOMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING},
		{name: "annual_compounding: short: upper case", data: `"ANNUAL_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING},
		{name: "annual_compounding: short: lower case", data: `"annual_compounding"`, exp: INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING},
		{name: "annual_compounding: short: mixed case", data: `"AnNuAl_CoMpOuNdInG"`, exp: INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING},
		{name: "annual_compounding: int", data: "6", exp: INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING},

		// INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING
		{name: "continuous_compounding: long: upper case", data: `"INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING},
		{name: "continuous_compounding: long: lower case", data: `"interest_accrual_method_continuous_compounding"`, exp: INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING},
		{name: "continuous_compounding: long: mixed case", data: `"interest_accrual_method_CONTINUOUS_compounding"`, exp: INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING},
		{name: "continuous_compounding: short: upper case", data: `"CONTINUOUS_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING},
		{name: "continuous_compounding: short: lower case", data: `"continuous_compounding"`, exp: INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING},
		{name: "continuous_compounding: short: mixed case", data: `"continuous_COMPOUNDING"`, exp: INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING},
		{name: "continuous_compounding: int", data: "7", exp: INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING},

		{
			name:   "unknown int: too large",
			data:   "8",
			expErr: "unknown interest_accrual_method integer value: 8",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var val InterestAccrualMethod
			var err error
			testFunc := func() {
				err = val.UnmarshalJSON([]byte(tc.data))
			}
			require.NotPanics(t, testFunc, "%T.UnmarshalJSON(%q)", val, tc.data)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.UnmarshalJSON(%q) error", val, tc.data)
			assert.Equal(t, tc.exp, val, "%T.UnmarshalJSON(%q) resulting value", val, tc.data)
		})
	}
}

func TestInterestAccrualMethod_Validate(t *testing.T) {
	tests := []struct {
		name   string
		i      InterestAccrualMethod
		expErr string
	}{
		{name: "negative one", i: -1, expErr: "unknown interest_accrual_method enum value: -1"},
		{name: "unspecified", i: INTEREST_ACCRUAL_METHOD_UNSPECIFIED},
		{name: "simple_interest", i: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST},
		{name: "compound_interest", i: INTEREST_ACCRUAL_METHOD_COMPOUND_INTEREST},
		{name: "daily_compounding", i: INTEREST_ACCRUAL_METHOD_DAILY_COMPOUNDING},
		{name: "monthly_compounding", i: INTEREST_ACCRUAL_METHOD_MONTHLY_COMPOUNDING},
		{name: "quarterly_compounding", i: INTEREST_ACCRUAL_METHOD_QUARTERLY_COMPOUNDING},
		{name: "annual_compounding", i: INTEREST_ACCRUAL_METHOD_ANNUAL_COMPOUNDING},
		{name: "continuous_compounding", i: INTEREST_ACCRUAL_METHOD_CONTINUOUS_COMPOUNDING},
		{name: "eight", i: 8, expErr: "unknown interest_accrual_method enum value: 8"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.i.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.i)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.i)
		})
	}
}

func TestPaymentFrequency_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		exp    PaymentFrequency
		expErr string
	}{
		{
			name:   "empty string",
			data:   "",
			expErr: "payment_frequency must be a string or integer, got: \"\"",
		},
		{
			name:   "unknown data type",
			data:   "not-right",
			expErr: "payment_frequency must be a string or integer, got: \"not-right\"",
		},
		{
			name:   "unknown int: negative",
			data:   "-1",
			expErr: "unknown payment_frequency integer value: -1",
		},
		{
			name:   "unknown name",
			data:   `"unknown"`,
			expErr: "unknown payment_frequency string value: \"unknown\"",
		},

		// PAYMENT_FREQUENCY_UNSPECIFIED
		{name: "unspecified: long: upper case", data: `"PAYMENT_FREQUENCY_UNSPECIFIED"`, exp: PAYMENT_FREQUENCY_UNSPECIFIED},
		{name: "unspecified: long: lower case", data: `"payment_frequency_unspecified"`, exp: PAYMENT_FREQUENCY_UNSPECIFIED},
		{name: "unspecified: long: mixed case", data: `"payment_freQUENCY_UNSPecified"`, exp: PAYMENT_FREQUENCY_UNSPECIFIED},
		{name: "unspecified: short: upper case", data: `"UNSPECIFIED"`, exp: PAYMENT_FREQUENCY_UNSPECIFIED},
		{name: "unspecified: short: lower case", data: `"unspecified"`, exp: PAYMENT_FREQUENCY_UNSPECIFIED},
		{name: "unspecified: short: mixed case", data: `"unspECified"`, exp: PAYMENT_FREQUENCY_UNSPECIFIED},
		{name: "unspecified: int", data: "0", exp: PAYMENT_FREQUENCY_UNSPECIFIED},

		// PAYMENT_FREQUENCY_DAILY
		{name: "daily: long: upper case", data: `"PAYMENT_FREQUENCY_DAILY"`, exp: PAYMENT_FREQUENCY_DAILY},
		{name: "daily: long: lower case", data: `"payment_frequency_daily"`, exp: PAYMENT_FREQUENCY_DAILY},
		{name: "daily: long: mixed case", data: `"payment_frequency_daily"`, exp: PAYMENT_FREQUENCY_DAILY},
		{name: "daily: short: upper case", data: `"DAILY"`, exp: PAYMENT_FREQUENCY_DAILY},
		{name: "daily: short: lower case", data: `"daily"`, exp: PAYMENT_FREQUENCY_DAILY},
		{name: "daily: short: mixed case", data: `"DaiLy"`, exp: PAYMENT_FREQUENCY_DAILY},
		{name: "daily: int", data: "1", exp: PAYMENT_FREQUENCY_DAILY},

		// PAYMENT_FREQUENCY_WEEKLY
		{name: "weekly: long: upper case", data: `"PAYMENT_FREQUENCY_WEEKLY"`, exp: PAYMENT_FREQUENCY_WEEKLY},
		{name: "weekly: long: lower case", data: `"payment_frequency_weekly"`, exp: PAYMENT_FREQUENCY_WEEKLY},
		{name: "weekly: long: mixed case", data: `"paymENT_FrequeNCY_Weekly"`, exp: PAYMENT_FREQUENCY_WEEKLY},
		{name: "weekly: short: upper case", data: `"WEEKLY"`, exp: PAYMENT_FREQUENCY_WEEKLY},
		{name: "weekly: short: lower case", data: `"weekly"`, exp: PAYMENT_FREQUENCY_WEEKLY},
		{name: "weekly: short: mixed case", data: `"wEEkly"`, exp: PAYMENT_FREQUENCY_WEEKLY},
		{name: "weekly: int", data: "2", exp: PAYMENT_FREQUENCY_WEEKLY},

		// PAYMENT_FREQUENCY_MONTHLY
		{name: "monthly: long: upper case", data: `"PAYMENT_FREQUENCY_MONTHLY"`, exp: PAYMENT_FREQUENCY_MONTHLY},
		{name: "monthly: long: lower case", data: `"payment_frequency_monthly"`, exp: PAYMENT_FREQUENCY_MONTHLY},
		{name: "monthly: long: mixed case", data: `"paymENT_FrequencY_MontHLy"`, exp: PAYMENT_FREQUENCY_MONTHLY},
		{name: "monthly: short: upper case", data: `"MONTHLY"`, exp: PAYMENT_FREQUENCY_MONTHLY},
		{name: "monthly: short: lower case", data: `"monthly"`, exp: PAYMENT_FREQUENCY_MONTHLY},
		{name: "monthly: short: mixed case", data: `"monTHly"`, exp: PAYMENT_FREQUENCY_MONTHLY},
		{name: "monthly: int", data: "3", exp: PAYMENT_FREQUENCY_MONTHLY},

		// PAYMENT_FREQUENCY_QUARTERLY
		{name: "quarterly: long: upper case", data: `"PAYMENT_FREQUENCY_QUARTERLY"`, exp: PAYMENT_FREQUENCY_QUARTERLY},
		{name: "quarterly: long: lower case", data: `"payment_frequency_quarterly"`, exp: PAYMENT_FREQUENCY_QUARTERLY},
		{name: "quarterly: long: mixed case", data: `"paYMEnt_frEQUEncy_quaRTerLy"`, exp: PAYMENT_FREQUENCY_QUARTERLY},
		{name: "quarterly: short: upper case", data: `"QUARTERLY"`, exp: PAYMENT_FREQUENCY_QUARTERLY},
		{name: "quarterly: short: lower case", data: `"quarterly"`, exp: PAYMENT_FREQUENCY_QUARTERLY},
		{name: "quarterly: short: mixed case", data: `"qUArtErly"`, exp: PAYMENT_FREQUENCY_QUARTERLY},
		{name: "quarterly: int", data: "4", exp: PAYMENT_FREQUENCY_QUARTERLY},

		// PAYMENT_FREQUENCY_ANNUALLY
		{name: "annually: long: upper case", data: `"PAYMENT_FREQUENCY_ANNUALLY"`, exp: PAYMENT_FREQUENCY_ANNUALLY},
		{name: "annually: long: lower case", data: `"payment_frequency_annually"`, exp: PAYMENT_FREQUENCY_ANNUALLY},
		{name: "annually: long: mixed case", data: `"paYMENT_FREQuency_AnNUAllY"`, exp: PAYMENT_FREQUENCY_ANNUALLY},
		{name: "annually: short: upper case", data: `"ANNUALLY"`, exp: PAYMENT_FREQUENCY_ANNUALLY},
		{name: "annually: short: lower case", data: `"annually"`, exp: PAYMENT_FREQUENCY_ANNUALLY},
		{name: "annually: short: mixed case", data: `"aNNuaLLy"`, exp: PAYMENT_FREQUENCY_ANNUALLY},
		{name: "annually: int", data: "5", exp: PAYMENT_FREQUENCY_ANNUALLY},

		{
			name:   "unknown int: too large",
			data:   "6",
			expErr: "unknown payment_frequency integer value: 6",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var val PaymentFrequency
			var err error
			testFunc := func() {
				err = val.UnmarshalJSON([]byte(tc.data))
			}
			require.NotPanics(t, testFunc, "%T.UnmarshalJSON(%q)", val, tc.data)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.UnmarshalJSON(%q) error", val, tc.data)
			assert.Equal(t, tc.exp, val, "%T.UnmarshalJSON(%q) resulting value", val, tc.data)
		})
	}
}

func TestPaymentFrequency_Validate(t *testing.T) {
	tests := []struct {
		name   string
		p      PaymentFrequency
		expErr string
	}{
		{name: "negative one", p: -1, expErr: "unknown payment_frequency enum value: -1"},
		{name: "unspecified", p: PAYMENT_FREQUENCY_UNSPECIFIED},
		{name: "daily", p: PAYMENT_FREQUENCY_DAILY},
		{name: "weekly", p: PAYMENT_FREQUENCY_WEEKLY},
		{name: "monthly", p: PAYMENT_FREQUENCY_MONTHLY},
		{name: "quarterly", p: PAYMENT_FREQUENCY_QUARTERLY},
		{name: "annually", p: PAYMENT_FREQUENCY_ANNUALLY},
		{name: "six", p: 6, expErr: "unknown payment_frequency enum value: 6"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.p.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.p)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.p)
		})
	}
}
