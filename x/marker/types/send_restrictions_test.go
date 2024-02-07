package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestKeysContainModuleName(t *testing.T) {
	assert.Contains(t, bypassKey, ModuleName, "bypassKey")
	assert.Contains(t, transferAgentKey, ModuleName, "transferAgentKey")
}

func TestContextCombos(t *testing.T) {
	newCtx := func() sdk.Context {
		return sdk.NewContext(nil, tmproto.Header{}, false, nil)
	}

	tests := []struct {
		name      string
		ctx       sdk.Context
		expBypass bool
		expTA     sdk.AccAddress
	}{
		{
			name:      "with transfer agent on with bypass",
			ctx:       WithTransferAgent(WithBypass(newCtx()), sdk.AccAddress("some_transfer_agent_")),
			expBypass: true,
			expTA:     sdk.AccAddress("some_transfer_agent_"),
		},
		{
			name:      "with bypass on with transfer agent",
			ctx:       WithBypass(WithTransferAgent(newCtx(), sdk.AccAddress("other_transfer_agent"))),
			expBypass: true,
			expTA:     sdk.AccAddress("other_transfer_agent"),
		},
		{
			name:      "without either on with transfer agent and bypass",
			ctx:       WithoutBypass(WithoutTransferAgent(WithBypass(WithTransferAgent(newCtx(), sdk.AccAddress("bad_transfer_agent__"))))),
			expBypass: false,
			expTA:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actBypass := HasBypass(tc.ctx)
			actTA := GetTransferAgent(tc.ctx)

			assert.Equal(t, tc.expBypass, actBypass, "HasBypass")
			assert.Equal(t, tc.expTA, actTA, "GetTransferAgent")
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
			ctx:  sdk.NewContext(nil, tmproto.Header{}, false, nil),
			exp:  false,
		},
		{
			name: "context with bypass",
			ctx:  WithBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil)),
			exp:  true,
		},
		{
			name: "context with bypass on one that originally was without it",
			ctx:  WithBypass(WithoutBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil))),
			exp:  true,
		},
		{
			name: "context with bypass twice",
			ctx:  WithBypass(WithBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil))),
			exp:  true,
		},
		{
			name: "context without bypass",
			ctx:  WithoutBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil)),
			exp:  false,
		},
		{
			name: "context without bypass on one that originally had it",
			ctx:  WithoutBypass(WithBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil))),
			exp:  false,
		},
		{
			name: "context without bypass twice",
			ctx:  WithoutBypass(WithoutBypass(sdk.NewContext(nil, tmproto.Header{}, false, nil))),
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := HasBypass(tc.ctx)
			assert.Equal(t, tc.exp, actual, "HasBypass")
		})
	}
}

func TestBypassFuncsDoNotModifyProvided(t *testing.T) {
	origCtx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
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
		return sdk.NewContext(nil, tmproto.Header{}, false, nil)
	}

	tests := []struct {
		name string
		ctx  sdk.Context
		exp  sdk.AccAddress
	}{
		{
			name: "brand new mostly empty context",
			ctx:  newCtx(),
			exp:  nil,
		},
		{
			name: "context with transfer agent",
			ctx:  WithTransferAgent(newCtx(), sdk.AccAddress("transfer_agent______")),
			exp:  sdk.AccAddress("transfer_agent______"),
		},
		{
			name: "context without transfer agent",
			ctx:  WithoutTransferAgent(newCtx()),
			exp:  nil,
		},
		{
			name: "context with transfer agent twice",
			ctx:  WithTransferAgent(WithTransferAgent(newCtx(), sdk.AccAddress("first_transfer_agent")), sdk.AccAddress("agent_2_of_transfer_")),
			exp:  sdk.AccAddress("agent_2_of_transfer_"),
		},
		{
			name: "context without transfer agent twice",
			ctx:  WithoutTransferAgent(WithoutTransferAgent(newCtx())),
			exp:  nil,
		},
		{
			name: "context with transfer agent on one that originally was without it",
			ctx:  WithTransferAgent(WithoutTransferAgent(newCtx()), sdk.AccAddress("agent_of_transfer___")),
			exp:  sdk.AccAddress("agent_of_transfer___"),
		},
		{
			name: "context without transfer agent on one that originally had it",
			ctx:  WithoutTransferAgent(WithTransferAgent(newCtx(), sdk.AccAddress("the_transfer_agent__"))),
			exp:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetTransferAgent(tc.ctx)
			assert.Equal(t, tc.exp, actual, "GetTransferAgent")
		})
	}
}

func TestTransferAgentFuncsDoNotModifyProvided(t *testing.T) {
	origCtx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	assert.Nil(t, GetTransferAgent(origCtx), "GetTransferAgent(origCtx)")

	ta := sdk.AccAddress("great_transfer_agent")
	afterWith := WithTransferAgent(origCtx, ta)
	assert.Equal(t, ta, GetTransferAgent(afterWith), "GetTransferAgent(afterWith)")
	assert.Nil(t, GetTransferAgent(origCtx), "GetTransferAgent(origCtx) after giving it to WithTransferAgent")

	afterWithout := WithoutTransferAgent(afterWith)
	assert.Nil(t, GetTransferAgent(afterWithout), "GetTransferAgent(afterWithout)")
	assert.Equal(t, ta, GetTransferAgent(afterWith), "GetTransferAgent(afterWith) after giving it to WithoutTransferAgent")
	assert.Nil(t, GetTransferAgent(origCtx), "GetTransferAgent(origCtx) after giving afterWith to WithoutTransferAgent")
}
