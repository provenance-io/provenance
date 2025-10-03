package types_test

import (
	"testing"

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
		{name: "valid", msg: MsgUpdatePaymentRequest{Signer: validAddr, Key: validKey, NextPmtAmt: math.NewInt(10), NextPmtDate: 20250101, PaymentFrequency: PAYMENT_FREQUENCY_MONTHLY}},
		{name: "neg amount", msg: MsgUpdatePaymentRequest{Signer: validAddr, Key: validKey, NextPmtAmt: math.NewInt(-1), NextPmtDate: 20250101, PaymentFrequency: PAYMENT_FREQUENCY_MONTHLY}, exp: []string{"invalid next_pmt_amt", "cannot be negative"}},
		{name: "bad date", msg: MsgUpdatePaymentRequest{Signer: validAddr, Key: validKey, NextPmtAmt: math.NewInt(10), NextPmtDate: 0, PaymentFrequency: PAYMENT_FREQUENCY_MONTHLY}, exp: []string{"invalid next_pmt_date", "positive"}},
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
	entry := &LedgerEntry{EntryTypeId: 1, PostedDate: 20240101, EffectiveDate: 20240101, TotalAmt: math.NewInt(100), AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 1, AppliedAmt: math.NewInt(100)}}, CorrelationId: "c1"}

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
		{name: "valid", msg: MsgUpdateBalancesRequest{Signer: validAddr, Key: validKey, CorrelationId: "c1", BalanceAmounts: []*BucketBalance{bal}, AppliedAmounts: []*LedgerBucketAmount{applied}}},
		{name: "empty balances", msg: MsgUpdateBalancesRequest{Signer: validAddr, Key: validKey, CorrelationId: "c1", BalanceAmounts: []*BucketBalance{}, AppliedAmounts: []*LedgerBucketAmount{applied}}, exp: []string{"invalid balance_amounts", "cannot be empty"}},
		{name: "empty applied", msg: MsgUpdateBalancesRequest{Signer: validAddr, Key: validKey, CorrelationId: "c1", BalanceAmounts: []*BucketBalance{bal}, AppliedAmounts: []*LedgerBucketAmount{}}, exp: []string{"invalid applied_amounts", "cannot be empty"}},
		{name: "bad correlation id", msg: MsgUpdateBalancesRequest{Signer: validAddr, Key: validKey, CorrelationId: "", BalanceAmounts: []*BucketBalance{bal}, AppliedAmounts: []*LedgerBucketAmount{applied}}, exp: []string{"invalid correlation_id", "between"}},
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
	si := &SettlementInstruction{Amount: sdk.NewInt64Coin("stake", 1), RecipientAddress: sdk.AccAddress("recipient_______________").String(), Status: FundingTransferStatus_FUNDING_TRANSFER_STATUS_PENDING}
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
	l := &Ledger{Key: lk, LedgerClassId: "lclass", StatusTypeId: 1}
	entry := &LedgerEntry{EntryTypeId: 1, PostedDate: 20240101, EffectiveDate: 20240101, TotalAmt: math.NewInt(100), AppliedAmounts: []*LedgerBucketAmount{{BucketTypeId: 1, AppliedAmt: math.NewInt(100)}}, CorrelationId: "c1"}
	le := &LedgerAndEntries{LedgerKey: lk, Ledger: l, Entries: []*LedgerEntry{entry}}

	tests := []struct {
		name string
		msg  MsgBulkCreateRequest
		exp  []string
	}{
		{name: "valid", msg: MsgBulkCreateRequest{Signer: validAddr, LedgerAndEntries: []*LedgerAndEntries{le}}},
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
	validAddr := sdk.AccAddress("entry_add_signer_______").String()
	et := &LedgerClassEntryType{Id: 1, Code: "ADJ", Description: "Adjustment"}

	tests := []struct {
		name string
		msg  MsgAddLedgerClassEntryTypeRequest
		exp  []string
	}{
		{name: "valid", msg: MsgAddLedgerClassEntryTypeRequest{Signer: validAddr, LedgerClassId: "lclass", EntryType: et}},
		{name: "nil entry type", msg: MsgAddLedgerClassEntryTypeRequest{Signer: validAddr, LedgerClassId: "lclass", EntryType: nil}, exp: []string{"invalid entry_type", "cannot be nil"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			assertions.RequireErrorContents(t, err, tc.exp)
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
