package antewrapper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

const (
	DefaultInsufficientFeeMsg = "not enough fees; after deducting fees required,got"
	SimAppChainID             = "simapp-unit-testing"
)

// MsgFeesDecorator will check if the transaction's fee is at least as large
// as tax + additional minimum gasFee (defined in msgfeeskeeper)
// and record additional fee proceeds to msgfees module to track additional fee proceeds.
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MsgFeesDecorator
type MsgFeesDecorator struct {
	msgFeeKeeper   msgfeestypes.MsgFeesKeeper
	bankKeeper     banktypes.Keeper
	accountKeeper  cosmosante.AccountKeeper
	feegrantKeeper msgfeestypes.FeegrantKeeper
}

func NewMsgFeesDecorator(bankKeeper banktypes.Keeper, accountKeeper cosmosante.AccountKeeper, feegrantKeeper msgfeestypes.FeegrantKeeper, keeper msgfeestypes.MsgFeesKeeper) MsgFeesDecorator {
	return MsgFeesDecorator{
		keeper,
		bankKeeper,
		accountKeeper,
		feegrantKeeper,
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
func (afd MsgFeesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, err := getFeeTx(tx)

	if err != nil {
		return ctx, err
	}

	ctx.Logger().Debug(fmt.Sprintf("In MsgFeesDecorator %d", ctx.GasMeter().GasConsumed()))

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Compute msg all additional fees
	msgFeesDistribution, err := CalculateAdditionalFeesToBePaid(ctx, afd.msgFeeKeeper, msgs...)
	totalAdditionalFees := msgFeesDistribution.TotalAdditionalFees
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, err.Error())
	}
	// floor gas price should be checked for all Tx's ( i.e nodes cannot set min-gas-price < floor gas price)
	// the chain id check is exclusively for not breaking all existing sim tests which freak out when denom is anything other than stake.
	if ctx.IsCheckTx() && !simulate && !isTestContext(ctx) && (totalAdditionalFees.IsZero() || totalAdditionalFees == nil) {
		err = checkFloorGasFees(gas, feeCoins, totalAdditionalFees, afd.msgFeeKeeper.GetFloorGasPrice(ctx))
		if err != nil {
			return ctx, err
		}
	}
	if !totalAdditionalFees.IsZero() {
		// ensure enough fees to cover mempool fee for base fee + additional fee
		// This is exact same logic as NewMempoolFeeDecorator except it accounts for additional Fees.
		if ctx.IsCheckTx() && !simulate {
			errFromMempoolCalc := EnsureSufficientMempoolFees(ctx, gas, feeCoins, totalAdditionalFees)
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
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %q does not exist", deductFeesFrom)
			if err != nil {
				return ctx, err
			}
		}

		// get all the coin balances for the fee payer account
		balancePerCoin := sdk.NewCoins()
		for _, fc := range feeCoins {
			balancePerCoin = balancePerCoin.Add(afd.bankKeeper.GetBalance(ctx, deductFeesFrom, fc.Denom))
		}

		if !simulate {
			if err = EnsureAccountHasSufficientFeesWithAcctBalanceCheck(gas, feeCoins, totalAdditionalFees, balancePerCoin,
				afd.msgFeeKeeper.GetFloorGasPrice(ctx)); err != nil {
				return ctx, err
			}
		}
	}
	return next(ctx, tx, simulate)
}

//   This check for chain-id is exclusively for not breaking all existing sim tests which freak out when denom is anything other than stake.
//   and some network tests won't work without a chain id being set(but they also setup everything with stake denom) so `simapp-unit-testing` chain id is skipped also.
//   This only needs to work to pio-testnet and pio-mainnet, so this is safe.
func isTestContext(ctx sdk.Context) bool {
	return !(len(ctx.ChainID()) != 0 && ctx.ChainID() != SimAppChainID && ctx.ChainID() != helpers.SimAppChainID)
}

// getFeeGranterIfExists checks if fee granter exists and returns account to deduct fees from
func getFeeGranterIfExists(ctx sdk.Context, feeGranter sdk.AccAddress, afd MsgFeesDecorator, feePayer sdk.AccAddress, deductFeesFrom sdk.AccAddress) (sdk.AccAddress, error) {
	if feeGranter != nil {
		if afd.feegrantKeeper == nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			grant, err := afd.feegrantKeeper.GetAllowance(ctx, feeGranter, feePayer)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "%q not allowed to pay fees from %q", feeGranter, feePayer)
			}
			if grant == nil {
				return nil, sdkerrors.Wrapf(err, "%q not allowed to pay fees from %q", feeGranter, feePayer)
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
	balancePerCoin sdk.Coins, minGasPriceForAdditionalFeeCalc sdk.Coin) error {
	err := EnsureSufficientFees(gas, feeCoins, additionalFees, minGasPriceForAdditionalFeeCalc)
	if err != nil {
		return err
	}
	_, hasNeg := balancePerCoin.SafeSub(feeCoins)
	if balancePerCoin.IsZero() || hasNeg {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fee payer account does not have enough balance to pay for %q", feeCoins)
	}
	return nil
}

