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

func (afd MsgFeeInvoker) Invoke(ctx sdk.Context, simulate bool) (sdk.Coins, sdk.Events, error) {
	chargedFees := sdk.Coins{}
	eventsToReturn := sdk.Events{}

	if len(ctx.TxBytes()) != 0 {
		tx, err := afd.txDecoder(ctx.TxBytes())
		if err != nil {
			panic(fmt.Errorf("error in MsgFeeInvoker.Invoke() while getting txBytes: %w", err))
		}

		feeTx, err := antewrapper.GetFeeTx(tx)
		if err != nil {
			// For provenance, should be a FeeTx since antehandler should enforce it,
			// but not adding complexity here.
			panic(err)
		}

		feeGasMeter, err := antewrapper.GetFeeGasMeter(ctx)
		if err != nil {
			// For provenance, should be a FeeGasMeter since antehandler should enforce it,
			// but not adding complexity here.
			panic(err)
		}

		// eat up the gas cost for charging fees. (This one is on us, Cheers!, mainly because we don't want to fail at this step, imo, but we can remove this is f necessary)
		ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

		consumedFees := feeGasMeter.FeeConsumed()
		if consumedFees.IsAnyNegative() {
			return nil, nil, sdkerrors.ErrInvalidCoins.Wrapf("consumed fees %v are negative, which should not be possible, aborting", chargedFees)
		}

		// this sweeps all extra fees too, 1. keeps current behavior 2. accounts for priority mempool
		unchargedFees, _ := feeTx.GetFee().SafeSub(feeGasMeter.BaseFeeConsumed()...)

		deductFeesFrom, err := antewrapper.GetFeePayerUsingFeeGrant(ctx, afd.feegrantKeeper, feeTx, unchargedFees, tx.GetMsgs())
		if err != nil {
			return nil, nil, err
		}

		deductFeesFromAcc := afd.accountKeeper.GetAccount(ctx, deductFeesFrom)
		if deductFeesFromAcc == nil {
			return nil, nil, sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %q does not exist", deductFeesFrom)
		}

		// If there's fees left to collect, or there were consumed fees, deduct/distribute them now.
		if !unchargedFees.IsZero() || !consumedFees.IsZero() {
			err = afd.msgFeeKeeper.DeductFeesDistributions(afd.bankKeeper, ctx, deductFeesFromAcc, unchargedFees, feeGasMeter.FeeConsumedDistributions())
			if err != nil {
				return nil, nil, err
			}
		}
		// the uncharged fees have now been charged.
		chargedFees = chargedFees.Add(unchargedFees...)

		// If there were msg based fees, add some events for them.
		if !consumedFees.IsZero() {
			// Add event for the total fees added by msg based fees.
			eventsToReturn = append(eventsToReturn, sdk.NewEvent(sdk.EventTypeTx,
				sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, consumedFees.String()),
				sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String())))

			// Add event with a breakdown of those fees.
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
