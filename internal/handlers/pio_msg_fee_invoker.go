package handlers

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

		if chargedFees != nil && chargedFees.IsValid() && !chargedFees.IsZero() {
			// eat up the gas cost for charging fees. (This one is on us, Cheers!, mainly because we don't want to fail at this step, imo, but we can remove this is f necessary)
			ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

			ctx.Logger().Debug(fmt.Sprintf("The Fee consumed by message types : %v", feeGasMeter.FeeConsumedByMsg()))

			err = afd.msgBasedFeeKeeper.DeductFees(afd.bankKeeper, ctx, deductFeesFromAcc, chargedFees)

			if err != nil {
				return nil, nil, err
			}
			events = sdk.Events{
				sdk.NewEvent(sdk.EventTypeTx,
					sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, chargedFees.String()),
				)}
		}
		events, err = SweepUpFees(ctx, feeGasMeter, chargedFees, afd, feeTx, err, deductFeesFromAcc, events)
		if err != nil {
			return nil, nil, err
		}
		// set back the original gasMeter
		ctx = ctx.WithGasMeter(originalGasMeter)
	}

	return chargedFees, events, nil
}

func SweepUpFees(ctx sdk.Context, feeGasMeter *antewrapper.FeeGasMeter, chargedFees sdk.Coins, afd MsgBasedFeeInvoker, feeTx sdk.FeeTx, err error, deductFeesFromAcc authtypes.AccountI, events sdk.Events) (sdk.Events, error) {
	consumedBaseFee := sdk.Coins{feeGasMeter.BaseFeeConsumed()}
	chargedFeesInDefaultDenom := getDenom(chargedFees, afd.msgBasedFeeKeeper.GetDefaultFeeDenom())
	totalBaseFee := consumedBaseFee.Add(chargedFeesInDefaultDenom)

	sweptUpFee, isNeg := feeTx.GetFee().SafeSub(totalBaseFee)
	if !isNeg {
		err = afd.msgBasedFeeKeeper.DeductFees(afd.bankKeeper, ctx, deductFeesFromAcc, sweptUpFee)
		if err != nil {
			return nil, err
		}
		totalBaseFee = totalBaseFee.Add(sweptUpFee[0])
	}
	events = events.AppendEvent(
		sdk.NewEvent(antewrapper.AttributeKeyBaseFee,
			sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, getDenom(feeTx.GetFee(), afd.msgBasedFeeKeeper.GetDefaultFeeDenom()).String()),
		))

	//temp event, delete after testing
	events = events.AppendEvent(
		sdk.NewEvent("tempbasefee",
			sdk.NewAttribute(antewrapper.AttributeKeyAdditionalFee, totalBaseFee.String()),
		))
	return events, err
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
