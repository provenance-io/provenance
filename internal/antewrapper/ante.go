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
type FeeMeterContextDecorator struct {
	SimulateFunc BaseAppSimulateFunc
	TxEncoder    sdk.TxEncoder
}

// NewFeeMeterContextDecorator creates a new FeeMeterContextDecorator
func NewFeeMeterContextDecorator(simulateFunc BaseAppSimulateFunc, txEncoder sdk.TxEncoder) FeeMeterContextDecorator {
	return FeeMeterContextDecorator{SimulateFunc: simulateFunc, TxEncoder: txEncoder}
}

var _ sdk.AnteDecorator = FeeMeterContextDecorator{}

// AnteHandle implements the AnteDecorator.AnteHandle method
func (r FeeMeterContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	bz, err := r.TxEncoder(tx)
	gasInfo, _, _, err := r.SimulateFunc(bz)
	println(gasInfo.GasUsed)
	if err != nil {
		return ctx, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	newCtx := ctx.WithGasMeter(NewFeeGasMeterWrapper(ctx.Logger(), ctx.GasMeter(), simulate))
	return next(newCtx, tx, simulate)
}

// GetFeeTx coverts the provided Tx to a FeeTx if possible.
func GetFeeTx(tx sdk.Tx) (sdk.FeeTx, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, sdkerrors.ErrTxDecode.Wrapf("Tx must be a FeeTx: %T", tx)
	}
	return feeTx, nil
}

// GetFeeGasMeter gets a FeeGasMeter from the provided context.
func GetFeeGasMeter(ctx sdk.Context) (*FeeGasMeter, error) {
	feeGasMeter, ok := ctx.GasMeter().(*FeeGasMeter)
	if !ok {
		return nil, sdkerrors.ErrLogic.Wrapf("gas meter is not a FeeGasMeter: %T", ctx.GasMeter())
	}
	return feeGasMeter, nil
}

// IsInitGenesis returns true if the context indicates we're in InitGenesis.
func IsInitGenesis(ctx sdk.Context) bool {
	// Note: This isn't fully accurate since you can initialize a chain at a height other than zero.
	// But it should be good enough for our stuff. Ideally we'd want something specifically set in
	// the context during InitGenesis to check, but that'd probably involve some SDK work.
	return ctx.BlockHeight() <= 0
}
