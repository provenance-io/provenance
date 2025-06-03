package antewrapper

import (
	"fmt"

	cerrs "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	internalsdk "github.com/provenance-io/provenance/internal/sdk"
)

// FlatFeePostHandler is an sdk.PostDecorator that collects the fee remainder once all the msgs have run.
//
// The up-front cost is collected by the DeductUpFrontCostDecorator.
type FlatFeePostHandler struct {
	bk BankKeeper
	fk ante.FeegrantKeeper
}

var _ sdk.PostDecorator = (*FlatFeePostHandler)(nil)

func NewFlatFeePostHandler(bk BankKeeper, fk ante.FeegrantKeeper) FlatFeePostHandler {
	return FlatFeePostHandler{bk: bk, fk: fk}
}

// PostHandle collects the rest of the fees for a tx.
func (h FlatFeePostHandler) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate, success bool, next sdk.PostHandler) (sdk.Context, error) {
	// If it wasn't successful, there's nothing to do in here.
	if !success {
		ctx.Logger().Debug("Skipping FlatFeePostHandler because tx was not successful.")
		return next(ctx, tx, simulate, success)
	}

	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return ctx, err
	}

	gasMeter, err := GetFlatFeeGasMeter(ctx)
	if err != nil {
		return ctx, err
	}
	err = gasMeter.Finalize(ctx)
	if err != nil {
		return ctx, fmt.Errorf("could not finalize gas meter: %w", err)
	}

	reqFee := gasMeter.GetRequiredFee()
	feeProvided := feeTx.GetFee()

	extraMsgsCost := gasMeter.GetExtraMsgsCost()
	addedFees := gasMeter.GetAddedFees()
	newCharges := addedFees.Add(extraMsgsCost...)
	if !newCharges.IsZero() && !simulate && !IsInitGenesis(ctx) {
		// There were extra msg costs added since we set up the gas meter. Re-check that
		// the full fee is enough to cover everything (now that we know what everything is).
		err = validateFeeAmount(reqFee, feeProvided)
		if err != nil {
			return ctx, err
		}
	}

	// We collect the entirety of the fee provided, even if it's more than what's required.
	// We've already collected the up-front cost, though, so we take that out of the full fee and collect what's left.
	upFrontCost := gasMeter.GetUpFrontCost()
	var uncharged sdk.Coins
	if !simulate && !IsInitGenesis(ctx) {
		// If not simulating, we want to collect all of the provided fee (we know it's at least what's required).
		uncharged = feeProvided.Sub(upFrontCost...)
		ctx.Logger().Debug("On-success cost calculated using fee provided (and up-front cost).",
			"fee provided", lazyCzStr(feeProvided), "up-front cost", lazyCzStr(upFrontCost), "on-success", lazyCzStr(uncharged))
	} else {
		// If simulating, pretend the reqFee is what was provided since there might not have been a fee provided.
		uncharged = reqFee.Sub(upFrontCost...)
		ctx.Logger().Debug("On-success cost calculated using required fee (and up-front cost).",
			"required fee", lazyCzStr(reqFee), "up-front cost", lazyCzStr(upFrontCost), "on-success", lazyCzStr(uncharged))
	}
	deductFeesFrom, usedFeeGrant, err := GetFeePayerUsingFeeGrant(ctx, h.fk, feeTx, uncharged, feeTx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	// Skip checking that the deductFeesFrom account exists since we checked that earlier and it shouldn't change.
	// We skip checking the balance again because the bank module will do that for us during the transfer.

	// Pay whatever is left.
	// When simulating, we don't care about the fees being paid.
	// During InitGenesis, there's no fees to pay (and no one to pay them).
	if !simulate && !IsInitGenesis(ctx) && !uncharged.IsZero() {
		ctx2 := ctx
		if usedFeeGrant {
			ctx2 = internalsdk.WithFeeGrantInUse(ctx)
		}
		if err = PayFee(ctx2, h.bk, deductFeesFrom, uncharged); err != nil {
			return ctx, cerrs.Wrapf(err, "could not collect remaining cost %q upon success", uncharged.String())
		}
		ctx.Logger().Debug("Collected remaining cost.", "remaining cost", lazyCzStr(uncharged))
	} else {
		ctx.Logger().Debug("Skipping collection of remaining cost.", "remaining cost", lazyCzStr(uncharged), "simulate", simulate)
	}

	var overage sdk.Coins
	if !simulate && !IsInitGenesis(ctx) {
		overage = feeProvided.Sub(reqFee...)
	}
	onSuccessCost := reqFee.Sub(upFrontCost...)
	ctx.EventManager().EmitEvent(CreateFeeEvent(deductFeesFrom, upFrontCost, onSuccessCost, overage))

	return next(ctx, tx, simulate, success)
}

// CreateFeeEvent creates the event we emit containing tx fee information.
func CreateFeeEvent(feePayer sdk.AccAddress, baseFee, additionalFee, overage sdk.Coins) sdk.Event {
	total := baseFee
	rv := sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFeePayer, feePayer.String()),
		sdk.NewAttribute(AttributeKeyBaseFee, baseFee.String()),
	)
	if !additionalFee.IsZero() {
		total = total.Add(additionalFee...)
		rv = rv.AppendAttributes(sdk.NewAttribute(AttributeKeyAdditionalFee, additionalFee.String()))
	}
	if !overage.IsZero() {
		total = total.Add(overage...)
		rv = rv.AppendAttributes(sdk.NewAttribute(AttributeKeyFeeOverage, overage.String()))
	}
	rv = rv.AppendAttributes(sdk.NewAttribute(AttributeKeyFeeTotal, total.String()))
	return rv
}
