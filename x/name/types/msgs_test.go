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
		func(signer string) sdk.Msg { return &MsgUpdateParamsRequest{Authority: signer} },
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

func TestMsgUpdateParamsRequestValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111").String()

	testCases := []struct {
		name                   string
		maxSegmentLength       uint32
		minSegmentLength       uint32
		maxNameLevels          uint32
		allowUnrestrictedNames bool
		authority              string
		shouldFail             bool
		expectedErr            string
	}{
		{
			"valid request",
			100,
			3,
			10,
			true,
			authority,
			false,
			"",
		},
		{
			"invalid authority",
			100,
			3,
			10,
			true,
			"blah",
			true,
			"decoding bech32 failed: invalid bech32 string length 4",
		},
	}

	for _, tc := range testCases {
		msg := NewMsgUpdateParamsRequest(tc.maxSegmentLength, tc.minSegmentLength, tc.maxNameLevels, tc.allowUnrestrictedNames, tc.authority)
		err := msg.ValidateBasic()
		if tc.shouldFail {
			require.EqualError(t, err, tc.expectedErr, "expected error for case: %s", tc.name)
		} else {
			require.NoError(t, err, "unexpected error for case: %s", tc.name)
		}
	}
}
