package antewrapper

import (
	"fmt"

	cerrs "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// ProvenanceDeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use ProvenanceDeductFeeDecorator
type ProvenanceDeductFeeDecorator struct {
	ak             authante.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	feegrantKeeper authante.FeegrantKeeper
	msgFeeKeeper   msgfeestypes.MsgFeesKeeper
}

var (
	AttributeKeyBaseFee       = "basefee"
	AttributeKeyAdditionalFee = "additionalfee"
	AttributeKeyMinFeeCharged = "min_fee_charged"
)

func NewProvenanceDeductFeeDecorator(ak authante.AccountKeeper, bk bankkeeper.Keeper, fk msgfeestypes.FeegrantKeeper, mbfk msgfeestypes.MsgFeesKeeper) ProvenanceDeductFeeDecorator {
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
		return ctx, sdkerrors.ErrTxDecode.Wrap("Tx must be a FeeTx")
	}

	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.FeeCollectorName))
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	feeGasMeter, ok := ctx.GasMeter().(*FeeGasMeter)
	if !ok {
		return ctx, sdkerrors.ErrTxDecode.Wrap("GasMeter not a FeeGasMeter")
	}

	payerAccount := feePayer

	// deduct the fees
	fee := feeTx.GetFee()
	msgs := feeTx.GetMsgs()
	feeDist, errFromCalculateAdditionalFeesToBePaid := CalculateAdditionalFeesToBePaid(ctx, dfd.msgFeeKeeper, msgs...)
	if errFromCalculateAdditionalFeesToBePaid != nil {
		return ctx, sdkerrors.ErrInvalidRequest.Wrap(errFromCalculateAdditionalFeesToBePaid.Error())
	}
	if feeDist != nil && len(feeDist.TotalAdditionalFees) > 0 {
		var hasNeg bool
		_, hasNeg = fee.SafeSub(feeDist.TotalAdditionalFees...)
		if hasNeg && !simulate {
			return ctx, sdkerrors.ErrInsufficientFee.Wrapf("invalid fee amount: %s", fee)
		}
	}

	if feeGranter != nil && !simulate {
		if dfd.feegrantKeeper == nil {
			return ctx, sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err = dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())

			if err != nil {
				return ctx, cerrs.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		payerAccount = feeGranter
	}

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, payerAccount)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", payerAccount)
	}

	// deduct minimum amount from fee, remainder will be swept on success
	baseFeeToConsume := CalculateBaseFee(ctx, feeTx, dfd.msgFeeKeeper)
	if !baseFeeToConsume.IsZero() && !simulate {
		err = DeductBaseFees(dfd.bankKeeper, ctx, deductFeesFromAcc, baseFeeToConsume)
		if err != nil {
			return ctx, err
		}
		feeGasMeter.ConsumeBaseFee(baseFeeToConsume)
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	),
		sdk.NewEvent(sdk.EventTypeTx,
			sdk.NewAttribute(AttributeKeyMinFeeCharged, baseFeeToConsume.String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

func CalculateBaseFee(ctx sdk.Context, feeTx sdk.FeeTx, msgfeekeeper msgfeestypes.MsgFeesKeeper) sdk.Coins {
	if !isTestContext(ctx) {
		gasWanted := feeTx.GetGas()
		floorPrice := msgfeekeeper.GetFloorGasPrice(ctx)
		amount := msgfeekeeper.GetFloorGasPrice(ctx).Amount.Mul(sdk.NewIntFromUint64(gasWanted))
		baseFeeToDeduct := sdk.NewCoins(sdk.NewCoin(floorPrice.Denom, amount))
		return baseFeeToDeduct
	}
	return DetermineTestBaseFeeAmount(ctx, feeTx)
}

// DetermineTestBaseFeeAmount determines the type of test that is running.  ChainID = "" is a simple unit
// We need this because of how tests are setup using atom and we have nhash specific code for msgfees
func DetermineTestBaseFeeAmount(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Coins {
	if len(ctx.ChainID()) == 0 {
		return feeTx.GetFee()
	}
	return sdk.NewCoins()
}

// DeductBaseFees deducts fees from the given account.
func DeductBaseFees(bankKeeper bankkeeper.Keeper, ctx sdk.Context, acc types.AccountI, fee sdk.Coins) error {
	if !fee.IsValid() {
		return sdkerrors.ErrInsufficientFee.Wrapf("invalid fee amount: %s", fee)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fee)
	if err != nil {
		return sdkerrors.ErrInsufficientFunds.Wrap(err.Error())
	}
	return nil
}
