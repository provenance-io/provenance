package antewrapper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// ProvenanceDeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use ProvenanceDeductFeeDecorator
type ProvenanceDeductFeeDecorator struct {
	ak             authante.AccountKeeper
	bankKeeper     types.BankKeeper
	feegrantKeeper authante.FeegrantKeeper
	msgFeeKeeper   msgfeestypes.MsgFeesKeeper
}

var (
	AttributeKeyBaseFee       = "basefee"
	AttributeKeyAdditionalFee = "additionalfee"
)

func NewProvenanceDeductFeeDecorator(ak authante.AccountKeeper, bk types.BankKeeper, fk msgfeestypes.FeegrantKeeper, mbfk msgfeestypes.MsgFeesKeeper) ProvenanceDeductFeeDecorator {
	return ProvenanceDeductFeeDecorator{
		ak:             ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		msgFeeKeeper:   mbfk,
	}
}

func (dfd ProvenanceDeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.FeeCollectorName))
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	feeGasMeter, ok := ctx.GasMeter().(*FeeGasMeter)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "GasMeter not a FeeGasMeter")
	}

	deductFeesFrom := feePayer

	// deduct the fees
	// Compute msg additionalFees
	msgs := feeTx.GetMsgs()
	additionalFees, errFromCalculateAdditionalFeesToBePaid := CalculateAdditionalFeesToBePaid(ctx, dfd.msgFeeKeeper, msgs...)
	if errFromCalculateAdditionalFeesToBePaid != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errFromCalculateAdditionalFeesToBePaid.Error())
	}
	feeToDeduct := feeTx.GetFee()
	if additionalFees != nil && len(additionalFees) > 0 {
		var hasNeg bool
		feeToDeduct, hasNeg = feeToDeduct.SafeSub(additionalFees)
		if hasNeg && !simulate {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %q", feeToDeduct)
		}
	}

	if feeGranter != nil && !simulate {
		if dfd.feegrantKeeper == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err = dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, feeToDeduct, tx.GetMsgs())

			if err != nil {
				return ctx, sdkerrors.Wrapf(err, "%q not allowed to pay fees from %q", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %q does not exist", deductFeesFrom)
	}
	// deduct total fees (base fee + additional fee)
	if !feeToDeduct.IsZero() && !simulate {
		err = DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, feeToDeduct)
		if err != nil {
			return ctx, err
		}
		feeGasMeter.ConsumeBaseFee(feeToDeduct)
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc types.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %q", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
