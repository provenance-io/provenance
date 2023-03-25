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
		    name: 	"invalid authority",
		    authority: 	"",
		    recordName: 	"human-readable-name",
		    address: 	address,
		    shouldFail: 	true,
		    expectedErr: 	"invalid account address",
		},
		{
		name: 	"invalid record address",
			authority: authority,
			recordName: "human-readable-name",
			address: "",
			shouldFail: true,
			expectedErr: "invalid account address",
		},
		{
			name: "invalid record name length",
			authority: authority,
			recordName: "",
			address: address,
			shouldFail: true,
			expectedErr: "segment of name is too short",
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
