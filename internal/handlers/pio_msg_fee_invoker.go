package handlers

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/provenance-io/provenance/internal/antewrapper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	msgbasedfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type MsgBasedFeeInvoker struct {
	msgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper
	bankKeeper        msgbasedfeetypes.BankKeeper
	accountKeeper     msgbasedfeetypes.AccountKeeper
	feegrantKeeper    msgbasedfeetypes.FeegrantKeeper
	txDecoder         sdk.TxDecoder
}

// NewMsgBasedFeeInvoker concrete impl of how to charge Msg Based Fees
func NewMsgBasedFeeInvoker(bankKeeper msgbasedfeetypes.BankKeeper, accountKeeper msgbasedfeetypes.AccountKeeper,
	feegrantKeeper msgbasedfeetypes.FeegrantKeeper, msgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper, decoder sdk.TxDecoder) MsgBasedFeeInvoker {
	return MsgBasedFeeInvoker{
		msgBasedFeeKeeper,
		bankKeeper,
		accountKeeper,
		feegrantKeeper,
		decoder,
	}
}

func (afd MsgBasedFeeInvoker) Invoke(ctx sdk.Context, simulate bool) (coins sdk.Coins, evets sdk.Events, err error) {
	chargedFees := sdk.Coins{}
	events := sdk.Events{}

	if ctx.TxBytes() != nil && len(ctx.TxBytes()) != 0 {
		originalGasMeter := ctx.GasMeter()

		tx, err := afd.txDecoder(ctx.TxBytes())
		if err != nil {
			panic(fmt.Errorf("error in chargeFees() while getting txBytes: %w", err))
		}

		// cast to FeeTx
		feeTx, ok := tx.(sdk.FeeTx)
		// only charge additional fee if of type FeeTx since it should give fee payer.
		// for provenance should be a FeeTx since antehandler should enforce it, but
		// not adding complexity here
		if !ok {
			panic("Provenance only supports feeTx for now")
		}
		feePayer := feeTx.FeePayer()
		feeGranter := feeTx.FeeGranter()
		deductFeesFrom := feePayer
		// if fee granter set deduct fee from feegranter account.
		// this works with only when feegrant enabled.

		feeGasMeter, ok := ctx.GasMeter().(*antewrapper.FeeGasMeter)
		if !ok {
			// all provenance tx's should have this set
			panic("GasMeter is not of type FeeGasMeter")
		}
		chargedFees = feeGasMeter.FeeConsumed()

		if chargedFees != nil && chargedFees.IsValid() && !chargedFees.IsZero() {
			// eat up the gas cost for charging fees. (This one is on us, Cheers!, mainly because we don't want to fail at this step, imo, but we can remove this is f necessary)
			ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
			// if feegranter set deduct fee from feegranter account.
			// this works with only when feegrant enabled.
			if feeGranter != nil {
				if afd.feegrantKeeper == nil {
					return nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
				} else if !feeGranter.Equals(feePayer) {
					err = afd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, chargedFees, tx.GetMsgs())
					if err != nil {
						return nil, nil, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
					}
				}
				deductFeesFrom = feeGranter
			}
			deductFeesFromAcc := afd.accountKeeper.GetAccount(ctx, deductFeesFrom)
			if deductFeesFromAcc == nil {
				return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
			}

			ctx.Logger().Debug(fmt.Sprintf("The Fee consumed by message types : %v", feeGasMeter.FeeConsumedByMsg()))

			var baseFeePaidJustForGasTx sdk.Coins
			// if base denom and additional fee in default base denom (for now nhash)
			if !getDenom(feeGasMeter.BaseFeeConsumed(), afd.msgBasedFeeKeeper.GetDefaultFeeDenom()).IsNil() && !getDenom(chargedFees, afd.msgBasedFeeKeeper.GetDefaultFeeDenom()).IsNil() {
				// for non authz/wasmd calls this will be zero but for wasmd/authz this will have a value.
				baseFeePaidJustForGasTx, err = baseFeePaidBasedOnGas(feeGasMeter, afd.msgBasedFeeKeeper.GetDefaultFeeDenom(), afd.msgBasedFeeKeeper.GetMinGasPrice(ctx), feeTx.GetGas())
				if err != nil {
					return nil, nil, err
				}
			} else {
				baseFeePaidJustForGasTx = feeGasMeter.BaseFeeConsumed()
			}

			var isNeg bool
			// this sweeps all extra fees too, 1. keeps current behavior
			chargedFees, isNeg = feeTx.GetFee().SafeSub(baseFeePaidJustForGasTx)
			// for e.g if authz paid 900 hotdog and additional message was 800 hotdog, already paid
			if isNeg {
				chargedFees = removeNegativeCoins(chargedFees)
			}
			if len(chargedFees) > 0 {
				err = afd.msgBasedFeeKeeper.DeductFees(afd.bankKeeper, ctx, deductFeesFromAcc, chargedFees)
				if err != nil {
					return nil, nil, err
				}
			}
			events = sdk.Events{
				sdk.NewEvent(sdk.EventTypeTx,
					sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, feeGasMeter.FeeConsumed().String()),
				),
				sdk.NewEvent(sdk.EventTypeTx,
					sdk.NewAttribute(antewrapper.AttributeKeyBaseFee, feeGasMeter.BaseFeeConsumed().Add(chargedFees...).Sub(feeGasMeter.FeeConsumed()).String()),
				)}
		}

		// set back the original gasMeter
		ctx = ctx.WithGasMeter(originalGasMeter)
	}

	return chargedFees, events, nil
}

func baseFeePaidBasedOnGas(meter *antewrapper.FeeGasMeter, defaultDenom string, minGasPriceForAdditionalFeeCalc uint32, gas uint64) (sdk.Coins, error) {
	minGasprice := sdk.NewCoin(defaultDenom, sdk.NewIntFromUint64(uint64(minGasPriceForAdditionalFeeCalc)))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	fee := minGasprice.Amount.Mul(sdk.NewIntFromUint64(gas))
	baseFeesForGas := sdk.NewCoin(minGasprice.Denom, fee)
	_, isNeg := meter.BaseFeeConsumed().SafeSub(sdk.Coins{baseFeesForGas})
	if isNeg {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "This should never happen if it passed current ante decorators")
	}
	return sdk.Coins{baseFeesForGas}, nil
}

func getDenom(coins sdk.Coins, denom string) sdk.Coin {
	if len(coins) == 0 {
		return sdk.Coin{}
	}
	amt := coins.AmountOf(denom)
	if !amt.IsZero() {
		return sdk.Coin{
			Denom:  denom,
			Amount: amt,
		}
	}
	return sdk.Coin{}
}

// removeNegativeCoins removes all zero coins from the given coin set in-place.
func removeNegativeCoins(coins sdk.Coins) sdk.Coins {
	var result []sdk.Coin
	if len(coins) > 0 {
		result = make([]sdk.Coin, 0, len(coins)-1)
	}

	for _, coin := range coins {
		if !coin.IsNegative() {
			result = append(result, coin)
		}
	}

	return result
}
