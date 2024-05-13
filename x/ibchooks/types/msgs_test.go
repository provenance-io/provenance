package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/ibchooks/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgEmitIBCAck{Sender: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}
