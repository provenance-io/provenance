package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/ledger/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgCreateLedgerRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateStatusRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateInterestRateRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdatePaymentRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateMaturityDateRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgAppendRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateBalancesRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgTransferFundsWithSettlementRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgDestroyRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgCreateLedgerClassRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgAddLedgerClassStatusTypeRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgAddLedgerClassEntryTypeRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgAddLedgerClassBucketTypeRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgBulkCreateRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}
