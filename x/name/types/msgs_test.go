package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgCreateRootNameRequestGetSignBytes(t *testing.T) {
	owner := sdk.AccAddress("input")
	msg := MsgCreateRootNameRequest{
		Name:       "hooman",
		Owner:      owner.String(),
		Restricted: false,
	}
	res := msg.GetSignBytes()

	expected := `{"type":"provenance/MsgCreateRootNameRequest","value":{"name":"hooman","owner":"cosmos1d9h8qat57ljhcm"}}`
	require.Equal(t, expected, string(res))
}

func TestMsgCreateRootNameRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	name := "Proposal Title"
	owner := "Proposal description"
	msg := NewMsgCreateRootNameRequest(authority.String(), name, owner, false)
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}
