package antewrapper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/keeper"
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
	bankKeeper        banktypes.Keeper
	accountKeeper     cosmosante.AccountKeeper
	feegrantKeeper    msgbasedfeetypes.FeegrantKeeper
}

func NewMsgBasedFeeDecorator(bankKeeper banktypes.Keeper, accountKeeper cosmosante.AccountKeeper, feegrantKeeper msgbasedfeetypes.FeegrantKeeper, keeper msgbasedfeetypes.MsgBasedFeeKeeper) MsgBasedFeeDecorator {
	return MsgBasedFeeDecorator{
		keeper,
		bankKeeper,
		accountKeeper,
		feegrantKeeper,
	}
}

// AnteHandle handles msg tax fee checking
func (afd MsgBasedFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// has to be FeeTx type
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	ctx.Logger().Info(fmt.Sprintf("NOTICE: here in MsgBasedFeeDecorator %d", ctx.GasMeter().GasConsumed()))

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Compute msg additionalFees
	additionalFees, err := CalculateAdditionalFeesToBePaid(ctx, afd.msgBasedFeeKeeper, msgs...)
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, err.Error())
	}
	if !additionalFees.IsZero() {
		// mempool fee validation tx
		// this is because we want to make sure if additional additionalFees in hash then there is enough
		if ctx.IsCheckTx() {
			if err := EnsureSufficientMempoolFees(ctx, gas, feeCoins, additionalFees); err != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, err.Error())
			}
		}
		feePayer := feeTx.FeePayer()
		feeGranter := feeTx.FeeGranter()

		deductFeesFrom := feePayer

		// if feegranter set deduct fee from feegranter account.

		// if feegranter check if grant exists.
		if feeGranter != nil {
			if afd.feegrantKeeper == nil {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
			} else if !feeGranter.Equals(feePayer) {
				grant, err := afd.feegrantKeeper.GetAllowance(ctx, feeGranter, feePayer)

				if err != nil {
					return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
				}
				if grant == nil {
					return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
				}
			}

			deductFeesFrom = feeGranter
		}

		// this works with only when feegrant enabled.

		deductFeesFromAcc := afd.accountKeeper.GetAccount(ctx, deductFeesFrom)
		if deductFeesFromAcc == nil {
			sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
		}

		// get all the coin balances for the fee payer account
		balancePerCoin := make(sdk.Coins, len(feeCoins))
		for i, fc := range feeCoins {
			balancePerCoin[i] = afd.bankKeeper.GetBalance(ctx, feePayer, fc.Denom)
		}

		if !simulate {
			if err = EnsureAccountHasSufficientFees(ctx, feeCoins, additionalFees, balancePerCoin); err != nil {
				return ctx, err
			}
		}

		// Ensure paid fee is enough to cover taxes
		if _, hasNeg := feeCoins.SafeSub(additionalFees); hasNeg {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee,
				"insufficient additionalFees; got: %s required: %s", feeCoins, additionalFees)
		}
	}
	return next(ctx, tx, simulate)
}

// EnsureSufficientMempoolFees verifies that the given transaction has supplied
// enough fees(gas + additional fees) to cover a proposer's minimum fees. A result object is returned
// indicating success or failure.
//
// Contract: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientMempoolFees(ctx sdk.Context, gas uint64, feeCoins sdk.Coins, additionalFees sdk.Coins) error {
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
	if feeCoins, hasNeg = feeCoins.SafeSub(additionalFees); hasNeg {
		return fmt.Errorf("insufficient fees; got: %q, required: %q = %q(gas fees) +%q(additional msg fees)", feeCoins, requiredFees.Add(additionalFees...), requiredFees, additionalFees)
	}

	if !requiredFees.IsZero() && !feeCoins.IsAnyGTE(requiredFees) {
		return fmt.Errorf("insufficient fees; got: %q, required: %q = %q(gas fees) +%q(additional msg fees)", feeCoins, requiredFees.Add(additionalFees...),
			requiredFees, additionalFees)
	}

	return nil
}

func EnsureAccountHasSufficientFees(ctx sdk.Context, feeCoins sdk.Coins, additionalFees sdk.Coins,
	balancePerCoin sdk.Coins) error {

	originalFees := feeCoins
	// Step 1. Check if fees has enough money to pay additional fees.
	var hasNeg bool
	if feeCoins, hasNeg = feeCoins.SafeSub(additionalFees); hasNeg {
		return fmt.Errorf("insufficient fees; got: %q, required additional fee: %q", feeCoins, additionalFees)
	}
	// Step 2: Check if account has enough to pay all fees.
	if !balancePerCoin.IsZero() && !balancePerCoin.IsAnyGTE(originalFees) {
		return fmt.Errorf("fee payer account does not have enough balance to pay for %q", feeCoins)
	}
	return nil
}

// CalculateAdditionalFeesToBePaid computes the stability tax on MsgSend and MsgMultiSend.
func CalculateAdditionalFeesToBePaid(ctx sdk.Context, mbfk msgbasedfeetypes.MsgBasedFeeKeeper, msgs ...sdk.Msg) (sdk.Coins, error) {
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
		if msgFees.AdditionalFee.IsPositive() {
			additionalFees = additionalFees.Add(sdk.NewCoin(msgFees.AdditionalFee.Denom, msgFees.AdditionalFee.Amount))
		}

	}

	return additionalFees, nil
}
