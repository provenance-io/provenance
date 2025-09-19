package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"

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
