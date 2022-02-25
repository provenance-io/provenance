package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TxGasLimitDecorator will check if the transaction's gas amount is higher than
// 5% of the maximum gas allowed per block.
// If gas is too high, decorator returns error and tx is rejected from mempool.
// If gas is below the limit, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use TxGasLimitDecorator
type TxGasLimitDecorator struct{}

// MinTxPerBlock is used to determine the maximum amount of gas that any given transaction can use based on the block gas limit.
const MinTxPerBlock = 20

func NewTxGasLimitDecorator() TxGasLimitDecorator {
	return TxGasLimitDecorator{}
}

func (mfd TxGasLimitDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Ensure that the requested gas does not exceed the configured block maximum
	gas := feeTx.GetGas()
	gasTxLimit := ctx.BlockGasMeter().Limit() * (1 / MinTxPerBlock)

	// TODO - remove "gasTxLimit > 0" with SDK 0.46 which fixes the infinite gas meter to use max int vs zero for the limit.
	if gasTxLimit > 0 && gas > gasTxLimit {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrTxTooLarge, "transaction gas exceeds maximum allowed; got: %d max allowed: %d", gas, gasTxLimit)
	}

	return next(ctx, tx, simulate)
}
