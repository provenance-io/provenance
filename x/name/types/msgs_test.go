package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgCreateRootNameRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	name := "human-readable-name"
	owner := "owner"
	msg := NewMsgCreateRootNameRequest(authority.String(), name, owner, false)
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
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
			"proto: negative length found during unmarshaling",
		},
		{
			"record name contains segment",
			authority,
			"...",
			address,
			true,
			"invalid name: \".\" is reserved",
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
