package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GasTracerContextDecorator is an AnteDecorator that wraps the current
// context gas meter with one that outputs debug logging and telemetry
// whenever gas is consumed on the meter.
type GasTracerContextDecorator struct{}

// NewGasTracerContextDecorator creates a new GasTracerContextDecorator
func NewGasTracerContextDecorator() GasTracerContextDecorator {
	return GasTracerContextDecorator{}
}

var _ sdk.AnteDecorator = GasTracerContextDecorator{}

// AnteHandle implements the AnteDecorator.AnteHandle method
func (r GasTracerContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	baseMeter := (ctx.GasMeter()).(sdk.GasMeter)
	newCtx = ctx.WithGasMeter(NewTracingMeterWrapper(ctx.Logger(), baseMeter))

	return next(newCtx, tx, simulate)
}

// FeeMeterContextDecorator is an AnteDecorator that wraps the current
// context gas meter with a msg based fee meter.
type FeeMeterContextDecorator struct{}

// NewFeeMeterContextDecorator creates a new FeeMeterContextDecorator
func NewFeeMeterContextDecorator() FeeMeterContextDecorator {
	return FeeMeterContextDecorator{}
}

var _ sdk.AnteDecorator = FeeMeterContextDecorator{}

// AnteHandle implements the AnteDecorator.AnteHandle method
func (r FeeMeterContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	newCtx = ctx.WithGasMeter(NewFeeTracingMeterWrapper(ctx.Logger(), ctx.GasMeter()))

	return next(newCtx, tx, simulate)
}
