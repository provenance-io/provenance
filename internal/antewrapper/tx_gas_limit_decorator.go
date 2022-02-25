package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TxGasLimitDecorator will check if the transaction's gas amount is higher than
// 5% of the maximum gas allowed per block.
// If gas is too high, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If gas is below the limit or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use TxGasLimitDecorator
type TxGasLimitDecorator struct{}

// MIN_TX_PER_BLOCK is used to determine the maximum amount of gas that any given transaction can use based on the block gas limit.
const MIN_TX_PER_BLOCK = 20

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
	gas_tx_limit := ctx.BlockGasMeter().Limit() * (1 / MIN_TX_PER_BLOCK)

	// TODO - remove "gas_tx_limit > 0" with SDK 0.46 which fixes the infinite gas meter to use max int vs zero for the limit.
	if gas_tx_limit > 0 && gas > gas_tx_limit {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrTxTooLarge, "transaction gas exceeds maximum allowed; got: %s max allowed: %s", gas, gas_tx_limit)
	}

	return next(ctx, tx, simulate)
}
