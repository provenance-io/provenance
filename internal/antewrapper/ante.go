package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FeeMeterContextDecorator is an AnteDecorator that wraps the current
// context gas meter with a msg based fee meter.
// Also, it merges functionality from GasTracerContextDecorator in previous versions
// which provided an AnteDecorator that wraps the current
// context gas meter with one that outputs debug logging and telemetry
// whenever gas is consumed on the meter.
type FeeMeterContextDecorator struct{}

// NewFeeMeterContextDecorator creates a new FeeMeterContextDecorator
func NewFeeMeterContextDecorator() FeeMeterContextDecorator {
	return FeeMeterContextDecorator{}
}

var _ sdk.AnteDecorator = FeeMeterContextDecorator{}

// AnteHandle implements the AnteDecorator.AnteHandle method
func (r FeeMeterContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	newCtx = ctx.WithGasMeter(NewFeeGasMeterWrapper(ctx.Logger(), ctx.GasMeter(), simulate))
	return next(newCtx, tx, simulate)
}
