package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	bypassKey        = "bypass-marker-restriction"
	transferAgentKey = "marker-transfer-agent"
)

// WithBypass returns a new context that will cause the marker bank send restriction to be skipped.
func WithBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(bypassKey, true)
}

// WithoutBypass returns a new context that will cause the marker bank send restriction to not be skipped.
func WithoutBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(bypassKey, false)
}

// HasBypass checks the context to see if the marker bank send restriction should be skipped.
func HasBypass(ctx sdk.Context) bool {
	bypassValue := ctx.Value(bypassKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}

// WithTransferAgent returns a new context that contains the provided marker transfer agent.
func WithTransferAgent(ctx sdk.Context, transferAgent sdk.AccAddress) sdk.Context {
	return ctx.WithValue(transferAgentKey, transferAgent)
}

// WithoutTransferAgent returns a new context with a nil marker transfer agent.
func WithoutTransferAgent(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(transferAgentKey, sdk.AccAddress(nil))
}

// GetTransferAgent gets the marker transfer agent from the provided context.
func GetTransferAgent(ctx sdk.Context) sdk.AccAddress {
	val := ctx.Value(transferAgentKey)
	if val == nil {
		return nil
	}
	rv, _ := val.(sdk.AccAddress)
	return rv
}
