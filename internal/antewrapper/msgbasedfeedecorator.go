package antewrapper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	cosmosauthtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	msgbasedfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// MsgBasedFeeDecorator will check if the transaction's fee is at least as large
// as tax + additional minimum gasFee (defined in msgfeeskeeper)
// and record additional fee proceeds to msgfees module to track additional fee proceeds.
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MsgBasedFeeDecorator
type MsgBasedFeeDecorator struct {
	msgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper
}

func NewMsgBasedFeeDecorator(bankKeeper cosmosauthtypes.BankKeeper, accountKeeper cosmosante.AccountKeeper, feegrantKeeper cosmosante.FeegrantKeeper, keeper msgbasedfeetypes.MsgBasedFeeKeeper) MsgBasedFeeDecorator {
	return MsgBasedFeeDecorator{
		keeper,
	}
}

// AnteHandle handles msg tax fee checking
func (afd MsgBasedFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// has to be FeeTx type
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	ctx.Logger().Info("NOTICE: here in MsgBasedFeeDecorator {}",ctx.GasMeter().GasConsumed())
	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	if !simulate {
		// Compute taxes
		fees, _ := FilterMsgAndComputeTax(ctx, afd.msgBasedFeeKeeper, msgs...)

		// Mempool fee validation
		// No fee validation for oracle txs
		if ctx.IsCheckTx() {
			if err := EnsureSufficientMempoolFees(ctx, gas, feeCoins, fees); err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, err.Error())
			}
		}

		// Ensure paid fee is enough to cover taxes
		if _, hasNeg := feeCoins.SafeSub(fees); hasNeg {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, fees)
		}

		// emit event of additional fees
		// TODO  should we record this in the keeper.
		if !fees.IsZero() {
			// emit event.

			// maybe record additional fees charged.
		}
	}


	return next(ctx, tx, simulate)
}

// EnsureSufficientMempoolFees verifies that the given transaction has supplied
// enough fees(gas + stability) to cover a proposer's minimum fees. A result object is returned
// indicating success or failure.
//
// Contract: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientMempoolFees(ctx sdk.Context, gas uint64, feeCoins sdk.Coins, taxes sdk.Coins) error {
	requiredFees := sdk.Coins{}
	minGasPrices := ctx.MinGasPrices()
	if !minGasPrices.IsZero() {
		requiredFees = make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	// Before checking gas prices, remove taxed from fee
	var hasNeg bool
	if feeCoins, hasNeg = feeCoins.SafeSub(taxes); hasNeg {
		return fmt.Errorf("insufficient fees; got: %q, required: %q = %q(gas) +%q(stability)", feeCoins.Add(taxes...), requiredFees.Add(taxes...), requiredFees, taxes)
	}

	if !requiredFees.IsZero() && !feeCoins.IsAnyGTE(requiredFees) {
		return fmt.Errorf("insufficient fees; got: %q, required: %q = %q(gas) +%q(stability)", feeCoins.Add(taxes...), requiredFees.Add(taxes...), requiredFees, taxes)
	}

	return nil
}



// FilterMsgAndComputeTax computes the stability tax on MsgSend and MsgMultiSend.
func FilterMsgAndComputeTax(ctx sdk.Context, mbfk msgbasedfeetypes.MsgBasedFeeKeeper, msgs ...sdk.Msg) (sdk.Coins, error) {
	taxes := sdk.Coins{}
	// get the msg fee

	additionalFees := sdk.Coins{}

	for _, msg := range msgs {
		typeURL := sdk.MsgTypeURL(msg)
		msgFees, err := mbfk.GetMsgBasedFee(ctx, typeURL)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}

		if msgFees == nil {
			continue
		}
		if msgFees.MinAdditionalFee.IsPositive() {
			additionalFees = additionalFees.Add(sdk.NewCoin(msgFees.MinAdditionalFee.Denom, msgFees.MinAdditionalFee.Amount))
		}

	}


	return taxes, nil
}
