package types_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// TODO: func TestLedger_Validate(t *testing.T) {}

// TODO: func TestValidatePmtFields(t *testing.T) {}

// TODO: func TestLedgerClassBucketType_Validate(t *testing.T) {}

// TODO: func TestLedgerEntry_Compare(t *testing.T) {}

// TODO: func TestLedgerEntry_Validate(t *testing.T) {}

// TODO: func TestValidateSequence(t *testing.T) {}

// TODO: func TestValidateLedgerEntryAmounts(t *testing.T) {}

// TODO: func TestValidateEntryAmounts(t *testing.T) {}

// TODO: func TestLedgerBucketAmount_Validate(t *testing.T) {}

// TODO: func TestBucketBalance_Validate(t *testing.T) {}

// TODO: func TestLedgerAndEntries_Validate(t *testing.T) {}

// TODO: func TestDayCountConvention_UnmarshalJSON(t *testing.T) {}

// TODO: func TestDayCountConvention_Validate(t *testing.T) {}

// TODO: func TestInterestAccrualMethod_UnmarshalJSON(t *testing.T) {}

// TODO: func TestInterestAccrualMethod_Validate(t *testing.T) {}

// TODO: func TestPaymentFrequency_UnmarshalJSON(t *testing.T) {}

// TODO: func TestPaymentFrequency_Validate(t *testing.T) {}
