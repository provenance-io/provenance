package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/name/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgBindNameRequest{Parent: NameRecord{Address: signer}} },
		func(signer string) sdk.Msg { return &MsgDeleteNameRequest{Record: NameRecord{Address: signer}} },
		func(signer string) sdk.Msg { return &MsgModifyNameRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgCreateRootNameRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestMsgCreateRootNameRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()
	address := "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"

	testCases := []struct {
		name        string
		authority   string
		recordName  string
		address     string
		shouldFail  bool
		expectedErr string
	}{
		{
			"invalid authority",
			"",
			"human-readable-name",
			address,
			true,
			"invalid account address",
		},
		{
			"invalid record address",
			authority,
			"human-readable-name",
			"",
			true,
			"invalid account address",
		},
		{
			"invalid record name length",
			authority,
			"",
			address,
			true,
			"segment of name is too short",
		},
	}

	for _, tc := range testCases {
		msg := NewMsgCreateRootNameRequest(tc.authority, tc.recordName, tc.address, false)
		err := msg.ValidateBasic()
		if tc.shouldFail {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}
