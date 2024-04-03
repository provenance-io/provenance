package sanction

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var bypassKey = "bypass-sanction-restriction"

// WithBypass returns a new context that will cause the sanction bank send restriction to be skipped.
func WithBypass[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(bypassKey, true)
	return context.Context(sdkCtx).(C)
}

// WithoutBypass returns a new context that will cause the sanction bank send restriction to not be skipped.
func WithoutBypass[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(bypassKey, false)
	return context.Context(sdkCtx).(C)
}

// HasBypass checks the context to see if the sanction bank send restriction should be skipped.
func HasBypass[C context.Context](ctx C) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	bypassValue := sdkCtx.Value(bypassKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}
