package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MinGasPricesDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MinGasPricesDecorator
type MinGasPricesDecorator struct{}

func NewMinGasPricesDecorator() MinGasPricesDecorator {
	return MinGasPricesDecorator{}
}

func (mfd MinGasPricesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !simulate {
		err := checkTxFeeWithValidatorMinGasPrices(ctx, tx)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// checkTxFeeWithValidatorMinGasPrices makes sure one or more of the fee coins has enough to cover
// the validator's min gas fee.
func checkTxFeeWithValidatorMinGasPrices(ctx sdk.Context, tx sdk.Tx) error {
	// Note: This is copied from Cosmos-SDK:x/auth/ante/validator_tx_fee.go and tweaked as follows:
	//	1. The priority return value and call to calculate the priority has been removed because we
	//		probably don't want the naive approach they have, and we don't know what we want yet.
	//		Also, the priority mempool isn't fully ready yet.
	//	2. The Coins return value has been removed because we use our network's floor gas price instead of
	//		the validators min gas prices when deciding the fee to deduct.
	//	3. Use of deprecated sdkerrors.Wrap and .Wrapf has been fixed.
	//	4. The first lines were updated to use GetFeeTx.
	//	5. The content of the final error message was updated to hopefully avoid confusion with the floor gas price.
	//  6. The comment above the function was fixed.
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return err
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() {
		minGasPrices := ctx.MinGasPrices()
		if !minGasPrices.IsZero() {
			requiredFees := make(sdk.Coins, len(minGasPrices))

			// Determine the required fees by multiplying each required minimum gas
			// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
			glDec := sdk.NewDec(int64(gas))
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
			}

			if !feeCoins.IsAnyGTE(requiredFees) {
				return sdkerrors.ErrInsufficientFee.Wrapf("min-gas-prices not met; got: %s required: %s", feeCoins, requiredFees)
			}
		}
	}

	return nil
}
