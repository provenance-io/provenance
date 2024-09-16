package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestKeysContainModuleName(t *testing.T) {
	assert.Contains(t, bypassKey, ModuleName, "bypassKey")
	assert.Contains(t, transferAgentKey, ModuleName, "transferAgentKey")
}

func TestContextCombos(t *testing.T) {
	newCtx := func() sdk.Context {
		return sdk.NewContext(nil, cmtproto.Header{}, false, nil)
	}

	tests := []struct {
		name      string
		ctx       sdk.Context
		expBypass bool
		expTAs    []sdk.AccAddress
	}{
		{
			name:      "with transfer agent on with bypass",
			ctx:       WithTransferAgents(WithBypass(newCtx()), sdk.AccAddress("some_transfer_agent_")),
			expBypass: true,
			expTAs:    []sdk.AccAddress{sdk.AccAddress("some_transfer_agent_")},
		},
		{
			name:      "with bypass on with transfer agents",
			ctx:       WithBypass(WithTransferAgents(newCtx(), sdk.AccAddress("other_transfer_agent"), sdk.AccAddress("third_transfer_agent"))),
			expBypass: true,
			expTAs:    []sdk.AccAddress{sdk.AccAddress("other_transfer_agent"), sdk.AccAddress("third_transfer_agent")},
		},
		{
			name:      "without either on with transfer agent and bypass",
			ctx:       WithoutBypass(WithoutTransferAgents(WithBypass(WithTransferAgents(newCtx(), sdk.AccAddress("bad_transfer_agent__"))))),
			expBypass: false,
			expTAs:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actBypass bool
			testHasBypass := func() {
				actBypass = HasBypass(tc.ctx)
			}
			require.NotPanics(t, testHasBypass, "HasBypass")
			var actTAs []sdk.AccAddress
			testGetTransferAgents := func() {
				actTAs = GetTransferAgents(tc.ctx)
			}
			require.NotPanics(t, testGetTransferAgents, "GetTransferAgents")
			assert.Equal(t, tc.expBypass, actBypass, "HasBypass")
			assert.Equal(t, tc.expTAs, actTAs, "GetTransferAgents")
		})
	}
}

func TestBypassFuncs(t *testing.T) {
	tests := []struct {
		name string
		ctx  sdk.Context
		exp  bool
	}{
		{
			name: "brand new mostly empty context",
			ctx:  sdk.NewContext(nil, cmtproto.Header{}, false, nil),
			exp:  false,
		},
		{
			name: "context with bypass",
			ctx:  WithBypass(sdk.NewContext(nil, cmtproto.Header{}, false, nil)),
			exp:  true,
		},
		{
			name: "context with bypass on one that originally was without it",
			ctx:  WithBypass(WithoutBypass(sdk.NewContext(nil, cmtproto.Header{}, false, nil))),
			exp:  true,
		},
		{
			name: "context with bypass twice",
			ctx:  WithBypass(WithBypass(sdk.NewContext(nil, cmtproto.Header{}, false, nil))),
			exp:  true,
		},
		{
			name: "context without bypass",
			ctx:  WithoutBypass(sdk.NewContext(nil, cmtproto.Header{}, false, nil)),
			exp:  false,
		},
		{
			name: "context without bypass on one that originally had it",
			ctx:  WithoutBypass(WithBypass(sdk.NewContext(nil, cmtproto.Header{}, false, nil))),
			exp:  false,
		},
		{
			name: "context without bypass twice",
			ctx:  WithoutBypass(WithoutBypass(sdk.NewContext(nil, cmtproto.Header{}, false, nil))),
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = HasBypass(tc.ctx)
			}
			require.NotPanics(t, testFunc, "HasBypass")
			assert.Equal(t, tc.exp, actual, "HasBypass")
		})
	}
}

func TestBypassFuncsDoNotModifyProvided(t *testing.T) {
	origCtx := sdk.NewContext(nil, cmtproto.Header{}, false, nil)
	assert.False(t, HasBypass(origCtx), "HasBypass(origCtx)")
	afterWith := WithBypass(origCtx)
	assert.True(t, HasBypass(afterWith), "HasBypass(afterWith)")
	assert.False(t, HasBypass(origCtx), "HasBypass(origCtx) after giving it to WithBypass")
	afterWithout := WithoutBypass(afterWith)
	assert.False(t, HasBypass(afterWithout), "HasBypass(afterWithout)")
	assert.True(t, HasBypass(afterWith), "HasBypass(afterWith) after giving it to WithoutBypass")
	assert.False(t, HasBypass(origCtx), "HasBypass(origCtx) after giving afterWith to WithoutBypass")
}

