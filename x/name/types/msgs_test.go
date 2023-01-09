package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMsgCreateRootNameRequestGetSignBytes(t *testing.T) {
	owner := sdk.AccAddress("input")
	msg := MsgCreateRootNameRequest{
		Title:       "title",
		Description: "description",
		Metadata: &Metadata{
			Name:       "hooman",
			Owner:      owner.String(),
			Restricted: false,
		},
	}
	res := msg.GetSignBytes()

	expected := `{"type":"provenance/MsgCreateRootNameRequest","value":{"description":"description","metadata":{"name":"hooman","owner":"cosmos1d9h8qat57ljhcm"},"title":"title"}}`
	require.Equal(t, expected, string(res))
}

func TestMsgCreateRootNameRequestGetSigners(t *testing.T) {
	authority := sdk.AccAddress("input111111111111111")
	title := "Proposal Title"
	description := "Proposal description"
	metadata := Metadata{}
	msg := NewMsgCreateRootNameRequest(title, description, &metadata, authority.String())
	res := msg.GetSigners()
	require.Equal(t, 1, len(res))
	require.True(t, authority.Equals(res[0]))
}
