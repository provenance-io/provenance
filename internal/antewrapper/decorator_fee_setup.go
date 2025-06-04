package antewrapper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FlatFeeSetupDecorator is an AnteHandler that calculates costs for the msgs, and ensures a sufficient fee is provided.
type FlatFeeSetupDecorator struct{}

func NewFlatFeeSetupDecorator() FlatFeeSetupDecorator {
	return FlatFeeSetupDecorator{}
}

func (d FlatFeeSetupDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return ctx, err
	}

	gasMeter, err := GetFlatFeeGasMeter(ctx)
	if err != nil {
		return ctx, err
	}

	// Calculate and set the costs in the gas meter.
	err = gasMeter.SetCosts(ctx, feeTx.GetMsgs())
	if err != nil {
		return ctx, fmt.Errorf("could not calculate msg costs: %w", err)
	}
	feeProvided := feeTx.GetFee()
	gasMeter.adjustCostsForUnitTests(ctx.ChainID(), feeProvided)

	// Make sure the fee provided is enough. There's a chance that costs/fees are added during the execution
	// of the tx, so we'll check this again later. We check it here too, though, in order to skip processing
	// the tx msgs if we know now that there isn't enough fee for what's been provided.
	// Skip if simulating since the fee is probably what they're trying to find out.
	// Skip during init genesis too since those should be free (and there's no one to pay).
	if !simulate && !IsInitGenesis(ctx) {
		reqFee := gasMeter.GetRequiredFee()
		err = validateFeeAmount(reqFee, feeProvided)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}
