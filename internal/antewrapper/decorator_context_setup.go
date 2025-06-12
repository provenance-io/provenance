package antewrapper

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// ProvSetUpContextDecorator creates and sets the flat-fee GasMeter in the Context and wraps the
// next AnteHandler with a defer clause to recover from any downstream OutOfGas panics in the
// AnteHandler chain to return an error with information on gas provided and gas used.
// CONTRACT: Must be first decorator in the chain.
// CONTRACT: Tx must implement GasTx interface.
// This is similar to "github.com/cosmos/cosmos-sdk/x/auth/ante".SetUpContextDecorator
// except we set and check the gas limits a little differently.
type ProvSetUpContextDecorator struct {
	ffk FlatFeesKeeper
}

func NewProvSetUpContextDecorator(ffk FlatFeesKeeper) ProvSetUpContextDecorator {
	return ProvSetUpContextDecorator{ffk: ffk}
}

func (d ProvSetUpContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	ctx.Logger().Debug("Starting ProvSetUpContextDecorator.AnteHandle.", "simulate", simulate, "IsCheckTx", ctx.IsCheckTx())

	// All transactions must implement FeeTx.
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		// Set a gas meter with limit 0 as to prevent an infinite gas meter attack during runTx.
		newCtx = ante.SetGasMeter(simulate, ctx, 0)
		return newCtx, err
	}

	// Get the actual gas wanted for this tx (accounting for our custom simulation process).
	gasWanted, err := GetGasWanted(ctx.Logger(), feeTx)
	if err != nil {
		// Set a gas meter with limit 0 as to prevent an infinite gas meter attack during runTx.
		newCtx = ante.SetGasMeter(simulate, ctx, 0)
		return newCtx, err
	}

	// Set a generic gas meter in the context with the appropriate amount of gas.
	// Note that SetGasMeter uses an infinite gas meter if simulating or at height 0 (init genesis).
	newCtx = ante.SetGasMeter(simulate, ctx, gasWanted)
	// Now wrap that gas meter in our flat-fee gas meter.
	newCtx = ctx.WithGasMeter(NewFlatFeeGasMeter(newCtx.GasMeter(), newCtx.Logger(), d.ffk))
	// Note: We don't set the costs yet, because we want to check a few more things before doing that work.

	// Ensure that the requested gas does not exceed either the configured block maximum, or the tx maximum.
	// If there's no block maximum defined, we can't do that check, and we interpret that as an indication
	// that there shouldn't be a tx limit either.
	if bp := ctx.ConsensusParams().Block; bp != nil {
		maxBlockGas := bp.GetMaxGas()
		if maxBlockGas > 0 {
			if gasWanted > uint64(maxBlockGas) {
				return newCtx, sdkerrors.ErrInvalidGasLimit.Wrapf("tx gas limit %d exceeds block max gas %d", gasWanted, maxBlockGas)
			}
			if txGasLimitShouldApply(ctx.ChainID(), tx.GetMsgs()) && gasWanted > TxGasLimit {
				return newCtx, sdkerrors.ErrInvalidGasLimit.Wrapf("tx gas limit %d exceeds tx max gas %d", gasWanted, TxGasLimit)
			}
		}
	}

	// Decorator will catch an OutOfGasPanic caused in the next antehandler
	// AnteHandlers must have their own defer/recover in order for the BaseApp
	// to know how much gas was used! This is because the GasMeter is created in
	// the AnteHandler, but if it panics the context won't be set properly in
	// runTx's recover call.
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case storetypes.ErrorOutOfGas:
				err = sdkerrors.ErrOutOfGas.Wrapf("out of gas in location: %v; gasWanted: %d, gasUsed: %d",
					rType.Descriptor, newCtx.GasMeter().Limit(), newCtx.GasMeter().GasConsumed())
			default:
				panic(r)
			}
		}
	}()

	return next(newCtx, tx, simulate)
}
