package antewrapper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	msgbasedfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

const (
	DefaultInsufficientFeeMsg = "not enough fees; after deducting fees required,got"
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

// AnteHandle handles msg fee checking
// has two functions ensures,
// 1. has enough fees to add to Mempool (this involves CheckTx)
// 2. Makes sure enough fees are present for additional message fees
// Let z be the Total Fees to be paid
// Let x be the Base gas Fees to be paid
// Let y is the additional fees to be paid per MsgType
// then z = x + y
// This Fee Decorator makes sure that z is >= to x + y
func (afd MsgBasedFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, err := getFeeTx(tx)

	if err != nil {
		return ctx, err
	}

	ctx.Logger().Debug(fmt.Sprintf("In MsgBasedFeeDecorator %d", ctx.GasMeter().GasConsumed()))

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Compute msg additionalFees
	additionalFees, err := CalculateAdditionalFeesToBePaid(ctx, afd.msgBasedFeeKeeper, msgs...)
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, err.Error())
	}
	if !additionalFees.IsZero() {
		// ensure enough fees to cover mempool fee for base fee + additional fee
		// This is exact same logic as NewMempoolFeeDecorator except it accounts for additional Fees.
		if ctx.IsCheckTx() && !simulate {
			errFromMempoolCalc := EnsureSufficientMempoolFees(ctx, gas, feeCoins, additionalFees)
			if errFromMempoolCalc != nil {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, errFromMempoolCalc.Error())
			}
		}
		feePayer := feeTx.FeePayer()
		feeGranter := feeTx.FeeGranter()

		deductFeesFrom := feePayer

		deductFeesFrom, errorFromFeeGrant := getFeeGranterIfExists(ctx, feeGranter, afd, feePayer, deductFeesFrom)
		if errorFromFeeGrant != nil {
			return ctx, errorFromFeeGrant
		}

		deductFeesFromAcc := afd.accountKeeper.GetAccount(ctx, deductFeesFrom)
		if deductFeesFromAcc == nil {
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
			if err != nil {
				return ctx, err
			}
		}

		// get all the coin balances for the fee payer account
		balancePerCoin := make(sdk.Coins, len(feeCoins))
		for i, fc := range feeCoins {
			balancePerCoin[i] = afd.bankKeeper.GetBalance(ctx, deductFeesFrom, fc.Denom)
		}

		if !simulate {
			if err = EnsureAccountHasSufficientFeesWithAcctBalanceCheck(gas, feeCoins, additionalFees, balancePerCoin,
				afd.msgBasedFeeKeeper.GetMinGasPrice(ctx), afd.msgBasedFeeKeeper.GetDefaultFeeDenom()); err != nil {
				return ctx, err
			}
		}
	}
	return next(ctx, tx, simulate)
}

func getFeeGranterIfExists(ctx sdk.Context, feeGranter sdk.AccAddress, afd MsgBasedFeeDecorator, feePayer sdk.AccAddress, deductFeesFrom sdk.AccAddress) (sdk.AccAddress, error) {
	// if feegranter set deduct fee from feegranter account.
	// if fee granter check if grant exists.
	if feeGranter != nil {
		if afd.feegrantKeeper == nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			grant, err := afd.feegrantKeeper.GetAllowance(ctx, feeGranter, feePayer)

			if err != nil {
				return nil, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
			if grant == nil {
				return nil, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}
	return deductFeesFrom, nil
}

func getFeeTx(tx sdk.Tx) (sdk.FeeTx, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	return feeTx, nil
}

// EnsureSufficientMempoolFees verifies that the given transaction has supplied
// enough fees(gas + additional fees) to cover a proposer's minimum fees. A result object is returned
// indicating success or failure.
//
// Contract: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientMempoolFees(ctx sdk.Context, gas uint64, feeCoins sdk.Coins, additionalFees sdk.Coins) error {
	feeCoinsOriginal := feeCoins
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
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, DefaultInsufficientFeeMsg+": %q, required fees: %q = %q(base-fee) +%q(additional-fees)", feeCoins, requiredFees.Add(additionalFees...), requiredFees, additionalFees)
	}

	if !requiredFees.IsZero() && !feeCoins.IsAnyGTE(requiredFees) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "Base Fee+additional fee cannot be paid with fee value passed in "+": %q, required: %q = %q(base-fee) +%q(additional-fees)", feeCoinsOriginal, requiredFees.Add(additionalFees...),
			requiredFees, additionalFees)
	}

	return nil
}

func EnsureAccountHasSufficientFeesWithAcctBalanceCheck(gas uint64, feeCoins sdk.Coins, additionalFees sdk.Coins,
	balancePerCoin sdk.Coins, minGasPriceForAdditionalFeeCalc uint32, defaultDenom string) error {
	err := EnsureSufficientFees(gas, feeCoins, additionalFees, minGasPriceForAdditionalFeeCalc, defaultDenom)
	if err != nil {
		return err
	}
	// Step 3: Check if account has enough to pay all fees.
	if balancePerCoin.IsZero() || !balancePerCoin.IsAnyGTE(feeCoins) {
		return fmt.Errorf("fee payer account does not have enough balance to pay for %q", feeCoins)
	}
	return nil
}

// EnsureSufficientFees to be used by msg_service_router
func EnsureSufficientFees(gas uint64, feeCoins sdk.Coins, additionalFees sdk.Coins,
	minGasPriceForAdditionalFeeCalc uint32, defaultDenom string) error {
	// Step 1. Check if fees has enough money to pay additional fees.
	var hasNeg bool
	if feeCoins, hasNeg = feeCoins.SafeSub(additionalFees); hasNeg {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, DefaultInsufficientFeeMsg+": %q, required additional fee: %q", feeCoins, additionalFees)
	}
	// Step 2: check if additional fees in nhash, that base fees and additional fees can be paid
	// total fees in hash - gas limit * price per gas >= additional fees in hash
	if containsDenom(additionalFees, defaultDenom) {
		minGasprice := sdk.NewCoin(defaultDenom, sdk.NewIntFromUint64(uint64(minGasPriceForAdditionalFeeCalc)))
		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		fee := minGasprice.Amount.Mul(sdk.NewIntFromUint64(gas))
		baseFees := sdk.NewCoin(minGasprice.Denom, fee)
		if feeCoins, hasNeg = feeCoins.SafeSub(sdk.Coins{baseFees}); hasNeg {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, DefaultInsufficientFeeMsg+": %q, required additional fee: %q", feeCoins, additionalFees)
		}
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

func containsDenom(coins sdk.Coins, denom string) bool {
	if len(coins) == 0 {
		return false
	}
	amt := coins.AmountOf(denom)
	return !amt.IsZero()
}