// EnsureSufficientFees to be used by msg_service_router
func EnsureSufficientFees(gas uint64, feeCoins sdk.Coins, additionalFees sdk.Coins,
	minGasPriceForAdditionalFeeCalc sdk.Coin) error {
	// Step 1. Check if fees has enough money to pay additional fees.
	var hasNeg bool
	if feeCoins, hasNeg = feeCoins.SafeSub(additionalFees); hasNeg {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, DefaultInsufficientFeeMsg+": %q, required additional fee: %q", feeCoins, additionalFees)
	}
	// Step 2: check if additional fees in nhash, that base fees and additional fees can be paid
	// total fees in hash - gas limit * price per gas >= additional fees in hash
	if !additionalFees.AmountOf(minGasPriceForAdditionalFeeCalc.Denom).IsZero() {
		// Determine the required fees by multiplying each required minimum gas
		err := checkFloorGasFees(gas, feeCoins, additionalFees, minGasPriceForAdditionalFeeCalc)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkFloorGasFees(gas uint64, feeCoins sdk.Coins, additionalFees sdk.Coins, minGasPriceForAdditionalFeeCalc sdk.Coin) error {
	// where fee = ceil(floorGasPrice * gasLimit).
	fee := minGasPriceForAdditionalFeeCalc.Amount.Mul(sdk.NewIntFromUint64(gas))
	baseFees := sdk.NewCoin(minGasPriceForAdditionalFeeCalc.Denom, fee)
	if _, hasNeg := feeCoins.SafeSub(sdk.Coins{baseFees}); hasNeg {
		// for tx without additional fees.
		if additionalFees == nil || additionalFees.IsZero() {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "not enough fees based on floor gas price: %q; required base fees >=%q: Supplied fee was %q", minGasPriceForAdditionalFeeCalc, baseFees, feeCoins)
		}
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "not enough fees based on floor gas price: %q; after deducting (total fee supplied fees - additional fees(%q)) required base fees >=%q: Supplied fee was %q", minGasPriceForAdditionalFeeCalc, additionalFees, baseFees, feeCoins)
	}
	return nil
}

// CalculateAdditionalFeesToBePaid computes the stability tax on MsgSend and MsgMultiSend.
func CalculateAdditionalFeesToBePaid(ctx sdk.Context, mbfk msgfeestypes.MsgFeesKeeper, msgs ...sdk.Msg) (*MsgFeesDistribution, error) {
	// get the msg fee
	msgFeesDistribution := MsgFeesDistribution{
		RecipientDistributions: make(map[string]sdk.Coin),
	}
	assessCustomMsgTypeURL := sdk.MsgTypeURL(&msgfeestypes.MsgAssessCustomMsgFeeRequest{})
	for _, msg := range msgs {
		typeURL := sdk.MsgTypeURL(msg)
		msgFees, err := mbfk.GetMsgFee(ctx, typeURL)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}

		if msgFees != nil {
			if msgFees.AdditionalFee.IsPositive() {
				msgFeesDistribution.AdditionalModuleFees = msgFeesDistribution.AdditionalModuleFees.Add(msgFees.AdditionalFee)
				msgFeesDistribution.TotalAdditionalFees = msgFeesDistribution.TotalAdditionalFees.Add(msgFees.AdditionalFee)
			}
		}
		if typeURL == assessCustomMsgTypeURL {
			assessFee, ok := msg.(*msgfeestypes.MsgAssessCustomMsgFeeRequest)
			if !ok {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "unable to convert msg to MsgAssessCustomMsgFeeRequest")
			}
			msgFeeCoin, err := mbfk.ConvertDenomToHash(ctx, assessFee.Amount)
			if err != nil {
				return nil, err
			}
			if msgFeeCoin.IsPositive() {
				if len(assessFee.Recipient) != 0 {
					recipientCoin, feePayoutCoin := msgfeestypes.SplitAmount(msgFeeCoin)
					if len(msgFeesDistribution.RecipientDistributions[assessFee.Recipient].Denom) == 0 {
						msgFeesDistribution.RecipientDistributions[assessFee.Recipient] = recipientCoin
					} else {
						msgFeesDistribution.RecipientDistributions[assessFee.Recipient] = msgFeesDistribution.RecipientDistributions[assessFee.Recipient].Add(recipientCoin)
					}
					msgFeesDistribution.AdditionalModuleFees = msgFeesDistribution.AdditionalModuleFees.Add(feePayoutCoin)
					msgFeesDistribution.TotalAdditionalFees = msgFeesDistribution.TotalAdditionalFees.Add(msgFeeCoin)
				} else {
					msgFeesDistribution.AdditionalModuleFees = msgFeesDistribution.AdditionalModuleFees.Add(msgFeeCoin)
					msgFeesDistribution.TotalAdditionalFees = msgFeesDistribution.TotalAdditionalFees.Add(msgFeeCoin)
				}
			}
		}
	}

	return &msgFeesDistribution, nil
}
