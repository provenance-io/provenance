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

func (afd MsgBasedFeeInvoker) Invoke(ctx sdk.Context, simulate bool) (coins sdk.Coins, err error) {
	chargedFees := sdk.Coins{}

	if ctx.TxBytes() != nil && len(ctx.TxBytes()) != 0 {
		ctx.Logger().Debug("In chargeFees()")
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
			panic("Provenenance only supports feeTx for now")
		}
		feePayer := feeTx.FeePayer()
		feeGranter := feeTx.FeeGranter()
		deductFeesFrom := feePayer
		// if fee granter set deduct fee from feegranter account.
		// this works with only when feegrant enabled.

		deductFeesFromAcc := afd.accountKeeper.GetAccount(ctx, deductFeesFrom)
		if deductFeesFromAcc == nil {
			panic("fee payer address: %s does not exist")
		}

		feeGasMeter, ok := ctx.GasMeter().(*antewrapper.FeeGasMeter)
		if !ok {
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
					return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
				} else if !feeGranter.Equals(feePayer) {
					err := afd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, chargedFees, tx.GetMsgs())
					if err != nil {
						return nil, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
					}
				}
				deductFeesFrom = feeGranter
			}
		}
		ctx.Logger().Debug(fmt.Sprintf("The Fee consumed by message type : %v", feeGasMeter.FeeConsumedByMsg()))
		err = afd.msgBasedFeeKeeper.DeductFees(afd.bankKeeper, ctx, deductFeesFromAcc, chargedFees)
		if err != nil {
			return nil, err
		}
		// set back the original gasMeter
		ctx = ctx.WithGasMeter(originalGasMeter)
	}

	return chargedFees, nil
}
