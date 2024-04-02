package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	bypassKey        = "bypass-marker-restriction"
	transferAgentKey = "marker-transfer-agent"
)

// WithBypass returns a new context that will cause the marker bank send restriction to be skipped.
func WithBypass[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(bypassKey, true)
	return context.Context(sdkCtx).(C)
}

// WithoutBypass returns a new context that will cause the marker bank send restriction to not be skipped.
func WithoutBypass[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(bypassKey, false)
	return context.Context(sdkCtx).(C)
}

// HasBypass checks the context to see if the marker bank send restriction should be skipped.
func HasBypass[C context.Context](ctx C) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	bypassValue := sdkCtx.Value(bypassKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}

// WithTransferAgent returns a new context that contains the provided marker transfer agent.
func WithTransferAgent[C context.Context](ctx C, transferAgent sdk.AccAddress) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(transferAgentKey, transferAgent)
	return context.Context(sdkCtx).(C)
}

// WithoutTransferAgent returns a new context with a nil marker transfer agent.
func WithoutTransferAgent[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(transferAgentKey, sdk.AccAddress(nil))
	return context.Context(sdkCtx).(C)
}

// GetTransferAgent gets the marker transfer agent from the provided context.
func GetTransferAgent[C context.Context](ctx C) sdk.AccAddress {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	val := sdkCtx.Value(transferAgentKey)
	if val == nil {
		return nil
	}
	rv, _ := val.(sdk.AccAddress)
	return rv
}
