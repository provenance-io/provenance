package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TxGasLimitDecorator set the gas limit to be 4m for every tx
type TxGasLimitDecorator struct{}

func NewTxGasLimitDecorator() TxGasLimitDecorator {
	return TxGasLimitDecorator{}
}

func (mfd TxGasLimitDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	_, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	// Set GasLimit to 4m, if a tx exceeds it, then it will give out of gas error.
	gasTxLimit := uint64(4_000_000)
	return next(ctx.WithGasMeter(sdk.NewGasMeter(gasTxLimit)), tx, simulate)
}
