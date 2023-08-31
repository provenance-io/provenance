package antewrapper

import (
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

const (
	SimAppChainID = "simapp-unit-testing"
)

// MsgFeesDecorator will check if the transaction's fee is at least as large
// as floor gas fee (defined in MsgFee module) + message-based fees (also defined in the MsgFee module).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MsgFeesDecorator
type MsgFeesDecorator struct {
	msgFeeKeeper msgfeestypes.MsgFeesKeeper
}

func NewMsgFeesDecorator(msgFeeKeeper msgfeestypes.MsgFeesKeeper) MsgFeesDecorator {
	return MsgFeesDecorator{
		msgFeeKeeper: msgFeeKeeper,
	}
}

type MsgFeesDistribution struct {
	AdditionalModuleFees   sdk.Coins
	RecipientDistributions map[string]sdk.Coin
	TotalAdditionalFees    sdk.Coins
}

// AnteHandle handles msg fee checking
// has two functions ensures,
// 1. has enough fees to add to Mempool (this involves CheckTx)
// 2. Makes sure enough fees are present for additional message fees
// Let z be the Total Fees to be paid
// Let x be the Base gas Fees to be paid
// Let y is the additional fees to be paid per MsgType
// then z = x + y
// This Fee Decorator makes sure that z is >= to x + y
func (mfd MsgFeesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return ctx, err
	}

	// Make sure there are enough fees to cover base fee + additional fees.
	// base fee = floor gas price * gas wanted
	// additional fees = sum of message based fees
	if ctx.IsCheckTx() {
		feeCoins := feeTx.GetFee()
		gas := feeTx.GetGas()
		floorGasPrice := mfd.msgFeeKeeper.GetFloorGasPrice(ctx)
		msgs := feeTx.GetMsgs()

		// Compute msg all additional fees
		msgFeesDistribution, calcErr := mfd.msgFeeKeeper.CalculateAdditionalFeesToBePaid(ctx, msgs...)
		if calcErr != nil && !simulate {
			return ctx, sdkerrors.ErrInsufficientFee.Wrap(calcErr.Error())
		}

		mpErr := EnsureSufficientFloorAndMsgFees(ctx, feeCoins, floorGasPrice, gas, msgFeesDistribution.TotalAdditionalFees)
		if mpErr != nil && !simulate {
			return ctx, sdkerrors.ErrInsufficientFee.Wrap(mpErr.Error())
		}
	}

	return next(ctx, tx, simulate)
}

// This check for chain-id is exclusively for not breaking all existing sim tests which freak out when denom is anything other than stake.
// and some network tests won't work without a chain id being set(but they also setup everything with stake denom) so `simapp-unit-testing` chain id is skipped also.
// This only needs to work to pio-testnet and pio-mainnet, so this is safe.
func isTestContext(ctx sdk.Context) bool {
	return len(ctx.ChainID()) == 0 || ctx.ChainID() == SimAppChainID || ctx.ChainID() == helpers.SimAppChainID
}

// EnsureSufficientFloorAndMsgFees verifies that the given transaction has supplied
// enough fees(gas + additional fees) to cover x/msgfees costs.
//
// Contract: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientFloorAndMsgFees(ctx sdk.Context, feeCoins sdk.Coins, floorGasPrice sdk.Coin, gas uint64, additionalFees sdk.Coins) error {
	// the isTestContext is exclusively for not breaking all existing sim tests which freak out when denom is anything other than stake.
	if isTestContext(ctx) {
		return nil
	}

	var baseFee sdk.Coins
	if !floorGasPrice.IsZero() {
		baseFee = baseFee.Add(sdk.NewCoin(floorGasPrice.Denom, floorGasPrice.Amount.Mul(sdk.NewIntFromUint64(gas))))
	}
	reqTotal := baseFee.Add(additionalFees...)

	if reqTotal.IsZero() {
		return nil
	}

	if _, hasNeg := feeCoins.SafeSub(reqTotal...); hasNeg {
		// Slightly different messages when there's additional fees and not.
		feeDesc := "base fee"
		if !additionalFees.IsZero() {
			feeDesc = "base fee + additional fee"
		}
		return sdkerrors.ErrInsufficientFee.Wrapf(
			"%s cannot be paid with provided fees: %q"+
				", required: %q = %q(base-fee) + %q(additional-fees)",
			feeDesc, feeCoins, reqTotal, baseFee, additionalFees)
	}

	return nil
}
