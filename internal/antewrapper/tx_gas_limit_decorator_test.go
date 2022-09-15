package antewrapper

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	nametypes "github.com/provenance-io/provenance/x/name/types"
)

func TestIsGovernanceMessage(t *testing.T) {
	tests := []struct {
		msg sdk.Msg
		exp bool
	}{
		{&govtypesv1beta1.MsgSubmitProposal{}, true},
		{&govtypesv1.MsgSubmitProposal{}, true},
		{&govtypesv1.MsgExecLegacyContent{}, true},
		{&govtypesv1beta1.MsgVote{}, true},
		{&govtypesv1.MsgVote{}, true},
		{&govtypesv1beta1.MsgDeposit{}, true},
		{&govtypesv1.MsgDeposit{}, true},
		{&govtypesv1beta1.MsgVoteWeighted{}, true},
		{&govtypesv1.MsgVoteWeighted{}, true},
		{nil, false},
		{&authztypes.MsgGrant{}, false},
		{&nametypes.MsgBindNameRequest{}, false},
	}

	for _, tc := range tests {
		name := "nil"
		if tc.msg != nil {
			name = sdk.MsgTypeURL(tc.msg)[1:]
		}
		t.Run(name, func(tt *testing.T) {
			act := isGovMessage(tc.msg)
			assert.Equal(tt, tc.exp, act, "isGovMessage")
		})
	}
}

func TestIsOnlyGovMsgs(t *testing.T) {
	tests := []struct {
		name string
		msgs []sdk.Msg
		exp  bool
	}{
		{
			name: "empty",
			msgs: []sdk.Msg{},
			exp:  false,
		},
		{
			name: "only MsgSubmitProposal",
			msgs: []sdk.Msg{&govtypesv1beta1.MsgSubmitProposal{}},
			exp:  true,
		},
		{
			name: "all the gov messages",
			msgs: []sdk.Msg{
				&govtypesv1beta1.MsgSubmitProposal{},
				&govtypesv1.MsgSubmitProposal{},
				&govtypesv1.MsgExecLegacyContent{},
				&govtypesv1beta1.MsgVote{},
				&govtypesv1.MsgVote{},
				&govtypesv1beta1.MsgDeposit{},
				&govtypesv1.MsgDeposit{},
				&govtypesv1beta1.MsgVoteWeighted{},
				&govtypesv1.MsgVoteWeighted{},
			},
			exp: true,
		},
		{
			name: "MsgSubmitProposal and MsgStoreCode",
			msgs: []sdk.Msg{
				&govtypesv1beta1.MsgSubmitProposal{},
				&wasmtypes.MsgStoreCode{},
			},
			exp: false,
		},
		{
			name: "MsgStoreCode and MsgSubmitProposal",
			msgs: []sdk.Msg{
				&wasmtypes.MsgStoreCode{},
				&govtypesv1beta1.MsgSubmitProposal{},
			},
			exp: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			act := isOnlyGovMsgs(tc.msgs)
			assert.Equal(tt, tc.exp, act, "isOnlyGovMsgs")
		})
	}
}

func TestFoo(t *testing.T) {
	var coins sdk.Coins
	assert.Nil(t, coins)
	assert.True(t, coins.IsZero())
}
