package antewrapper

import (
	cerrs "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	internalsdk "github.com/provenance-io/provenance/internal/sdk"
)

// DeductUpFrontCostDecorator is an AnteHandler that collects the up-front cost and ensures the fee
// payer account has enough in it to cover the entire fee that has been provided.
//
// The remainder of the fee is collected by the FlatFeePostHandler.
type DeductUpFrontCostDecorator struct {
	ak ante.AccountKeeper
	bk BankKeeper
	fk FeegrantKeeper
}

func NewDeductUpFrontCostDecorator(ak ante.AccountKeeper, bk BankKeeper, fk FeegrantKeeper) DeductUpFrontCostDecorator {
	return DeductUpFrontCostDecorator{ak: ak, bk: bk, fk: fk}
}

func (d DeductUpFrontCostDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if err := d.checkDeductUpFrontCost(ctx, tx, simulate); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

// checkDeductUpFrontCost identifies the fee payer (possibly using a fee grant), makes sure they have enough
// in their account to cover the entire fee that was provided, and collects the up-front cost from them.
func (d DeductUpFrontCostDecorator) checkDeductUpFrontCost(ctx sdk.Context, tx sdk.Tx, simulate bool) error {
	if addr := d.ak.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		return sdkerrors.ErrLogic.Wrapf("%s module account has not been set", authtypes.FeeCollectorName)
	}

	feeTx, err := GetFeeTx(tx)
	if err != nil {
		return err
	}

	gasMeter, err := GetFlatFeeGasMeter(ctx)
	if err != nil {
		return err
	}

	upFrontCost := gasMeter.GetUpFrontCost()
	deductFeesFrom, usedFeeGrant, err := getFeePayerUsingFeeGrant(ctx, d.fk, feeTx, upFrontCost, feeTx.GetMsgs())
	if err != nil {
		return err
	}

	// Make sure the account has enough to cover the whole cost. We do this check in order to prevent the futile
	// execution of the tx msgs just to find out at the end that the account doesn't have enough to pay for it.
	// When simulating, we don't care about the fees being paid.
	fullFee := feeTx.GetFee()
	if !simulate {
		// Make sure the paying account exists before we check their balance.
		if d.ak.GetAccount(ctx, deductFeesFrom) == nil {
			return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address %q does not exist", deductFeesFrom.String())
		}
		// Now we can check their balance for the full fee they're expected to pay.
		// We check the full balance here (instead of just the up-front cost) in order to prevent
		// the extra work of running all the Msgs when it's unlikely they'll be able to pay.
		// This means that a tx cannot be paid for using funds received during that tx.
		// I feel like it will be really rare that a tx sends funds to the fee payer.
		if err = validateHasBalance(ctx, d.bk, deductFeesFrom, fullFee); err != nil {
			return err
		}
	}

	// Pay the up-front cost.
	// When simulating, we don't care about the fees being paid.
	// During InitGenesis, there's no fees to pay (and no one to pay them).
	if !simulate && !isInitGenesis(ctx) && !upFrontCost.IsZero() {
		ctx2 := ctx
		if usedFeeGrant {
			ctx2 = internalsdk.WithFeeGrantInUse(ctx)
		}
		if err = PayFee(ctx2, d.bk, deductFeesFrom, upFrontCost); err != nil {
			return cerrs.Wrapf(err, "could not collect up-front fee of %q", upFrontCost.String())
		}
		ctx.Logger().Debug("Up Front cost collected.", "up-front cost", lazyCzStr(upFrontCost))
	} else {
		ctx.Logger().Debug("Skipping collection of up-front cost.", "up-front cost", lazyCzStr(upFrontCost), "simulate", simulate)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, fullFee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(AttributeKeyMinFeeCharged, upFrontCost.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
	})

	return nil
}

// validateHasBalance returns an error if the provided account does not have at least the required funds.
func validateHasBalance(ctx sdk.Context, bk BankKeeper, addr sdk.AccAddress, required sdk.Coins) error {
	if required.IsZero() {
		return nil
	}

	var bal sdk.Coins
	for _, coin := range required {
		bal = append(bal, bk.GetBalance(ctx, addr, coin.Denom))
	}
	_, hasNeg := bal.SafeSub(required...)
	if hasNeg {
		return sdkerrors.ErrInsufficientFunds.Wrapf("account %q balance: %q, required: %q",
			addr.String(), bal.String(), required.String())
	}

	return nil
}
