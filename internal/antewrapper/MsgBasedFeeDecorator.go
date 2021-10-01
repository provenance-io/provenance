package antewrapper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	k "github.com/provenance-io/provenance/x/msgfees/keeper"
)

// AdditionalFeeDecorator will check if the transaction's fee is at least as large
// as tax + additional minimum gasFee (defined in msgfeeskeeper)
// and record additional fee proceeds to msgfees module to track additional fee proceeds.
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use AdditionalFeeDecorator
type AdditionalFeeDecorator struct {
	FeeKeeper k.MsgBasedFeeKeeperI
}

// AnteHandle handles msg tax fee checking
func (afd AdditionalFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// has to be FeeTx type
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	if !simulate {
		// Compute taxes
		fees := FilterMsgAndComputeTax(ctx, afd.FeeKeeper, msgs...)

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

// DeductFees deducts additional fees from the given account.
// same as deduct fees for now.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc types.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}

// FilterMsgAndComputeTax computes the stability tax on MsgSend and MsgMultiSend.
func FilterMsgAndComputeTax(ctx sdk.Context, tk k.MsgBasedFeeKeeperI, msgs ...sdk.Msg) sdk.Coins {
	taxes := sdk.Coins{}
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *banktypes.MsgSend:
			taxes = taxes.Add(computeFees(ctx, tk, msg.Amount)...)

		case *banktypes.MsgMultiSend:
			for _, input := range msg.Inputs {
				taxes = taxes.Add(computeFees(ctx, tk, input.Coins)...)
			}

		case *authz.MsgExec:
			messages, err := msg.GetMessages()
			if err != nil {
				panic(err)
			}

			taxes = taxes.Add(FilterMsgAndComputeTax(ctx, tk, messages...)...)
		}
	}

	return taxes
}

// computes the fees
func computeFees(ctx sdk.Context, tk k.MsgBasedFeeKeeperI, principal sdk.Coins) sdk.Coins {
	return nil
}
