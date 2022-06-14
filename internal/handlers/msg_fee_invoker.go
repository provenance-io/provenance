package handlers

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/provenance-io/provenance/internal/antewrapper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type MsgFeeInvoker struct {
	msgFeeKeeper   msgfeestypes.MsgFeesKeeper
	bankKeeper     bankkeeper.Keeper
	accountKeeper  msgfeestypes.AccountKeeper
	feegrantKeeper msgfeestypes.FeegrantKeeper
	txDecoder      sdk.TxDecoder
}

// NewMsgFeeInvoker concrete impl of how to charge Msg Based Fees
func NewMsgFeeInvoker(bankKeeper bankkeeper.Keeper, accountKeeper msgfeestypes.AccountKeeper,
	feegrantKeeper msgfeestypes.FeegrantKeeper, msgFeeKeeper msgfeestypes.MsgFeesKeeper, decoder sdk.TxDecoder) MsgFeeInvoker {
	return MsgFeeInvoker{
		msgFeeKeeper,
		bankKeeper,
		accountKeeper,
		feegrantKeeper,
		decoder,
	}
}

func (afd MsgFeeInvoker) Invoke(ctx sdk.Context, simulate bool) (coins sdk.Coins, events sdk.Events, err error) {
	chargedFees := sdk.Coins{}
	eventsToReturn := sdk.Events{}

	if ctx.TxBytes() != nil && len(ctx.TxBytes()) != 0 {
		tx, err := afd.txDecoder(ctx.TxBytes())
		if err != nil {
			panic(fmt.Errorf("error in chargeFees() while getting txBytes: %w", err))
		}

		feeTx, ok := tx.(sdk.FeeTx)
		// only charge additional fee if of type FeeTx since it should give fee payer.
		// for provenance should be a FeeTx since antehandler should enforce it, but
		// not adding complexity here
		if !ok {
			panic("Transaction not of type FeeTx.  Provenance only supports feeTx for now.")
		}

		feeGasMeter, ok := ctx.GasMeter().(*antewrapper.FeeGasMeter)
		if !ok {
			// all provenance tx's should have this set
			panic("GasMeter is not of type FeeGasMeter")
		}

		chargedFees = feeGasMeter.FeeConsumed()
		// check chargedFees is not nil && is not all zero(IsZero returns true if there are no coins or all coins are zero.)
		if chargedFees != nil && !chargedFees.IsZero() {
			// there should not be any negative coins, just to be very sure here
			if chargedFees.IsAnyNegative() {
				return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "charged fees %v are negative, which should not be possible, aborting", chargedFees)
			}
			// eat up the gas cost for charging fees. (This one is on us, Cheers!, mainly because we don't want to fail at this step, imo, but we can remove this is f necessary)
			ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

			feePayer := feeTx.FeePayer()
			feeGranter := feeTx.FeeGranter()
			deductFeesFrom := feePayer
			// if feegranter set deduct fee from feegranter account.
			// this works with only when feegrant enabled.
			if feeGranter != nil {
				if afd.feegrantKeeper == nil {
					return nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
				} else if !feeGranter.Equals(feePayer) {
					err = afd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, chargedFees, tx.GetMsgs())
					if err != nil {
						return nil, nil, sdkerrors.Wrapf(err, "%q not allowed to pay fees from %q", feeGranter, feePayer)
					}
				}
				deductFeesFrom = feeGranter
			}
			deductFeesFromAcc := afd.accountKeeper.GetAccount(ctx, deductFeesFrom)
			if deductFeesFromAcc == nil {
				return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %q does not exist", deductFeesFrom)
			}

			ctx.Logger().Debug(fmt.Sprintf("The Fee consumed by message types : %v", feeGasMeter.FeeConsumedByMsg()))

			baseFeeConsumedAtAnteHandler := feeGasMeter.BaseFeeConsumed()

			var isNeg bool
			// this sweeps all extra fees too, 1. keeps current behavior 2. accounts for priority mempool
			chargedFees, isNeg = feeTx.GetFee().SafeSub(baseFeeConsumedAtAnteHandler)

			if isNeg {
				return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "This should never happen, because fee charged in ante handlers should never have been more than fee promised without error'ing out.")
			}

			if len(chargedFees) > 0 && chargedFees.IsAllPositive() {
				// deduct fees from chargedFees and distribute
				err = afd.msgFeeKeeper.DeductFeesDistributions(afd.bankKeeper, ctx, deductFeesFromAcc, chargedFees, feeGasMeter.FeeConsumedDistributions())
				if err != nil {
					return nil, nil, err
				}
			}

			eventsToReturn = sdk.Events{
				sdk.NewEvent(sdk.EventTypeTx,
					sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, feeGasMeter.FeeConsumed().String()),
				),
				sdk.NewEvent(sdk.EventTypeTx,
					sdk.NewAttribute(antewrapper.AttributeKeyBaseFee, feeGasMeter.BaseFeeConsumed().Add(chargedFees...).Sub(feeGasMeter.FeeConsumed()).String()),
				)}
			msgFeesSummaryEvent, err := sdk.TypedEventToEvent(feeGasMeter.EventFeeSummary())
			if err != nil {
				return nil, nil, err
			}
			if len(msgFeesSummaryEvent.Attributes) > 0 {
				eventsToReturn = append(eventsToReturn, msgFeesSummaryEvent)
			}
		}
	}

	return chargedFees, eventsToReturn, nil
}
