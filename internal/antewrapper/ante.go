package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
func (r FeeMeterContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	newCtx := ctx.WithGasMeter(NewFeeGasMeterWrapper(ctx.Logger(), ctx.GasMeter(), simulate))
	return next(newCtx, tx, simulate)
}

// GetFeeTx coverts the provided Tx to a FeeTx if possible.
func GetFeeTx(tx sdk.Tx) (sdk.FeeTx, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, sdkerrors.ErrTxDecode.Wrap("Tx must be a FeeTx")
	}
	return feeTx, nil
}

// GetFeeGasMeter gets a FeeGasMeter from the provided context.
func GetFeeGasMeter(ctx sdk.Context) (*FeeGasMeter, error) {
	feeGasMeter, ok := ctx.GasMeter().(*FeeGasMeter)
	if !ok {
		return nil, sdkerrors.ErrLogic.Wrap("gas meter is not a FeeGasMeter")
	}
	return feeGasMeter, nil
}