func TestTransferAgentFuncs(t *testing.T) {
	newCtx := func() sdk.Context {
		return sdk.NewContext(nil, cmtproto.Header{}, false, nil)
	}

	tests := []struct {
		name string
		ctx  sdk.Context
		exp  []sdk.AccAddress
	}{
		{
			name: "brand new mostly empty context",
			ctx:  newCtx(),
			exp:  nil,
		},
		{
			name: "one transfer agent",
			ctx:  WithTransferAgents(newCtx(), sdk.AccAddress("transfer_agent______")),
			exp:  []sdk.AccAddress{sdk.AccAddress("transfer_agent______")},
		},
		{
			name: "two transfer agents",
			ctx:  WithTransferAgents(newCtx(), sdk.AccAddress("transfer_agent_one__"), sdk.AccAddress("transfer_agent_two__")),
			exp:  []sdk.AccAddress{sdk.AccAddress("transfer_agent_one__"), sdk.AccAddress("transfer_agent_two__")},
		},
		{
			name: "without transfer agents",
			ctx:  WithoutTransferAgents(newCtx()),
			exp:  nil,
		},
		{
			name: "one transfer agent on context already with one",
			ctx: WithTransferAgents(
				WithTransferAgents(newCtx(), sdk.AccAddress("first_transfer_agent")),
				sdk.AccAddress("agent_2_of_transfer_"),
			),
			exp: []sdk.AccAddress{sdk.AccAddress("agent_2_of_transfer_")},
		},
		{
			name: "two transfer agents on context already with one",
			ctx: WithTransferAgents(
				WithTransferAgents(newCtx(), sdk.AccAddress("first_transfer_agent")),
				sdk.AccAddress("agent_of_transfer_2_"), sdk.AccAddress("agent_of_chaos______"),
			),
			exp: []sdk.AccAddress{sdk.AccAddress("agent_of_transfer_2_"), sdk.AccAddress("agent_of_chaos______")},
		},
		{
			name: "one transfer agent on context already with two",
			ctx: WithTransferAgents(
				WithTransferAgents(newCtx(), sdk.AccAddress("1st_transfer_agent__"), sdk.AccAddress("2nd_transfer_agent__")),
				sdk.AccAddress("another_agent_______"),
			),
			exp: []sdk.AccAddress{sdk.AccAddress("another_agent_______")},
		},
		{
			name: "two transfer agents on context already with two",
			ctx: WithTransferAgents(
				WithTransferAgents(newCtx(), sdk.AccAddress("transfer_agent_one__"), sdk.AccAddress("transfer_agent_two__")),
				sdk.AccAddress("transfer_agent_three"), sdk.AccAddress("transfer_agent_four_"),
			),
			exp: []sdk.AccAddress{sdk.AccAddress("transfer_agent_three"), sdk.AccAddress("transfer_agent_four_")},
		},
		{
			name: "context without transfer agent twice",
			ctx:  WithoutTransferAgents(WithoutTransferAgents(newCtx())),
			exp:  nil,
		},
		{
			name: "context with one transfer agent on one that originally was without it",
			ctx:  WithTransferAgents(WithoutTransferAgents(newCtx()), sdk.AccAddress("agent_of_transfer___")),
			exp:  []sdk.AccAddress{sdk.AccAddress("agent_of_transfer___")},
		},
		{
			name: "context with two transfer agent on one that originally was without it",
			ctx:  WithTransferAgents(WithoutTransferAgents(newCtx()), sdk.AccAddress("agent_of_transfer_1_"), sdk.AccAddress("agent_of_transfer_2_")),
			exp:  []sdk.AccAddress{sdk.AccAddress("agent_of_transfer_1_"), sdk.AccAddress("agent_of_transfer_2_")},
		},
		{
			name: "context without transfer agent on one that originally had one",
			ctx:  WithoutTransferAgents(WithTransferAgents(newCtx(), sdk.AccAddress("the_transfer_agent__"))),
			exp:  nil,
		},
		{
			name: "context without transfer agent on one that originally had two",
			ctx:  WithoutTransferAgents(WithTransferAgents(newCtx(), sdk.AccAddress("the_transfer_agent__"), sdk.AccAddress("other_transfer_agent"))),
			exp:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []sdk.AccAddress
			testFunc := func() {
				actual = GetTransferAgents(tc.ctx)
			}
			require.NotPanics(t, testFunc, "GetTransferAgents")
			assert.Equal(t, tc.exp, actual, "GetTransferAgents")
		})
	}
}

func TestTransferAgentFuncsDoNotModifyProvided(t *testing.T) {
	origCtx := sdk.NewContext(nil, cmtproto.Header{}, false, nil)
	assert.Nil(t, GetTransferAgents(origCtx), "GetTransferAgents(origCtx)")

	ta := sdk.AccAddress("great_transfer_agent")
	expAgents := []sdk.AccAddress{ta}
	afterWith := WithTransferAgents(origCtx, ta)
	assert.Equal(t, expAgents, GetTransferAgents(afterWith), "GetTransferAgents(afterWith)")
	assert.Nil(t, GetTransferAgents(origCtx), "GetTransferAgents(origCtx) after giving it to WithTransferAgents")

	afterWithout := WithoutTransferAgents(afterWith)
	assert.Nil(t, GetTransferAgents(afterWithout), "GetTransferAgents(afterWithout)")
	assert.Equal(t, expAgents, GetTransferAgents(afterWith), "GetTransferAgents(afterWith) after giving it to WithoutTransferAgents")
	assert.Nil(t, GetTransferAgents(origCtx), "GetTransferAgents(origCtx) after giving afterWith to WithoutTransferAgents")
}
