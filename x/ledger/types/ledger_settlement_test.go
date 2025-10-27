package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil/assertions"

	. "github.com/provenance-io/provenance/x/ledger/types"
)

func TestFundTransferWithSettlement_Validate(t *testing.T) {
	tests := []struct {
		name   string
		ft     *FundTransferWithSettlement
		expErr string
	}{
		{
			name:   "nil",
			ft:     nil,
			expErr: "fund_transfer_with_settlement cannot be nil",
		},
		{
			name: "valid",
			ft: &FundTransferWithSettlement{
				Key:                      &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerEntryCorrelationId: "abcd1234-ef56",
				SettlementInstructions: []*SettlementInstruction{
					{
						Amount:           sdk.NewInt64Coin("banana", 99),
						RecipientAddress: sdk.AccAddress("recipient_1_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_PENDING,
					},
					{
						Amount:           sdk.NewInt64Coin("apple", 12),
						RecipientAddress: sdk.AccAddress("recipient_2_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_COMPLETED,
					},
				},
			},
		},
		{
			name: "invalid key",
			ft: &FundTransferWithSettlement{
				Key:                      &LedgerKey{NftId: "", AssetClassId: "the-asset-class-id"},
				LedgerEntryCorrelationId: "abcd1234-ef56",
				SettlementInstructions: []*SettlementInstruction{
					{
						Amount:           sdk.NewInt64Coin("banana", 99),
						RecipientAddress: sdk.AccAddress("recipient_1_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_PENDING,
					},
					{
						Amount:           sdk.NewInt64Coin("apple", 12),
						RecipientAddress: sdk.AccAddress("recipient_2_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_COMPLETED,
					},
				},
			},
			expErr: "key: invalid nft_id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "no correlation id",
			ft: &FundTransferWithSettlement{
				Key:                      &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerEntryCorrelationId: "",
				SettlementInstructions: []*SettlementInstruction{
					{
						Amount:           sdk.NewInt64Coin("banana", 99),
						RecipientAddress: sdk.AccAddress("recipient_1_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_PENDING,
					},
					{
						Amount:           sdk.NewInt64Coin("apple", 12),
						RecipientAddress: sdk.AccAddress("recipient_2_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_COMPLETED,
					},
				},
			},
			expErr: "ledger_entry_correlation_id: must be between 1 and 50 characters",
		},
		{
			name: "correlation id too long",
			ft: &FundTransferWithSettlement{
				Key:                      &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerEntryCorrelationId: strings.Repeat("L", 51),
				SettlementInstructions: []*SettlementInstruction{
					{
						Amount:           sdk.NewInt64Coin("banana", 99),
						RecipientAddress: sdk.AccAddress("recipient_1_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_PENDING,
					},
					{
						Amount:           sdk.NewInt64Coin("apple", 12),
						RecipientAddress: sdk.AccAddress("recipient_2_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_COMPLETED,
					},
				},
			},
			expErr: "ledger_entry_correlation_id: must be between 1 and 50 characters",
		},
		{
			name: "invalid settlement instructions",
			ft: &FundTransferWithSettlement{
				Key:                      &LedgerKey{NftId: "the-nft-id", AssetClassId: "the-asset-class-id"},
				LedgerEntryCorrelationId: "abcd1234-ef56",
				SettlementInstructions: []*SettlementInstruction{
					{
						Amount:           sdk.NewInt64Coin("banana", 99),
						RecipientAddress: "bad-recip",
						Status:           FUNDING_TRANSFER_STATUS_PENDING,
					},
					{
						Amount:           sdk.NewInt64Coin("apple", 12),
						RecipientAddress: sdk.AccAddress("recipient_2_________").String(),
						Status:           12,
					},
				},
			},
			expErr: joinErrs("settlement_instructions[0]: recipient_address: decoding bech32 failed: invalid separator index -1",
				"settlement_instructions[1]: unknown funding_transfer_status enum value: 12"),
		},
		{
			name: "multiple errors",
			ft: &FundTransferWithSettlement{
				Key:                      &LedgerKey{NftId: "", AssetClassId: "the-asset-class-id"},
				LedgerEntryCorrelationId: "",
				SettlementInstructions: []*SettlementInstruction{
					{
						Amount:           sdk.NewInt64Coin("banana", 99),
						RecipientAddress: sdk.AccAddress("recipient_1_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_PENDING,
					},
					{
						Amount:           sdk.NewInt64Coin("banana", 99),
						RecipientAddress: "bad-recip",
						Status:           FUNDING_TRANSFER_STATUS_PENDING,
					},
					{
						Amount:           sdk.NewInt64Coin("apple", 12),
						RecipientAddress: sdk.AccAddress("recipient_2_________").String(),
						Status:           FUNDING_TRANSFER_STATUS_COMPLETED,
					},
					{
						Amount:           sdk.NewInt64Coin("apple", 12),
						RecipientAddress: sdk.AccAddress("recipient_2_________").String(),
						Status:           12,
					},
				},
			},
			expErr: joinErrs("key: invalid nft_id: must be between 1 and 128 characters: invalid field",
				"ledger_entry_correlation_id: must be between 1 and 50 characters",
				"settlement_instructions[1]: recipient_address: decoding bech32 failed: invalid separator index -1",
				"settlement_instructions[3]: unknown funding_transfer_status enum value: 12"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.ft.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.ft)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.ft)
		})
	}
}

func TestSettlementInstruction_Validate(t *testing.T) {
	tests := []struct {
		name   string
		si     *SettlementInstruction
		expErr string
	}{
		{
			name:   "nil",
			si:     nil,
			expErr: "settlement_instruction cannot be nil",
		},
		{
			name: "valid",
			si: &SettlementInstruction{
				Amount:           sdk.NewInt64Coin("banana", 35),
				RecipientAddress: sdk.AccAddress("recipient_address___").String(),
				Status:           FUNDING_TRANSFER_STATUS_PROCESSING,
				Memo:             "Just a little note to say you're awesome.",
			},
		},
		{
			name: "invalid amount",
			si: &SettlementInstruction{
				Amount:           sdk.Coin{Denom: "x", Amount: sdkmath.NewInt(35)},
				RecipientAddress: sdk.AccAddress("recipient_address___").String(),
				Status:           FUNDING_TRANSFER_STATUS_PROCESSING,
				Memo:             "Just a little note to say you're awesome.",
			},
			expErr: "amount: invalid denom: x",
		},
		{
			name: "negative amount",
			si: &SettlementInstruction{
				Amount:           sdk.Coin{Denom: "banana", Amount: sdkmath.NewInt(-3)},
				RecipientAddress: sdk.AccAddress("recipient_address___").String(),
				Status:           FUNDING_TRANSFER_STATUS_PROCESSING,
				Memo:             "Just a little note to say you're awesome.",
			},
			expErr: "amount: negative coin amount: -3",
		},
		{
			name: "no memo",
			si: &SettlementInstruction{
				Amount:           sdk.NewInt64Coin("banana", 35),
				RecipientAddress: sdk.AccAddress("recipient_address___").String(),
				Status:           FUNDING_TRANSFER_STATUS_PROCESSING,
				Memo:             "",
			},
		},
		{
			name: "memo too long",
			si: &SettlementInstruction{
				Amount:           sdk.NewInt64Coin("banana", 35),
				RecipientAddress: sdk.AccAddress("recipient_address___").String(),
				Status:           FUNDING_TRANSFER_STATUS_PROCESSING,
				Memo:             strings.Repeat("m", 51),
			},
			expErr: "memo: must be between 0 and 50 characters",
		},
		{
			name: "no recipient",
			si: &SettlementInstruction{
				Amount:           sdk.NewInt64Coin("banana", 35),
				RecipientAddress: "",
				Status:           FUNDING_TRANSFER_STATUS_PROCESSING,
				Memo:             "Just a little note to say you're awesome.",
			},
			expErr: "recipient_address: empty address string is not allowed",
		},
		{
			name: "invalid recipient",
			si: &SettlementInstruction{
				Amount:           sdk.NewInt64Coin("banana", 35),
				RecipientAddress: "recipient_address___",
				Status:           FUNDING_TRANSFER_STATUS_PROCESSING,
				Memo:             "Just a little note to say you're awesome.",
			},
			expErr: "recipient_address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "invalid status",
			si: &SettlementInstruction{
				Amount:           sdk.NewInt64Coin("banana", 35),
				RecipientAddress: sdk.AccAddress("recipient_address___").String(),
				Status:           12,
				Memo:             "Just a little note to say you're awesome.",
			},
			expErr: "unknown funding_transfer_status enum value: 12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.si.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.si)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.si)
		})
	}
}

func TestFundingTransferStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		exp    FundingTransferStatus
		expErr string
	}{
		{
			name:   "empty string",
			data:   "",
			expErr: "funding_transfer_status must be a string or integer, got: \"\"",
		},
		{
			name:   "unknown data type",
			data:   "not-right",
			expErr: "funding_transfer_status must be a string or integer, got: \"not-right\"",
		},
		{
			name:   "unknown int: negative",
			data:   "-1",
			expErr: "unknown funding_transfer_status integer value: -1",
		},
		{
			name:   "unknown name",
			data:   `"unknown"`,
			expErr: "unknown funding_transfer_status string value: \"unknown\"",
		},

		// FUNDING_TRANSFER_STATUS_UNSPECIFIED
		{name: "unspecified: long: upper case", data: `"FUNDING_TRANSFER_STATUS_UNSPECIFIED"`, exp: FUNDING_TRANSFER_STATUS_UNSPECIFIED},
		{name: "unspecified: long: lower case", data: `"funding_transfer_status_unspecified"`, exp: FUNDING_TRANSFER_STATUS_UNSPECIFIED},
		{name: "unspecified: long: mixed case", data: `"funding_tranSFer_statUS_UNspecIFied"`, exp: FUNDING_TRANSFER_STATUS_UNSPECIFIED},
		{name: "unspecified: short: upper case", data: `"UNSPECIFIED"`, exp: FUNDING_TRANSFER_STATUS_UNSPECIFIED},
		{name: "unspecified: short: lower case", data: `"unspecified"`, exp: FUNDING_TRANSFER_STATUS_UNSPECIFIED},
		{name: "unspecified: short: mixed case", data: `"unspEcifiEd"`, exp: FUNDING_TRANSFER_STATUS_UNSPECIFIED},
		{name: "unspecified: int", data: "0", exp: FUNDING_TRANSFER_STATUS_UNSPECIFIED},

		// FUNDING_TRANSFER_STATUS_PENDING
		{name: "pending: long: upper case", data: `"FUNDING_TRANSFER_STATUS_PENDING"`, exp: FUNDING_TRANSFER_STATUS_PENDING},
		{name: "pending: long: lower case", data: `"funding_transfer_status_pending"`, exp: FUNDING_TRANSFER_STATUS_PENDING},
		{name: "pending: long: mixed case", data: `"Funding_Transfer_Status_Pending"`, exp: FUNDING_TRANSFER_STATUS_PENDING},
		{name: "pending: short: upper case", data: `"PENDING"`, exp: FUNDING_TRANSFER_STATUS_PENDING},
		{name: "pending: short: lower case", data: `"pending"`, exp: FUNDING_TRANSFER_STATUS_PENDING},
		{name: "pending: short: mixed case", data: `"Pending"`, exp: FUNDING_TRANSFER_STATUS_PENDING},
		{name: "pending: int", data: "1", exp: FUNDING_TRANSFER_STATUS_PENDING},

		// FUNDING_TRANSFER_STATUS_PROCESSING
		{name: "processing: long: upper case", data: `"FUNDING_TRANSFER_STATUS_PROCESSING"`, exp: FUNDING_TRANSFER_STATUS_PROCESSING},
		{name: "processing: long: lower case", data: `"funding_transfer_status_processing"`, exp: FUNDING_TRANSFER_STATUS_PROCESSING},
		{name: "processing: long: mixed case", data: `"funding_trANSfer_status_proceSSing"`, exp: FUNDING_TRANSFER_STATUS_PROCESSING},
		{name: "processing: short: upper case", data: `"PROCESSING"`, exp: FUNDING_TRANSFER_STATUS_PROCESSING},
		{name: "processing: short: lower case", data: `"processing"`, exp: FUNDING_TRANSFER_STATUS_PROCESSING},
		{name: "processing: short: mixed case", data: `"prOCEssINg"`, exp: FUNDING_TRANSFER_STATUS_PROCESSING},
		{name: "processing: int", data: "2", exp: FUNDING_TRANSFER_STATUS_PROCESSING},

		// FUNDING_TRANSFER_STATUS_COMPLETED
		{name: "completed: long: upper case", data: `"FUNDING_TRANSFER_STATUS_COMPLETED"`, exp: FUNDING_TRANSFER_STATUS_COMPLETED},
		{name: "completed: long: lower case", data: `"funding_transfer_status_completed"`, exp: FUNDING_TRANSFER_STATUS_COMPLETED},
		{name: "completed: long: mixed case", data: `"funding_tRANsfer_statuS_COMpletED"`, exp: FUNDING_TRANSFER_STATUS_COMPLETED},
		{name: "completed: short: upper case", data: `"COMPLETED"`, exp: FUNDING_TRANSFER_STATUS_COMPLETED},
		{name: "completed: short: lower case", data: `"completed"`, exp: FUNDING_TRANSFER_STATUS_COMPLETED},
		{name: "completed: short: mixed case", data: `"cOmpLETed"`, exp: FUNDING_TRANSFER_STATUS_COMPLETED},
		{name: "completed: int", data: "3", exp: FUNDING_TRANSFER_STATUS_COMPLETED},

		// FUNDING_TRANSFER_STATUS_FAILED
		{name: "completed: long: upper case", data: `"FUNDING_TRANSFER_STATUS_FAILED"`, exp: FUNDING_TRANSFER_STATUS_FAILED},
		{name: "completed: long: lower case", data: `"funding_transfer_status_failed"`, exp: FUNDING_TRANSFER_STATUS_FAILED},
		{name: "completed: long: mixed case", data: `"fUndIng_traNSfer_status_fAilEd"`, exp: FUNDING_TRANSFER_STATUS_FAILED},
		{name: "completed: short: upper case", data: `"FAILED"`, exp: FUNDING_TRANSFER_STATUS_FAILED},
		{name: "completed: short: lower case", data: `"failed"`, exp: FUNDING_TRANSFER_STATUS_FAILED},
		{name: "completed: short: mixed case", data: `"fAILEd"`, exp: FUNDING_TRANSFER_STATUS_FAILED},
		{name: "completed: int", data: "4", exp: FUNDING_TRANSFER_STATUS_FAILED},

		{
			name:   "unknown int: too large",
			data:   "5",
			expErr: "unknown funding_transfer_status integer value: 5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var val FundingTransferStatus
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

func TestFundingTransferStatus_Validate(t *testing.T) {
	tests := []struct {
		name   string
		s      FundingTransferStatus
		expErr string
	}{
		{name: "negative one", s: -1, expErr: "unknown funding_transfer_status enum value: -1"},
		{name: "unspecified", s: FUNDING_TRANSFER_STATUS_UNSPECIFIED},
		{name: "pending", s: FUNDING_TRANSFER_STATUS_PENDING},
		{name: "processing", s: FUNDING_TRANSFER_STATUS_PROCESSING},
		{name: "completed", s: FUNDING_TRANSFER_STATUS_COMPLETED},
		{name: "failed", s: FUNDING_TRANSFER_STATUS_FAILED},
		{name: "five", s: 5, expErr: "unknown funding_transfer_status enum value: 5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.s.Validate()
			}
			require.NotPanics(t, testFunc, "%T.Validate()", tc.s)
			assertions.AssertErrorValue(t, err, tc.expErr, "%T.Validate() error", tc.s)
		})
	}
}
