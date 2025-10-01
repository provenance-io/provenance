package sdk

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	feeGranteeKey = "pio-feegrant-in-use"
)

// WithFeeGrantInUse returns a new context that will indicate that a feegrant is being used.
func WithFeeGrantInUse[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(feeGranteeKey, true)
	return context.Context(sdkCtx).(C)
}

// WithoutFeeGrantInUse returns a new context that will indicate that a feegrant is NOT being used.
func WithoutFeeGrantInUse[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(feeGranteeKey, false)
	return context.Context(sdkCtx).(C)
}

// HasFeeGrantInUse checks the context to see if the a feegrant is being used.
func HasFeeGrantInUse[C context.Context](ctx C) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	bypassValue := sdkCtx.Value(feeGranteeKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}
