package sdk

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFeeGrantContextFuncs(t *testing.T) {
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
			name: "context with fee grant in use",
			ctx:  WithFeeGrantInUse(sdk.NewContext(nil, cmtproto.Header{}, false, nil)),
			exp:  true,
		},
		{
			name: "context with fee grant in use on one that originally was without it",
			ctx:  WithFeeGrantInUse(WithoutFeeGrantInUse(sdk.NewContext(nil, cmtproto.Header{}, false, nil))),
			exp:  true,
		},
		{
			name: "context with fee grant in use twice",
			ctx:  WithFeeGrantInUse(WithFeeGrantInUse(sdk.NewContext(nil, cmtproto.Header{}, false, nil))),
			exp:  true,
		},
		{
			name: "context without fee grant in use",
			ctx:  WithoutFeeGrantInUse(sdk.NewContext(nil, cmtproto.Header{}, false, nil)),
			exp:  false,
		},
		{
			name: "context without fee grant in use on one that originally had it",
			ctx:  WithoutFeeGrantInUse(WithFeeGrantInUse(sdk.NewContext(nil, cmtproto.Header{}, false, nil))),
			exp:  false,
		},
		{
			name: "context without fee grant in use twice",
			ctx:  WithoutFeeGrantInUse(WithoutFeeGrantInUse(sdk.NewContext(nil, cmtproto.Header{}, false, nil))),
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := HasFeeGrantInUse(tc.ctx)
			assert.Equal(t, tc.exp, actual, "HasFeeGrantInUse")
		})
	}
}

func TestFeeGrantInUseFuncsDoNotModifyProvided(t *testing.T) {
	origCtx := sdk.NewContext(nil, cmtproto.Header{}, false, nil)
	assert.False(t, HasFeeGrantInUse(origCtx), "HasFeeGrantInUse(origCtx)")
	afterWith := WithFeeGrantInUse(origCtx)
	assert.True(t, HasFeeGrantInUse(afterWith), "HasFeeGrantInUse(afterWith)")
	assert.False(t, HasFeeGrantInUse(origCtx), "HasFeeGrantInUse(origCtx) after giving it to WithFeeGrantInUse")
	afterWithout := WithoutFeeGrantInUse(afterWith)
	assert.False(t, HasFeeGrantInUse(afterWithout), "HasFeeGrantInUse(afterWithout)")
	assert.True(t, HasFeeGrantInUse(afterWith), "HasFeeGrantInUse(afterWith) after giving it to WithoutFeeGrantInUse")
	assert.False(t, HasFeeGrantInUse(origCtx), "HasFeeGrantInUse(origCtx) after giving afterWith to WithoutFeeGrantInUse")
}
