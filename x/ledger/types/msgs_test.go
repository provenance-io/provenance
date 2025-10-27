package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"

	. "github.com/provenance-io/provenance/x/ledger/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgCreateLedgerRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateStatusRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateInterestRateRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgUpdatePaymentRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateMaturityDateRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgAppendRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateBalancesRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgTransferFundsWithSettlementRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgDestroyRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreateLedgerClassRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgAddLedgerClassStatusTypeRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgAddLedgerClassEntryTypeRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgAddLedgerClassBucketTypeRequest{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgBulkCreateRequest{Signer: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestMsgUpdateStatus_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("status_signer___________").String()
	validKey := &LedgerKey{AssetClassId: "aclass", NftId: "nft1"}

	tests := []struct {
		name string
		msg  MsgUpdateStatusRequest
		exp  []string
	}{
		{name: "valid", msg: MsgUpdateStatusRequest{Signer: validAddr, Key: validKey, StatusTypeId: 1}},
		{name: "empty signer", msg: MsgUpdateStatusRequest{Signer: "", Key: validKey, StatusTypeId: 1}, exp: []string{"invalid signer"}},
		{name: "invalid key", msg: MsgUpdateStatusRequest{Signer: validAddr, Key: &LedgerKey{AssetClassId: "", NftId: "nft1"}, StatusTypeId: 1}, exp: []string{"invalid key", "asset_class_id", "must be between"}},
		{name: "non-positive status", msg: MsgUpdateStatusRequest{Signer: validAddr, Key: validKey, StatusTypeId: 0}, exp: []string{"invalid status_type_id", "positive"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgUpdateInterestRate_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("rate_signer____________").String()
	validKey := &LedgerKey{AssetClassId: "aclass", NftId: "nft1"}

	tests := []struct {
		name string
		msg  MsgUpdateInterestRateRequest
		exp  []string
	}{
		{name: "valid", msg: MsgUpdateInterestRateRequest{Signer: validAddr, Key: validKey, InterestRate: 5_000_000, InterestDayCountConvention: DAY_COUNT_CONVENTION_THIRTY_360, InterestAccrualMethod: INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST}},
		{name: "empty signer", msg: MsgUpdateInterestRateRequest{Signer: "", Key: validKey}, exp: []string{"invalid signer"}},
		{name: "invalid key", msg: MsgUpdateInterestRateRequest{Signer: validAddr, Key: &LedgerKey{AssetClassId: "", NftId: "nft1"}}, exp: []string{"invalid key", "asset_class_id"}},
		{name: "rate out of bounds", msg: MsgUpdateInterestRateRequest{Signer: validAddr, Key: validKey, InterestRate: 101_000_000}, exp: []string{"invalid interest_rate", "between"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgUpdatePayment_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("payment_signer__________").String()
	validKey := &LedgerKey{AssetClassId: "aclass", NftId: "nft1"}

	tests := []struct {
		name string
		msg  MsgUpdatePaymentRequest
		exp  []string
	}{
		{name: "valid", msg: MsgUpdatePaymentRequest{Signer: validAddr, Key: validKey, NextPmtAmt: math.NewInt(10), NextPmtDate: 20089, PaymentFrequency: PAYMENT_FREQUENCY_MONTHLY}},
		{name: "neg amount", msg: MsgUpdatePaymentRequest{Signer: validAddr, Key: validKey, NextPmtAmt: math.NewInt(-1), NextPmtDate: 20089, PaymentFrequency: PAYMENT_FREQUENCY_MONTHLY}, exp: []string{"next_pmt_amt", "must be a non-negative integer"}},
		{name: "bad date", msg: MsgUpdatePaymentRequest{Signer: validAddr, Key: validKey, NextPmtAmt: math.NewInt(10), NextPmtDate: -1, PaymentFrequency: PAYMENT_FREQUENCY_MONTHLY}, exp: []string{"next_pmt_date", "must be after 1970-01-01"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgAppend_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("append_signer___________").String()
	validKey := &LedgerKey{AssetClassId: "aclass", NftId: "nft1"}
	entry := &LedgerEntry{
		CorrelationId:  "c1",
		EntryTypeId:    1,
		PostedDate:     20240101,
		EffectiveDate:  20240101,
		TotalAmt:       math.NewInt(100),
		AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 1, AppliedAmt: math.NewInt(100)}},
		BalanceAmounts: []*BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
	}

	tests := []struct {
		name string
		msg  MsgAppendRequest
		exp  []string
	}{
		{name: "valid", msg: MsgAppendRequest{Signer: validAddr, Key: validKey, Entries: []*LedgerEntry{entry}}},
		{name: "empty entries", msg: MsgAppendRequest{Signer: validAddr, Key: validKey, Entries: []*LedgerEntry{}}, exp: []string{"invalid entries", "cannot be empty"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgDestroy_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("destroy_signer_________").String()
	validKey := &LedgerKey{AssetClassId: "aclass", NftId: "nft1"}

	tests := []struct {
		name string
		msg  MsgDestroyRequest
		exp  []string
	}{
		{name: "valid", msg: MsgDestroyRequest{Signer: validAddr, Key: validKey}},
		{name: "nil key", msg: MsgDestroyRequest{Signer: validAddr, Key: nil}, exp: []string{"invalid key", "cannot be nil"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgUpdateBalances_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("balances_signer________").String()
	validKey := &LedgerKey{AssetClassId: "aclass", NftId: "nft1"}

	bal := &BucketBalance{BucketTypeId: 1, BalanceAmt: math.NewInt(10)}
	applied := &LedgerBucketAmount{BucketTypeId: 1, AppliedAmt: math.NewInt(10)}

	tests := []struct {
		name string
		msg  MsgUpdateBalancesRequest
		exp  []string
	}{
		{
			name: "valid",
			msg: MsgUpdateBalancesRequest{
				Signer:         validAddr,
				Key:            validKey,
				CorrelationId:  "c1",
				TotalAmt:       applied.AppliedAmt,
				AppliedAmounts: []*LedgerBucketAmount{applied},
				BalanceAmounts: []*BucketBalance{bal},
			},
		},
		{
			name: "wrong total",
			msg: MsgUpdateBalancesRequest{
				Signer:         validAddr,
				Key:            validKey,
				CorrelationId:  "c1",
				TotalAmt:       applied.AppliedAmt.SubRaw(1),
				AppliedAmounts: []*LedgerBucketAmount{applied},
				BalanceAmounts: []*BucketBalance{bal},
			},
			exp: []string{"applied_amounts", "total amount must equal abs(sum of applied amounts)"},
		},
		{
			name: "empty balances",
			msg: MsgUpdateBalancesRequest{
				Signer:         validAddr,
				Key:            validKey,
				CorrelationId:  "c1",
				TotalAmt:       applied.AppliedAmt,
				AppliedAmounts: []*LedgerBucketAmount{applied},
				BalanceAmounts: []*BucketBalance{},
			},
		},
		{
			name: "empty applied",
			msg: MsgUpdateBalancesRequest{
				Signer:         validAddr,
				Key:            validKey,
				CorrelationId:  "c1",
				TotalAmt:       applied.AppliedAmt,
				AppliedAmounts: []*LedgerBucketAmount{},
				BalanceAmounts: []*BucketBalance{bal},
			},
			exp: []string{"applied_amounts", "cannot be empty"},
		},
		{
			name: "bad correlation id",
			msg: MsgUpdateBalancesRequest{
				Signer:         validAddr,
				Key:            validKey,
				CorrelationId:  "",
				TotalAmt:       applied.AppliedAmt,
				AppliedAmounts: []*LedgerBucketAmount{applied},
				BalanceAmounts: []*BucketBalance{bal},
			},
			exp: []string{"correlation_id", "must be between"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgTransferFundsWithSettlement_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("transfer_signer________").String()
	key := &LedgerKey{AssetClassId: "aclass", NftId: "nft1"}
	si := &SettlementInstruction{Amount: sdk.NewInt64Coin("stake", 1), RecipientAddress: sdk.AccAddress("recipient_______________").String(), Status: FUNDING_TRANSFER_STATUS_PENDING}
	ft := &FundTransferWithSettlement{Key: key, LedgerEntryCorrelationId: "c1", SettlementInstructions: []*SettlementInstruction{si}}

	tests := []struct {
		name string
		msg  MsgTransferFundsWithSettlementRequest
		exp  []string
	}{
		{name: "valid", msg: MsgTransferFundsWithSettlementRequest{Signer: validAddr, Transfers: []*FundTransferWithSettlement{ft}}},
		{name: "no transfers", msg: MsgTransferFundsWithSettlementRequest{Signer: validAddr, Transfers: []*FundTransferWithSettlement{}}, exp: []string{"invalid transfers", "cannot be empty"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgBulkCreate_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("bulk_signer____________").String()
	lk := &LedgerKey{AssetClassId: "aclass", NftId: "nft1"}
	l := &Ledger{
		Key:                        lk,
		LedgerClassId:              "lclass",
		StatusTypeId:               1,
		NextPmtDate:                20250101,
		NextPmtAmt:                 math.NewInt(100),
		InterestRate:               5_000_000,
		InterestDayCountConvention: DAY_COUNT_CONVENTION_THIRTY_360,
		InterestAccrualMethod:      INTEREST_ACCRUAL_METHOD_SIMPLE_INTEREST,
		PaymentFrequency:           PAYMENT_FREQUENCY_MONTHLY,
		MaturityDate:               20260101,
	}
	entry := &LedgerEntry{
		CorrelationId:  "c1",
		EntryTypeId:    1,
		PostedDate:     20240101,
		EffectiveDate:  20240101,
		TotalAmt:       math.NewInt(100),
		AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 1, AppliedAmt: math.NewInt(100)}},
		BalanceAmounts: []*BucketBalance{{BucketTypeId: 1, BalanceAmt: math.NewInt(100)}},
	}
	le := &LedgerAndEntries{LedgerKey: lk, Ledger: l, Entries: []*LedgerEntry{entry}}

	tests := []struct {
		name string
		msg  MsgBulkCreateRequest
		exp  []string
	}{
		{name: "valid", msg: MsgBulkCreateRequest{Signer: validAddr, LedgerAndEntries: []*LedgerAndEntries{le}}},
		{name: "empty signer", msg: MsgBulkCreateRequest{Signer: "", LedgerAndEntries: []*LedgerAndEntries{le}}, exp: []string{"invalid signer"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgCreateLedgerClass_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("class_signer___________").String()
	lc := &LedgerClass{LedgerClassId: "class1", AssetClassId: "aclass", MaintainerAddress: validAddr, Denom: "stake"}

	tests := []struct {
		name string
		msg  MsgCreateLedgerClassRequest
		exp  []string
	}{
		{name: "valid", msg: MsgCreateLedgerClassRequest{Signer: validAddr, LedgerClass: lc}},
		{name: "maintainer mismatch", msg: MsgCreateLedgerClassRequest{Signer: validAddr, LedgerClass: &LedgerClass{LedgerClassId: "class1", AssetClassId: "aclass", MaintainerAddress: "cosmos1other", Denom: "stake"}}, exp: []string{"unauthorized access", "maintainer address"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgAddLedgerClassStatusType_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("status_add_signer______").String()
	st := &LedgerClassStatusType{Id: 1, Code: "IN_PROGRESS", Description: "In Progress"}

	tests := []struct {
		name string
		msg  MsgAddLedgerClassStatusTypeRequest
		exp  []string
	}{
		{name: "valid", msg: MsgAddLedgerClassStatusTypeRequest{Signer: validAddr, LedgerClassId: "lclass", StatusType: st}},
		{name: "nil status type", msg: MsgAddLedgerClassStatusTypeRequest{Signer: validAddr, LedgerClassId: "lclass", StatusType: nil}, exp: []string{"invalid status_type", "cannot be nil"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}

func TestMsgAddLedgerClassEntryType_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgAddLedgerClassEntryTypeRequest
		exp  []string
	}{
		{
			name: "valid",
			msg: MsgAddLedgerClassEntryTypeRequest{
				Signer:        sdk.AccAddress("signer______________").String(),
				LedgerClassId: "lclass",
				EntryType:     &LedgerClassEntryType{Id: 1, Code: "ADJ", Description: "Adjustment"},
			},
		},
		{
			name: "no signer",
			msg: MsgAddLedgerClassEntryTypeRequest{
				Signer:        "",
				LedgerClassId: "lclass",
				EntryType:     &LedgerClassEntryType{Id: 1, Code: "ADJ", Description: "Adjustment"},
			},
			exp: []string{"invalid signer", "empty address string is not allowed", "invalid field"},
		},
		{
			name: "bad signer",
			msg: MsgAddLedgerClassEntryTypeRequest{
				Signer:        "not-an-addr",
				LedgerClassId: "lclass",
				EntryType:     &LedgerClassEntryType{Id: 1, Code: "ADJ", Description: "Adjustment"},
			},
			exp: []string{"invalid signer", "decoding bech32 failed: invalid separator index -1", "invalid field"},
		},
		{
			name: "no ledger class",
			msg: MsgAddLedgerClassEntryTypeRequest{
				Signer:        sdk.AccAddress("signer______________").String(),
				LedgerClassId: "",
				EntryType:     &LedgerClassEntryType{Id: 1, Code: "ADJ", Description: "Adjustment"},
			},
			exp: []string{"invalid ledger_class_id", "must be between 1 and 50 characters", "invalid field"},
		},
		{
			name: "ledger class too long",
			msg: MsgAddLedgerClassEntryTypeRequest{
				Signer:        sdk.AccAddress("signer______________").String(),
				LedgerClassId: "l" + strings.Repeat("c", MaxLenLedgerClassID),
				EntryType:     &LedgerClassEntryType{Id: 1, Code: "ADJ", Description: "Adjustment"},
			},
			exp: []string{"invalid ledger_class_id", "must be between 1 and 50 characters", "invalid field"},
		},
		{
			name: "nil entry type",
			msg: MsgAddLedgerClassEntryTypeRequest{
				Signer:        sdk.AccAddress("signer______________").String(),
				LedgerClassId: "lclass",
				EntryType:     nil,
			},
			exp: []string{"invalid entry_type", "cannot be nil", "invalid field"},
		},
		{
			name: "invalid entry type",
			msg: MsgAddLedgerClassEntryTypeRequest{
				Signer:        sdk.AccAddress("signer______________").String(),
				LedgerClassId: "lclass",
				EntryType:     &LedgerClassEntryType{Id: -1, Code: "ADJ", Description: "Adjustment"},
			},
			exp: []string{"invalid entry_type", "id: -1 must be a non-negative integer", "invalid field"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "%T.ValidateBasic()", tc.msg)
			assertions.RequireErrorContents(t, err, tc.exp, "%T.ValidateBasic() error", tc.msg)
		})
	}
}

func TestMsgAddLedgerClassBucketType_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("bucket_add_signer______").String()
	bt := &LedgerClassBucketType{Id: 1, Code: "ESCROW", Description: "Escrow"}

	tests := []struct {
		name string
		msg  MsgAddLedgerClassBucketTypeRequest
		exp  []string
	}{
		{name: "valid", msg: MsgAddLedgerClassBucketTypeRequest{Signer: validAddr, LedgerClassId: "lclass", BucketType: bt}},
		{name: "nil bucket type", msg: MsgAddLedgerClassBucketTypeRequest{Signer: validAddr, LedgerClassId: "lclass", BucketType: nil}, exp: []string{"invalid bucket_type", "cannot be nil"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
		})
	}
}
