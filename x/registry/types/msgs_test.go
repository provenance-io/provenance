package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/registry/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgRegisterNFT{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgGrantRole{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgRevokeRole{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgUnregisterNFT{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgRegistryBulkUpdate{Signer: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}
